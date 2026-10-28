package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/internal/sp"
)

// DLR sources recorded on message_segments.dlr_source
const (
	DLRSourceManual = "manual"
	DLRSourceAuto   = "auto"
)

// ManualDLRStatuses are the SMPP DLR stat values an admin or auto DLR may set.
var ManualDLRStatuses = []string{"DELIVRD", "UNDELIV", "EXPIRED", "REJECTD", "DELETED"}

// ErrInvalidDLRStatus is returned when the requested status is not in ManualDLRStatuses.
var ErrInvalidDLRStatus = errors.New("invalid DLR status")

// ManualDLRRequest describes a DLR that did not come from the MNO.
type ManualDLRRequest struct {
	MessageID int64
	Status    string
	ErrorCode string
	Source    string // DLRSourceManual or DLRSourceAuto
	// SentBefore limits the update to segments sent to the MNO before this time. Zero means all pending segments.
	SentBefore time.Time
}

// NormalizeManualDLRStatus upper cases status and checks it is allowed.
func NormalizeManualDLRStatus(status string) (string, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, s := range ManualDLRStatuses {
		if s == status {
			return status, nil
		}
	}
	return "", fmt.Errorf("%w: %q (allowed: %s)", ErrInvalidDLRStatus, status, strings.Join(ManualDLRStatuses, ", "))
}

// ManualDLRResult reports what ApplyManualDLR did.
type ManualDLRResult struct {
	SegmentsUpdated int
	// QueueID is the dlr_forwarding_queue job that forwards the DLR to the SP; 0 when nothing was pending.
	QueueID int64
	// ForwardWarning is set when the SP is not configured to receive DLRs, so the
	// queued job will fail until the SP's callback/connection is fixed.
	ForwardWarning string
}

// ApplyManualDLR completes the segments of a message that the MNO accepted but
// that have no final DLR (none, no_dlr or an intermediate status) and queues the
// DLR for forwarding to the service provider in the same statement, so a DLR is
// never applied without being forwarded. It then updates the message final
// status (reversing the debit once on failure). Segments with a final DLR are
// left alone; SegmentsUpdated 0 means nothing was pending and nothing was queued.
// Sender and route health are not touched because the status did not come from the MNO.
func (s *DLRService) ApplyManualDLR(ctx context.Context, req ManualDLRRequest) (ManualDLRResult, error) {
	logCtx := logging.ContextWithMessageID(ctx, req.MessageID)
	var result ManualDLRResult

	status, err := NormalizeManualDLRStatus(req.Status)
	if err != nil {
		return result, err
	}

	var errorCode *string
	if req.ErrorCode != "" {
		errorCode = &req.ErrorCode
	}
	source := req.Source
	if source == "" {
		source = DLRSourceManual
	}
	var sentBefore pgtype.Timestamptz
	if !req.SentBefore.IsZero() {
		sentBefore = pgtype.Timestamptz{Time: req.SentBefore, Valid: true}
	}

	dlrInfo := mno.DLRInfo{
		Status:    status,
		ErrorCode: req.ErrorCode,
		Timestamp: time.Now(),
	}
	// Build the forwarding payload first: if it cannot be built nothing is changed.
	payload, spDetails, err := s.buildDLRForwardingPayload(logCtx, req.MessageID, dlrInfo)
	if err != nil {
		return result, err
	}

	row, err := s.dbQueries.ApplyManualDLRAndEnqueue(logCtx, database.ApplyManualDLRAndEnqueueParams{
		DlrStatus:   &status,
		ErrorCode:   errorCode,
		DlrSource:   &source,
		MessageID:   req.MessageID,
		SentBefore:  sentBefore,
		Payload:     payload,
		MaxAttempts: DLRMaxRetries,
	})
	if err != nil {
		return result, fmt.Errorf("failed to apply %s DLR and queue it for forwarding: %w", source, err)
	}
	result.SegmentsUpdated = int(row.SegmentsUpdated)
	result.QueueID = row.QueueID
	if result.SegmentsUpdated == 0 {
		slog.InfoContext(logCtx, "No pending segments for DLR", slog.String("dlr_source", source))
		return result, nil
	}
	result.ForwardWarning = forwardWarning(spDetails)

	slog.InfoContext(logCtx, "Applied DLR to pending segments and queued it for forwarding",
		slog.String("dlr_source", source),
		slog.String("dlr_status", status),
		slog.Int("segments", result.SegmentsUpdated),
		slog.Int64("queue_id", result.QueueID),
	)
	if result.ForwardWarning != "" {
		slog.WarnContext(logCtx, "Queued DLR cannot be forwarded until SP config is fixed", slog.String("warning", result.ForwardWarning))
	}

	s.aggregateMessageStatus(logCtx, req.MessageID)

	return result, nil
}

// forwardWarning explains why a DLR for this SP cannot be delivered, or returns "".
func forwardWarning(spDetails sp.SPDetails) string {
	switch spDetails.Protocol {
	case "http":
		if spDetails.HTTPCallbackURL == "" {
			return "service provider has no HTTP callback URL configured; the DLR is queued but cannot be delivered until one is set and the job is retried"
		}
	case "smpp":
		if spDetails.SMPPSystemID == "" {
			return "service provider SMPP credential has no system ID; the DLR is queued but cannot be delivered"
		}
	}
	return ""
}
