package mnosim

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	mathrand "math/rand/v2"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

const (
	idleTimeout     = 2 * time.Minute // a bound client sends enquire_link well within this
	receiptTextSize = 20
	receiptDate     = "0601021504" // YYMMDDhhmm, as in SMPP 3.4 appendix B
)

// Server is the simulated MNO.
type Server struct {
	cfg Config
	mix []outcome

	mu       sync.Mutex
	sessions map[*session]struct{}
	listener net.Listener
	wg       sync.WaitGroup
}

type session struct {
	conn       net.Conn
	writeMu    sync.Mutex
	bound      bool
	systemID   string
	canReceive bool // bound as receiver or transceiver, so it can take DLRs
	closed     bool
}

// New returns a server for the configuration.
func New(cfg Config) (*Server, error) {
	mix, err := parseMix(cfg.DLRMix)
	if err != nil {
		return nil, err
	}
	return &Server{cfg: cfg, mix: mix, sessions: make(map[*session]struct{})}, nil
}

// Run listens until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("mnosim listen on %s: %w", s.cfg.Addr, err)
	}
	return s.Serve(ctx, ln)
}

// Serve accepts connections on ln until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	slog.Info("Simulated MNO listening", slog.String("addr", ln.Addr().String()),
		slog.String("dlr_mix", s.cfg.DLRMix),
		slog.Float64("reject_percent", s.cfg.RejectPercent),
		slog.Float64("no_response_percent", s.cfg.NoResponsePercent))

	go func() {
		<-ctx.Done()
		_ = ln.Close()
		s.mu.Lock()
		for ss := range s.sessions {
			_ = ss.conn.Close()
		}
		s.mu.Unlock()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				s.wg.Wait()
				return nil
			}
			slog.Warn("Simulated MNO accept failed", slog.Any("error", err))
			time.Sleep(100 * time.Millisecond)
			continue
		}
		ss := &session{conn: conn}
		s.mu.Lock()
		s.sessions[ss] = struct{}{}
		s.mu.Unlock()
		s.wg.Add(1)
		go s.handle(ss)
	}
}

func (s *Server) handle(ss *session) {
	defer s.wg.Done()
	defer func() {
		s.mu.Lock()
		ss.closed = true
		delete(s.sessions, ss)
		s.mu.Unlock()
		_ = ss.conn.Close()
	}()
	remote := ss.conn.RemoteAddr().String()
	r := bufio.NewReader(ss.conn)

	for {
		_ = ss.conn.SetReadDeadline(time.Now().Add(idleTimeout))
		p, err := pdu.Parse(r)
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				slog.Info("Simulated MNO connection ended", slog.String("remote", remote), slog.Any("error", err))
			}
			return
		}

		switch req := p.(type) {
		case *pdu.BindRequest:
			s.handleBind(ss, req)
		case *pdu.SubmitSM:
			s.handleSubmit(ss, req)
		case *pdu.EnquireLink:
			s.write(ss, req.GetResponse())
		case *pdu.Unbind:
			s.write(ss, req.GetResponse())
			return
		case *pdu.DeliverSMResp, *pdu.EnquireLinkResp, *pdu.GenericNack:
			// Responses are never answered.
		default:
			if uint32(p.GetHeader().CommandID)&0x80000000 != 0 {
				continue
			}
			nack := pdu.NewGenericNack()
			nack.SetSequenceNumber(p.GetSequenceNumber())
			setStatus(nack, data.ESME_RINVCMDID)
			s.write(ss, nack)
		}
	}
}

func (s *Server) handleBind(ss *session, req *pdu.BindRequest) {
	resp := req.GetResponse().(*pdu.BindResp)
	switch {
	case ss.bound:
		setStatus(resp, data.ESME_RALYBND)
	case req.SystemID != s.cfg.SystemID:
		setStatus(resp, data.ESME_RINVSYSID)
	case req.Password != s.cfg.Password:
		setStatus(resp, data.ESME_RINVPASWD)
	default:
		s.mu.Lock()
		ss.bound = true
		ss.systemID = req.SystemID
		ss.canReceive = req.BindingType != pdu.Transmitter
		s.mu.Unlock()
		resp.SystemID = "MNOSIM"
		slog.Info("Simulated MNO bind accepted", slog.String("remote", ss.conn.RemoteAddr().String()),
			slog.String("system_id", req.SystemID), slog.Bool("can_receive", ss.canReceive))
	}
	if resp.CommandStatus != data.ESME_ROK {
		slog.Warn("Simulated MNO bind refused", slog.String("system_id", req.SystemID), slog.String("status", resp.CommandStatus.String()))
	}
	s.write(ss, resp)
}

