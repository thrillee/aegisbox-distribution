package sms

import (
	"context"
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
	"github.com/thrillee/aegisbox/pkg/codes"
)

// ========================================================================================
// Interfaces this Processor relies on (Implementations should be injected)
// ========================================================================================

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
	Send(
		ctx context.Context,
		msgDetails database.GetMessageDetailsForSendingRow,
	) (*mno.SubmitResult, error)
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

// Processor orchestrates the SMS processing lifecycle steps using specialized services.
type Processor struct {
	dbQueries           database.Querier
	dbPool              *pgxpool.Pool
	router              Router
	validator           Validator
	pricer              Pricer
	sender              Sender
	walletService       WalletService
	dlrForwarder        DLRForwarder
	senderHealthTracker *SenderHealthTracker
	routeHealthTracker  *RouteHealthTracker

	// Extracted Services
	routingService *RoutingService
	sendingService *SendingService
	dlrService     *DLRService
}

// ProcessorDependencies holds all dependencies needed to create a Processor.
type ProcessorDependencies struct {
	DBPool              *pgxpool.Pool
	DBQueries           database.Querier
	Router              Router
	Validator           Validator
	Pricer              Pricer
	Sender              Sender
	WalletService       WalletService
	DLRForwarder        DLRForwarder
	SenderHealthTracker *SenderHealthTracker
	RouteHealthTracker  *RouteHealthTracker

	// Batch workers for performance optimization
	BatchReversalWorker *BatchReversalWorker
	BatchHealthWorker   *BatchHealthWorker
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
	if deps.WalletService == nil {
		missing = append(missing, "WalletService")
	}

	if len(missing) > 0 {
		panic(fmt.Sprintf(
			"Missing required dependencies for SMS Processor: %s",
			strings.Join(missing, ", "),
		))
	}

	routingService := NewRoutingService(RoutingServiceDependencies{
		DBQueries: deps.DBQueries,
		Router:    deps.Router,
		Validator: deps.Validator,
	})

	sendingService := NewSendingService(SendingServiceDependencies{
		DBQueries:           deps.DBQueries,
		Sender:              deps.Sender,
		WalletService:       deps.WalletService,
		RouteHealthTracker:  deps.RouteHealthTracker,
		BatchReversalWorker: deps.BatchReversalWorker,
	})

	dlrService := NewDLRService(DLRServiceDependencies{
		DBQueries:           deps.DBQueries,
		WalletService:       deps.WalletService,
		SenderHealthTracker: deps.SenderHealthTracker,
		RouteHealthTracker:  deps.RouteHealthTracker,
		BatchHealthWorker:   deps.BatchHealthWorker,
	})

	return &Processor{
		dbPool:              deps.DBPool,
		dbQueries:           deps.DBQueries,
		router:              deps.Router,
		validator:           deps.Validator,
		pricer:              deps.Pricer,
		sender:              deps.Sender,
		walletService:       deps.WalletService,
		dlrForwarder:        deps.DLRForwarder,
		senderHealthTracker: deps.SenderHealthTracker,
		routeHealthTracker:  deps.RouteHealthTracker,
		routingService:      routingService,
		sendingService:      sendingService,
		dlrService:          dlrService,
	}
}

func (p *Processor) SetSender(sender Sender) {
	p.sender = sender
	if p.sendingService != nil {
		p.sendingService.SetSender(sender)
	}
}

// ========================================================================================
// Processing Steps (Called by Worker Manager)
// ========================================================================================

