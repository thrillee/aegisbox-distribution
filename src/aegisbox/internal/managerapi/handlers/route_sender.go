package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

var (
	routeSenderStatuses  = map[string]bool{"active": true, "inactive": true}
	senderPoolStrategies = map[string]bool{"round_robin": true, "random": true, "ordered": true}
)

// RouteSenderHandler manages route sender pools: the sender IDs each connection accepts, and
// which service providers send with them.
type RouteSenderHandler struct {
	dbQueries database.Querier
}

func NewRouteSenderHandler(q database.Querier) *RouteSenderHandler {
	return &RouteSenderHandler{dbQueries: q}
}

func toRouteSenderResponse(rs database.RouteSender) dto.RouteSenderResponse {
	resp := dto.RouteSenderResponse{
		ID:              rs.ID,
		MnoConnectionID: rs.MnoConnectionID,
		SenderID:        rs.SenderIDString,
		Status:          rs.Status,
		Position:        rs.Position,
		UsageCount:      rs.UsageCount,
		CreatedAt:       rs.CreatedAt.Time,
		UpdatedAt:       rs.UpdatedAt.Time,
	}
	if rs.LastUsedAt.Valid {
		t := rs.LastUsedAt.Time
		resp.LastUsedAt = &t
	}
	return resp
}

// GetRouteSenderPool handles GET /mno-connections/:id/route-senders
func (h *RouteSenderHandler) GetRouteSenderPool(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetRouteSenderPool")
	connID, err := parseIDParam(c)
	if err != nil {
		return
	}

	conn, err := h.dbQueries.GetMNOConnectionSenderPoolStrategy(logCtx, connID)
	if err != nil {
		handleGetError(c, logCtx, "MNO connection", err)
		return
	}
	senders, err := h.dbQueries.ListRouteSenders(logCtx, connID)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list route senders", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list route senders"})
		return
	}

	resp := dto.RouteSenderPoolResponse{
		MnoConnectionID: conn.ID,
		MnoID:           conn.MnoID,
		Strategy:        conn.SenderPoolStrategy,
		Senders:         make([]dto.RouteSenderResponse, 0, len(senders)),
	}
	for _, rs := range senders {
		resp.Senders = append(resp.Senders, toRouteSenderResponse(rs))
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateSenderPoolStrategy handles PUT /mno-connections/:id/route-senders/strategy
func (h *RouteSenderHandler) UpdateSenderPoolStrategy(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateSenderPoolStrategy")
	connID, err := parseIDParam(c)
	if err != nil {
		return
	}

	var req dto.UpdateSenderPoolStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if !senderPoolStrategies[req.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "strategy must be one of round_robin, random, ordered"})
		return
	}

	conn, err := h.dbQueries.UpdateMNOConnectionSenderPoolStrategy(logCtx, database.UpdateMNOConnectionSenderPoolStrategyParams{
		ID:                 connID,
		SenderPoolStrategy: req.Strategy,
	})
	if err != nil {
		handleGetError(c, logCtx, "MNO connection", err)
		return
	}
	slog.InfoContext(logCtx, "Updated route sender pool strategy",
		slog.Int("mno_connection_id", int(conn.ID)), slog.String("strategy", conn.SenderPoolStrategy))
	c.JSON(http.StatusOK, gin.H{"mno_connection_id": conn.ID, "mno_id": conn.MnoID, "strategy": conn.SenderPoolStrategy})
}

// CreateRouteSender handles POST /mno-connections/:id/route-senders
func (h *RouteSenderHandler) CreateRouteSender(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateRouteSender")
	connID, err := parseIDParam(c)
	if err != nil {
		return
	}

	var req dto.CreateRouteSenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	senderID := strings.TrimSpace(req.SenderID)
	if senderID == "" || len(senderID) > 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender_id must be 1 to 16 characters"})
		return
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	if !routeSenderStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be active or inactive"})
		return
	}

	if _, err := h.dbQueries.GetMNOConnectionSenderPoolStrategy(logCtx, connID); err != nil {
		handleGetError(c, logCtx, "MNO connection", err)
		return
	}

	rs, err := h.dbQueries.CreateRouteSender(logCtx, database.CreateRouteSenderParams{
		MnoConnectionID: connID,
		SenderIDString:  senderID,
		Status:          status,
		Position:        req.Position,
	})
	if err != nil {
		if isUniqueViolationError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "This sender ID is already in the connection's pool"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create route sender", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add route sender"})
		return
	}
	slog.InfoContext(logCtx, "Added route sender",
		slog.Int("mno_connection_id", int(connID)), slog.String("sender_id", rs.SenderIDString))
	c.JSON(http.StatusCreated, toRouteSenderResponse(rs))
}

// UpdateRouteSender handles PUT /route-senders/:id
func (h *RouteSenderHandler) UpdateRouteSender(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateRouteSender")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}

	var req dto.UpdateRouteSenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if !routeSenderStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be active or inactive"})
		return
	}

	rs, err := h.dbQueries.UpdateRouteSender(logCtx, database.UpdateRouteSenderParams{
		ID:       id,
		Status:   req.Status,
		Position: req.Position,
	})
	if err != nil {
		handleGetError(c, logCtx, "Route sender", err)
		return
	}
	c.JSON(http.StatusOK, toRouteSenderResponse(rs))
}

// DeleteRouteSender handles DELETE /route-senders/:id
func (h *RouteSenderHandler) DeleteRouteSender(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteRouteSender")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}

	rows, err := h.dbQueries.DeleteRouteSender(logCtx, id)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete route sender", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete route sender"})
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route sender not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetRouteSenderPoolSetting handles GET /service-providers/:id/route-sender-pool
func (h *RouteSenderHandler) GetRouteSenderPoolSetting(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetRouteSenderPoolSetting")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}

	enabled, err := h.dbQueries.UsesRouteSenderPool(logCtx, id)
	if err != nil {
		handleGetError(c, logCtx, "Service provider", err)
		return
	}
	c.JSON(http.StatusOK, dto.RouteSenderPoolSettingResponse{ServiceProviderID: id, Enabled: enabled})
}

// UpdateRouteSenderPoolSetting handles PUT /service-providers/:id/route-sender-pool
func (h *RouteSenderHandler) UpdateRouteSenderPoolSetting(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateRouteSenderPoolSetting")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, id)

	var req dto.UpdateRouteSenderPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	row, err := h.dbQueries.UpdateServiceProviderRouteSenderPool(logCtx, database.UpdateServiceProviderRouteSenderPoolParams{
		ID:                 id,
		UseRouteSenderPool: req.Enabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service provider not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to update route sender pool setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update route sender pool setting"})
		return
	}
	slog.InfoContext(logCtx, "Updated route sender pool setting", slog.Bool("enabled", row.UseRouteSenderPool))
	c.JSON(http.StatusOK, dto.RouteSenderPoolSettingResponse{ServiceProviderID: row.ID, Enabled: row.UseRouteSenderPool})
}
