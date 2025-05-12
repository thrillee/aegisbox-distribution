package sms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/pkg/errormapper"
)

// ========================================================================================
// Interfaces this Processor relies on (Implementations should be injected)
// ========================================================================================

// Router finds the appropriate MNO for a destination MSISDN.
type Router interface {
	Route(ctx context.Context, msisdn string) (*mno.RoutingResult, error)
}

// Validator checks SenderID, templates, etc.
type Validator interface {
	Validate(ctx context.Context, spID int32, senderID string) (*ValidationResult, error)
}

// Pricer handles debiting the SP's wallet for the message cost.
type Pricer interface {
	PriceAndDebit(ctx context.Context, msgID int64) (*PricingResult, error)
}

// Sender handles the segmentation and submission of the message to the MNO Connector.
type Sender interface {
	Send(ctx context.Context, msgDetails database.GetMessageDetailsForSendingRow) (*mno.SubmitResult, error)
}

// WalletService provides wallet operations, including reversals.
type WalletService interface {
	ReverseDebit(ctx context.Context, messageID int64, reason string) error
}

// DLRForwarder handles sending DLRs back to the Service Provider.
type DLRForwarder interface {
	ForwardDLR(ctx context.Context, spDetails sp.SPDetails, dlr sp.ForwardedDLRInfo) error
}

// ========================================================================================
// Result Structs (Consider moving to specific interface packages if preferred)
// ========================================================================================

// ValidationResult holds the outcome of the validation step.
type ValidationResult struct {
	ApprovedSenderID int32
	TemplateID       int32
	IsValid          bool
	Error            error // Describes the logical validation failure, if any
}

// PricingResult holds the outcome of the pricing/debiting step.
type PricingResult struct {
	Debited bool
	Error   error // Describes logical failure (e.g., insufficient funds) or internal error
}

// ========================================================================================
// Core Processor Implementation
// ========================================================================================

// Processor orchestrates the SMS processing lifecycle steps.
type Processor struct {
	dbQueries     database.Querier // Interface for DB operations
	dbPool        *pgxpool.Pool    // Pool needed for direct Tx management if required
	router        Router
	validator     Validator
	pricer        Pricer
	sender        Sender
	walletService WalletService // Added for reversals
	dlrForwarder  DLRForwarder  // Added for DLR forwarding
}

// ProcessorDependencies holds all dependencies needed to create a Processor.
type ProcessorDependencies struct {
	DBPool        *pgxpool.Pool
	DBQueries     database.Querier
	Router        Router
	Validator     Validator
	Pricer        Pricer
	Sender        Sender
	WalletService WalletService
	DLRForwarder  DLRForwarder
}

// NewProcessor creates a new SMS Processor with explicit dependencies.
func NewProcessor(deps ProcessorDependencies) *Processor {
	var missing []string

	if deps.DBPool == nil {
		missing = append(missing, "DBPool")
	}
	if deps.DBQueries == nil {
		missing = append(missing, "DBQueries")
	}
	if deps.Router == nil {
		missing = append(missing, "Router")
	}
	if deps.Validator == nil {
		missing = append(missing, "Validator")
	}
	if deps.Pricer == nil {
		missing = append(missing, "Pricer")
	}
	// if deps.Sender == nil {
	// 	missing = append(missing, "Sender")
	// }
	if deps.WalletService == nil {
		missing = append(missing, "WalletService")
	}
	// if deps.DLRForwarder == nil {
	// 	missing = append(missing, "DLRForwarder")
	// }

	if len(missing) > 0 {
		// you can panic or return an error here
		panic(fmt.Sprintf(
			"Missing required dependencies for SMS Processor: %s",
			strings.Join(missing, ", "),
		))
	}

	if deps.WalletService == nil {
		panic("Missing WalletService dependency")
	}
	// if deps.DLRForwarder == nil {
	// 	panic("Missing DLRForwarder dependency")
	// }

	return &Processor{
		dbPool:        deps.DBPool,
		dbQueries:     deps.DBQueries,
		router:        deps.Router,
		validator:     deps.Validator,
		pricer:        deps.Pricer,
		sender:        deps.Sender,
		walletService: deps.WalletService,
		dlrForwarder:  deps.DLRForwarder,
	}
}

