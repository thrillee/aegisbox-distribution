package sms

import (
	"context"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno"
)

// segmentMockQuerier records which segment rows the sender updates.
type segmentMockQuerier struct {
	database.Querier

	nextSegID  int64
	sentRowIDs []int64
}

func (m *segmentMockQuerier) UsesRouteSenderPool(ctx context.Context, id int32) (bool, error) {
	return false, nil
}

func (m *segmentMockQuerier) CreateMessageSegment(ctx context.Context, arg database.CreateMessageSegmentParams) (int64, error) {
	m.nextSegID++
	return m.nextSegID, nil
}

func (m *segmentMockQuerier) UpdateSegmentSent(ctx context.Context, arg database.UpdateSegmentSentParams) error {
	m.sentRowIDs = append(m.sentRowIDs, arg.ID)
	return nil
}

// smppLikeConnector reports success per segment by sequence number only (no DB segment ID),
// as the SMPP connector does; the MNO message ID arrives later in the submit_sm_resp.
type smppLikeConnector struct {
	mno.Connector
}

func (c *smppLikeConnector) ConnectionID() int32 { return 9 }

func (c *smppLikeConnector) SubmitMessage(ctx context.Context, msg mno.PreparedMessage) (mno.SubmitResult, error) {
	res := mno.SubmitResult{WasSubmitted: true}
	for i := range msg.DBSeqs {
		res.Segments = append(res.Segments, mno.SegmentSubmitInfo{Seqn: int32(i + 1), IsSuccess: true})
	}
	return res, nil
}

type fixedConnectorProvider struct{ conn mno.Connector }

func (p fixedConnectorProvider) GetConnector(ctx context.Context, mnoID int32) (mno.Connector, error) {
	return p.conn, nil
}

func TestSend_UpdatesTheMessagesOwnSegmentRows(t *testing.T) {
	// Segment rows 88 and 89 belong to this message; rows 1 and 2 belong to older messages.
	q := &segmentMockQuerier{nextSegID: 87}
	s := &DefaultSender{dbQueries: q, mnoConnectorProvider: fixedConnectorProvider{conn: &smppLikeConnector{}}}
	mnoID := int32(12)

	if _, err := s.Send(context.Background(), database.GetMessageDetailsForSendingRow{
		ID: 99, ServiceProviderID: 6, TotalSegments: 2, RoutedMnoID: &mnoID, OriginalSourceAddr: "HKKF",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(q.sentRowIDs) != 2 || q.sentRowIDs[0] != 88 || q.sentRowIDs[1] != 89 {
		t.Fatalf("expected segment rows [88 89] to be updated, got %v", q.sentRowIDs)
	}
}
