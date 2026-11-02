package sms

import (
	"context"
	"errors"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno"
)

type recordingEnqueuer struct {
	calls []string
	err   error
}

func (r *recordingEnqueuer) MessageSent(ctx context.Context, messageID int64, spID int32, mnoID int32, text string) error {
	r.calls = append(r.calls, text)
	return r.err
}

// rejectingConnector refuses every segment, as an MNO rejecting the submit does.
type rejectingConnector struct{ smppLikeConnector }

func (c *rejectingConnector) SubmitMessage(ctx context.Context, msg mno.PreparedMessage) (mno.SubmitResult, error) {
	res := mno.SubmitResult{WasSubmitted: true}
	for i := range msg.DBSeqs {
		res.Segments = append(res.Segments, mno.SegmentSubmitInfo{Seqn: int32(i + 1), IsSuccess: false, Error: errors.New("rejected")})
	}
	return res, nil
}

func (m *segmentMockQuerier) UpdateSegmentSendFailed(context.Context, database.UpdateSegmentSendFailedParams) error {
	return nil
}

func TestSend_QueuesLinksOnlyAfterASuccessfulSubmit(t *testing.T) {
	mnoID := int32(12)
	msg := database.GetMessageDetailsForSendingRow{
		ID: 99, ServiceProviderID: 6, TotalSegments: 2, RoutedMnoID: &mnoID, OriginalSourceAddr: "HKKF",
		ShortMessage: "Win at https://example.com/x",
	}

	enq := &recordingEnqueuer{}
	s := &DefaultSender{dbQueries: &segmentMockQuerier{}, mnoConnectorProvider: fixedConnectorProvider{conn: &smppLikeConnector{}}}
	s.SetLinkEnqueuer(enq)
	if _, err := s.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(enq.calls) != 1 || enq.calls[0] != msg.ShortMessage {
		t.Fatalf("expected one enqueue with the full text, got %v", enq.calls)
	}

	enq = &recordingEnqueuer{}
	s = &DefaultSender{dbQueries: &segmentMockQuerier{}, mnoConnectorProvider: fixedConnectorProvider{conn: &rejectingConnector{}}}
	s.SetLinkEnqueuer(enq)
	if _, err := s.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(enq.calls) != 0 {
		t.Fatalf("a rejected submit must not queue links, got %v", enq.calls)
	}
}

func TestSend_LinkQueueErrorDoesNotFailTheSend(t *testing.T) {
	mnoID := int32(12)
	enq := &recordingEnqueuer{err: errors.New("db down")}
	s := &DefaultSender{dbQueries: &segmentMockQuerier{}, mnoConnectorProvider: fixedConnectorProvider{conn: &smppLikeConnector{}}}
	s.SetLinkEnqueuer(enq)
	res, err := s.Send(context.Background(), database.GetMessageDetailsForSendingRow{
		ID: 1, ServiceProviderID: 6, RoutedMnoID: &mnoID, OriginalSourceAddr: "HKKF", ShortMessage: "https://a.com",
	})
	if err != nil || res == nil || res.Error != nil {
		t.Fatalf("send must succeed, got res=%+v err=%v", res, err)
	}
}
