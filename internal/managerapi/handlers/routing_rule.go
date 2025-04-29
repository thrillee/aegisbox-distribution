package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

type RoutingRuleHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool // Needed for checking MNO existence maybe?
}

func NewRoutingRuleHandler(q database.Querier, pool *pgxpool.Pool) *RoutingRuleHandler {
	return &RoutingRuleHandler{dbQueries: q, dbPool: pool}
}

// CreateRule handles POST /routing-rules
func (h *RoutingRuleHandler) CreateRule(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateRoutingRule")
	var req dto.CreateRoutingRuleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// --- Optional: Validate MNO exists ---
	_, err := h.dbQueries.GetMNOByID(logCtx, req.MnoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid mno_id: MNO %d not found", req.MnoID)})
		} else {
			slog.ErrorContext(logCtx, "Failed to validate MNO existence", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate MNO"})
		}
		return
	}
	// --- End Validation ---

	params := database.CreateRoutingRuleParams{
		Prefix:  req.Prefix,
		MnoID:   req.MnoID,
		Column3: req.Priority, // Use Column3 for nullable priority $3
	}

	createdRule, err := h.dbQueries.CreateRoutingRule(logCtx, params)
	if err != nil {
		if isUniqueViolationError(err) { // Check for duplicate prefix
			slog.WarnContext(logCtx, "Routing rule creation failed: Duplicate prefix", slog.String("prefix", req.Prefix))
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Routing rule with prefix '%s' already exists", req.Prefix)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create routing rule", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create routing rule"})
		return
	}

	// Refetch with JOIN to get MNO name for response consistency
	detailedRule, err := h.dbQueries.GetRoutingRuleByID(logCtx, createdRule.ID)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to fetch created routing rule details", slog.Any("error", err))
		// Fallback: return basic info from 'createdRule'
		resp := mapDBRoutingRuleToResponse(createdRule) // Use basic mapping
		slog.InfoContext(logging.ContextWithRuleID(logCtx, createdRule.ID), "Routing rule created successfully (details fetch failed)")
		c.JSON(http.StatusCreated, resp)
		return
	}

	resp := mapDBRoutingRuleDetailToResponse(detailedRule) // Use detailed mapping
	slog.InfoContext(logging.ContextWithRuleID(logCtx, createdRule.ID), "Routing rule created successfully")
	c.JSON(http.StatusCreated, resp)
}

// ListRules handles GET /routing-rules
func (h *RoutingRuleHandler) ListRules(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListRoutingRules")
	limit, offset := parsePagination(c)
	mnoIDFilter := parseNullInt32Filter(c, "mno_id") // Optional filter by MNO

	listParams := database.ListRoutingRulesParams{
		Column1: *mnoIDFilter, // Use $1 for mno_id filter
		Limit:   limit,
		Offset:  offset,
	}

	total, err := h.dbQueries.CountRoutingRules(logCtx, *mnoIDFilter)
	if err != nil {
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK, dto.PaginatedListResponse{Data: []dto.RoutingRuleResponse{}, Pagination: dto.PaginationResponse{Total: 0, Limit: limit, Offset: offset}})
		return
	}

	rules, err := h.dbQueries.ListRoutingRules(logCtx, listParams)
	if err != nil { /* ... handle list error ... */
		return
	}

	respData := make([]dto.RoutingRuleResponse, len(rules))
	for i, rule := range rules {
		respData[i] = mapDBRoutingRuleListToResponse(rule) // Use mapping helper
	}
	paginatedResp := dto.PaginatedListResponse{Data: respData, Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset}}
	c.JSON(http.StatusOK, paginatedResp)
}

// GetRule handles GET /routing-rules/:id
func (h *RoutingRuleHandler) GetRule(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetRoutingRule")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithRuleID(logCtx, id) // Add helper

	rule, err := h.dbQueries.GetRoutingRuleByID(logCtx, id)
	if err != nil {
		handleGetError(c, logCtx, "Routing Rule", err)
		return
	}

	resp := mapDBRoutingRuleDetailToResponse(rule)
	c.JSON(http.StatusOK, resp)
}

