package sms

import (
	"context"
	"log/slog"

	"github.com/thrillee/aegisbox/internal/database"
)

// OtpSenderResetWorker resets depleted OTP alternative senders once their
// reset_interval_hours has elapsed, making them active for rotation again.
type OtpSenderResetWorker struct {
	dbQueries database.Querier
}

// NewOtpSenderResetWorker creates a new OtpSenderResetWorker.
func NewOtpSenderResetWorker(q database.Querier) *OtpSenderResetWorker {
	return &OtpSenderResetWorker{dbQueries: q}
}

// ResetDepletedSenders satisfies the workers.WorkerFunc signature.
// It queries for senders whose usage count has reached the max and whose
// reset interval has elapsed, then resets them to active with a zeroed count.
func (w *OtpSenderResetWorker) ResetDepletedSenders(ctx context.Context, _ int) (int, error) {
	rows, err := w.dbQueries.ResetDepletedOtpAlternativeSenders(ctx)
	if err != nil {
		return 0, err
	}

	if len(rows) > 0 {
		ids := make([]int32, len(rows))
		for i, r := range rows {
			ids[i] = r.ID
		}
		slog.InfoContext(ctx, "Reset depleted OTP alternative senders",
			slog.Int("count", len(rows)),
			slog.Any("sender_ids", ids),
		)
	}

	return len(rows), nil
}