func (p *Processor) SetSender(sender Sender) {
	p.sender = sender
}

// ========================================================================================
// Processing Steps (Called by Worker Manager)
// ========================================================================================

// ProcessRoutingStep fetches messages needing routing/validation and processes them.
func (p *Processor) ProcessRoutingStep(ctx context.Context, batchSize int) (processedCount int, err error) {
	defer func() { // Add recovery for panics within the loop
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "PANIC recovered in ProcessRoutingStep loop", slog.Any("panic_info", r))
			err = fmt.Errorf("panic occurred during routing step: %v", r) // Report error back to worker manager
		}
	}()

	messages, fetchErr := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: "received",
		Limit:            int32(batchSize),
	})
	if fetchErr != nil && !errors.Is(fetchErr, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for routing", slog.Any("error", fetchErr))
		return 0, fmt.Errorf("fetching messages for routing: %w", fetchErr)
	}
	if errors.Is(fetchErr, pgx.ErrNoRows) || len(messages) == 0 {
		slog.DebugContext(ctx, "No messages found in 'received' state for routing.")
		return 0, nil // No error, just no messages
	}

	for _, msg := range messages {
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)

		processErr := p.validateAndRouteMessage(logCtx, msg)
		if processErr != nil {
			// Logging is handled within validateAndRouteMessage or its sub-steps
			// Decide if this error requires stopping the batch or just continuing
			// For now, continue processing other messages in the batch
			continue
		}
		processedCount++
	}
	return processedCount, nil
}

