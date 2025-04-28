package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
	// "regexp" // If needed for validation
)

// Compile-time check
var _ sp.IncomingMessageHandler = (*DefaultIncomingMessageHandler)(nil)

// DefaultIncomingMessageHandler provides the standard implementation for handling incoming messages.
type DefaultIncomingMessageHandler struct {
	dbQueries database.Querier
	// Add Pool if transactions become needed here later
	// dbPool *pgxpool.Pool
}

// NewDefaultIncomingMessageHandler creates a new handler instance.
func NewDefaultIncomingMessageHandler(q database.Querier) *DefaultIncomingMessageHandler {
	return &DefaultIncomingMessageHandler{dbQueries: q}
}

// HandleIncomingMessage implements the sp.IncomingMessageHandler interface.
// It validates basic inputs, determines currency, inserts the message into the DB,
// and returns an Acknowledgement.
func (h *DefaultIncomingMessageHandler) HandleIncomingMessage(ctx context.Context, msg sp.IncomingSPMessage) (sp.Acknowledgement, error) {
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
		slog.ErrorContext(logCtx, "Failed to fetch service provider details for currency lookup", slog.Any("error", err))
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  "Internal error: could not verify service provider",
		}, fmt.Errorf("failed to get SP details %d: %w", msg.ServiceProviderID, err) // Return internal error
	}
	currencyCode := spDetails.DefaultCurrencyCode
	if currencyCode == "" {
		slog.ErrorContext(logCtx, "Service provider is missing a default currency code", slog.Int("sp_id", int(msg.ServiceProviderID)))
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

	params := database.InsertMessageInParams{
		ServiceProviderID:       msg.ServiceProviderID,
		SpCredentialID:          msg.CredentialID,      // Use correct FK name from schema/sqlc
		ClientMessageID:         &msg.ClientMessageRef, // Map ClientRef to client_message_id
		ClientRef:               nil,                   // If UDH/concat ref is parsed, put it here
		OriginalSourceAddr:      msg.SenderID,
		OriginalDestinationAddr: msg.DestinationMSISDN,
		ShortMessage:            msg.MessageContent,
		TotalSegments:           msg.TotalSegments,
		CurrencyCode:            currencyCode,
		SubmittedAt:             pgtype.Timestamptz{Time: msg.ReceivedAt, Valid: true}, // Use SP submit time
	}
	if msg.ConcatRef > 0 { // Example if concat ref parsed from UDH
		concatRef := fmt.Sprintf("%d", msg.ConcatRef)
		params.ClientRef = &concatRef
	}

	// 4. Insert into Database
	insertedID, err := h.dbQueries.InsertMessageIn(logCtx, params)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to insert incoming message into database", slog.Any("error", err))
		// Determine if it's a retryable DB error or something else?
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  "Internal Server Error: Failed to store message",
		}, fmt.Errorf("db insert failed: %w", err) // Return internal error
	}

	// 5. Return Success Acknowledgement
	slog.InfoContext(logCtx, "Incoming message accepted and stored", slog.Int64("internal_msg_id", insertedID))
	return sp.Acknowledgement{
		InternalMessageID: insertedID,
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