// ProcessRoutingStep fetches messages needing routing/validation and processes them.
func (p *Processor) ProcessRoutingStep(
	ctx context.Context,
	batchSize int,
) (processedCount int, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(
				ctx,
				"PANIC recovered in ProcessRoutingStep loop",
				slog.Any("panic_info", r),
			)
			err = fmt.Errorf(
				"panic occurred during routing step: %v",
				r,
			)
		}
	}()

	messages, fetchErr := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: codes.MsgStatusReceived,
		Limit:            int32(batchSize),
	})
	if fetchErr != nil && !errors.Is(fetchErr, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "Failed fetching messages for routing", slog.Any("error", fetchErr))
		return 0, fmt.Errorf("fetching messages for routing: %w", fetchErr)
	}
	if errors.Is(fetchErr, pgx.ErrNoRows) || len(messages) == 0 {
		slog.DebugContext(ctx, "No messages found in 'received' state for routing.")
		return 0, nil
	}

	for _, msg := range messages {
		logCtx := logging.ContextWithMessageID(ctx, msg.ID)
		logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)
		if msg.SpName != nil {
			logCtx = logging.ContextWithSPName(logCtx, *msg.SpName)
		}
		if msg.SpCredentialSystemID != nil {
			logCtx = logging.ContextWithSystemID(logCtx, *msg.SpCredentialSystemID)
		}
		if msg.MnoName != nil {
			logCtx = logging.ContextWithMOName(logCtx, *msg.MnoName)
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(logCtx, "PANIC in RouteMessage, marking message as failed",
						slog.Any("panic_info", r),
					)
					errDesc := fmt.Sprintf("panic during routing: %v", r)
					errCode := "PANIC_ERR"
					p.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
						ID:               msg.ID,
						FinalStatus:      codes.MsgFinalStatusFailed,
						ErrorCode:        &errCode,
						ErrorDescription: &errDesc,
					})
				}
			}()

			outcome := p.routingService.RouteMessage(logCtx, msg)
			if outcome.IsValid {
				processedCount++
			}
		}()
	}
	return processedCount, nil
}

// ProcessPricingStep fetches messages needing pricing and processes them.
func (p *Processor) ProcessPricingStep(
	ctx context.Context,
	batchSize int,
) (processedCount int, err error) {
	defer func() { // Add recovery
		if r := recover(); r != nil {
			slog.ErrorContext(
				ctx,
				"PANIC recovered in ProcessPricingStep loop",
				slog.Any("panic_info", r),
			)
			err = fmt.Errorf("panic occurred during pricing step: %v", r)
		}
	}()

	messages, fetchErr := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: codes.MsgStatusRouted,
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
		if msg.SpName != nil {
			logCtx = logging.ContextWithSPName(logCtx, *msg.SpName)
		}
		if msg.SpCredentialSystemID != nil {
			logCtx = logging.ContextWithSystemID(logCtx, *msg.SpCredentialSystemID)
		}
		if msg.MnoName != nil {
			logCtx = logging.ContextWithMOName(logCtx, *msg.MnoName)
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(logCtx, "PANIC in PriceAndDebit, marking message as failed",
						slog.Any("panic_info", r),
					)
					errDesc := fmt.Sprintf("panic during pricing: %v", r)
					errCode := "PANIC_ERR"
					p.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
						ID:               msg.ID,
						FinalStatus:      codes.MsgFinalStatusFailed,
						ErrorCode:        &errCode,
						ErrorDescription: &errDesc,
					})
				}
			}()

			pricingRes, pricingErr := p.pricer.PriceAndDebit(logCtx, msg.ID)
			if pricingErr != nil {
				slog.ErrorContext(
					logCtx,
					"Error processing pricing/debiting step",
					slog.Any("error", pricingErr),
				)
				return
			}
			if pricingRes != nil && pricingRes.Debited {
				processedCount++
			} else if pricingRes != nil {
				slog.WarnContext(logCtx, "Pricing/debiting failed logically", slog.Any("reason", pricingRes.Error))
			}
		}()
	}
	return processedCount, nil
}

