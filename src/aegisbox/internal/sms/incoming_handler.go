package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/pkg/codes"
	"github.com/thrillee/aegisbox/pkg/segmenter"
)

// Compile-time check
var _ sp.IncomingMessageHandler = (*DefaultIncomingMessageHandler)(nil)

// DefaultIncomingMessageHandler provides the standard implementation for handling incoming messages.
type DefaultIncomingMessageHandler struct {
	dbQueries     database.Querier
	preprocessors []sp.MessagePreprocessor
	segmenter     segmenter.Segmenter
}

// NewDefaultIncomingMessageHandler creates a new handler instance.
func NewDefaultIncomingMessageHandler(
	q database.Querier,
	preprocessors []sp.MessagePreprocessor,
) *DefaultIncomingMessageHandler {
	return &DefaultIncomingMessageHandler{dbQueries: q, preprocessors: preprocessors, segmenter: segmenter.NewDefaultSegmenter()}
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
	// Currency is needed to store the message; check it before accepting anything.
	if spDetails.DefaultCurrencyCode == "" {
		slog.ErrorContext(logCtx, "Service provider is missing a default currency code")
		return sp.Acknowledgement{
			Status: "rejected",
			Error:  "Internal configuration error: Billing currency not set",
		}, errors.New("missing default currency for SP")
	}

	// A part of a long SMS sent as separate PDUs: store it and join the parts into one message.
	if msg.TotalSegments > 1 && msg.SegmentSeqn >= 1 && msg.SegmentSeqn <= msg.TotalSegments {
		return h.handleConcatPart(logCtx, msg, spDetails, credDetails)
	}

	ack, _, err := h.storeMessage(logCtx, msg, spDetails, credDetails, "", nil)
	return ack, err
}

// storeMessage runs the preprocessors on a complete message and stores it. clientMessageID is the
// ID returned to the client ("" for a new one); extraClientMessageIDs are the IDs of the other
// parts of a joined message, which receive its DLR too. It returns the stored message's ID.
func (h *DefaultIncomingMessageHandler) storeMessage(
	ctx context.Context,
	msg sp.IncomingSPMessage,
	spDetails database.ServiceProvider,
	credDetails database.SpCredential,
	clientMessageID string,
	extraClientMessageIDs []string,
) (sp.Acknowledgement, int64, error) {
	logCtx := ctx
	processedMsg := msg
	var messageModifiedByPipeline bool

	for _, p := range h.preprocessors {
		logCtxPreproc := logging.ContextWithService(
			ctx,
			p.Name(),
		) // Add preprocessor name to context
		slog.InfoContext(logCtxPreproc, "Applying preprocessor", slog.String("Preprocessor name", p.Name()), slog.Any("Credentials", credDetails))
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
			}, 0, nil // Logical rejection due to preprocessing error
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
	logCtx = logging.ContextWithCurrency(logCtx, currencyCode)

	// 3. Prepare DB Insert Parameters
	// total_segments is the number of SMS segments the stored text needs: what is billed and
	// reported. It is counted from the text itself, never taken from the client's UDH part count.
	totalSegments, segErr := h.countSegments(processedMsg.MessageContent)
	if segErr != nil {
		slog.WarnContext(logCtx, "Could not segment message", slog.Any("error", segErr))
		return sp.Acknowledgement{Status: "rejected", Error: "Invalid message content: " + segErr.Error()}, 0, nil
	}
	msg.TotalSegments = totalSegments

	processingStatus := processedMsg.ProcessingStatus
	if processingStatus == "" {
		processingStatus = codes.MsgStatusReceived
	}

	clientMessageId := clientMessageID
	if clientMessageId == "" {
		clientMessageId = uuid.New().String()
	}
	params := database.InsertMessageInParams{
		ServiceProviderID:       msg.ServiceProviderID,
		SpCredentialID:          msg.CredentialID, // Use correct FK name from schema/sqlc
		ClientMessageID:         &clientMessageId, // Map ClientRef to client_message_id
		ClientRef:               nil,              // If UDH/concat ref is parsed, put it here
		OriginalSourceAddr:      processedMsg.SenderID,
		OriginalDestinationAddr: msg.DestinationMSISDN,
		ShortMessage:            processedMsg.MessageContent,
		TotalSegments:           msg.TotalSegments,
		RoutedMnoID:             processedMsg.MnoID,
		CurrencyCode:            currencyCode,
		ProcessingStatus:        processingStatus,
		ExtraClientMessageIds:   extraClientMessageIDs,
		SubmittedAt: pgtype.Timestamptz{
			Time:  msg.ReceivedAt,
			Valid: true,
		}, // Use SP submit time
	}
	if msg.ConcatRef > 0 { // Example if concat ref parsed from UDH
		concatRef := fmt.Sprintf("%d", msg.ConcatRef)
		params.ClientRef = &concatRef
	}

	slog.DebugContext(logCtx, "New SMS", slog.Any("Params", params))

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
		}, 0, fmt.Errorf("db insert failed: %w", err) // Return internal error
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
	}, insertedID, nil
}

