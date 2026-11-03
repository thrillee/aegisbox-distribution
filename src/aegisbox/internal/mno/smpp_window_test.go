package mno

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/pkg/codes"
)

// The connector keeps one window slot for enquire_link, waits for a free slot instead of letting
// gosmpp drop the PDU, and frees a slot exactly once when the submit's outcome is known.
func TestSubmitWindowSlots(t *testing.T) {
	q := &failedSegmentRecorder{}
	c, err := NewSMPPMNOConnector(SMPPConnectorConfig{Host: "h", Port: 1, SystemID: "s", MaxWindowSize: 3, RequestTimeout: 50 * time.Millisecond}, nil, q)
	if err != nil {
		t.Fatal(err)
	}
	if cap(c.inflight) != 2 {
		t.Fatalf("expected 2 submit slots for a window of 3, got %d", cap(c.inflight))
	}
	ctx := context.Background()
	for seq := int32(1); seq <= 2; seq++ {
		if err := c.acquireSlot(ctx); err != nil {
			t.Fatalf("slot %d: %v", seq, err)
		}
		c.pendingSubmits.Store(seq, &submitJob{dbSegmentID: int64(seq)})
	}
	if err := c.acquireSlot(ctx); !errors.Is(err, gosmpp.ErrWindowsFull) {
		t.Fatalf("expected window full after the timeout, got %v", err)
	}

	// A response frees a slot, and a waiting submit gets it.
	got := make(chan error, 1)
	go func() { got <- c.acquireSlot(ctx) }()
	time.Sleep(10 * time.Millisecond)
	if _, ok := c.takePending(1); !ok {
		t.Fatal("expected pending submit 1")
	}
	if err := <-got; err != nil {
		t.Fatalf("waiting submit should get the freed slot: %v", err)
	}
	c.pendingSubmits.Store(int32(3), &submitJob{dbSegmentID: 3})

	// Taking the same submit twice frees one slot only.
	if _, ok := c.takePending(1); ok {
		t.Fatal("submit 1 was already taken")
	}
	if len(c.inflight) != 2 {
		t.Fatalf("expected 2 slots in use, got %d", len(c.inflight))
	}

	// An asynchronous submit error marks the segment failed and frees its slot.
	sm := pdu.NewSubmitSM().(*pdu.SubmitSM)
	sm.SequenceNumber = 2
	c.onSubmitError(sm, gosmpp.ErrWindowsFull)
	if len(q.failed) != 1 || q.failed[0].ID != 2 || *q.failed[0].ErrorCode != string(codes.ErrorCodeMnoWindowFull) {
		t.Fatalf("expected segment 2 marked failed with window full, got %+v", q.failed)
	}
	if len(c.inflight) != 1 {
		t.Fatalf("expected 1 slot in use after the error, got %d", len(c.inflight))
	}
}

type sentSegmentRecorder struct {
	failedSegmentRecorder
	sent []database.UpdateSegmentSentParams
}

func (r *sentSegmentRecorder) UpdateSegmentSent(_ context.Context, arg database.UpdateSegmentSentParams) error {
	r.sent = append(r.sent, arg)
	return nil
}

// A submit_sm_resp that arrives before gosmpp has recorded the request is still matched, and its
// slot stays taken until gosmpp expires the request, without the submit being failed.
func TestEarlySubmitResponseIsMatched(t *testing.T) {
	q := &sentSegmentRecorder{}
	c, err := NewSMPPMNOConnector(SMPPConnectorConfig{Host: "h", Port: 1, SystemID: "s", MaxWindowSize: 2, RequestTimeout: 50 * time.Millisecond}, nil, q)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.acquireSlot(context.Background()); err != nil {
		t.Fatal(err)
	}
	c.pendingSubmits.Store(int32(5), &submitJob{dbSegmentID: 9})

	resp := pdu.NewSubmitSMResp().(*pdu.SubmitSMResp)
	resp.SequenceNumber = 5
	resp.MessageID = "MNO-9"
	c.onUnexpectedPduResponse(resp)
	if len(q.sent) != 1 || q.sent[0].ID != 9 || *q.sent[0].MnoMessageID != "MNO-9" {
		t.Fatalf("expected segment 9 recorded as sent with MNO-9, got %+v", q.sent)
	}
	if len(c.inflight) != 1 {
		t.Fatal("slot should stay taken until gosmpp expires the request")
	}

	c.processSubmitSMTimeout(context.Background(), 5)
	if len(q.failed) != 0 {
		t.Fatalf("an answered submit must not be failed on expiry, got %+v", q.failed)
	}
	if len(c.inflight) != 0 {
		t.Fatal("expiry should free the slot")
	}
}