// validateAndRouteMessage performs validation and routing for a single message.
func (p *Processor) validateAndRouteMessage(ctx context.Context, msg database.GetMessagesByStatusRow) error {
	slog.DebugContext(ctx, "Starting validation and routing")
	var finalStatus string
	var logicalError error  // Describes the logical validation/routing failure reason
	var internalError error // Describes unexpected errors during the process

	// 1. Validate
	validationRes, valErr := p.validator.Validate(ctx, msg.ServiceProviderID, msg.OriginalSourceAddr)
	if valErr != nil {
		internalError = fmt.Errorf("validation internal error: %w", valErr)
		slog.ErrorContext(ctx, "Internal error during validation step", slog.Any("error", valErr), slog.Int("sp_id", int(msg.ServiceProviderID)), slog.String("Source Addr", msg.OriginalDestinationAddr))
	}
	if validationRes == nil {
		// Should not happen if validator implementation is correct
		validationRes = &ValidationResult{IsValid: false, Error: errors.New("validation step returned nil result")}
		slog.ErrorContext(ctx, "Validation step returned unexpected nil result")
		if internalError == nil {
			internalError = validationRes.Error
		}
	}

	// 2. Route (only if validation didn't cause immediate logical failure)
	var routingRes *mno.RoutingResult
	var routeErr error
	if validationRes.IsValid {
		routingRes, routeErr = p.router.Route(ctx, msg.OriginalDestinationAddr)
		if routeErr != nil {
			if internalError == nil {
				internalError = fmt.Errorf("routing internal error: %w", routeErr)
			} else {
				internalError = fmt.Errorf("%w; routing internal error: %w", internalError, routeErr)
			}
			slog.ErrorContext(ctx, "Internal error during routing step", slog.Any("error", routeErr))
		}
		if routingRes == nil {
			routingRes = &mno.RoutingResult{ShouldUse: false, Error: errors.New("routing step returned nil result")}
			slog.ErrorContext(ctx, "Routing step returned unexpected nil result")
			if internalError == nil {
				internalError = routingRes.Error
			}
		}
	} else {
		routingRes = &mno.RoutingResult{ShouldUse: false, Error: validationRes.Error} // Carry over validation error
	}

	// 3. Determine Final Status
	if !validationRes.IsValid {
		finalStatus = "failed_validation"
		logicalError = validationRes.Error
		if logicalError == nil {
			logicalError = errors.New("validation failed")
		}
	} else if !routingRes.ShouldUse {
		finalStatus = "failed_routing"
		logicalError = routingRes.Error
		if logicalError == nil {
			logicalError = fmt.Errorf("no route found")
		}
	} else {
		finalStatus = "routed" // Success!
		if routingRes.MNOID == 0 {
			ctx = logging.ContextWithMNOID(ctx, routingRes.MNOID)
		}
	}

	// 4. Prepare and Update DB
	var dbErrMsg string
	var errorCode string
	if logicalError != nil {
		dbErrMsg = logicalError.Error()
		// TODO: Map logicalError to a specific error code (e.g., "NO_ROUTE", "INVALID_SENDER")
		errorCode = errormapper.MapErrorCode(dbErrMsg, "smpp")
	} else if internalError != nil {
		// Log internal errors differently if needed, but might still use generic failure status
		dbErrMsg = fmt.Sprintf("Internal processing error: %v", internalError) // Provide internal error info
		// errorCode.String = "SYS_ERR"
		// errorCode.Valid = true
	}

	params := database.UpdateMessageValidatedRoutedParams{
		ProcessingStatus: finalStatus,
		RoutedMnoID:      &routingRes.MNOID,
		ApprovedSenderID: &validationRes.ApprovedSenderID,
		TemplateID:       nil,
		ErrorDescription: &dbErrMsg,
		ErrorCode:        &errorCode,
		ID:               msg.ID,
	}

	updateErr := p.dbQueries.UpdateMessageValidatedRouted(ctx, params)
	if updateErr != nil {
		slog.ErrorContext(ctx, "CRITICAL: Failed to update message status after validation/routing",
			slog.Any("db_error", updateErr),
			slog.String("attempted_status", finalStatus),
		)
		// This is a critical failure - return error to worker
		return fmt.Errorf("failed to update DB status for msg ID %d: %w", msg.ID, updateErr)
	}

	// 5. Trigger Failure Notification/DLR if needed
	logLevel := slog.LevelInfo
	if finalStatus != "routed" {
		logLevel = slog.LevelWarn
		// Generate and forward a failure DLR back to the SP
		// Need SP details (protocol, callback/systemID etc) -> requires fetching from DB
		go p.generateAndEnqueueFailureDLR(context.Background(), msg.ID, finalStatus, logicalError, &errorCode)
	}

	// Log overall outcome
	slog.LogAttrs(ctx, logLevel, "Routing Step Completed",
		slog.String("final_status", finalStatus),
		slog.Any("routed_mno_id", routingRes.MNOID),
		slog.Any("reason", logicalError),          // Log the logical reason for success/failure
		slog.Any("internal_error", internalError), // Log any unexpected errors separately
	)

	// If DB update succeeded, return nil even if logical checks failed (failure is recorded)
	// Return internalError only if it prevented DB update? No, DB update error is returned above.
	return nil
}

// ProcessPricingStep fetches messages needing pricing and processes them.
func (p *Processor) ProcessPricingStep(ctx context.Context, batchSize int) (processedCount int, err error) {
	defer func() { // Add recovery
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "PANIC recovered in ProcessPricingStep loop", slog.Any("panic_info", r))
			err = fmt.Errorf("panic occurred during pricing step: %v", r)
		}
	}()

	messages, fetchErr := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: "routed",
		Limit:            int32(batchSize),
	})
	if fetchErr != nil && !errors.Is(fetchErr, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for pricing", slog.Any("error", fetchErr))
		return 0, fmt.Errorf("fetching messages for pricing: %w", fetchErr)
	}
	if errors.Is(fetchErr, pgx.ErrNoRows) || len(messages) == 0 {
		slog.DebugContext(ctx, "No messages found in 'routed' state for pricing.")
		return 0, nil
	}

	for _, msg := range messages {
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)
		// Pricer handles its own transaction, internal logging, and status updates.
		pricingRes, pricingErr := p.pricer.PriceAndDebit(logCtx, msg.ID)
		if pricingErr != nil {
			// Log internal errors from the pricer step itself (e.g., begin Tx error)
			// Logical errors like insufficient funds are logged within PriceAndDebit.
			slog.ErrorContext(logCtx, "Error processing pricing/debiting step", slog.Any("error", pricingErr))
			// Should we trigger a failure DLR here? Pricer already updated status to failed_pricing.
			// go p.generateAndForwardFailureDLR(context.Background(), msg.ID, finalStatus, logicalError, errorCode) // Run async
			continue // Continue with next message
		}
		if pricingRes != nil && pricingRes.Debited {
			processedCount++
		} else if pricingRes != nil { // Debited is false, likely logical failure
			// Pricer already logged and set status to failed_pricing.
			// Generate failure DLR? Status is already updated. Maybe not needed here.
			slog.WarnContext(logCtx, "Pricing/debiting failed logically", slog.Any("reason", pricingRes.Error))
		}
	}
	return processedCount, nil
}

