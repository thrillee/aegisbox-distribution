package sms

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
)

// --- Step Results (Assuming types from previous version are correct) ---

type RoutingResult struct {
	MNOID     int32 // Use sql.NullInt32 for nullable FK
	ShouldUse bool
	Error     error
}

type ValidationResult struct {
	ApprovedSenderID int32 // Use sql.NullInt32 for nullable FK
	TemplateID       int32 // Use sql.NullInt32 for nullable FK
	IsValid          bool
	Error            error
}

type PricingResult struct {
	Debited bool
	Error   error
}

type SendingResult struct {
	MNOConnectionID sql.NullInt32  // Use sql.NullInt32
	MNOMessageID    sql.NullString // Use sql.NullString for nullable Varchar
	Sent            bool
	Error           error
}

// --- Step Interfaces (Assuming interfaces from previous version are correct) ---

type Router interface {
	Route(ctx context.Context, msisdn string) (*RoutingResult, error)
}

type Validator interface {
	Validate(ctx context.Context, spID int32, senderID string) (*ValidationResult, error)
}

type Pricer interface {
	PriceAndDebit(ctx context.Context, msgID int64) (*PricingResult, error)
}

type Sender interface {
	Send(ctx context.Context, mnoID int32, params database.GetMessageDetailsForSendingRow) (*SendingResult, error)
}

// SMPPServer Interface for DLR Forwarding dependency
type SMPPServer interface {
	GetBoundClientWriter(systemID string) (pdu.Writer, error) // Assuming pdu.Writer interface exists
}

// --- Processor ---

type Processor struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool // Needed by Pricer
	router    Router
	validator Validator
	pricer    Pricer
	sender    Sender
}

// NewProcessor creates a new SMS Processor.
func NewProcessor(pool *pgxpool.Pool, queries database.Querier, clientMgr *smppclient.Manager) *Processor {
	// Initialize steps
	router := NewDefaultRouter(queries)
	validator := NewDefaultValidator(queries)
	pricer := NewDefaultPricer(pool) // Corrected: Pricer only needs the pool
	sender := NewDefaultSender(queries, clientMgr)

	return &Processor{
		dbPool:    pool,
		dbQueries: queries, // Keep for steps not needing direct Tx control
		router:    router,
		validator: validator,
		pricer:    pricer,
		sender:    sender,
	}
}

// ProcessRoutingStep fetches messages needing routing/validation and processes them.
func (p *Processor) ProcessRoutingStep(ctx context.Context, batchSize int) (int, error) {
	// Fetch messages in 'received' state
	messages, err := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: "received",
		Limit:            int32(batchSize),
	})
	// Handle pgx.ErrNoRows gracefully - it's not an error, just no work
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for routing", slog.Any("error", err))
		return 0, fmt.Errorf("fetching messages for routing: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		slog.DebugContext(ctx, "No messages found in 'received' state for routing.")
		return 0, nil // No error, just no messages
	}

	processedCount := 0
	for _, msg := range messages {
		// Process each message, log errors individually
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)
		err := p.validateAndRouteMessage(logCtx, msg) // Pass context with IDs
		if err != nil {
			// Log the error encountered during processing this specific message
			slog.ErrorContext(logCtx, "Failed to process routing/validation for message", slog.Any("error", err))
			// Decide on retry logic or permanent failure marking here if needed
		} else {
			processedCount++
		}
	}
	return processedCount, nil
}