func (s *Server) handleSubmit(ss *session, req *pdu.SubmitSM) {
	resp := req.GetResponse().(*pdu.SubmitSMResp)
	if !ss.bound {
		setStatus(resp, data.ESME_RINVBNDSTS)
		s.write(ss, resp)
		return
	}

	text, _ := req.Message.GetMessage()
	part := ""
	if udh := req.Message.UDH(); udh != nil {
		if total, num, ref, ok := udh.GetConcatInfo(); ok {
			part = fmt.Sprintf("%d/%d ref %d", num, total, ref)
		}
	}
	src, dst := req.SourceAddr.Address(), req.DestAddr.Address()

	roll := mathrand.Float64() * 100
	switch {
	case roll < s.cfg.RejectPercent:
		setStatus(resp, data.CommandStatusType(s.cfg.RejectStatus))
		s.write(ss, resp)
		slog.Info("Simulated MNO rejected submit", slog.String("from", src), slog.String("to", dst),
			slog.String("status", fmt.Sprintf("0x%X", s.cfg.RejectStatus)))
		return
	case roll < s.cfg.RejectPercent+s.cfg.NoResponsePercent:
		slog.Info("Simulated MNO left submit unanswered", slog.String("from", src), slog.String("to", dst))
		return
	}

	id := newMessageID()
	resp.MessageID = id
	s.write(ss, resp)
	slog.Info("Simulated MNO accepted submit", slog.String("message_id", id), slog.String("from", src),
		slog.String("to", dst), slog.String("part", part), slog.String("text", text))

	status := s.pickOutcome()
	if status == "" || req.RegisteredDelivery&0x03 == 0 {
		return // no DLR for this one, or none was requested
	}
	submitted := time.Now()
	delay := s.cfg.DLRDelayMin
	if span := s.cfg.DLRDelayMax - s.cfg.DLRDelayMin; span > 0 {
		delay += time.Duration(mathrand.Int64N(int64(span) + 1))
	}
	time.AfterFunc(delay, func() { s.sendDLR(ss, id, src, dst, text, status, submitted) })
}

// sendDLR delivers the receipt on the session that submitted, or another receiving session of the
// same login if that one is gone.
func (s *Server) sendDLR(origin *session, id, src, dst, text, status string, submitted time.Time) {
	target := s.receivingSession(origin)
	if target == nil {
		slog.Warn("Simulated MNO has no receiving session for DLR", slog.String("message_id", id))
		return
	}

	delivered, errCode := 0, "000"
	if status == "DELIVRD" {
		delivered = 1
	} else {
		errCode = "001"
	}
	receipt := fmt.Sprintf("id:%s sub:001 dlvrd:%03d submit date:%s done date:%s stat:%s err:%s text:%s",
		id, delivered, submitted.Format(receiptDate), time.Now().Format(receiptDate), status, errCode, receiptText(text))

	dlr := pdu.NewDeliverSM().(*pdu.DeliverSM)
	_ = dlr.SourceAddr.SetAddress(dst) // the receipt comes from the recipient
	_ = dlr.DestAddr.SetAddress(src)
	dlr.EsmClass = 0x04 // SMSC delivery receipt
	msg, err := pdu.NewShortMessage(receipt)
	if err != nil {
		slog.Error("Simulated MNO could not build DLR", slog.Any("error", err))
		return
	}
	dlr.Message = msg
	s.write(target, dlr)
	slog.Info("Simulated MNO sent DLR", slog.String("message_id", id), slog.String("stat", status))
}

func (s *Server) receivingSession(origin *session) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !origin.closed && origin.canReceive {
		return origin
	}
	for ss := range s.sessions {
		if ss.bound && ss.canReceive && ss.systemID == origin.systemID {
			return ss
		}
	}
	return nil
}

func (s *Server) pickOutcome() string {
	total := 0.0
	for _, o := range s.mix {
		total += o.weight
	}
	roll := mathrand.Float64() * total
	for _, o := range s.mix {
		if roll < o.weight {
			return o.status
		}
		roll -= o.weight
	}
	return s.mix[len(s.mix)-1].status
}

func (s *Server) write(ss *session, p pdu.PDU) {
	buf := pdu.NewBuffer(make([]byte, 0, 64))
	p.Marshal(buf)
	ss.writeMu.Lock()
	defer ss.writeMu.Unlock()
	_ = ss.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := ss.conn.Write(buf.Bytes()); err != nil {
		slog.Warn("Simulated MNO write failed", slog.Any("error", err))
	}
}

// setStatus sets a response's command status.
func setStatus(p pdu.PDU, status data.CommandStatusType) {
	switch r := p.(type) {
	case *pdu.BindResp:
		r.CommandStatus = status
	case *pdu.SubmitSMResp:
		r.CommandStatus = status
	case *pdu.GenericNack:
		r.CommandStatus = status
	}
}

func newMessageID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// receiptText is the first characters of the message, on one line, as receipts carry.
func receiptText(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if r := []rune(text); len(r) > receiptTextSize {
		text = string(r[:receiptTextSize])
	}
	return text
}