// ProcessSendingStep fetches messages ready to be sent and processes them.
func (p *Processor) ProcessSendingStep(ctx context.Context, batchSize int) (processedCount int, err error) {
	defer func() { // Add recovery
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "PANIC recovered in ProcessSendingStep loop", slog.Any("panic_info", r))
			err = fmt.Errorf("panic occurred during sending step: %v", r)
		}
	}()

	messages, fetchErr := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: "queued_for_send", // Process messages successfully priced/debited
		Limit:            int32(batchSize),
	})
	if fetchErr != nil && !errors.Is(fetchErr, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for sending", slog.Any("error", fetchErr))
		return 0, fmt.Errorf("fetching messages for sending: %w", fetchErr)
	}
	if errors.Is(fetchErr, pgx.ErrNoRows) || len(messages) == 0 {
		slog.DebugContext(ctx, "No messages found in 'queued_for_send' state for sending.")
		return 0, nil
	}

	for _, msg := range messages {
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)
		var sendErr error                  // To store error from sender.Send or pre-send failures
		var submitResult *mno.SubmitResult // Store result from sender

		// 1. Fetch full details needed for sending
		fullMsg, dbErr := p.dbQueries.GetMessageDetailsForSending(logCtx, msg.ID)
		if dbErr != nil {
			slog.ErrorContext(logCtx, "CRITICAL: Error fetching full details for sending message", slog.Any("error", dbErr))
			sendErr = fmt.Errorf("failed to fetch details: %w", dbErr)
			// Attempt reversal as we can't send, and debit already happened.
			p.attemptReversal(logCtx, msg.ID, sendErr.Error())
			_ = p.markAsFailed(logCtx, msg.ID, sendErr, "SYS_ERR") // Mark as failed
			continue                                               // Next message
		}

		// 2. Basic check before calling sender
		if fullMsg.RoutedMnoID == nil {
			sendErr = errors.New("missing Routed MNO ID")
			slog.ErrorContext(logCtx, "Cannot send message", slog.Any("reason", sendErr))
			p.attemptReversal(logCtx, msg.ID, sendErr.Error())
			_ = p.markAsFailed(logCtx, msg.ID, sendErr, "ROUTE_ERR") // Mark as failed
			continue                                                 // Next message
		}
		logCtx = logging.ContextWithMNOID(logCtx, *fullMsg.RoutedMnoID)

		// 3. Call the Sender (handles segmentation & MNO submission attempts)
		submitResult, sendErr = p.sender.Send(logCtx, fullMsg) // Pass fullMsg directly
		// sendErr here represents errors *within* the Send logic (e.g., getting MNO client, segmentation issues)
		// submitResult contains outcome per segment and overall MNO submission errors/success.

		slog.InfoContext(logCtx, "Submit Result", slog.Any("WasSubmitted", submitResult.WasSubmitted))

		// 4. Process Result / Handle Errors
		if sendErr != nil {
			// Internal error within the Sender itself before/during submission attempts
			slog.ErrorContext(logCtx, "Internal error during send attempt", slog.Any("error", sendErr))
			p.attemptReversal(logCtx, msg.ID, fmt.Sprintf("Internal sender error: %v", sendErr))
			_ = p.markAsFailed(logCtx, msg.ID, sendErr, "SYS_ERR")
			continue // Next message
		}

		// Check the SubmitResult
		if !submitResult.WasSubmitted || submitResult.Error != nil {
			// Submission attempt failed definitively (e.g., MNO connection down, immediate rejection)
			failureReason := submitResult.Error
			if failureReason == nil {
				failureReason = errors.New("MNO submission failed or was not attempted")
			}
			errorCode := "MNO_FAIL" // TODO: Map specific MNO errors if possible
			slog.WarnContext(logCtx, "MNO submission failed", slog.Any("reason", failureReason))
			p.attemptReversal(logCtx, msg.ID, fmt.Sprintf("MNO submission failed: %v", failureReason))
			_ = p.markAsFailed(logCtx, msg.ID, failureReason, errorCode)
			// Generate failure DLR? Yes, as MNO rejected/failed submission
			go p.generateAndEnqueueFailureDLR(context.Background(), msg.ID, "failed", failureReason, &errorCode)
			continue // Next message
		}

		// If WasSubmitted is true and submitResult.Error is nil, it implies *at least* an attempt was made.
		// Success/failure per segment is recorded by the Sender via DB updates.
		// Update main message status to indicate send was attempted.
		updateErr := p.markSendAttempted(logCtx, msg.ID)
		if updateErr != nil {
			slog.ErrorContext(logCtx, "Failed to update message status to send_attempted", slog.Any("error", updateErr))
			// Don't trigger reversal here, as MNO submission might have happened. Needs manual check.
		} else {
			processedCount++
			slog.InfoContext(logCtx, "Message successfully processed for sending (segments submitted/attempted)")
		}
	}
	return processedCount, nil
}

