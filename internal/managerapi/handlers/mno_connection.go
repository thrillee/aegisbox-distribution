package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

type MNOConnectionHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool // May not be needed here
}

func NewMNOConnectionHandler(q database.Querier, pool *pgxpool.Pool) *MNOConnectionHandler {
	return &MNOConnectionHandler{dbQueries: q, dbPool: pool}
}

// CreateMNOConnection handles POST /mnos/:mno_id/connections (and potentially direct POST /mno-connections)
func (h *MNOConnectionHandler) CreateMNOConnection(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateMNOConnection")
	var req dto.CreateMNOConnectionRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Get MNO ID from path OR body
	mnoIDStr := c.Param("mno_id") // For nested route /mnos/:mno_id/connections
	var mnoID int32
	var err error
	if mnoIDStr != "" {
		id64, pErr := strconv.ParseInt(mnoIDStr, 10, 32)
		if pErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mno_id in URL path"})
			return
		}
		mnoID = int32(id64)
		logCtx = logging.ContextWithMNOID(logCtx, mnoID)
		// If nested, ensure body doesn't contradict path
		if req.MnoID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mno_id in body contradicts URL path"})
			return
		}
	} else if req.MnoID > 0 {
		// For direct POST /mno-connections
		mnoID = req.MnoID
		logCtx = logging.ContextWithMNOID(logCtx, mnoID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required field: mno_id (in path or body)"})
		return
	}

	// --- Validate MNO Exists ---
	_, err = h.dbQueries.GetMNOByID(logCtx, mnoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: mno_id %d not found", mnoID)})
		} else {
			slog.ErrorContext(logCtx, "Failed to validate MNO existence", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate MNO"})
		}
		return
	}

	// Prepare Params
	params := database.CreateMNOConnectionParams{
		MnoID:    mnoID,
		Protocol: req.Protocol,
		Column3:  req.Status, // SQL defaults to 'active' if NULL
		// SMPP fields
		SystemID:                req.SystemID,
		Password:                req.Password, // Store plain text as per schema comment (verify security reqs)
		Host:                    req.Host,
		Port:                    req.Port,
		UseTls:                  req.UseTLS,
		BindType:                req.BindType,
		SystemType:              req.SystemType,
		EnquireLinkIntervalSecs: req.EnquireLinkIntervalSecs,
		RequestTimeoutSecs:      req.RequestTimeoutSecs,
		ConnectRetryDelaySecs:   req.ConnectRetryDelaySecs,
		MaxWindowSize:           req.MaxWindowSize,
		DefaultDataCoding:       req.DefaultDataCoding,
		SourceAddrTon:           req.SourceAddrTON,
		SourceAddrNpi:           req.SourceAddrNPI,
		DestAddrTon:             req.DestAddrTON,
		DestAddrNpi:             req.DestAddrNPI,
		// HTTP fields (flattened)
		ApiKey:          req.ApiKey,
		BaseUrl:         *req.BaseUrl,
		Username:        req.Username,
		HttpPassword:    req.HttpPassword,
		SecretKey:       req.SecretKey,
		TimeoutSecs:     req.TimeoutSecs,
		SupportsWebhook: req.SupportsWebhook,
		WebhookPath:     req.WebhookPath,
	}

	// Create Connection
	createdConn, err := h.dbQueries.CreateMNOConnection(logCtx, params)
	if err != nil {
		// Handle potential errors (e.g., FK violation if MNO deleted concurrently?)
		slog.ErrorContext(logCtx, "Failed to create MNO connection", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MNO connection"})
		return
	}

	resp := mapDBMNOConnToResponse(createdConn)
	slog.InfoContext(logging.ContextWithMNOConnID(logCtx, createdConn.ID), "MNO Connection created successfully")
	c.JSON(http.StatusCreated, resp)
}

// ListMNOConnections handles GET /mnos/:mno_id/connections
func (h *MNOConnectionHandler) ListMNOConnections(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListMNOConnections")
	mnoID, err := parseMnoIDParam(c) // Use specific helper for /:mno_id
	if err != nil {
		return
	}
	logCtx = logging.ContextWithMNOID(logCtx, mnoID)

	limit, offset := parsePagination(c)

	// Need sqlc CountMNOConnectionsByMNO(ctx, mnoID)
	total, err := h.dbQueries.CountMNOConnectionsByMNO(logCtx, mnoID)
	if err != nil { /* ... handle count error ... */
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK, dto.PaginatedListResponse{Data: []dto.MNOConnectionResponse{}, Pagination: dto.PaginationResponse{Total: 0, Limit: limit, Offset: offset}})
		return
	}

	// Need sqlc ListMNOConnectionsByMNO(ctx, {MnoID, Limit, Offset})
	conns, err := h.dbQueries.ListMNOConnectionsByMNO(logCtx, database.ListMNOConnectionsByMNOParams{
		MnoID:  mnoID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil { /* ... handle list error ... */
		return
	}

	respData := make([]dto.MNOConnectionResponse, len(conns))
	for i, conn := range conns {
		respData[i] = mapDBMNOConnToResponse(conn)
	}
	c.JSON(http.StatusOK, conns)
}

// GetMNOConnection handles GET /mno-connections/:conn_id
func (h *MNOConnectionHandler) GetMNOConnection(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetMNOConnection")
	connID, err := parseConnIDParam(c) // Use specific helper for /:conn_id
	if err != nil {
		return
	}
	logCtx = logging.ContextWithMNOConnID(logCtx, connID)

	conn, err := h.dbQueries.GetMNOConnectionByID(logCtx, connID)
	if err != nil {
		handleGetError(c, logCtx, "MNO Connection", err)
		return
	}

	resp := mapDBMNOConnToResponse(conn)
	c.JSON(http.StatusOK, resp)
}

// UpdateMNOConnection handles PUT /mno-connections/:conn_id
func (h *MNOConnectionHandler) UpdateMNOConnection(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateMNOConnection")
	connID, err := parseConnIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithMNOConnID(logCtx, connID)

	// Optional: Fetch existing to log MNO ID?
	// existing, err := h.dbQueries.GetMNOConnectionByID(logCtx, connID) ... handle ...
	// logCtx = logging.ContextWithMNOID(logCtx, existing.MnoID)

	var req dto.UpdateMNOConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Prepare update params
	params := database.UpdateMNOConnectionParams{
		ID:                      connID,
		Status:                  req.Status,
		SystemID:                req.SystemID,
		Password:                req.Password, // Update password if provided
		Host:                    req.Host,
		Port:                    req.Port,
		UseTls:                  req.UseTLS,
		BindType:                req.BindType,
		SystemType:              req.SystemType,
		EnquireLinkIntervalSecs: req.EnquireLinkIntervalSecs,
		RequestTimeoutSecs:      req.RequestTimeoutSecs,
		ConnectRetryDelaySecs:   req.ConnectRetryDelaySecs,
		MaxWindowSize:           req.MaxWindowSize,
		DefaultDataCoding:       req.DefaultDataCoding,
		SourceAddrTon:           req.SourceAddrTON,
		SourceAddrNpi:           req.SourceAddrNPI,
		DestAddrTon:             req.DestAddrTON,
		DestAddrNpi:             req.DestAddrNPI,
		// HTTP fields (flattened)
		ApiKey:          req.ApiKey,
		BaseUrl:         req.BaseUrl,
		Username:        req.Username,
		HttpPassword:    req.HttpPassword,
		SecretKey:       req.SecretKey,
		TimeoutSecs:     req.TimeoutSecs,
		SupportsWebhook: req.SupportsWebhook,
		WebhookPath:     req.WebhookPath,
	}

	updatedConn, err := h.dbQueries.UpdateMNOConnection(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "MNO Connection not found"})
		} else {
			slog.ErrorContext(logCtx, "Failed to update MNO connection", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update MNO connection"})
		}
		return
	}

	resp := mapDBMNOConnToResponse(updatedConn)
	slog.InfoContext(logCtx, "MNO Connection updated successfully")
	c.JSON(http.StatusOK, resp)
}

