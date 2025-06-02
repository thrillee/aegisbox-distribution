package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
)

// Compile-time check
var _ sp.IncomingMessageHandler = (*DefaultIncomingMessageHandler)(nil)

// DefaultIncomingMessageHandler provides the standard implementation for handling incoming messages.
type DefaultIncomingMessageHandler struct {
	dbQueries     database.Querier
	preprocessors []sp.MessagePreprocessor
}

// NewDefaultIncomingMessageHandler creates a new handler instance.
func NewDefaultIncomingMessageHandler(
	q database.Querier,
	preprocessors []sp.MessagePreprocessor,
) *DefaultIncomingMessageHandler {
	return &DefaultIncomingMessageHandler{dbQueries: q, preprocessors: preprocessors}
}

// HandleIncomingMessage implements the sp.IncomingMessageHandler interface.
// It validates basic inputs, determines currency, inserts the message into the DB,
// and returns an Acknowledgement.
func (h *DefaultIncomingMessageHandler) HandleIncomingMessage(
	ctx context.Context,
	msg sp.IncomingSPMessage,
) (sp.Acknowledgement, error) {
	// Enrich context - SPID and CredentialID should already be in ctx from auth middleware/handler
	logCtx := logging.ContextWithSenderID(ctx, msg.SenderID) // Add logging helpers as needed
	logCtx = logging.ContextWithMSISDN(logCtx, msg.DestinationMSISDN)
	slog.InfoContext(logCtx, "Handling incoming message", slog.String("protocol", msg.Protocol))
	fmt.Println("SenderID: ", msg.SenderID)
	fmt.Println("DestinationMSISDN: ", msg.DestinationMSISDN)

	// 1. Basic Input Validation (Add more specific rules as needed)
	if err := validateIncomingMessage(msg); err != nil {
		slog.WarnContext(logCtx, "Incoming message validation failed", slog.Any("error", err))
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  err.Error(),
		}, nil // Return nil error, as rejection is handled logically
	}

	// 2. Determine Currency (Fetch SP's default currency)
	// This assumes SPID is reliably passed in msg or context
	// Consider adding default_currency_code to the spAuthInfo context value during auth
	spDetails, err := h.dbQueries.GetServiceProviderByID(logCtx, msg.ServiceProviderID)
	if err != nil {
		slog.ErrorContext(
			logCtx,
			"Failed to fetch service provider details for currency lookup",
			slog.Any("error", err),
		)
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  "Internal error: could not verify service provider",
		}, fmt.Errorf("failed to get SP details %d: %w", msg.ServiceProviderID, err) // Return internal error
	}

	credDetails, err := h.dbQueries.GetSPCredentialByID(
		ctx,
		msg.CredentialID,
	)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Failed to fetch SP credential details for preprocessing",
			slog.Any("error", err),
		)
		return sp.Acknowledgement{
				Status: "rejected",
				Error:  "Internal error: Credential lookup failed",
			}, fmt.Errorf(
				"failed to get SP Credential details %d: %w",
				msg.CredentialID,
				err,
			)
	}

	// --- Preprocessing Pipeline ---
	// Create a mutable copy for preprocessors if they need to modify it.
	processedMsg := msg
	var messageModifiedByPipeline bool

	for _, p := range h.preprocessors {
		logCtxPreproc := logging.ContextWithService(
			ctx,
			p.Name(),
		) // Add preprocessor name to context
		slog.DebugContext(logCtxPreproc, "Applying preprocessor")
		modified, err := p.Process(
			logCtxPreproc,
			&processedMsg,
			spDetails,
			credDetails,
		) // Pass pointer to the copy
		if err != nil {
			slog.ErrorContext(logCtxPreproc, "Preprocessor failed", slog.Any("error", err))
			return sp.Acknowledgement{
				Status: "rejected",
				Error:  fmt.Sprintf("Preprocessing error (%s): %s", p.Name(), err.Error()),
			}, nil // Logical rejection due to preprocessing error
		}
		if modified {
			messageModifiedByPipeline = true
			slog.InfoContext(logCtxPreproc, "Message modified by preprocessor")
		}
	}
	// --- End Preprocessing Pipeline ---

	// Use 'processedMsg' from here on.
	// Enrich context based on potentially modified message
	if messageModifiedByPipeline {
		slog.InfoContext(logCtx, "Message content after pipeline",
			slog.String("final_sender", processedMsg.SenderID),
			slog.String("final_content_preview", firstNChars(processedMsg.MessageContent, 30)))
	}

	currencyCode := spDetails.DefaultCurrencyCode
	if currencyCode == "" {
		slog.ErrorContext(
			logCtx,
			"Service provider is missing a default currency code",
			slog.Int("sp_id", int(msg.ServiceProviderID)),
		)
		// Fallback or reject? Reject for now.
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  "Internal configuration error: Billing currency not set",
		}, errors.New("missing default currency for SP")
	}
	logCtx = logging.ContextWithCurrency(logCtx, currencyCode) // Add helper

	// 3. Prepare DB Insert Parameters
	// TODO: Handle incoming multipart detection (UDH) here if needed.
	// If multipart is detected, buffer segments until complete OR insert segment details?
	// For now, assumes single segment or already reassembled message content.
	if msg.TotalSegments == 0 { // Ensure total_segments is at least 1
		msg.TotalSegments = 1
	}

	clientMessageId := uuid.New().String()
	params := database.InsertMessageInParams{
		ServiceProviderID:       msg.ServiceProviderID,
		SpCredentialID:          msg.CredentialID, // Use correct FK name from schema/sqlc
		ClientMessageID:         &clientMessageId, // Map ClientRef to client_message_id
		ClientRef:               nil,              // If UDH/concat ref is parsed, put it here
		OriginalSourceAddr:      msg.SenderID,
		OriginalDestinationAddr: msg.DestinationMSISDN,
		ShortMessage:            msg.MessageContent,
		TotalSegments:           msg.TotalSegments,
		CurrencyCode:            currencyCode,
		SubmittedAt: pgtype.Timestamptz{
			Time:  msg.ReceivedAt,
			Valid: true,
		}, // Use SP submit time
	}
	if msg.ConcatRef > 0 { // Example if concat ref parsed from UDH
		concatRef := fmt.Sprintf("%d", msg.ConcatRef)
		params.ClientRef = &concatRef
	}

	// 4. Insert into Database
	insertedID, err := h.dbQueries.InsertMessageIn(logCtx, params)
	if err != nil {
		slog.ErrorContext(
			logCtx,
			"Failed to insert incoming message into database",
			slog.Any("error", err),
		)
		// Determine if it's a retryable DB error or something else?
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  "Internal Server Error: Failed to store message",
		}, fmt.Errorf("db insert failed: %w", err) // Return internal error
	}

	// 5. Return Success Acknowledgement
	slog.InfoContext(
		logCtx,
		"Incoming message accepted and stored",
		slog.Int64("internal_msg_id", insertedID),
	)
	return sp.Acknowledgement{
		InternalMessageID: clientMessageId,
		Status:            "accepted",
	}, nil
}

// validateIncomingMessage performs basic checks on submitted message data.
func validateIncomingMessage(msg sp.IncomingSPMessage) error {
	// Example Validations (Customize these based on requirements)
	if len(msg.SenderID) == 0 || len(msg.SenderID) > 16 {
		return fmt.Errorf("invalid 'from' (sender ID) length: %d", len(msg.SenderID))
	}
	// Could add regex check for alphanumeric/numeric based on length

	if len(msg.DestinationMSISDN) == 0 || len(msg.DestinationMSISDN) > 20 {
		return fmt.Errorf("invalid 'to' (destination) length: %d", len(msg.DestinationMSISDN))
	}
	// Could add regex check for digits, optional '+'

	if msg.MessageContent == "" && msg.TotalSegments <= 1 {
		// Allow empty content only if potentially part of multipart sequence? Risky.
		// Generally reject empty messages.
		return errors.New("invalid 'text': message content cannot be empty")
	}
	// Could check message length against maximum allowed?

	return nil
}