// attemptReversal tries to reverse a previous debit, logging errors.
func (p *Processor) attemptReversal(ctx context.Context, msgID int64, reason string) {
	slog.WarnContext(ctx, "Attempting wallet debit reversal", slog.String("reason", reason))
	reverseErr := p.walletService.ReverseDebit(ctx, msgID, reason)
	if reverseErr != nil {
		slog.ErrorContext(ctx, "CRITICAL: Debit reversal FAILED",
			slog.Any("reversal_error", reverseErr),
			slog.String("reason", reason),
		)
		// ALERTING is essential here - money might be stuck!
	} else {
		slog.InfoContext(ctx, "Debit reversal successful")
	}
}

// markAsFailed updates the main message record to a failed state.
func (p *Processor) markAsFailed(ctx context.Context, msgID int64, failureReason error, errorCode string) error {
	errMsg := "processing failed"
	if failureReason != nil {
		errMsg = failureReason.Error()
	}
	dbErrMsg := errMsg
	dbErrCode := errorCode

	// Update main message status to failed
	// Use UpdateMessageFinalStatus directly? Or a specific failure update?
	// Let's assume UpdateMessageFinalStatus sets completed_at as well.
	err := p.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
		ID:               msgID,
		FinalStatus:      "failed", // Terminal failed state
		ErrorCode:        &dbErrCode,
		ErrorDescription: &dbErrMsg,
	})
	if err == nil {
		slog.WarnContext(ctx, "Message marked as failed", slog.String("reason", errMsg), slog.String("error_code", errorCode))
	} else {
		slog.ErrorContext(ctx, "Failed to update message status to failed", slog.Any("db_error", err))
	}
	return err
}

// markSendAttempted updates the main message status after segments are sent/attempted.
func (p *Processor) markSendAttempted(ctx context.Context, msgID int64) error {
	// This query might update processing_status to 'send_attempted'
	err := p.dbQueries.UpdateMessageSendAttempted(ctx, msgID) // Need this query
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update message status to send_attempted", slog.Any("db_error", err))
	}
	return err
}

// ========================================================================================
// DLR Processing and Forwarding
// ========================================================================================