// Long SMS sent over SMPP as separate PDUs with a concatenation UDH.
const (
	concatJoinTimeout = 2 * time.Minute    // parts still missing after this are sent without them
	concatRetention   = 7 * 24 * time.Hour // joined parts are kept this long
)

// handleConcatPart stores one part, answering it with its own message ID, and joins the group into
// one message as soon as every part is in.
func (h *DefaultIncomingMessageHandler) handleConcatPart(
	ctx context.Context,
	msg sp.IncomingSPMessage,
	spDetails database.ServiceProvider,
	credDetails database.SpCredential,
) (sp.Acknowledgement, error) {
	partID, err := h.dbQueries.UpsertConcatPart(ctx, database.UpsertConcatPartParams{
		ServiceProviderID: msg.ServiceProviderID,
		SpCredentialID:    msg.CredentialID,
		SourceAddr:        msg.SenderID,
		DestinationAddr:   msg.DestinationMSISDN,
		ConcatRef:         msg.ConcatRef,
		TotalParts:        msg.TotalSegments,
		PartNum:           msg.SegmentSeqn,
		Content:           msg.MessageContent,
		ClientMessageID:   uuid.New().String(),
		RequestDlr:        msg.RequestDLR,
		IsFlash:           msg.IsFlash,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store message part", slog.Any("error", err))
		return sp.Acknowledgement{Status: "rejected", Error: "Internal Server Error: Failed to store message"}, err
	}
	slog.InfoContext(ctx, "Stored message part",
		slog.Int("part", int(msg.SegmentSeqn)), slog.Int("of", int(msg.TotalSegments)),
		slog.Int("ref", int(msg.ConcatRef)), slog.String("client_message_id", partID))

	parts, err := h.dbQueries.ClaimCompleteConcatGroup(ctx, database.ClaimCompleteConcatGroupParams{
		SpCredentialID:  msg.CredentialID,
		SourceAddr:      msg.SenderID,
		DestinationAddr: msg.DestinationMSISDN,
		ConcatRef:       msg.ConcatRef,
		TotalParts:      msg.TotalSegments,
	})
	if err != nil {
		// The part is stored; the stale part worker joins the group later.
		slog.ErrorContext(ctx, "Failed to check whether all message parts are in", slog.Any("error", err))
	} else if len(parts) > 0 {
		joined := joinedPartsFromComplete(parts)
		h.storeJoined(ctx, msg, spDetails, credDetails, joined)
	}
	return sp.Acknowledgement{InternalMessageID: partID, Status: "accepted"}, nil
}

type joinedParts struct {
	ids        []int64
	text       string
	clientIDs  []string // in part order; the first is the message's own client ID
	partsCount int
}

func joinedPartsFromComplete(rows []database.ClaimCompleteConcatGroupRow) joinedParts {
	sort.Slice(rows, func(i, j int) bool { return rows[i].PartNum < rows[j].PartNum })
	var j joinedParts
	var b strings.Builder
	for _, r := range rows {
		j.ids = append(j.ids, r.ID)
		b.WriteString(r.Content)
		j.clientIDs = append(j.clientIDs, r.ClientMessageID)
	}
	j.text, j.partsCount = b.String(), len(rows)
	return j
}

// storeJoined stores the joined text as one message and links the parts to it. If storing fails
// the parts are released, so the stale part worker tries again.
func (h *DefaultIncomingMessageHandler) storeJoined(
	ctx context.Context,
	msg sp.IncomingSPMessage,
	spDetails database.ServiceProvider,
	credDetails database.SpCredential,
	j joinedParts,
) {
	msg.MessageContent = j.text
	msg.TotalSegments, msg.SegmentSeqn, msg.ConcatRef = 1, 1, 0
	ack, messageID, err := h.storeMessage(ctx, msg, spDetails, credDetails, j.clientIDs[0], j.clientIDs[1:])
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store joined message; parts released for retry", slog.Any("error", err))
		if relErr := h.dbQueries.ReleaseConcatParts(ctx, j.ids); relErr != nil {
			slog.ErrorContext(ctx, "Failed to release message parts", slog.Any("error", relErr))
		}
		return
	}
	if ack.Status != "accepted" {
		// A permanent rejection (e.g. a preprocessor refused the text): the parts stay claimed and are
		// not retried. The client already has IDs for them and will not get a DLR.
		slog.ErrorContext(ctx, "Joined message rejected after its parts were accepted",
			slog.String("reason", ack.Error), slog.Any("client_message_ids", j.clientIDs))
		return
	}
	if err := h.dbQueries.SetConcatPartsMessage(ctx, database.SetConcatPartsMessageParams{MessageID: &messageID, Ids: j.ids}); err != nil {
		slog.WarnContext(ctx, "Failed to link message parts to the joined message", slog.Any("error", err))
	}
	slog.InfoContext(ctx, "Joined message parts into one message",
		slog.Int64("message_id", messageID), slog.Int("parts", j.partsCount), slog.Any("client_message_ids", j.clientIDs))
}

