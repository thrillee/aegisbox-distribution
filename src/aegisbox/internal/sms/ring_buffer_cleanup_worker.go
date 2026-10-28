package sms

import (
	"context"
	"log/slog"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
)

// RingBufferCleanupWorker periodically cleans up old delivery attempt records from ring buffer tables
type RingBufferCleanupWorker struct {
	dbQueries     *database.Queries
	logger        *slog.Logger
	interval      time.Duration
	retentionDays int32
}

// RingBufferCleanupWorkerConfig holds configuration for the ring buffer cleanup worker
type RingBufferCleanupWorkerConfig struct {
	Interval      time.Duration // How often to run cleanup (e.g., 24 hours)
	RetentionDays int32         // How many days to retain data (e.g., 90)
}

// NewRingBufferCleanupWorker creates a new ring buffer cleanup worker
func NewRingBufferCleanupWorker(
	q *database.Queries,
	logger *slog.Logger,
	config RingBufferCleanupWorkerConfig,
) *RingBufferCleanupWorker {
	// Apply defaults if not provided
	if config.Interval == 0 {
		config.Interval = 24 * time.Hour // Daily
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 90 // 90 days
	}

	return &RingBufferCleanupWorker{
		dbQueries:     q,
		logger:        logger,
		interval:      config.Interval,
		retentionDays: config.RetentionDays,
	}
}

// Start begins the ring buffer cleanup worker loop
func (w *RingBufferCleanupWorker) Start(ctx context.Context) {
	w.logger.Info("Ring buffer cleanup worker started",
		slog.Duration("interval", w.interval),
		slog.Int("retention_days", int(w.retentionDays)))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run immediately on start
	w.performCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Ring buffer cleanup worker stopping")
			return
		case <-ticker.C:
			w.performCleanup(ctx)
		}
	}
}

// performCleanup executes the cleanup operations
func (w *RingBufferCleanupWorker) performCleanup(ctx context.Context) {
	startTime := time.Now()
	w.logger.Info("Starting ring buffer cleanup",
		slog.Int("retention_days", int(w.retentionDays)))

	// Get cleanup configuration from database (if available)
	configs, err := w.dbQueries.GetRingBufferCleanupConfig(ctx)
	if err != nil {
		w.logger.Warn("Failed to get cleanup config, using defaults",
			slog.Any("error", err))
	}

	retentionDays := w.retentionDays
	// Override with database config if available
	if len(configs) > 0 {
		// Use the first enabled config (they should all have the same retention period)
		retentionDays = configs[0].RetentionDays
	}

	// Clean up sender delivery attempts
	senderErr := w.cleanupSenderDeliveryAttempts(ctx, retentionDays)
	if senderErr != nil {
		w.logger.Error("Failed to cleanup sender delivery attempts",
			slog.Any("error", senderErr))
	}

	// Clean up route delivery attempts
	routeErr := w.cleanupRouteDeliveryAttempts(ctx, retentionDays)
	if routeErr != nil {
		w.logger.Error("Failed to cleanup route delivery attempts",
			slog.Any("error", routeErr))
	}

	// Update cleanup timestamps if we have configs
	if len(configs) > 0 {
		for _, config := range configs {
			err := w.dbQueries.UpdateRingBufferCleanupTimestamp(ctx, config.TableName)
			if err != nil {
				w.logger.Warn("Failed to update cleanup timestamp",
					slog.String("table_name", config.TableName),
					slog.Any("error", err))
			}
		}
	}

	duration := time.Since(startTime)
	w.logger.Info("Ring buffer cleanup completed",
		slog.Duration("duration", duration),
		slog.Int("retention_days", int(retentionDays)),
		slog.Bool("sender_cleanup_success", senderErr == nil),
		slog.Bool("route_cleanup_success", routeErr == nil))
}

// cleanupSenderDeliveryAttempts deletes old sender delivery attempt records
func (w *RingBufferCleanupWorker) cleanupSenderDeliveryAttempts(ctx context.Context, retentionDays int32) error {
	w.logger.Debug("Cleaning up sender delivery attempts",
		slog.Int("retention_days", int(retentionDays)))

	err := w.dbQueries.CleanupOldDeliveryAttempts(ctx, retentionDays)
	if err != nil {
		return err
	}

	w.logger.Info("Sender delivery attempts cleanup completed",
		slog.Int("retention_days", int(retentionDays)))

	return nil
}

// cleanupRouteDeliveryAttempts deletes old route delivery attempt records
func (w *RingBufferCleanupWorker) cleanupRouteDeliveryAttempts(ctx context.Context, retentionDays int32) error {
	w.logger.Debug("Cleaning up route delivery attempts",
		slog.Int("retention_days", int(retentionDays)))

	err := w.dbQueries.CleanupOldRouteAttempts(ctx, retentionDays)
	if err != nil {
		return err
	}

	w.logger.Info("Route delivery attempts cleanup completed",
		slog.Int("retention_days", int(retentionDays)))

	return nil
}
