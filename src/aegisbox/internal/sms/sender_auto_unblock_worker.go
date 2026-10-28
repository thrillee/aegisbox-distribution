package sms

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
)

// SenderAutoUnblockWorker periodically unblocks senders that have been inactive for a specified duration
type SenderAutoUnblockWorker struct {
	dbQueries *database.Queries
	logger    *slog.Logger
	interval  time.Duration
}

// SenderAutoUnblockWorkerConfig holds configuration for the sender auto-unblock worker
type SenderAutoUnblockWorkerConfig struct {
	Interval time.Duration // How often to check for auto-unblock (e.g., 6 hours)
}

// NewSenderAutoUnblockWorker creates a new sender auto-unblock worker
func NewSenderAutoUnblockWorker(
	q *database.Queries,
	logger *slog.Logger,
	config SenderAutoUnblockWorkerConfig,
) *SenderAutoUnblockWorker {
	// Apply defaults if not provided
	if config.Interval == 0 {
		config.Interval = 6 * time.Hour
	}

	return &SenderAutoUnblockWorker{
		dbQueries: q,
		logger:    logger,
		interval:  config.Interval,
	}
}

// Start begins the sender auto-unblock worker loop
func (w *SenderAutoUnblockWorker) Start(ctx context.Context) {
	w.logger.Info("Sender auto-unblock worker started",
		slog.Duration("interval", w.interval))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run immediately on start
	w.processAutoUnblock(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Sender auto-unblock worker stopping")
			return
		case <-ticker.C:
			w.processAutoUnblock(ctx)
		}
	}
}

// processAutoUnblock finds and unblocks senders that have been inactive for 7+ days
func (w *SenderAutoUnblockWorker) processAutoUnblock(ctx context.Context) {
	startTime := time.Now()
	w.logger.Debug("Processing sender auto-unblock")

	// Get blocked senders eligible for auto-unblock (inactive for 7+ days)
	senders, err := w.dbQueries.GetBlockedSendersForAutoUnblock(ctx)
	if err != nil {
		w.logger.Error("Failed to get blocked senders for auto-unblock", slog.Any("error", err))
		return
	}

	if len(senders) == 0 {
		w.logger.Debug("No blocked senders eligible for auto-unblock")
		return
	}

	w.logger.Info("Found blocked senders eligible for auto-unblock",
		slog.Int("count", len(senders)))

	// Process each sender
	unblockedCount := 0
	errorCount := 0

	for _, sender := range senders {
		err := w.unblockSender(ctx, sender)
		if err != nil {
			w.logger.Error("Failed to auto-unblock sender",
				slog.Int("sender_id", int(sender.SenderID)),
				slog.Int("mno_id", int(sender.MnoID)),
				slog.String("sender_id_string", sender.SenderIDString),
				slog.Any("error", err))
			errorCount++
			continue
		}
		unblockedCount++
	}

	duration := time.Since(startTime)
	w.logger.Info("Sender auto-unblock processing completed",
		slog.Int("unblocked", unblockedCount),
		slog.Int("errors", errorCount),
		slog.Duration("duration", duration))
}

// unblockSender unblocks a single sender by updating their health record
func (w *SenderAutoUnblockWorker) unblockSender(ctx context.Context, sender database.GetBlockedSendersForAutoUnblockRow) error {
	// Unblock the sender by updating their health record
	falseVal := false
	reason := fmt.Sprintf("Auto-unblocked after 7 days of inactivity. Last attempt: %v", sender.LastAttemptAt)

	_, err := w.dbQueries.UpsertSenderIdHealth(ctx, database.UpsertSenderIdHealthParams{
		SenderID:        sender.SenderID,
		MnoID:           sender.MnoID,
		TotalSent:       sender.TotalSent,
		TotalDelivered:  sender.TotalDelivered,
		TotalFailed:     sender.TotalFailed,
		TotalNoDlr:      sender.TotalNoDlr,
		DeliveryRate:    sender.DeliveryRate,
		DlrResponseRate: sender.DlrResponseRate,
		HealthScore:     sender.HealthScore,
		IsBlocked:       &falseVal, // Unblock
		BlockedReason:   &reason,
		AutoBlocked:     &falseVal,
		LastSuccessAt:   sender.LastSuccessAt,
		LastFailureAt:   sender.LastFailureAt,
		LastNoDlrAt:     sender.LastNoDlrAt,
		LastAttemptAt:   sender.LastAttemptAt,
		WindowSize:      sender.WindowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to update sender health: %w", err)
	}

	w.logger.Info("Sender auto-unblocked after 7 days of inactivity",
		slog.Int("sender_id", int(sender.SenderID)),
		slog.Int("mno_id", int(sender.MnoID)),
		slog.String("sender_id_string", sender.SenderIDString),
		slog.Time("last_attempt_at", sender.LastAttemptAt.Time),
		slog.Float64("previous_delivery_rate", numericToFloat64(sender.DeliveryRate)))

	return nil
}
