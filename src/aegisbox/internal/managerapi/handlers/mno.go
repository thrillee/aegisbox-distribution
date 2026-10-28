package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool" // Only needed if using Transactions here
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

// MNOHandler holds dependencies for MNO handlers.
type MNOHandler struct {
	dbQueries database.Querier
	// dbPool    *pgxpool.Pool // Not needed for simple MNO CRUD
}

// NewMNOHandler creates a new handler instance.
func NewMNOHandler(q database.Querier, pool *pgxpool.Pool) *MNOHandler {
	// Pool might not be strictly needed here unless complex operations arise
	return &MNOHandler{dbQueries: q}
}

// CreateMNO handles POST /mnos
func (h *MNOHandler) CreateMNO(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateMNO")
	var req dto.CreateMNORequest

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(logCtx, "Failed to bind request JSON", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Prepare params
	params := database.CreateMNOParams{
		Name:        req.Name,
		CountryCode: req.CountryCode,
		NetworkCode: req.NetworkCode,
		Column4:     req.Status,
	}

	createdMNO, err := h.dbQueries.CreateMNO(logCtx, params)
	if err != nil {
		// Handle potential unique constraint violation on name
		fmt.Println("Err: ", err)
		if isUniqueViolationError(err) { // Use helper to check DB error
			slog.WarnContext(logCtx, "MNO creation failed: Duplicate name", slog.String("name", req.Name))
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("MNO with name '%s' already exists", req.Name)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create MNO in DB", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MNO"})
		return
	}

	resp := mapDBMNOToResponse(createdMNO)
	slog.InfoContext(logging.ContextWithMNOID(logCtx, createdMNO.ID), "MNO created successfully")
	c.JSON(http.StatusCreated, resp)
}

// GetMNO handles GET /mnos/:id
func (h *MNOHandler) GetMNO(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetMNO")
	id, err := parseIDParam(c)
	if err != nil {
		return
	} // Error response handled in parseIDParam
	logCtx = logging.ContextWithMNOID(logCtx, id)

	mno, err := h.dbQueries.GetMNOByID(logCtx, id)
	if err != nil {
		handleGetError(c, logCtx, "MNO", err) // Use helper for Not Found / Internal Server Error
		return
	}

	resp := mapDBMNOToResponse(mno)
	c.JSON(http.StatusOK, resp)
}

// ListMNOs handles GET /mnos with pagination
func (h *MNOHandler) ListMNOs(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListMNOs")
	limit, offset := parsePagination(c)

	total, err := h.dbQueries.CountMNOs(logCtx)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count MNOs", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve MNO count"})
		return
	}

	if total == 0 || offset >= int32(total) {
		c.JSON(http.StatusOK, dto.PaginatedListResponse{Data: []dto.MNOResponse{}, Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset}})
		return
	}

	dbMNOs, err := h.dbQueries.ListMNOs(logCtx, database.ListMNOsParams{Limit: limit, Offset: offset})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list MNOs from DB", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve MNOs"})
		return
	}

	respData := make([]dto.MNOResponse, len(dbMNOs))
	for i, dbMNO := range dbMNOs {
		respData[i] = mapDBMNOToResponse(dbMNO)
	}

	paginatedResp := dto.PaginatedListResponse{
		Data:       respData,
		Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
	}
	c.JSON(http.StatusOK, paginatedResp)
}

// UpdateMNO handles PUT /mnos/:id
func (h *MNOHandler) UpdateMNO(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateMNO")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithMNOID(logCtx, id)

	var req dto.UpdateMNORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(logCtx, "Failed to bind request JSON for MNO update", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Prepare params using sqlc.narg helper for optional fields
	params := database.UpdateMNOParams{
		ID:          id,
		Name:        req.Name,
		CountryCode: req.CountryCode,
		NetworkCode: req.NetworkCode,
		Status:      req.Status,
	}

	updatedMNO, err := h.dbQueries.UpdateMNO(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "MNO not found for update")
			c.JSON(http.StatusNotFound, gin.H{"error": "MNO not found"})
		} else if isUniqueViolationError(err) { // Check for unique name violation
			slog.WarnContext(logCtx, "MNO update failed: Duplicate name", slog.String("name", *req.Name))
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("MNO with name '%s' already exists", *req.Name)})
		} else {
			slog.ErrorContext(logCtx, "Failed to update MNO", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update MNO"})
		}
		return
	}

	resp := mapDBMNOToResponse(updatedMNO)
	slog.InfoContext(logCtx, "MNO updated successfully")
	c.JSON(http.StatusOK, resp)
}

// DeleteMNO handles DELETE /mnos/:id
func (h *MNOHandler) DeleteMNO(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteMNO")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithMNOID(logCtx, id)

	// WARNING: Deleting an MNO might fail due to Foreign Key constraints
	// (ON DELETE RESTRICT) from mno_connections or routing_rules.
	// Consider using a soft delete (setting status to 'inactive') instead.
	// We will proceed with hard delete for this example.

	err = h.dbQueries.DeleteMNO(logCtx, id)
	if err != nil {
		// Check for foreign key violation error (e.g., pgconn.PgError code 23503)
		if isForeignKeyViolationError(err) {
			slog.WarnContext(logCtx, "Cannot delete MNO: It is referenced by other resources (connections, routes)", slog.Any("error", err))
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete MNO: It is currently in use by connections or routing rules. Try deactivating it instead."})
			return
		}
		slog.ErrorContext(logCtx, "Failed to delete MNO", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete MNO"})
		return
	}

	slog.InfoContext(logCtx, "MNO deleted successfully")
	c.Status(http.StatusNoContent) // 204 No Content on successful DELETE
}

// --- Helpers ---

// mapDBMNOToResponse converts the database model to the API response DTO.
func mapDBMNOToResponse(mno database.Mno) dto.MNOResponse {
	resp := dto.MNOResponse{
		ID:          mno.ID,
		Name:        mno.Name,
		CountryCode: mno.CountryCode,
		Status:      mno.Status,
		CreatedAt:   mno.CreatedAt.Time, // Adjust if using pgtype
		UpdatedAt:   mno.UpdatedAt.Time, // Adjust if using pgtype
	}
	// Handle nullable NetworkCode
	if mno.NetworkCode != nil {
		resp.NetworkCode = mno.NetworkCode
	}
	return resp
}