// UpdateSegmentDLRStatus processes an incoming DLR for a specific segment.
// This is intended to be called by the MNO Connector's DLR handler.
func (p *Processor) UpdateSegmentDLRStatus(ctx context.Context, connectorID int32, dlrInfo mno.DLRInfo) error {
	logCtx := ctx                                                      // Use incoming context
	logCtx = logging.ContextWithMNOConnID(logCtx, connectorID)         // Add connector ID
	logCtx = logging.ContextWithMNOMsgID(logCtx, dlrInfo.MnoMessageID) // Add MNO Msg ID

	slog.InfoContext(logCtx, "Processing received DLR",
		slog.String("mno_msg_id", dlrInfo.MnoMessageID),
		slog.Int64("og_msg_id", dlrInfo.OriginalMsgID),
		slog.String("dlr_status", dlrInfo.Status),
		slog.Any("dlr_error_code", dlrInfo.ErrorCode),
		slog.Time("dlr_timestamp", dlrInfo.Timestamp),
	)

	// 1. Find the specific segment record using MNO Message ID
	segmentInfo, err := p.dbQueries.FindSegmentByMnoMessageID(logCtx, &dlrInfo.MnoMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "Received DLR for unknown/unmatched MNO message ID")
			return nil // Ignore DLR for unknown ID
		}
		slog.ErrorContext(logCtx, "DB error finding segment for DLR", slog.Any("error", err))
		return fmt.Errorf("db error finding segment for DLR (MNO ID %s): %w", dlrInfo.MnoMessageID, err)
	}

	// Enrich context further
	logCtx = logging.ContextWithMessageID(logCtx, segmentInfo.MessageID)
	logCtx = logging.ContextWithSegmentID(logCtx, segmentInfo.SegmentID) // Add segment ID

	// 2. Update the specific segment's DLR status in DB
	err = p.dbQueries.UpdateSegmentDLR(logCtx, database.UpdateSegmentDLRParams{
		ID:        segmentInfo.SegmentID,
		DlrStatus: &dlrInfo.Status,
		ErrorCode: &dlrInfo.ErrorCode, // Pass through MNO error code
	})
	if err != nil {
		slog.ErrorContext(logCtx, "CRITICAL: Failed to update segment DLR status in DB", slog.Any("error", err))
		// Don't return error here? Or should we? If DB fails, we might reprocess DLR later.
		return fmt.Errorf("failed to update segment DLR status: %w", err)
	}
	slog.InfoContext(logCtx, "Updated segment DLR status in DB",
		slog.Int("segment_seqn", int(segmentInfo.SegmentSeqn)),
		slog.String("dlr_status", dlrInfo.Status),
	)

	// 3. Aggregate Status for Parent Message (Run asynchronously to avoid blocking DLR processing)
	// Pass a clean background context enriched with MessageID for tracing aggregation task
	go p.aggregateMessageStatus(logCtx, segmentInfo.MessageID)

	// 4. Forward DLR back to the originating SP (Run asynchronously)
	// Pass a clean background context enriched with MessageID for tracing forwarding task
	go p.generateAndEnqueueSuccessDLR(logging.ContextWithMessageID(context.Background(), segmentInfo.MessageID), segmentInfo.MessageID, dlrInfo)

	return nil // DLR processing for this segment successful from MNO perspective
}