// ProcessSendingStep fetches messages ready to be sent and processes them.
func (p *Processor) ProcessSendingStep(
	ctx context.Context,
	batchSize int,
) (processedCount int, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(
				ctx,
				"PANIC recovered in ProcessSendingStep loop",
				slog.Any("panic_info", r),
			)
			err = fmt.Errorf("panic occurred during sending step: %v", r)
		}
	}()

	messages, fetchErr := p.dbQueries.GetMessagesByStatus(ctx, database.GetMessagesByStatusParams{
		ProcessingStatus: codes.MsgStatusQueuedForSend,
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
		if msg.SpName != nil {
			logCtx = logging.ContextWithSPName(logCtx, *msg.SpName)
		}
		if msg.SpCredentialSystemID != nil {
			logCtx = logging.ContextWithSystemID(logCtx, *msg.SpCredentialSystemID)
		}
		if msg.MnoName != nil {
			logCtx = logging.ContextWithMOName(logCtx, *msg.MnoName)
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(logCtx, "PANIC in SendMessage, marking message as failed",
						slog.Any("panic_info", r),
					)
					errDesc := fmt.Sprintf("panic during send: %v", r)
					errCode := "PANIC_ERR"
					p.dbQueries.UpdateMessageSendAttempted(ctx, database.UpdateMessageSendAttemptedParams{
						ID:          msg.ID,
						RoutedMnoID: msg.RoutedMnoID,
					})
					p.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
						ID:               msg.ID,
						FinalStatus:      codes.MsgFinalStatusFailed,
						ErrorCode:        &errCode,
						ErrorDescription: &errDesc,
					})
				}
			}()

			success, sendErr := p.sendingService.SendMessage(logCtx, msg, p.router)
			if sendErr != nil {
				slog.ErrorContext(logCtx, "Error in sending service", slog.Any("error", sendErr))
				return
			}
			if success {
				processedCount++
			}
		}()
	}
	return processedCount, nil
}

// ========================================================================================
// DLR Processing and Forwarding
// ========================================================================================
// DLR Processing and Forwarding
// ========================================================================================

// UpdateSegmentDLRStatus processes an incoming DLR for a specific segment.
// This is intended to be called by the MNO Connector's DLR handler.
func (p *Processor) UpdateSegmentDLRStatus(
	ctx context.Context,
	connectorID int32,
	dlrInfo mno.DLRInfo,
) error {
	return p.dlrService.UpdateSegmentDLRStatus(ctx, connectorID, dlrInfo)
}

// aggregateMessageStatus checks all segments and updates the parent message status
func (p *Processor) aggregateMessageStatus(ctx context.Context, messageID int64) {
	// Context should already have messageID from caller
	slog.DebugContext(ctx, "Starting message status aggregation")
	time.Sleep(100 * time.Millisecond) // Small delay, consider better concurrency control

	segments, err := p.dbQueries.GetSegmentStatusesForMessage(ctx, messageID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(
			ctx,
			"Error fetching segments for status aggregation",
			slog.Any("error", err),
		)
		return
	}

	totalSegmentsDB, err := p.dbQueries.GetMessageTotalSegments(ctx, messageID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error fetching total segment count for status aggregation",
			slog.Any("error", err),
		)
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
	}

	// Trigger wallet reversal if message failed
	if finalStatus == "failed" {
		reverseErr := p.walletService.ReverseDebit(ctx, messageID, "Message delivery failed - DLR received with failure status")
		if reverseErr != nil {
			slog.ErrorContext(ctx, "CRITICAL: Failed to reverse wallet debit on DLR failure",
				slog.Any("error", reverseErr),
				slog.Int64("message_id", messageID),
			)
		} else {
			slog.InfoContext(ctx, "Wallet debit reversed due to DLR failure",
				slog.Int64("message_id", messageID),
			)
		}
	}

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
			slog.InfoContext(
				ctx,
				"Attempting to update final message status",
				slog.String("new_status", finalStatus),
			)
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
				slog.ErrorContext(
					ctx,
					"Error updating final message status",
					slog.String("attempted_status", finalStatus),
					slog.Any("error", err),
				)
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