// UpdateRule handles PUT /routing-rules/:id
func (h *RoutingRuleHandler) UpdateRule(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateRoutingRule")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithRuleID(logCtx, id)

	var req dto.UpdateRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil { /* ... handle bind error ... */
		return
	}

	// Optional: Validate MNO ID if provided
	if req.MnoID != nil {
		_, err := h.dbQueries.GetMNOByID(logCtx, *req.MnoID)
		if err != nil { /* ... handle MNO not found error ... */
			return
		}
	}

	params := database.UpdateRoutingRuleParams{
		ID:       id,
		Prefix:   req.Prefix,
		MnoID:    req.MnoID,
		Priority: req.Priority,
	}

	updatedRuleDB, err := h.dbQueries.UpdateRoutingRule(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routing Rule not found"})
		} else if isUniqueViolationError(err) {
			slog.WarnContext(logCtx, "Routing rule update failed: Duplicate prefix", slog.String("prefix", *req.Prefix))
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Routing rule with prefix '%s' already exists", *req.Prefix)})
		} else if isForeignKeyViolationError(err) {
			slog.WarnContext(logCtx, "Routing rule update failed: Invalid MNO ID", slog.Int("mno_id", int(*req.MnoID)))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid mno_id: MNO %d not found", *req.MnoID)})
		} else {
			slog.ErrorContext(logCtx, "Failed to update routing rule", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update routing rule"})
		}
		return
	}

	// Refetch with JOIN to get MNO name for response consistency
	detailedRule, err := h.dbQueries.GetRoutingRuleByID(logCtx, updatedRuleDB.ID)
	if err != nil {
	}

	resp := mapDBRoutingRuleDetailToResponse(detailedRule)
	slog.InfoContext(logCtx, "Routing rule updated successfully")
	c.JSON(http.StatusOK, resp)
}

// DeleteRule handles DELETE /routing-rules/:id
func (h *RoutingRuleHandler) DeleteRule(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteRoutingRule")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithRuleID(logCtx, id)

	err = h.dbQueries.DeleteRoutingRule(logCtx, id)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete routing rule", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete routing rule"})
		return
	}

	slog.InfoContext(logCtx, "Routing rule deleted successfully")
	c.Status(http.StatusNoContent)
}

// --- Helpers ---

// mapDBRoutingRuleDetailToResponse converts the detailed DB row (with MNO name) to DTO.
func mapDBRoutingRuleDetailToResponse(rule database.GetRoutingRuleByIDRow) dto.RoutingRuleResponse {
	resp := dto.RoutingRuleResponse{
		ID:        rule.ID,
		Prefix:    rule.Prefix,
		MnoID:     rule.MnoID,
		Priority:  rule.Priority,
		CreatedAt: rule.CreatedAt.Time,
		UpdatedAt: rule.UpdatedAt.Time,
	}
	if rule.MnoName != nil {
		resp.MnoName = rule.MnoName
	}
	return resp
}

// mapDBRoutingRuleListToResponse converts the list DB row (with MNO name) to DTO.
func mapDBRoutingRuleListToResponse(rule database.ListRoutingRulesRow) dto.RoutingRuleResponse {
	resp := dto.RoutingRuleResponse{
		ID:        rule.ID,
		Prefix:    rule.Prefix,
		MnoID:     rule.MnoID,
		Priority:  rule.Priority,
		CreatedAt: rule.CreatedAt.Time,
		UpdatedAt: rule.UpdatedAt.Time,
	}
	if rule.MnoName != nil {
		resp.MnoName = rule.MnoName
	}
	return resp
}

// mapDBRoutingRuleToResponse converts the basic DB row (no MNO name) to DTO.
func mapDBRoutingRuleToResponse(rule database.RoutingRule) dto.RoutingRuleResponse {
	return dto.RoutingRuleResponse{
		ID:        rule.ID,
		Prefix:    rule.Prefix,
		MnoID:     rule.MnoID,
		Priority:  rule.Priority,
		CreatedAt: rule.CreatedAt.Time,
		UpdatedAt: rule.UpdatedAt.Time,
		MnoName:   nil, // Name not available in basic struct
	}
}