// aggregateMessageStatus checks all segments and updates the parent message status
func (p *Processor) aggregateMessageStatus(ctx context.Context, messageID int64) {
	// Context should already have messageID from caller
	slog.DebugContext(ctx, "Starting message status aggregation")
	time.Sleep(100 * time.Millisecond) // Small delay, consider better concurrency control

	segments, err := p.dbQueries.GetSegmentStatusesForMessage(ctx, messageID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Error fetching segments for status aggregation", slog.Any("error", err))
		return
	}

	totalSegmentsDB, err := p.dbQueries.GetMessageTotalSegments(ctx, messageID)
	if err != nil {
		slog.ErrorContext(ctx, "Error fetching total segment count for status aggregation", slog.Any("error", err))
		return
	}

	if len(segments) < int(totalSegmentsDB) {
		slog.DebugContext(ctx, "Skipping aggregation: Not all segment statuses received yet",
			slog.Int("found", len(segments)), slog.Int("expected", int(totalSegmentsDB)))
		return
	}
	// Optional: Handle len(segments) > int(totalSegmentsDB) warning

	var deliveredCount, failedCount, pendingCount int
	finalStatus := "unknown"

	for _, seg := range segments {
		status := ""
		if seg.DlrStatus != nil {
			status = strings.ToUpper(*seg.DlrStatus)
		}
		switch status {
		case "DELIVRD":
			deliveredCount++
		case "EXPIRED", "DELETED", "UNDELIV", "REJECTD":
			failedCount++
		case "":
			pendingCount++ // Treat NULL/empty as pending
		default:
			pendingCount++ // Treat other non-terminal states as pending
		}
	}

	// Determine final status
	if pendingCount > 0 {
		finalStatus = "processing"
	} else if failedCount > 0 {
		finalStatus = "failed"
	} else if deliveredCount > 0 && deliveredCount >= int(totalSegmentsDB) {
		finalStatus = "delivered"
	} // Add warning/handling for mismatch if desired

	// Update only if a terminal state is reached and status changed
	if finalStatus == "delivered" || finalStatus == "failed" {
		currentStatusRec, err := p.dbQueries.GetMessageFinalStatus(ctx, messageID)
		statusChanged := true // Assume change unless proven otherwise
		if err == nil && currentStatusRec != "" && currentStatusRec == finalStatus {
			statusChanged = false
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(ctx, "Could not get current final status before aggregation update", slog.Any("error", err))
		}

		if statusChanged {
			slog.InfoContext(ctx, "Attempting to update final message status", slog.String("new_status", finalStatus))
			// --- Aggregate Error Code/Description ---
			// Find first segment error or combine them? Needs logic.
			// For now, leave ErrorCode/Desc update out of aggregation, handle on initial failure.
			aggErrCode := ""
			aggErrMsg := ""
			// --- End Aggregate Error ---
			err = p.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
				ID:               messageID,
				FinalStatus:      finalStatus,
				ErrorCode:        &aggErrCode, // Use aggregated error info if implemented
				ErrorDescription: &aggErrMsg,
			})
			if err != nil {
				slog.ErrorContext(ctx, "Error updating final message status", slog.String("attempted_status", finalStatus), slog.Any("error", err))
			} else {
				slog.InfoContext(ctx, "Aggregated final message status updated successfully", slog.String("final_status", finalStatus))
			}
		} else {
			slog.DebugContext(ctx, "Final status already matches target, skipping update", slog.String("status", finalStatus))
		}
	} else {
		slog.DebugContext(ctx, "Aggregation determined non-terminal status, no update performed", slog.String("status", finalStatus))
	}
}

