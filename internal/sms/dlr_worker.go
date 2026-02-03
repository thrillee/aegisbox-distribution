package sms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
)

const (
	DLRBaseBackoff  = 5 * time.Second
	DLRMaxBackoff   = 5 * time.Minute
	DLRMaxRetries   = 10
	DLRJitterFactor = 0.3
)

type DLRForwarderFactory interface {
	GetForwarder(protocol string) (sp.DLRForwarder, error)
}

// DLRWorker processes DLR forwarding jobs from the database queue.
type DLRWorker struct {
	workerID         string // Unique ID for this worker instance
	dbQueries        database.Querier
	forwarderFactory DLRForwarderFactory
}

// NewDLRWorker creates a new DLR forwarding worker.
func NewDLRWorker(q database.Querier, factory DLRForwarderFactory) *DLRWorker {
	hostname, _ := os.Hostname()
	workerUUID, _ := uuid.NewRandom()
	workerID := fmt.Sprintf("%s-%s", hostname, workerUUID.String())

	return &DLRWorker{
		workerID:         workerID,
		dbQueries:        q,
		forwarderFactory: factory,
	}
}

// ProcessDLRForwardingBatch is the WorkerFunc compatible function.
func (w *DLRWorker) ProcessDLRForwardingBatch(ctx context.Context, batchSize int) (processedCount int, err error) {
	logCtx := logging.ContextWithWorkerID(ctx, w.workerID)
	slog.DebugContext(logCtx, "Checking for pending DLR forwarding jobs...")

	// Use the backoff-aware query
	jobs, err := w.dbQueries.GetPendingDLRsToForward(logCtx, database.GetPendingDLRsToForwardParams{
		Limit:    int32(batchSize),
		LockedBy: &w.workerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.DebugContext(logCtx, "No pending DLR forwarding jobs found.")
			return 0, nil // Not an error
		}
		slog.ErrorContext(logCtx, "Failed to fetch pending DLR forwarding jobs", slog.Any("error", err))
		return 0, err // Return DB error
	}

	if len(jobs) == 0 {
		return 0, nil // No jobs to process
	}

	slog.InfoContext(logCtx, "Fetched DLR forwarding jobs", slog.Int("count", len(jobs)))

	// 2. Process Each Job
	for _, job := range jobs {
		jobCtx := logging.ContextWithMessageID(logCtx, job.MessageID) // Use job's MessageID
		jobCtx = logging.ContextWithJobID(jobCtx, job.ID)             // Add job ID if helper exists

		slog.InfoContext(jobCtx, "Processing DLR forwarding job", slog.Int("attempts", int(job.Attempts)))

		// Deserialize payload
		var payloadData map[string]json.RawMessage // Use RawMessage for flexibility initially
		var spDetails sp.SPDetails
		var dlrInfo sp.ForwardedDLRInfo

		if err := json.Unmarshal(job.Payload, &payloadData); err != nil {
			slog.ErrorContext(jobCtx, "Failed to unmarshal outer job payload JSON", slog.Any("error", err))
			errMsg := fmt.Sprintf("Unmarshal outer payload error: %v", err)
			_ = w.dbQueries.MarkDLRForwardingAttemptFailed(jobCtx, database.MarkDLRForwardingAttemptFailedParams{
				ErrorMessage: &errMsg,
				ID:           job.ID,
			})
			continue // Move to next job
		}

		if err := json.Unmarshal(payloadData["sp_details"], &spDetails); err != nil {
			slog.ErrorContext(jobCtx, "Failed to unmarshal SPDetails from job payload", slog.Any("error", err))
			errMsg := fmt.Sprintf("Unmarshal SPDetails error: %v", err)
			_ = w.dbQueries.MarkDLRForwardingAttemptFailed(jobCtx, database.MarkDLRForwardingAttemptFailedParams{
				ErrorMessage: &errMsg,
				ID:           job.ID,
			})
			continue
		}
		if err := json.Unmarshal(payloadData["dlr_info"], &dlrInfo); err != nil {
			slog.ErrorContext(jobCtx, "Failed to unmarshal ForwardedDLRInfo from job payload", slog.Any("error", err))
			errMsg := fmt.Sprintf("Unmarshal DLRInfo error: %v", err)
			_ = w.dbQueries.MarkDLRForwardingAttemptFailed(jobCtx, database.MarkDLRForwardingAttemptFailedParams{
				ErrorMessage: &errMsg,
				ID:           job.ID,
			})
			continue
		}

		// Get the correct forwarder
		forwarder, err := w.forwarderFactory.GetForwarder(spDetails.Protocol)
		if err != nil {
			slog.ErrorContext(jobCtx, "Failed to get DLR forwarder for protocol", slog.String("protocol", spDetails.Protocol), slog.Any("error", err))
			errMsg := fmt.Sprintf("Get forwarder error: %v", err)
			_ = w.dbQueries.MarkDLRForwardingAttemptFailed(jobCtx, database.MarkDLRForwardingAttemptFailedParams{
				ErrorMessage: &errMsg,
				ID:           job.ID,
			})
			continue
		}

		// Attempt to forward
		// Add a timeout per attempt? Context already has overall worker run timeout.
		forwardErr := forwarder.ForwardDLR(jobCtx, spDetails, dlrInfo)

		// Update Job Status
		if forwardErr != nil {
			slog.WarnContext(jobCtx, "Failed forwarding attempt for DLR job", slog.Any("error", forwardErr), slog.Int("attempts", int(job.Attempts+1)), slog.Int("max_attempts", int(job.MaxAttempts)))
			// Mark failure, potentially retrying if attempts remain
			forwardErrMsg := forwardErr.Error()
			_ = w.dbQueries.MarkDLRForwardingAttemptFailed(jobCtx, database.MarkDLRForwardingAttemptFailedParams{
				ErrorMessage: &forwardErrMsg,
				ID:           job.ID,
			})
		} else {
			slog.InfoContext(jobCtx, "Successfully forwarded DLR job",
				slog.String("status", dlrInfo.Status),
				slog.String("destination", dlrInfo.DestAddr))
			_ = w.dbQueries.MarkDLRForwardingSuccess(jobCtx, job.ID)
			processedCount++
		}
	}

	slog.InfoContext(logCtx, "DLR forwarding batch completed",
		slog.Int("total", len(jobs)),
		slog.Int("processed", processedCount),
	)

	return processedCount, nil
}