// validateAndRouteMessage performs validation and routing for a single message.
// Now takes database.GetMessagesByStatusRow assuming it has needed fields. Adjust if model is different.
func (p *Processor) validateAndRouteMessage(ctx context.Context, msg database.GetMessagesByStatusRow) error {
	slog.DebugContext(ctx, "Starting validation and routing")

	// 1. Validate
	// Pass enriched context down
	validationRes, valErr := p.validator.Validate(ctx, msg.ServiceProviderID, msg.SenderIDUsed)
	if valErr != nil {
		slog.WarnContext(ctx, "Internal error during validation step", slog.Any("error", valErr))
		// Continue even if there was an internal error, the result might still indicate logical failure
	}
	if validationRes == nil {
		validationRes = &ValidationResult{IsValid: false, Error: errors.New("validation step returned nil result")}
		slog.ErrorContext(ctx, "Validation step returned unexpected nil result")
	}

	// 2. Route (only if validation didn't result in immediate failure)
	var routingRes *RoutingResult
	var routeErr error
	if validationRes.IsValid {
		// Pass enriched context down
		routingRes, routeErr = p.router.Route(ctx, msg.DestinationMsisdn)
		if routeErr != nil {
			slog.WarnContext(ctx, "Internal error during routing step", slog.Any("error", routeErr))
		}
		if routingRes == nil {
			routingRes = &RoutingResult{ShouldUse: false, Error: errors.New("routing step returned nil result")}
			slog.ErrorContext(ctx, "Routing step returned unexpected nil result")
		}
	} else {
		routingRes = &RoutingResult{ShouldUse: false, Error: validationRes.Error} // Carry over validation error
	}

	// 3. Determine Final Status and Update DB
	var finalStatus string
	var combinedError error // The logical error describing why processing stopped

	if !validationRes.IsValid {
		finalStatus = "failed_validation"
		combinedError = validationRes.Error // This should contain the reason (e.g., "sender ID not approved")
		if combinedError == nil && valErr != nil {
			combinedError = valErr // Use internal error if logical error is nil
		} else if combinedError == nil {
			combinedError = errors.New("validation failed for unknown reason")
		}
	} else if !routingRes.ShouldUse {
		finalStatus = "failed_routing"
		combinedError = routingRes.Error // This should contain the reason (e.g., "no route rule matched")
		if combinedError == nil && routeErr != nil {
			combinedError = routeErr // Use internal error if logical error is nil
		} else if combinedError == nil {
			combinedError = fmt.Errorf("no applicable route found for %s", msg.DestinationMsisdn)
		}
	} else {
		finalStatus = "routed" // Success!
		if routingRes.MNOID > 0 {
			ctx = logging.ContextWithMNOID(ctx, routingRes.MNOID)
		}
	}

	var dbErrMsg string
	if combinedError != nil {
		dbErrMsg = combinedError.Error()
	}

	// Use sqlc generated params struct correctly
	params := database.UpdateMessageValidatedRoutedParams{
		ProcessingStatus: finalStatus,
		RoutedMnoID:      &routingRes.MNOID,
		ApprovedSenderID: &validationRes.ApprovedSenderID,
		// TemplateID: validationRes.TemplateID, // Pass sql.NullInt32 if using templates
		ErrorDescription: &dbErrMsg, // Pass sql.NullString directly
		ID:               msg.ID,
	}

	err := p.dbQueries.UpdateMessageValidatedRouted(ctx, params)
	if err != nil {
		// Log the DB update error critically
		slog.ErrorContext(ctx, "Failed to update message status after validation/routing",
			slog.Any("error", err),
			slog.String("attempted_status", finalStatus),
		)
		// Return the DB error, indicating failure to persist state
		return fmt.Errorf("failed to update status for msg ID %d after validation/routing: %w", msg.ID, err)
	}

	// Log overall outcome of the step
	logLevel := slog.LevelInfo
	if finalStatus != "routed" {
		logLevel = slog.LevelWarn // Log failures as warnings
	}
	slog.LogAttrs(ctx, logLevel, "Routing Step Completed",
		slog.String("final_status", finalStatus),
		slog.Int64("msg_id", msg.ID), // msg_id already in context, but can add explicitly
		slog.Any("sp_id", msg.ServiceProviderID),
		slog.Any("routed_mno_id", routingRes.MNOID),
		slog.Any("approved_sid", validationRes.ApprovedSenderID),
		slog.Any("reason", combinedError), // Use "reason" for the logical error
	)
	return nil // Return nil if DB update succeeded, even if logical validation/routing failed
}

// ProcessPricingStep fetches messages needing pricing and processes them.
func (p *Processor) ProcessPricingStep(ctx context.Context, batchSize int) (int, error) {
	messages, err := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: "routed", // Process messages that are successfully routed
		Limit:            int32(batchSize),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for pricing", slog.Any("error", err))
		return 0, fmt.Errorf("fetching messages for pricing: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		slog.DebugContext(ctx, "No messages found in 'routed' state for pricing.")
		return 0, nil
	}

	processedCount := 0
	for _, msg := range messages {
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)
		// Pricer handles its own transaction and logging internally
		pricingRes, err := p.pricer.PriceAndDebit(logCtx, msg.ID) // Pass context down
		if err != nil {
			// Log errors returned by the PriceAndDebit step itself (e.g., begin Tx error)
			slog.ErrorContext(logCtx, "Error processing pricing/debiting step", slog.Any("error", err))
			// Error is already logged within Pricer if it's a logical failure (e.g., insufficient funds)
		} else if pricingRes != nil && pricingRes.Debited {
			processedCount++
		}
	}
	return processedCount, nil
}

