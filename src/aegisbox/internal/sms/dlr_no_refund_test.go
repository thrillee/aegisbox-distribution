package sms

import (
	"context"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
)

// aggregationMockQuerier serves one message whose single segment got the given DLR status.
type aggregationMockQuerier struct {
	database.Querier

	dlrStatus   string
	finalStatus string
}

func (m *aggregationMockQuerier) GetSegmentStatusesForMessage(ctx context.Context, messageID int64) ([]database.GetSegmentStatusesForMessageRow, error) {
	return []database.GetSegmentStatusesForMessageRow{{SegmentSeqn: 1, DlrStatus: &m.dlrStatus}}, nil
}

func (m *aggregationMockQuerier) GetMessageTotalSegments(ctx context.Context, id int64) (int32, error) {
	return 1, nil
}

func (m *aggregationMockQuerier) GetMessageFinalStatus(ctx context.Context, id int64) (string, error) {
	return "processing", nil
}

func (m *aggregationMockQuerier) UpdateMessageFinalStatus(ctx context.Context, arg database.UpdateMessageFinalStatusParams) error {
	m.finalStatus = arg.FinalStatus
	return nil
}

type countingWallet struct{ reversals int }

func (w *countingWallet) ReverseDebit(ctx context.Context, messageID int64, reason string) error {
	w.reversals++
	return nil
}

func TestFailedDLRMarksFailedWithoutRefund(t *testing.T) {
	for _, status := range []string{"UNDELIV", "EXPIRED", "REJECTD", "DELETED"} {
		q := &aggregationMockQuerier{dlrStatus: status}
		w := &countingWallet{}

		(&DLRService{dbQueries: q, walletService: w}).aggregateMessageStatus(context.Background(), 1)
		if q.finalStatus != "failed" || w.reversals != 0 {
			t.Fatalf("DLRService %s: final status %q, reversals %d; want failed and no refund", status, q.finalStatus, w.reversals)
		}

		q = &aggregationMockQuerier{dlrStatus: status}
		(&Processor{dbQueries: q, walletService: w}).aggregateMessageStatus(context.Background(), 1)
		if q.finalStatus != "failed" || w.reversals != 0 {
			t.Fatalf("Processor %s: final status %q, reversals %d; want failed and no refund", status, q.finalStatus, w.reversals)
		}
	}
}
