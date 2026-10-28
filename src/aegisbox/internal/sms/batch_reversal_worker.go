package sms

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/wallet"
)

type ReversalRequest struct {
	MessageID int64
	Reason    string
}

type BatchReversalWorker struct {
	dbQueries     database.Querier
	walletSvc     wallet.Service
	reversalChan  chan ReversalRequest
	wg            sync.WaitGroup
	stopChan      chan struct{}
	flushInterval time.Duration
	batchSize     int
}

func NewBatchReversalWorker(
	dbQueries database.Querier,
	walletSvc wallet.Service,
	flushInterval time.Duration,
	batchSize int,
) *BatchReversalWorker {
	return &BatchReversalWorker{
		dbQueries:     dbQueries,
		walletSvc:     walletSvc,
		reversalChan:  make(chan ReversalRequest, 10000),
		stopChan:      make(chan struct{}),
		flushInterval: flushInterval,
		batchSize:     batchSize,
	}
}

func (w *BatchReversalWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
	slog.InfoContext(ctx, "Batch reversal worker started")
}

func (w *BatchReversalWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	slog.Info("Batch reversal worker stopped")
}

func (w *BatchReversalWorker) QueueReversal(messageID int64, reason string) {
	select {
	case w.reversalChan <- ReversalRequest{MessageID: messageID, Reason: reason}:
	default:
		slog.Warn("Reversal channel full, performing immediate reversal")
		go func() {
			_ = w.walletSvc.ReverseDebit(context.Background(), messageID, reason)
		}()
	}
}

func (w *BatchReversalWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	var batch []ReversalRequest

	for {
		select {
		case <-ctx.Done():
			w.flushBatch(ctx, batch)
			return

		case <-w.stopChan:
			w.flushBatch(ctx, batch)
			return

		case req := <-w.reversalChan:
			batch = append(batch, req)
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

func (w *BatchReversalWorker) flushBatch(ctx context.Context, batch []ReversalRequest) {
	if len(batch) == 0 {
		return
	}

	slog.InfoContext(ctx, "Flushing batch reversals", slog.Int("count", len(batch)))

	for _, req := range batch {
		err := w.walletSvc.ReverseDebit(ctx, req.MessageID, req.Reason)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to reverse debit in batch",
				slog.Int64("message_id", req.MessageID),
				slog.String("reason", req.Reason),
				slog.Any("error", err),
			)
		}
	}
}