// ProcessSendingStep fetches messages ready to be sent and processes them.
func (p *Processor) ProcessSendingStep(ctx context.Context, batchSize int) (int, error) {
	messages, err := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: "queued_for_send", // Process messages successfully priced/debited
		Limit:            int32(batchSize),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for sending", slog.Any("error", err))
		return 0, fmt.Errorf("fetching messages for sending: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		slog.DebugContext(ctx, "No messages found in 'queued_for_send' state for sending.")
		return 0, nil
	}

	processedCount := 0
	for _, msg := range messages {
		// Prepare context for this message
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)

		// Need full details for sending
		fullMsg, err := p.dbQueries.GetMessageDetailsForSending(logCtx, msg.ID)
		if err != nil {
			slog.ErrorContext(logCtx, "Error fetching full details for sending message", slog.Any("error", err))
			// Mark as failed permanently as we can't even get details
			failErr := fmt.Errorf("failed to fetch details: %w", err)
			_ = p.markSendFailed(logCtx, msg.ID, nil, failErr) // Use logCtx
			continue
		}

		// Use correct check for sql.NullInt32
		if fullMsg.RoutedMnoID == nil {
			failErr := errors.New("missing Routed MNO ID")
			slog.ErrorContext(logCtx, "Cannot send message", slog.Any("reason", failErr))
			_ = p.markSendFailed(logCtx, msg.ID, nil, failErr)
			continue
		}
		// Add MNO ID to context
		logCtx = logging.ContextWithMNOID(logCtx, *fullMsg.RoutedMnoID)

		// Sender handles segmentation and actual sending attempts
		sendResult, err := p.sender.Send(logCtx, *fullMsg.RoutedMnoID, fullMsg)
		if err != nil {
			// Log internal errors within the sender.Send call itself (e.g., client selection)
			slog.ErrorContext(logCtx, "Internal error trying to send message", slog.Any("error", err))
			_ = p.markSendFailed(logCtx, msg.ID, nil, err)
			continue
		}

		// Update status based on sending attempt result (success/failure logged within these)
		if sendResult.Sent {
			err = p.markSent(logCtx, msg.ID, sendResult)
			if err == nil {
				processedCount++
			}
		} else {
			err = p.markSendFailed(logCtx, msg.ID, sendResult, sendResult.Error) // Pass logical send error
		}
		// Log error only if DB update failed
		if err != nil {
			slog.ErrorContext(logCtx, "Error updating database after sending attempt", slog.Any("error", err))
		}
	}
	return processedCount, nil
}

// markSent updates DB after successful send attempt (potentially partial for multipart)
func (p *Processor) markSent(ctx context.Context, msgID int64, res *SendingResult) error {
	// This function might need adjustment depending on whether MarkMessageSent
	// still exists or if status updates are now purely per-segment.
	// Assuming MarkMessageSent updates the main message status to 'send_attempted' or similar.
	// Need to define MarkMessageSent query if it's still relevant.
	// Let's assume we only log here now, segment updates happen in Sender.Send
	slog.InfoContext(ctx, "Sending Step Completed (attempted)",
		slog.Int64("msg_id", msgID), // Already in context, but explicit is ok
	)
	return nil // Placeholder - actual DB update might be needed depending on final state machine
}

// markSendFailed updates DB after failed send attempt
func (p *Processor) markSendFailed(ctx context.Context, msgID int64, res *SendingResult, sendErr error) error {
	errMsg := "send failed"
	if sendErr != nil {
		errMsg = sendErr.Error()
	} else if res != nil && res.Error != nil {
		errMsg = res.Error.Error()
	}

	dbErrMsg := errMsg

	// Update main message status to indicate failure
	err := p.dbQueries.MarkMessageSendFailed(ctx, database.MarkMessageSendFailedParams{
		ID:               msgID,
		ErrorDescription: &dbErrMsg, // Use sql.NullString
		// ErrorCode: // Map SMPP error code if available
	})
	if err == nil {
		slog.WarnContext(ctx, "Sending Step Failed",
			slog.Int64("msg_id", msgID),
			slog.String("reason", errMsg),
		)
	} else {
		slog.ErrorContext(ctx, "Failed to update DB status after send failure", slog.Any("error", err))
	}
	return err
}

