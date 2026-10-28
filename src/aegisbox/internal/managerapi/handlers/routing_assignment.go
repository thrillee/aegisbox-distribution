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

type RoutingAssignmentHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool
}

func NewRoutingAssignmentHandler(q database.Querier, pool *pgxpool.Pool) *RoutingAssignmentHandler {
	return &RoutingAssignmentHandler{dbQueries: q, dbPool: pool}
}

func (h *RoutingAssignmentHandler) Create(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateRoutingAssignment")
	var req dto.CreateRoutingAssignmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	_, err := h.dbQueries.GetRoutingGroupByID(logCtx, req.RoutingGroupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid routing_group_id: Routing Group %d not found", req.RoutingGroupID)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to validate routing group", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate routing group"})
		return
	}

	_, err = h.dbQueries.GetMsisdnPrefixGroupByID(logCtx, req.MsisdnPrefixGroupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid msisdn_prefix_group_id: Prefix Group %d not found", req.MsisdnPrefixGroupID)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to validate prefix group", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate prefix group"})
		return
	}

	_, err = h.dbQueries.GetMNOByID(logCtx, req.MnoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid mno_id: MNO %d not found", req.MnoID)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to validate MNO", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate MNO"})
		return
	}

	params := database.CreateRoutingAssignmentParams{
		RoutingGroupID:      req.RoutingGroupID,
		MsisdnPrefixGroupID: req.MsisdnPrefixGroupID,
		MnoID:               req.MnoID,
		Status:              req.Status,
		Priority:            req.Priority,
		Comment:             req.Comment,
	}

	created, err := h.dbQueries.CreateRoutingAssignment(logCtx, params)
	if err != nil {
		if isUniqueViolationError(err) {
			slog.WarnContext(logCtx, "Routing assignment creation failed: Duplicate combination",
				slog.Int("routing_group_id", int(req.RoutingGroupID)),
				slog.Int("msisdn_prefix_group_id", int(req.MsisdnPrefixGroupID)))
			c.JSON(http.StatusConflict, gin.H{"error": "Routing assignment for this routing group and prefix group already exists"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create routing assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create routing assignment"})
		return
	}

	detailed, err := h.dbQueries.GetRoutingAssignmentByID(logCtx, created.ID)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to fetch created assignment", slog.Any("error", err))
		empty := database.GetRoutingAssignmentByIDRow{}
		c.JSON(http.StatusCreated, mapRoutingAssignmentToDetailResponse(&created, &empty))
		return
	}

	c.JSON(http.StatusCreated, mapRoutingAssignmentToDetailResponse(nil, &detailed))
}

func (h *RoutingAssignmentHandler) List(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListRoutingAssignments")

	limit, offset := parsePagination(c)
	limitVal := limit
	offsetVal := offset

	routingGroupIDStr := c.Query("routing_group_id")
	if routingGroupIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "routing_group_id is required"})
		return
	}

	routingGroupID, err := strconv.ParseInt(routingGroupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid routing_group_id format"})
		return
	}

	total, err := h.dbQueries.CountRoutingAssignmentsByRoutingGroup(logCtx, int32(routingGroupID))
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count assignments", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count assignments"})
		return
	}

	assignments, err := h.dbQueries.ListRoutingAssignmentsByRoutingGroup(logCtx, database.ListRoutingAssignmentsByRoutingGroupParams{
		RoutingGroupID: int32(routingGroupID),
		Limit:          &limitVal,
		Offset:         &offsetVal,
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list assignments", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list assignments"})
		return
	}

	resp := make([]dto.RoutingAssignmentResponse, len(assignments))
	for i, a := range assignments {
		resp[i] = mapRoutingAssignmentToListResponse(&a)
	}

	c.JSON(http.StatusOK, dto.PaginatedListResponse{
		Data: resp,
		Pagination: dto.PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *RoutingAssignmentHandler) GetByID(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetRoutingAssignment")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}

	assignment, err := h.dbQueries.GetRoutingAssignmentByID(logCtx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routing assignment not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to get assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignment"})
		return
	}

	c.JSON(http.StatusOK, mapRoutingAssignmentToDetailResponse(nil, &assignment))
}

func (h *RoutingAssignmentHandler) Update(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateRoutingAssignment")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}

	var req dto.UpdateRoutingAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	params := database.UpdateRoutingAssignmentParams{
		ID:       id,
		MnoID:    req.MnoID,
		Status:   req.Status,
		Priority: req.Priority,
		Comment:  req.Comment,
	}

	_, err = h.dbQueries.UpdateRoutingAssignment(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routing assignment not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to update assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
		return
	}

	updated, err := h.dbQueries.GetRoutingAssignmentByID(logCtx, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Routing assignment updated successfully"})
		return
	}

	c.JSON(http.StatusOK, mapRoutingAssignmentToDetailResponse(nil, &updated))
}

func (h *RoutingAssignmentHandler) Delete(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteRoutingAssignment")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}

	err = h.dbQueries.DeleteRoutingAssignment(logCtx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routing assignment not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to delete assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete assignment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Routing assignment deleted successfully"})
}

func mapRoutingAssignmentToDetailResponse(a *database.RoutingAssignment, detail *database.GetRoutingAssignmentByIDRow) dto.RoutingAssignmentResponse {
	resp := dto.RoutingAssignmentResponse{
		Status:    "active",
		Priority:  0,
		CreatedAt: detail.CreatedAt.Time,
		UpdatedAt: detail.UpdatedAt.Time,
	}

	if detail != nil {
		resp.ID = detail.ID
		resp.RoutingGroupID = detail.RoutingGroupID
		resp.MsisdnPrefixGroupID = detail.MsisdnPrefixGroupID
		resp.MnoID = detail.MnoID
		resp.Status = detail.Status
		resp.Priority = detail.Priority
		resp.Comment = detail.Comment
		resp.RoutingGroupName = stringPtr(detail.RoutingGroupName)
		resp.RoutingGroupReference = detail.RoutingGroupReference
		resp.PrefixGroupName = stringPtr(detail.PrefixGroupName)
		resp.PrefixGroupReference = detail.PrefixGroupReference
		resp.MnoName = stringPtr(detail.MnoName)
	} else if a != nil {
		resp.ID = a.ID
		resp.RoutingGroupID = a.RoutingGroupID
		resp.MsisdnPrefixGroupID = a.MsisdnPrefixGroupID
		resp.MnoID = a.MnoID
		resp.Status = a.Status
		resp.Priority = a.Priority
		resp.Comment = a.Comment
	}

	return resp
}

func mapRoutingAssignmentToListResponse(a *database.ListRoutingAssignmentsByRoutingGroupRow) dto.RoutingAssignmentResponse {
	resp := dto.RoutingAssignmentResponse{
		ID:                   a.ID,
		RoutingGroupID:       a.RoutingGroupID,
		MsisdnPrefixGroupID:  a.MsisdnPrefixGroupID,
		MnoID:                a.MnoID,
		Status:               a.Status,
		Priority:             a.Priority,
		Comment:              a.Comment,
		CreatedAt:            a.CreatedAt.Time,
		UpdatedAt:            a.UpdatedAt.Time,
		PrefixGroupName:      stringPtr(a.PrefixGroupName),
		PrefixGroupReference: a.PrefixGroupReference,
		MnoName:              stringPtr(a.MnoName),
	}
	return resp
}

func stringPtr(s string) *string {
	return &s
}
