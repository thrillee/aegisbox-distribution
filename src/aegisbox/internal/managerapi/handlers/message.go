package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

// MessageHandler struct definition (as before)
type MessageHandler struct {
	dbQueries database.Querier
}

// NewMessageHandler func definition (as before)
func NewMessageHandler(q database.Querier) *MessageHandler {
	return &MessageHandler{dbQueries: q}
}

// ListMessageDetails handles GET /messages/details with filters and pagination
func (h *MessageHandler) ListMessageDetails(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListMessageDetails")

	limit, offset := parsePagination(c)

	// --- Parse Optional Filters ---
	// Using helper functions defined in common.go (or define here)
	spIDFilter := parseNullInt32Filter(c, "sp_id")
	spCredIDFilter := parseNullInt32Filter(c, "sp_credential_id")
	mnoIDFilter := parseNullInt32Filter(c, "mno_id")
	senderIDFilter := parseNullStringFilter(c, "sender_id")
	destinationFilter := parseNullStringFilter(c, "destination")
	statusFilter := parseNullStringFilter(c, "status")
	clientMsgIDFilter := parseNullStringFilter(c, "client_message_id")
	clientRefFilter := parseNullStringFilter(c, "client_ref")
	mnoMsgIDFilter := parseNullStringFilter(c, "mno_message_id") // Filter param
	startDateFilter := parseNullDateFilter(c, "start_date")
	endDateFilter := parseNullDateFilter(c, "end_date")

	// Log filters
	slog.DebugContext(logCtx, "ListMessageDetails filters parsed",
		slog.Any("sp_id", spIDFilter), slog.Any("sp_credential_id", spCredIDFilter), slog.Any("mno_id", mnoIDFilter),
		slog.Any("sender_id", senderIDFilter), slog.Any("destination", destinationFilter),
		slog.Any("status", statusFilter), slog.Any("client_msg_id", clientMsgIDFilter),
		slog.Any("client_ref", clientRefFilter), slog.Any("mno_msg_id", mnoMsgIDFilter),
		slog.Any("start_date", startDateFilter), slog.Any("end_date", endDateFilter),
		slog.Int("limit", int(limit)), slog.Int("offset", int(offset)),
	)

	// --- Prepare Params for SQLC ---
	// Ensure order matches the query placeholders ($1 to $13)
	countParams := database.CountMessageDetailsParams{
		SpID:            spIDFilter,                                              // $1
		SpCredentialID:  spCredIDFilter,                                          // $2 - NEW filter based on thought process
		MnoID:           mnoIDFilter,                                             // $3
		SenderID:        senderIDFilter,                                          // $4
		Destination:     destinationFilter,                                       // $5
		FinalStatus:     statusFilter,                                            // $6
		ClientMessageID: clientMsgIDFilter,                                       // $7
		ClientRef:       clientRefFilter,                                         // $8
		MnoMessageID:    mnoMsgIDFilter,                                          // $9 - Filter value used twice in query (main + EXISTS)
		StartDate:       pgtype.Timestamptz{Time: *startDateFilter, Valid: true}, // $10
		EndDate:         pgtype.Timestamptz{Time: *endDateFilter, Valid: true},   // $11
	}
	listParams := database.ListMessageDetailsParams{
		SpID:            spIDFilter,            // $1
		SpCredentialID:  spCredIDFilter,        // $2
		MnoID:           mnoIDFilter,           // $3
		SenderID:        senderIDFilter,        // $4
		Destination:     destinationFilter,     // $5
		FinalStatus:     statusFilter,          // $6
		ClientMessageID: clientMsgIDFilter,     // $7
		ClientRef:       clientRefFilter,       // $8
		MnoMessageID:    mnoMsgIDFilter,        // $9
		StartDate:       countParams.StartDate, // $10
		EndDate:         countParams.EndDate,   // $11
		QueryLimit:      limit,                 // $12
		QueryOffset:     offset,                // $13
	}

	// --- Fetch Count ---
	total, err := h.dbQueries.CountMessageDetails(logCtx, countParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count message details", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count messages"})
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK, dto.PaginatedListResponse{Data: []dto.MessageDetailResponse{}, Pagination: dto.PaginationResponse{Total: 0, Limit: limit, Offset: offset}})
		return
	}

	// --- Fetch Data ---
	messages, err := h.dbQueries.ListMessageDetails(logCtx, listParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list message details", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	// --- Map & Respond ---
	respData := make([]dto.MessageDetailResponse, len(messages))
	for i, msg := range messages {
		respData[i] = mapDBMessageDetailToResponse(msg) // Use a mapping helper
	}
	paginatedResp := dto.PaginatedListResponse{
		Data:       respData,
		Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
	}
	c.JSON(http.StatusOK, paginatedResp)
}

// --- Helpers for Message Handler ---

// mapDBMessageDetailToResponse converts the detailed DB row to the API DTO.
func mapDBMessageDetailToResponse(msg database.ListMessageDetailsRow) dto.MessageDetailResponse {
	return dto.MessageDetailResponse{
		ID:                      msg.ID,
		ServiceProviderID:       msg.ServiceProviderID,
		ServiceProviderName:     msg.ServiceProviderName,
		SpCredentialID:          msg.SpCredentialID,
		CredentialIdentifier:    msg.CredentialIdentifier, // sql.NullString from COALESCE
		Protocol:                msg.Protocol,
		ClientMessageID:         msg.ClientMessageID,
		ClientRef:               msg.ClientRef,
		OriginalSourceAddr:      msg.OriginalSourceAddr,
		OriginalDestinationAddr: msg.OriginalDestinationAddr,
		TotalSegments:           msg.TotalSegments,
		SubmittedAt:             msg.SubmittedAt.Time,
		ReceivedAt:              msg.ReceivedAt.Time, // Adjust if needed
		RoutedMnoID:             msg.RoutedMnoID,
		MnoName:                 msg.MnoName, // sql.NullString
		CurrencyCode:            msg.CurrencyCode,
		Cost:                    msg.Cost, // decimal.Decimal
		ProcessingStatus:        msg.ProcessingStatus,
		FinalStatus:             msg.FinalStatus,
		ErrorCode:               msg.ErrorCode,
		ErrorDescription:        msg.ErrorDescription,
		CompletedAt:             msg.CompletedAt.Time,
	}
}

// parseNullInt32Filter parses an optional integer filter from query params.
func parseNullInt32Filter(c *gin.Context, key string) *int32 {
	valStr := c.Query(key)
	if valStr == "" {
		return nil
	}
	id, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return nil
	}
	result := int32(id)
	return &result
}

// parseNullStringFilter parses an optional string filter from query params.
func parseNullStringFilter(c *gin.Context, key string) *string {
	valStr := c.Query(key)
	if valStr == "" {
		return nil
	}
	return &valStr
}

// parseNullDateFilter parses an optional date filter (YYYY-MM-DD) from query params.
func parseNullDateFilter(c *gin.Context, key string) *time.Time {
	valStr := c.Query(key)
	if valStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", valStr)
	if err != nil {
		return nil
	}
	// Return time representing start of day UTC? Query uses date comparison.
	// Passing the parsed time directly should work with TIMESTAMPTZ comparison in query.
	return &t
}