// UpdateSegmentDLRStatus processes an incoming DLR for a specific segment.
func (p *Processor) UpdateSegmentDLRStatus(ctx context.Context, mnoMessageID, dlrStatus string, dlrErrorCode int32, server SMPPServer) error {
	logCtx := ctx // Use incoming context, may not have msg_id yet
	slog.InfoContext(logCtx, "Processing DLR",
		slog.String("mno_msg_id", mnoMessageID),
		slog.String("dlr_status", dlrStatus),
		slog.Any("dlr_error_code", dlrErrorCode),
	)

	// 1. Find the specific segment record
	segmentInfo, err := p.dbQueries.FindSegmentByMnoMessageID(logCtx, &mnoMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "Received DLR for unknown MNO message ID", slog.String("mno_msg_id", mnoMessageID))
			return nil // Not an error we can handle further
		}
		slog.ErrorContext(logCtx, "DB error finding segment for DLR", slog.String("mno_msg_id", mnoMessageID), slog.Any("error", err))
		return fmt.Errorf("db error finding segment for DLR (MNO ID %s): %w", mnoMessageID, err)
	}

	// Enrich context with MessageID and SegmentID for subsequent logs/calls
	logCtx = logging.ContextWithMessageID(logCtx, segmentInfo.MessageID)
	// logCtx = logging.ContextWithSegmentID(logCtx, segmentInfo.SegmentID) // Add if needed

	// 2. Update the segment's DLR status
	err = p.dbQueries.UpdateSegmentDLR(logCtx, database.UpdateSegmentDLRParams{
		ID:        segmentInfo.SegmentID,
		DlrStatus: &dlrStatus,
		ErrorCode: &dlrErrorCode,
	})
	if err != nil {
		// Log critical error, as we couldn't update the status
		slog.ErrorContext(logCtx, "CRITICAL: Failed to update segment DLR status in DB",
			slog.Int64("segment_id", segmentInfo.SegmentID),
			slog.String("mno_msg_id", mnoMessageID),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to update segment DLR status: %w", err)
	}
	slog.InfoContext(logCtx, "Updated segment DLR status",
		slog.Int64("segment_id", segmentInfo.SegmentID),
		slog.Any("segment_seqn", segmentInfo.SegmentSeqn),
		slog.String("dlr_status", dlrStatus),
	)

	// 3. Aggregate Status for Parent Message (Run asynchronously)
	// Pass enriched context if aggregateMessageStatus uses logging
	go p.aggregateMessageStatus(logging.ContextWithMessageID(context.Background(), segmentInfo.MessageID), segmentInfo.MessageID) // Pass clean context + ID

	// 4. Forward DLR (if needed) - Pass segmentInfo.MessageID
	// Pass enriched context if forwardDLR uses logging
	go p.forwardDLR(logging.ContextWithMessageID(context.Background(), segmentInfo.MessageID), segmentInfo.MessageID, dlrStatus, dlrErrorCode, server) // Pass clean context + ID

	return nil // DLR processing for this segment successful
}