// FlushStaleConcatParts joins groups whose missing parts did not arrive within concatJoinTimeout,
// sending what arrived, and deletes old joined parts. It returns the number of groups flushed.
func (h *DefaultIncomingMessageHandler) FlushStaleConcatParts(ctx context.Context) int {
	flushed := 0
	for range 100 { // bounded per run
		rows, err := h.dbQueries.ClaimStaleConcatGroup(ctx, pgtype.Timestamptz{Time: time.Now().Add(-concatJoinTimeout), Valid: true})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to claim stale message parts", slog.Any("error", err))
			break
		}
		if len(rows) == 0 {
			break
		}
		sort.Slice(rows, func(i, k int) bool { return rows[i].PartNum < rows[k].PartNum })
		first := rows[0]
		var j joinedParts
		var b strings.Builder
		for _, r := range rows {
			j.ids = append(j.ids, r.ID)
			b.WriteString(r.Content)
			j.clientIDs = append(j.clientIDs, r.ClientMessageID)
		}
		j.text, j.partsCount = b.String(), len(rows)

		logCtx := logging.ContextWithSPID(ctx, first.ServiceProviderID)
		slog.WarnContext(logCtx, "Sending a long message with missing parts",
			slog.Int("parts_received", len(rows)), slog.Int("parts_expected", int(first.TotalParts)))

		spDetails, err := h.dbQueries.GetServiceProviderByID(logCtx, first.ServiceProviderID)
		if err != nil {
			slog.ErrorContext(logCtx, "Failed to load service provider for stale parts", slog.Any("error", err))
			_ = h.dbQueries.ReleaseConcatParts(ctx, j.ids)
			break
		}
		credDetails, err := h.dbQueries.GetSPCredentialByID(logCtx, first.SpCredentialID)
		if err != nil {
			slog.ErrorContext(logCtx, "Failed to load credential for stale parts", slog.Any("error", err))
			_ = h.dbQueries.ReleaseConcatParts(ctx, j.ids)
			break
		}
		h.storeJoined(logCtx, sp.IncomingSPMessage{
			ServiceProviderID: first.ServiceProviderID,
			CredentialID:      first.SpCredentialID,
			Protocol:          "smpp",
			SenderID:          first.SourceAddr,
			DestinationMSISDN: first.DestinationAddr,
			IsFlash:           first.IsFlash,
			RequestDLR:        first.RequestDlr,
			ReceivedAt:        first.ReceivedAt.Time,
		}, spDetails, credDetails, j)
		flushed++
	}
	if n, err := h.dbQueries.DeleteOldConcatParts(ctx, pgtype.Timestamptz{Time: time.Now().Add(-concatRetention), Valid: true}); err != nil {
		slog.WarnContext(ctx, "Failed to delete old message parts", slog.Any("error", err))
	} else if n > 0 {
		slog.InfoContext(ctx, "Deleted old message parts", slog.Int64("rows", n))
	}
	return flushed
}

// RunConcatFlusher calls FlushStaleConcatParts every interval until ctx is cancelled.
func (h *DefaultIncomingMessageHandler) RunConcatFlusher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.FlushStaleConcatParts(ctx)
		}
	}
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

// countSegments is the number of SMS segments the text needs (at least 1).
func (h *DefaultIncomingMessageHandler) countSegments(text string) (int32, error) {
	segs, _, err := h.segmenter.GetSegments(text)
	if err != nil {
		return 0, err
	}
	if len(segs) == 0 {
		return 1, nil
	}
	return int32(len(segs)), nil
}