// generateAndEnqueueSuccessDLR: Enqueue to DB
func (p *Processor) generateAndEnqueueSuccessDLR(ctx context.Context, messageID int64, dlrInfo mno.DLRInfo) {
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	slog.DebugContext(logCtx, "Attempting to enqueue success/failure DLR", slog.String("dlr_status", dlrInfo.Status))

	spMsgInfo, err := p.dbQueries.GetSPMessageInfoForDLR(logCtx, messageID)
	if err != nil {
		slog.ErrorContext(logCtx, "Cannot forward DLR: Failed to get SP/Message info", slog.Any("error", err))
		return
	}

	// 2. Prepare SP Details
	spDetails := sp.SPDetails{
		ServiceProviderID: spMsgInfo.ServiceProviderID, // Need SP ID from query
		Protocol:          spMsgInfo.Protocol,
		SMPPSystemID:      *spMsgInfo.SystemID, // This is sql.NullString from DB
	}

	callbackUrl := sp.GetHttpCallBackURL(spMsgInfo.HttpConfig)
	authConfig := sp.GetHttpCallBackKey(spMsgInfo.HttpConfig)

	if callbackUrl != nil {
		spDetails.HTTPCallbackURL = *callbackUrl
	}

	if authConfig != nil {
		spDetails.HTTPAuthConfig = *authConfig
	}

	// 3. Construct Forwarded DLR Info
	forwardedDLR := sp.ForwardedDLRInfo{
		InternalMessageID: messageID,
		ClientMessageRef:  spMsgInfo.ClientRef, // Use client's reference
		ClientMessageID:   spMsgInfo.ClientMessageID,
		SourceAddr:        spMsgInfo.OriginalDestinationAddr, // DLR source is original MT destination
		DestAddr:          spMsgInfo.OriginalSourceAddr,      // DLR destination is original MT source (sender ID)
		SubmitDate:        spMsgInfo.SubmittedAt.Time,        // Use submitted_at
		DoneDate:          dlrInfo.Timestamp,                 // Use timestamp from MNO DLR
		Status:            dlrInfo.Status,                    // Pass through MNO status
		ErrorCode:         &dlrInfo.ErrorCode,                // Pass through MNO error code
		// NetworkCode:    dlrInfo.NetworkCode,          // Pass through if available
		TotalSegments: spMsgInfo.TotalSegments,
	}

	// Create combined payload for DB queue
	payloadData := map[string]interface{}{
		"sp_details": spDetails,
		"dlr_info":   forwardedDLR,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		slog.ErrorContext(logCtx, "CRITICAL: Failed to marshal payload for DLR forwarding queue", slog.Any("error", err))
		return // Cannot enqueue
	}

	// Enqueue to DB
	// Use default max attempts (e.g., 5) or make configurable
	err = p.dbQueries.EnqueueDLRForForwarding(logCtx, database.EnqueueDLRForForwardingParams{
		MessageID:   messageID,
		Payload:     payloadBytes,
		MaxAttempts: 5, // Example Max Attempts
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to enqueue DLR forwarding job to DB", slog.Any("error", err))
		// Handle DB enqueue failure (retry? alert?)
	} else {
		slog.InfoContext(logCtx, "Successfully enqueued DLR forwarding job to DB")
	}
}

// generateAndEnqueueFailureDLR: Enqueue to DB
func (p *Processor) generateAndEnqueueFailureDLR(ctx context.Context, messageID int64, failureStatus string, failureReason error, errorCode *string) {
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	slog.DebugContext(logCtx, "Attempting to enqueue internal failure DLR", slog.String("failure_status", failureStatus), slog.String("failure_reason", failureReason.Error()))

	spMsgInfo, err := p.dbQueries.GetSPMessageInfoForDLR(logCtx, messageID)
	if err != nil {
		slog.ErrorContext(logCtx, "Cannot forward DLR: Failed to get SP/Message info", slog.Any("error", err))
		return
	}
	if err != nil { /* ... handle error ... */
		return
	}

	// 2. Prepare SP Details
	spDetails := sp.SPDetails{
		ServiceProviderID: spMsgInfo.ServiceProviderID, // Need SP ID from query
		Protocol:          spMsgInfo.Protocol,
		SMPPSystemID:      *spMsgInfo.SystemID, // This is sql.NullString from DB
		HTTPCallbackURL:   *sp.GetHttpCallBackURL(spMsgInfo.HttpConfig),
		HTTPAuthConfig:    *sp.GetHttpCallBackKey(spMsgInfo.HttpConfig),
	}

	// Map internal failure status to DLR status code
	dlrStatusString := "UNDELIV" // Default for failures
	if failureStatus == "failed_validation" || failureStatus == "failed_routing" {
		dlrStatusString = "REJECTD"
	}

	// 3. Construct Forwarded DLR Info
	forwardedDLR := sp.ForwardedDLRInfo{
		InternalMessageID: messageID,
		ClientMessageRef:  spMsgInfo.ClientRef,
		ClientMessageID:   spMsgInfo.ClientMessageID,
		SourceAddr:        spMsgInfo.OriginalDestinationAddr,
		DestAddr:          spMsgInfo.OriginalSourceAddr,
		SubmitDate:        spMsgInfo.SubmittedAt.Time,
		DoneDate:          time.Now(), // Failure happened now
		Status:            dlrStatusString,
		ErrorCode:         errorCode, // Use the mapped internal error code
		TotalSegments:     spMsgInfo.TotalSegments,
		// Network code likely N/A for internal failures
	}

	// Create combined payload
	payloadData := map[string]interface{}{
		"sp_details": spDetails,
		"dlr_info":   forwardedDLR,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return
	}

	// Enqueue to DB
	err = p.dbQueries.EnqueueDLRForForwarding(logCtx, database.EnqueueDLRForForwardingParams{
		MessageID:   messageID,
		Payload:     payloadBytes,
		MaxAttempts: 3, // Maybe fewer retries for internal failures?
	})
	if err != nil {
	} else {
		slog.InfoContext(logCtx, "Successfully enqueued internal failure DLR job to DB")
	}
}
