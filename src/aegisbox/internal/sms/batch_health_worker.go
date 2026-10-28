package sms

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/thrillee/aegisbox/internal/database"
)

type HealthRecord struct {
	SenderID         int32
	MNOID            int32
	MessageSegmentID *int64
	Status           string
	ErrorCode        *string
	SentAt           time.Time
	DLRReceivedAt    *time.Time
}

type BatchHealthWorker struct {
	dbQueries     database.Querier
	logger        *slog.Logger
	recordChan    chan HealthRecord
	wg            sync.WaitGroup
	stopChan      chan struct{}
	flushInterval time.Duration
	batchSize     int
}

func NewBatchHealthWorker(
	dbQueries database.Querier,
	logger *slog.Logger,
	flushInterval time.Duration,
	batchSize int,
) *BatchHealthWorker {
	return &BatchHealthWorker{
		dbQueries:     dbQueries,
		logger:        logger,
		recordChan:    make(chan HealthRecord, 10000),
		stopChan:      make(chan struct{}),
		flushInterval: flushInterval,
		batchSize:     batchSize,
	}
}

func (w *BatchHealthWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
	slog.InfoContext(ctx, "Batch health worker started")
}

func (w *BatchHealthWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	slog.Info("Batch health worker stopped")
}

func (w *BatchHealthWorker) RecordHealth(record HealthRecord) {
	select {
	case w.recordChan <- record:
	default:
		w.logger.Warn("Health record channel full, recording immediately")
		w.recordHealthSync(context.Background(), record)
	}
}

func (w *BatchHealthWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	var batch []HealthRecord

	for {
		select {
		case <-ctx.Done():
			w.flushBatch(ctx, batch)
			return

		case <-w.stopChan:
			w.flushBatch(ctx, batch)
			return

		case record := <-w.recordChan:
			batch = append(batch, record)
			if len(batch) >= w.batchSize {
				w.flushBatch(ctx, batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				w.flushBatch(ctx, batch)
				batch = nil
			}
		}
	}
}

func (w *BatchHealthWorker) flushBatch(ctx context.Context, batch []HealthRecord) {
	if len(batch) == 0 {
		return
	}

	w.logger.InfoContext(ctx, "Flushing health records", slog.Int("count", len(batch)))

	for _, record := range batch {
		w.recordHealthSync(ctx, record)
	}
}

func (w *BatchHealthWorker) recordHealthSync(ctx context.Context, record HealthRecord) {
	dlrTime := pgtype.Timestamptz{}
	if record.DLRReceivedAt != nil {
		dlrTime = pgtype.Timestamptz{Time: *record.DLRReceivedAt, Valid: true}
	}

	_, err := w.dbQueries.RecordSenderDeliveryAttempt(ctx, database.RecordSenderDeliveryAttemptParams{
		SenderID:         record.SenderID,
		MnoID:            record.MNOID,
		MessageSegmentID: record.MessageSegmentID,
		Status:           record.Status,
		ErrorCode:        record.ErrorCode,
		SentAt:           pgtype.Timestamptz{Time: record.SentAt, Valid: true},
		DlrReceivedAt:    dlrTime,
		RecordedDate:     pgtype.Date{Time: record.SentAt, Valid: true},
	})
	if err != nil {
		w.logger.Error("Failed to record health metric",
			slog.Int("sender_id", int(record.SenderID)),
			slog.Int("mno_id", int(record.MNOID)),
			slog.Any("error", err),
		)
	}
}
