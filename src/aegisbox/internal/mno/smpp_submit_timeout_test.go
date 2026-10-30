package mno

import (
	"context"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
)

// failedSegmentRecorder records UpdateSegmentSendFailed calls; other Querier methods are unused.
type failedSegmentRecorder struct {
	database.Querier
	failed []database.UpdateSegmentSendFailedParams
}

func (r *failedSegmentRecorder) UpdateSegmentSendFailed(_ context.Context, arg database.UpdateSegmentSendFailedParams) error {
	r.failed = append(r.failed, arg)
	return nil
}

// A submit stored under its int32 sequence number must be found when it times out, so the segment
// is recorded as failed instead of staying "sent" with no MNO message ID.
func TestProcessSubmitSMTimeoutFindsPendingSubmit(t *testing.T) {
	q := &failedSegmentRecorder{}
	c := &SMPPMNOConnector{dbQueries: q}
	var seq int32 = 42
	c.pendingSubmits.Store(seq, &submitJob{dbSegmentID: 7, segmentSeqn: 1})

	c.processSubmitSMTimeout(context.Background(), seq)

	if len(q.failed) != 1 || q.failed[0].ID != 7 {
		t.Fatalf("expected segment 7 to be marked failed once, got %+v", q.failed)
	}
	if _, still := c.pendingSubmits.Load(seq); still {
		t.Fatal("expected the timed out submit to be removed from pendingSubmits")
	}
}
