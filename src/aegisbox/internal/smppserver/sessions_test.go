package smppserver

import (
	"bufio"
	"context"
	"io"
	"net"
	"testing"
	"time"
)

// newTestSession returns a bound session whose client end is returned for reading.
func newTestSession(t *testing.T, systemID string, canReceive bool) (*sessionState, net.Conn) {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() { _ = server.Close(); _ = client.Close() })
	return &sessionState{systemID: systemID, conn: server, writer: bufio.NewWriter(server), canReceive: canReceive}, client
}

func newTestServer(sessions ...*sessionState) *Server {
	s := &Server{sessions: make(map[string][]*sessionState), nextSession: make(map[string]int)}
	for _, ss := range sessions {
		s.sessions[ss.systemID] = append(s.sessions[ss.systemID], ss)
	}
	return s
}

// readPDU reads len(want) bytes from conn, failing the test on timeout.
func readPDU(t *testing.T, conn net.Conn, n int) []byte {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("expected a PDU on this session: %v", err)
	}
	return buf
}

func TestForwardRawPDUSkipsTransmitterAndRotates(t *testing.T) {
	tx, _ := newTestSession(t, "SP1", false)
	rx1, c1 := newTestSession(t, "SP1", true)
	rx2, c2 := newTestSession(t, "SP1", true)
	s := newTestServer(tx, rx1, rx2)
	pdu := []byte("dlr-pdu")

	got := make(chan string, 2)
	go func() { readPDU(t, c1, len(pdu)); got <- "rx1" }()
	go func() { readPDU(t, c2, len(pdu)); got <- "rx2" }()
	for i := 0; i < 2; i++ {
		if err := s.ForwardRawPDU(context.Background(), "SP1", pdu); err != nil {
			t.Fatalf("forward %d: %v", i, err)
		}
	}
	first, second := <-got, <-got
	if first == second {
		t.Fatalf("expected the two DLRs on different receiving sessions, both went to %s", first)
	}
}

func TestForwardRawPDUFailsOverToNextSession(t *testing.T) {
	dead, deadClient := newTestSession(t, "SP1", true)
	live, liveClient := newTestSession(t, "SP1", true)
	_ = deadClient.Close() // writes to dead now fail
	s := newTestServer(dead, live)
	pdu := []byte("dlr-pdu")

	done := make(chan struct{})
	go func() { readPDU(t, liveClient, len(pdu)); close(done) }()
	if err := s.ForwardRawPDU(context.Background(), "SP1", pdu); err != nil {
		t.Fatalf("forward: %v", err)
	}
	<-done
}

func TestForwardRawPDUNeedsReceivingSession(t *testing.T) {
	tx, _ := newTestSession(t, "SP1", false)
	s := newTestServer(tx)
	if err := s.ForwardRawPDU(context.Background(), "SP1", []byte("x")); err == nil {
		t.Fatal("expected an error with only a transmitter bound")
	}
}

// A reconnect must not lose the new session when the old connection's handler exits afterwards.
func TestRemoveSessionKeepsOtherSessions(t *testing.T) {
	old, _ := newTestSession(t, "SP1", true)
	fresh, freshClient := newTestSession(t, "SP1", true)
	s := newTestServer(old, fresh)

	s.removeSession(old)
	if got := s.sessions["SP1"]; len(got) != 1 || got[0] != fresh {
		t.Fatalf("expected only the new session to remain, got %d sessions", len(got))
	}

	pdu := []byte("dlr-pdu")
	done := make(chan struct{})
	go func() { readPDU(t, freshClient, len(pdu)); close(done) }()
	if err := s.ForwardRawPDU(context.Background(), "SP1", pdu); err != nil {
		t.Fatalf("forward after reconnect: %v", err)
	}
	<-done

	s.removeSession(fresh)
	if _, ok := s.sessions["SP1"]; ok {
		t.Fatal("expected the login to be gone once its last session is removed")
	}
}