// aggregateMessageStatus checks all segments and updates the parent message status
func (p *Processor) aggregateMessageStatus(ctx context.Context, messageID int64) {
	// Add message ID if not already in context (it should be from the caller)
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	slog.DebugContext(logCtx, "Starting message status aggregation")

	// Add slight delay to allow concurrent DLR updates for the same message to potentially settle? Risky. Consider locking or deterministic aggregation trigger.
	time.Sleep(100 * time.Millisecond)

	segments, err := p.dbQueries.GetSegmentStatusesForMessage(logCtx, messageID)
	if err != nil {
		// Don't log ErrNoRows as error here, could happen if message deleted?
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.ErrorContext(logCtx, "Error fetching segments for status aggregation", slog.Any("error", err))
		}
		return
	}

	totalSegmentsDB, err := p.dbQueries.GetMessageTotalSegments(logCtx, messageID)
	if err != nil {
		slog.ErrorContext(logCtx, "Error fetching total segment count for status aggregation", slog.Any("error", err))
		return
	}

	if len(segments) < int(totalSegmentsDB) { // Check if less than expected
		// Not all segments have reported yet (or DB inconsistency)
		slog.DebugContext(logCtx, "Skipping aggregation: Not all segment statuses received yet",
			slog.Int("found", len(segments)),
			slog.Any("expected", totalSegmentsDB),
		)
		return
	}
	if len(segments) > int(totalSegmentsDB) {
		slog.WarnContext(logCtx, "Inconsistent segment count during aggregation",
			slog.Int("found", len(segments)),
			slog.Any("expected", totalSegmentsDB),
		)
		// Proceed cautiously? Or fail? Let's proceed.
	}

	deliveredCount := 0
	failedCount := 0
	pendingCount := 0
	finalStatus := "unknown" // Default

	for _, seg := range segments {
		// Correctly handle sql.NullString
		status := ""
		if seg.DlrStatus != nil {
			status = strings.ToUpper(*seg.DlrStatus)
		}

		switch status {
		case "DELIVRD":
			deliveredCount++
		case "EXPIRED", "DELETED", "UNDELIV", "REJECTD":
			failedCount++
		case "": // Treat NULL/empty as pending
			pendingCount++
		default: // ACCEPTD, UNKNOWN etc. also treated as pending for final status aggregation
			pendingCount++
		}
	}

	// Determine final status based on counts
	if pendingCount > 0 {
		finalStatus = "processing" // Still waiting for some DLRs
	} else if failedCount > 0 {
		finalStatus = "failed" // If any segment failed, mark whole message failed
	} else if deliveredCount > 0 && deliveredCount >= int(totalSegmentsDB) { // Check delivered >= expected
		finalStatus = "delivered" // All expected segments delivered
	} else if deliveredCount > 0 && failedCount == 0 && pendingCount == 0 {
		// Edge case: Fewer segments reported than total_segments, but all are DELIVRD. Risky. Mark delivered?
		slog.WarnContext(logCtx, "Marking message delivered despite segment count mismatch", slog.Int("delivered", deliveredCount), slog.Int32("expected", totalSegmentsDB))
		finalStatus = "delivered"
	} else {
		slog.WarnContext(logCtx, "Inconsistent segment states during aggregation logic",
			slog.Int("delivered", deliveredCount),
			slog.Int("failed", failedCount),
			slog.Int("pending", pendingCount),
			slog.Any("expected", totalSegmentsDB),
		)
		finalStatus = "unknown"
	}

	// Only update if a terminal state is reached and status changed
	if finalStatus == "delivered" || finalStatus == "failed" {
		// Optional: Check current status before updating
		currentStatusRec, err := p.dbQueries.GetMessageFinalStatus(logCtx, messageID) // Needs query: SELECT final_status FROM messages WHERE id=$1
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "Could not get current final status before aggregation update", slog.Any("error", err))
		}

		// Only update if status is actually changing or not yet set correctly
		if err == nil && currentStatusRec.FinalStatus == finalStatus {
			slog.DebugContext(logCtx, "Final status already matches target, skipping update", slog.String("status", finalStatus))
		} else {
			slog.InfoContext(logCtx, "Attempting to update final message status", slog.String("new_status", finalStatus))
			err = p.dbQueries.UpdateMessageFinalStatus(logCtx, database.UpdateMessageFinalStatusParams{
				ID:          messageID,
				FinalStatus: finalStatus,
			})
			if err != nil {
				slog.ErrorContext(logCtx, "Error updating final message status", slog.String("attempted_status", finalStatus), slog.Any("error", err))
			} else {
				slog.InfoContext(logCtx, "Aggregated final message status updated successfully", slog.String("final_status", finalStatus))
			}
		}
	} else {
		slog.DebugContext(logCtx, "Aggregation determined non-terminal status, no update performed", slog.String("status", finalStatus))
	}
}

// forwardDLR (Extracted forwarding logic)
func (p *Processor) forwardDLR(ctx context.Context, messageID int64, dlrStatus string, dlrErrorCode sql.NullInt32, server SMPPServer) {
	logCtx := logging.ContextWithMessageID(ctx, messageID) // Use context with MsgID
	slog.DebugContext(logCtx, "Attempting to forward DLR")

	// TODO: Implement DLR forwarding logic using slog for logging
	// 1. Get SP Info (SystemID, ClientRef, Original Src/Dest Addr) from DB using messageID
	//    spInfo, err := p.dbQueries.GetSPInfoForForwarding(logCtx, messageID) ... handle error with slog
	// 2. Get active connection writer for SP's systemID from the server
	//    connWriter, err := server.GetBoundClientWriter(spInfo.SystemID) ... handle error with slog
	// 3. Construct deliver_sm PDU
	//    dlrPdu := pdu.NewDeliverSM() ... set fields ... use spInfo.ClientRef for 'id:' field
	// 4. Write PDU
	//    err = connWriter.Write(dlrPdu) ... handle error with slog
	slog.WarnContext(logCtx, "DLR forwarding not fully implemented") // Placeholder log
}

// --- Helper ---
// Interface for smppclient.Manager needed by NewProcessor
type SMPPClientManager interface {
	GetClient(ctx context.Context, mnoID int32) (*smppclient.MNOClient, error)
}

// Interface for SMPPServer needed by DLR forwarding
// type SMPPServer interface { ... } defined above
