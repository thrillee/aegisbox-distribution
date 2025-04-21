package workers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// WorkerFunc defines the function signature for work performed by a worker loop.
// It returns the number of items processed and any critical error encountered.
type WorkerFunc func(ctx context.Context, batchSize int) (int, error)

// runWorkerLoop runs a generic worker function periodically.
func runWorkerLoop(ctx context.Context, name string, interval time.Duration, batchSize int, workerFunc WorkerFunc) {
	log.Printf("%s worker starting with interval %v, batch size %d", name, interval, batchSize)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run once immediately at startup? Optional.
	// runWork(ctx, name, batchSize, workerFunc)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s worker stopping...", name)
			return
		case <-ticker.C:
			runWork(ctx, name, batchSize, workerFunc)
		}
	}
}

// runWork executes a single batch of work with a timeout.
func runWork(ctx context.Context, name string, batchSize int, workerFunc WorkerFunc) {
	// Create a timeout for this specific run
	runCtx, cancel := context.WithTimeout(ctx, 1*time.Minute) // TODO: Make timeout configurable
	defer cancel()

	processedCount, err := workerFunc(runCtx, batchSize)

	if err != nil {
		// Don't log expected "no rows" error unless debugging
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Error in %s worker run: %v", name, err)
		}
		// Potentially implement backoff or circuit breaker logic here on persistent errors
	} else if processedCount > 0 {
		log.Printf("%s worker processed %d items.", name, processedCount)
	}
	// If processedCount is 0 and err is nil or pgx.ErrNoRows, it just means no work was available.
}
