package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal" // Import decimal
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

type PricingRuleHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool
}

func NewPricingRuleHandler(q database.Querier, pool *pgxpool.Pool) *PricingRuleHandler {
	return &PricingRuleHandler{dbQueries: q, dbPool: pool}
}

// CreatePricingRule handles POST /pricing-rules
func (h *PricingRuleHandler) CreatePricingRule(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreatePricingRule")
	var req dto.CreatePricingRuleRequest

	// Need custom binding/validation for decimal or validate manually
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	// Manual validation for decimal
	if req.PricePerSms.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price_per_sms: must be positive"})
		return
	}

	logCtx = logging.ContextWithSPID(logCtx, req.ServiceProviderID)

	// --- Validate SP and MNO (if MnoID provided) ---
	_, err := h.dbQueries.GetServiceProviderByID(logCtx, req.ServiceProviderID)
	if err != nil { /* ... handle SP not found ... */
		return
	}
	if req.MnoID != nil {
		logCtx = logging.ContextWithMNOID(logCtx, *req.MnoID)
		_, err := h.dbQueries.GetMNOByID(logCtx, *req.MnoID)
		if err != nil { /* ... handle MNO not found ... */
			return
		}
	}
	// --- End Validation ---

	params := database.CreatePricingRuleParams{
		ServiceProviderID: req.ServiceProviderID,
		MnoID:             req.MnoID,
		PricePerSms:       req.PricePerSms,
		CurrencyCode:      req.CurrencyCode,
	}

	createdRule, err := h.dbQueries.CreatePricingRule(logCtx, params)
	if err != nil {
		if isUniqueViolationError(err) { // Check for duplicate rule (SP, Currency, MNO)
			errMsg := "Pricing rule conflict: A rule already exists for this provider/currency"
			if req.MnoID != nil {
				errMsg += fmt.Sprintf(" and MNO %d", *req.MnoID)
			} else {
				errMsg += " (default)"
			}
			slog.WarnContext(logCtx, errMsg)
			c.JSON(http.StatusConflict, gin.H{"error": errMsg})
			return
		}
		if isForeignKeyViolationError(err) {
			slog.WarnContext(logCtx, "Pricing rule creation failed: Invalid SP ID or MNO ID FK", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service_provider_id or mno_id"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create pricing rule", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pricing rule"})
		return
	}

	// Refetch with JOIN to get MNO name for response consistency
	detailedRule, err := h.dbQueries.GetPricingRuleByID(logCtx, createdRule.ID)
	if err != nil { /* ... handle fetch error ... */
	}

	resp := mapDBPricingRuleToResponse(detailedRule) // Use mapping helper
	slog.InfoContext(logging.ContextWithPricingRuleID(logCtx, createdRule.ID), "Pricing rule created successfully")
	c.JSON(http.StatusCreated, resp)
}

// ListPricingRules handles GET /pricing-rules (filtered by sp_id)
func (h *PricingRuleHandler) ListPricingRules(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListPricingRules")

	// --- Filter by Service Provider ID (REQUIRED) ---
	spIDStr := c.Query("sp_id")
	if spIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameter: sp_id"})
		return
	}
	spID, err := strconv.ParseInt(spIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sp_id format"})
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, int32(spID))

	limit, offset := parsePagination(c)

	total, err := h.dbQueries.CountPricingRulesBySP(logCtx, int32(spID))
	if err != nil {
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK,
			dto.PaginatedListResponse{
				Data:       []dto.PricingRuleResponse{},
				Pagination: dto.PaginationResponse{Total: 0, Limit: limit, Offset: offset},
			})
		return
	}

	rules, err := h.dbQueries.ListPricingRulesBySP(logCtx, database.ListPricingRulesBySPParams{
		ServiceProviderID: int32(spID),
		Limit:             limit,
		Offset:            offset,
	})
	if err != nil {
		return
	}

	respData := make([]dto.PricingRuleResponse, len(rules))
	for i, rule := range rules {
		respData[i] = mapDBPricingRuleListToResponse(rule) // Use mapping helper
	}
	paginatedResp := dto.PaginatedListResponse{Data: respData, Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset}}
	c.JSON(http.StatusOK, paginatedResp)
}

// DeletePricingRule handles DELETE /pricing-rules/:id
func (h *PricingRuleHandler) DeletePricingRule(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeletePricingRule")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithPricingRuleID(logCtx, id) // Add helper if needed

	// Optional: Fetch rule first to check ownership if needed

	err = h.dbQueries.DeletePricingRule(logCtx, id)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete pricing rule", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pricing rule"})
		return
	}

	slog.InfoContext(logCtx, "Pricing rule deleted successfully")
	c.Status(http.StatusNoContent)
}

// --- Helpers ---
func mapDBPricingRuleToResponse(rule database.GetPricingRuleByIDRow) dto.PricingRuleResponse {
	resp := dto.PricingRuleResponse{
		ID:                rule.ID,
		ServiceProviderID: rule.ServiceProviderID,
		MnoID:             rule.MnoID,
		CurrencyCode:      rule.CurrencyCode,
		PricePerSms:       rule.PricePerSms,
		CreatedAt:         rule.CreatedAt.Time,
		UpdatedAt:         rule.UpdatedAt.Time,
	}
	if rule.MnoName != nil {
		resp.MnoName = rule.MnoName
	}
	return resp
}

func mapDBPricingRuleListToResponse(rule database.ListPricingRulesBySPRow) dto.PricingRuleResponse {
	resp := dto.PricingRuleResponse{
		ID:                rule.ID,
		ServiceProviderID: rule.ServiceProviderID,
		MnoID:             rule.MnoID,
		CurrencyCode:      rule.CurrencyCode,
		PricePerSms:       rule.PricePerSms,
		CreatedAt:         rule.CreatedAt.Time,
		UpdatedAt:         rule.UpdatedAt.Time,
	}
	if rule.MnoName != nil {
		resp.MnoName = rule.MnoName
	}
	return resp
}
