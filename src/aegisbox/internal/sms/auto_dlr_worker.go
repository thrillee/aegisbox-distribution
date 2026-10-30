package sms

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/thrillee/aegisbox/internal/database"
)

// maxAutoDLRBatchesPerRun caps how many batches one run drains, so a backlog is
// worked through quickly without a single run holding on indefinitely.
const maxAutoDLRBatchesPerRun = 20

// AutoDLRWorker periodically completes messages of service providers with auto DLR
// enabled whose segments got no DLR from the MNO within the SP's configured delay.
type AutoDLRWorker struct {
	dbQueries  *database.Queries
	dlrService *DLRService
	logger     *slog.Logger
	interval   time.Duration
	batchSize  int32
	maxAge     time.Duration
}

// AutoDLRWorkerConfig holds configuration for the auto DLR worker
type AutoDLRWorkerConfig struct {
	Interval  time.Duration // How often to check for messages to auto complete (e.g., 1 minute)
	BatchSize int32         // How many messages to fetch per batch (e.g., 500)
	MaxAge    time.Duration // Ignore segments sent longer ago than this (e.g., 7 days)
}

// NewAutoDLRWorker creates a new auto DLR worker
func NewAutoDLRWorker(
	q *database.Queries,
	dlrService *DLRService,
	logger *slog.Logger,
	config AutoDLRWorkerConfig,
) *AutoDLRWorker {
	if config.Interval == 0 {
		config.Interval = 1 * time.Minute
	}
	if config.BatchSize == 0 {
		config.BatchSize = 500
	}
	if config.MaxAge == 0 {
		config.MaxAge = 7 * 24 * time.Hour
	}

	return &AutoDLRWorker{
		dbQueries:  q,
		dlrService: dlrService,
		logger:     logger,
		interval:   config.Interval,
		batchSize:  config.BatchSize,
		maxAge:     config.MaxAge,
	}
}

// Start begins the auto DLR worker loop
func (w *AutoDLRWorker) Start(ctx context.Context) {
	w.logger.Info("Auto DLR worker started",
		slog.Duration("interval", w.interval),
		slog.Int("batch_size", int(w.batchSize)),
		slog.Duration("max_age", w.maxAge))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.processAutoDLRs(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Auto DLR worker stopping")
			return
		case <-ticker.C:
			w.processAutoDLRs(ctx)
		}
	}
}

// processAutoDLRs finds messages due for an auto DLR and applies it, draining
// up to maxAutoDLRBatchesPerRun batches
func (w *AutoDLRWorker) processAutoDLRs(ctx context.Context) {
	startTime := time.Now()
	processedCount := 0
	errorCount := 0

	for batch := 0; batch < maxAutoDLRBatchesPerRun; batch++ {
		candidates, err := w.dbQueries.GetAutoDLRCandidates(ctx, database.GetAutoDLRCandidatesParams{
			NotBefore:  pgtype.Timestamptz{Time: time.Now().Add(-w.maxAge), Valid: true},
			MaxResults: w.batchSize,
		})
		if err != nil {
			w.logger.Error("Failed to get auto DLR candidates", slog.Any("error", err))
			break
		}

		batchProcessed := 0
		for _, candidate := range candidates {
			if ctx.Err() != nil {
				return
			}
			result, err := w.dlrService.ApplyManualDLR(ctx, ManualDLRRequest{
				MessageID:  candidate.MessageID,
				Status:     candidate.AutoDlrStatus,
				Source:     DLRSourceAuto,
				SentBefore: candidate.SentBefore.Time,
				// Only when the SP opted in; the candidate query applies the same rule.
				IncludeUnconfirmed: candidate.AutoDlrIncludeUnconfirmed,
			})
			if err != nil {
				w.logger.Error("Failed to apply auto DLR",
					slog.Int64("message_id", candidate.MessageID),
					slog.Any("error", err))
				errorCount++
				continue
			}
			if result.SegmentsUpdated > 0 {
				batchProcessed++
			}
		}
		processedCount += batchProcessed

		// Stop when the backlog is drained, or when nothing in a full batch could be
		// applied, so failing messages are not refetched in a tight loop.
		if len(candidates) < int(w.batchSize) || batchProcessed == 0 {
			break
		}
	}

	if processedCount == 0 && errorCount == 0 {
		w.logger.Debug("No messages due for auto DLR")
		return
	}
	w.logger.Info("Auto DLR processing completed",
		slog.Int("processed", processedCount),
		slog.Int("errors", errorCount),
		slog.Duration("duration", time.Since(startTime)))
}