// DeleteMNOConnection handles DELETE /mno-connections/:conn_id
func (h *MNOConnectionHandler) DeleteMNOConnection(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteMNOConnection")
	connID, err := parseConnIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithMNOConnID(logCtx, connID)

	// Fetch existing first maybe to log MNO ID? Optional.

	// Delete the connection - NOTE: This might affect active sessions in MNOManager.
	// The manager should detect the deletion on its next sync cycle and shut down the connector.
	err = h.dbQueries.DeleteMNOConnection(logCtx, connID)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete MNO connection", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete MNO connection"})
		return
	}

	slog.InfoContext(logCtx, "MNO Connection deleted successfully")
	c.Status(http.StatusNoContent)
}

// --- Helper ---

// mapDBMNOConnToResponse converts DB model to API DTO.
func mapDBMNOConnToResponse(conn database.MnoConnection) dto.MNOConnectionResponse {
	return dto.MNOConnectionResponse{
		ID:                      conn.ID,
		MnoID:                   conn.MnoID,
		Protocol:                conn.Protocol,
		Status:                  conn.Status,
		SystemID:                conn.SystemID, // sql.NullString maps to pointer via omitempty? Check marshal. Explicit pointers might be better.
		Host:                    conn.Host,
		Port:                    conn.Port,
		UseTLS:                  conn.UseTls,
		BindType:                conn.BindType,
		SystemType:              conn.SystemType,
		EnquireLinkIntervalSecs: conn.EnquireLinkIntervalSecs,
		RequestTimeoutSecs:      conn.RequestTimeoutSecs,
		ConnectRetryDelaySecs:   conn.ConnectRetryDelaySecs,
		MaxWindowSize:           conn.MaxWindowSize,
		DefaultDataCoding:       conn.DefaultDataCoding,
		SourceAddrTON:           conn.SourceAddrTon,
		SourceAddrNPI:           conn.SourceAddrNpi,
		DestAddrTON:             conn.DestAddrTon,
		DestAddrNPI:             conn.DestAddrNpi,
		// HTTP fields (flattened)
		ApiKey:          conn.ApiKey,
		BaseUrl:         &conn.BaseUrl,
		Username:        conn.Username,
		HttpPassword:    conn.HttpPassword,
		SecretKey:       conn.SecretKey,
		TimeoutSecs:     conn.TimeoutSecs,
		SupportsWebhook: conn.SupportsWebhook,
		WebhookPath:     conn.WebhookPath,
		CreatedAt:       conn.CreatedAt.Time,
		UpdatedAt:       conn.UpdatedAt.Time,
	}
}
