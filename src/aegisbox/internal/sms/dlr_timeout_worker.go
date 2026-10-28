package sms

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"

	"github.com/thrillee/aegisbox/internal/database"
)

// DLRTimeoutWorker periodically checks for segments with missing DLRs and marks them as "no_dlr"
type DLRTimeoutWorker struct {
	dbQueries           *database.Queries
	senderHealthTracker *SenderHealthTracker
	logger              *slog.Logger
	interval            time.Duration
	timeoutSeconds      int32
	batchSize           int32
}

// DLRTimeoutWorkerConfig holds configuration for the DLR timeout worker
type DLRTimeoutWorkerConfig struct {
	Interval       time.Duration // How often to check for timeouts (e.g., 1 minute)
	TimeoutSeconds int32         // How long to wait for DLR before marking as timeout (e.g., 300 seconds)
	BatchSize      int32         // How many segments to process per batch (e.g., 1000)
}

// NewDLRTimeoutWorker creates a new DLR timeout worker
func NewDLRTimeoutWorker(
	q *database.Queries,
	senderHealthTracker *SenderHealthTracker,
	logger *slog.Logger,
	config DLRTimeoutWorkerConfig,
) *DLRTimeoutWorker {
	// Apply defaults if not provided
	if config.Interval == 0 {
		config.Interval = 1 * time.Minute
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 300 // 5 minutes
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}

	return &DLRTimeoutWorker{
		dbQueries:           q,
		senderHealthTracker: senderHealthTracker,
		logger:              logger,
		interval:            config.Interval,
		timeoutSeconds:      config.TimeoutSeconds,
		batchSize:           config.BatchSize,
	}
}

// Start begins the DLR timeout worker loop
func (w *DLRTimeoutWorker) Start(ctx context.Context) {
	w.logger.Info("DLR timeout worker started",
		slog.Duration("interval", w.interval),
		slog.Int("timeout_seconds", int(w.timeoutSeconds)),
		slog.Int("batch_size", int(w.batchSize)))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run immediately on start
	w.processTimeouts(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("DLR timeout worker stopping")
			return
		case <-ticker.C:
			w.processTimeouts(ctx)
		}
	}
}

// processTimeouts finds and processes segments with missing DLRs
func (w *DLRTimeoutWorker) processTimeouts(ctx context.Context) {
	startTime := time.Now()
	w.logger.Debug("Processing DLR timeouts")

	// The DLR timeout set in the global routing config (manager API / portal) wins over the default
	timeoutSeconds := w.timeoutSeconds
	if config, err := w.dbQueries.GetGlobalRoutingConfig(ctx); err == nil &&
		config.SenderIDDlrTimeoutSecs != nil && *config.SenderIDDlrTimeoutSecs >= 60 {
		timeoutSeconds = *config.SenderIDDlrTimeoutSecs
	}

	// Get pending segments that have exceeded timeout threshold
	segments, err := w.dbQueries.GetPendingDLRSegments(ctx, database.GetPendingDLRSegmentsParams{
		TimeoutSeconds: decimal.NewFromInt32(timeoutSeconds),
		MaxResults:     w.batchSize,
	})
	if err != nil {
		w.logger.Error("Failed to get pending DLR segments", slog.Any("error", err))
		return
	}

	if len(segments) == 0 {
		w.logger.Debug("No pending DLR segments found")
		return
	}

	w.logger.Info("Found pending DLR segments",
		slog.Int("count", len(segments)),
		slog.Duration("timeout", time.Duration(timeoutSeconds)*time.Second))

	// Process each segment
	processedCount := 0
	errorCount := 0

	for _, segment := range segments {
		err := w.processSegment(ctx, segment)
		if err != nil {
			w.logger.Error("Failed to process timed-out segment",
				slog.Int64("segment_id", segment.ID),
				slog.Int64("message_id", segment.MessageID),
				slog.Any("error", err))
			errorCount++
			continue
		}
		processedCount++
	}

	duration := time.Since(startTime)
	w.logger.Info("DLR timeout processing completed",
		slog.Int("processed", processedCount),
		slog.Int("errors", errorCount),
		slog.Duration("duration", duration))
}

// processSegment marks a single segment as "no_dlr" and updates health metrics
func (w *DLRTimeoutWorker) processSegment(ctx context.Context, segment database.GetPendingDLRSegmentsRow) error {
	// Update segment DLR status to "no_dlr" using the existing UpdateSegmentDLR query
	dlrStatus := "no_dlr"
	errorCode := "TIMEOUT"
	err := w.dbQueries.UpdateSegmentDLR(ctx, database.UpdateSegmentDLRParams{
		DlrStatus: &dlrStatus,
		ErrorCode: &errorCode,
		ID:        segment.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update segment DLR status: %w", err)
	}

	// Record in sender health tracker if available
	if w.senderHealthTracker != nil && segment.ApprovedSenderID != nil && segment.RoutedMnoID != nil {
		sentAt := time.Now()
		if segment.SentToMnoAt.Valid {
			sentAt = segment.SentToMnoAt.Time
		}

		err = w.senderHealthTracker.RecordDeliveryResult(
			ctx,
			*segment.ApprovedSenderID,
			*segment.RoutedMnoID,
			segment.ID,
			"no_dlr",
			&errorCode,
			sentAt,
			nil, // No DLR received
		)
		if err != nil {
			// Log warning but don't fail the operation
			w.logger.Warn("Failed to record no-DLR in health tracker",
				slog.Int64("segment_id", segment.ID),
				slog.Int("sender_id", int(*segment.ApprovedSenderID)),
				slog.Int("mno_id", int(*segment.RoutedMnoID)),
				slog.Any("error", err))
		}
	}

	w.logger.Debug("Marked segment as no-DLR",
		slog.Int64("segment_id", segment.ID),
		slog.Int64("message_id", segment.MessageID),
		slog.Any("sender_id", segment.ApprovedSenderID),
		slog.Any("mno_id", segment.RoutedMnoID))

	return nil
}

// stringPtr is a helper to create string pointers
func stringPtr(s string) *string {
	return &s
}
