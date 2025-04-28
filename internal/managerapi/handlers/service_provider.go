package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5" // For DB error checks
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
	"github.com/thrillee/aegisbox/internal/wallet"
)

// ServiceProviderHandler holds dependencies for SP handlers.
type ServiceProviderHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool  // Needed for transaction (Create SP + Wallet)
	walletSvc wallet.Service // Potentially needed? Or just call CreateWallet query? Let's use query for now.
}

// NewServiceProviderHandler creates a new handler instance.
func NewServiceProviderHandler(q database.Querier, pool *pgxpool.Pool) *ServiceProviderHandler {
	return &ServiceProviderHandler{dbQueries: q, dbPool: pool}
}

// CreateServiceProvider handles POST /service-providers
func (h *ServiceProviderHandler) CreateServiceProvider(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateServiceProvider") // Use context helper
	var req dto.CreateServiceProviderRequest

	if err := c.ShouldBindJSON(&req); err != nil { /* ... handle bind error ... */
		return
	}

	// --- Transaction: Create SP and default Wallet ---
	tx, err := h.dbPool.BeginTx(logCtx, pgx.TxOptions{}) // Start Tx
	if err != nil {                                      /* ... handle begin error ... */
		return
	}
	defer tx.Rollback(context.Background()) // Ensure rollback on panic or error return

	qtx := database.New(tx) // <<< Create querier attached to transaction

	// Call database query to create SP using qtx
	spParams := database.CreateServiceProviderParams{
		Name:                req.Name,
		Email:               req.Email,
		Status:              "active",
		DefaultCurrencyCode: req.DefaultCurrencyCode,
	}
	createdSP, err := qtx.CreateServiceProvider(logCtx, spParams) // Use qtx
	if err != nil {                                               /* ... handle unique constraint or other DB errors ... */
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, createdSP.ID)

	// Create default wallet for the SP using qtx
	// Make sure decimal.Zero is used correctly if needed
	zeroBalance, _ := decimal.NewFromString("0.000000") // Safer way to ensure precision
	walletParams := database.CreateWalletParams{
		ServiceProviderID:   createdSP.ID,
		CurrencyCode:        createdSP.DefaultCurrencyCode,
		Balance:             zeroBalance, // Use decimal
		LowBalanceThreshold: zeroBalance, // Use decimal
	}
	_, err = qtx.CreateWallet(logCtx, walletParams) // Use qtx
	if err != nil {                                 /* ... handle wallet creation error ... */
		return
	}
	slog.InfoContext(logCtx, "Created default wallet")

	// Commit Transaction
	if err := tx.Commit(logCtx); err != nil { /* ... handle commit error ... */
		return
	}

	// Map DB result to response DTO
	slog.InfoContext(logCtx, "Service provider created successfully")
	c.JSON(http.StatusCreated, createdSP)
}

// ListServiceProviders handles GET /service-providers with pagination
func (h *ServiceProviderHandler) ListServiceProviders(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListServiceProviders")

	limit, offset := parsePagination(c) // Use helper

	// Fetch total count for pagination info
	total, err := h.dbQueries.CountServiceProviders(logCtx)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count service providers", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve service provider count"})
		return
	}

	if total == 0 || offset >= int32(total) {
		// Return empty list if offset is beyond total or no records exist
		c.JSON(http.StatusOK, dto.PaginatedListResponse{
			Data:       []dto.ServiceProviderResponse{},
			Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
		})
		return
	}

	// Fetch the paginated list
	sps, err := h.dbQueries.ListServiceProviders(logCtx, database.ListServiceProvidersParams{Limit: limit, Offset: offset})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list service providers from DB", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve service providers"})
		return
	}

	// Map results to response DTO
	respData := make([]dto.ServiceProviderResponse, len(sps))
	for i, sp := range sps {
		// Use pgtype helper for time conversion if needed: sp.CreatedAt.Time
		respData[i] = dto.ServiceProviderResponse{
			ID:                  sp.ID,
			Name:                sp.Name,
			Email:               sp.Email,
			Status:              sp.Status,
			DefaultCurrencyCode: sp.DefaultCurrencyCode,
			CreatedAt:           sp.CreatedAt.Time,
			UpdatedAt:           sp.UpdatedAt.Time,
		}
	}

	// Prepare paginated response
	paginatedResp := dto.PaginatedListResponse{
		Data: respData,
		Pagination: dto.PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	c.JSON(http.StatusOK, paginatedResp)
}

// GetServiceProvider handles GET /service-providers/:id
func (h *ServiceProviderHandler) GetServiceProvider(c *gin.Context) {
	logCtx := context.WithValue(c.Request.Context(), "handler", "GetServiceProvider")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		slog.WarnContext(logCtx, "Invalid service provider ID format", slog.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, int32(id)) // Add ID to context

	sp, err := h.dbQueries.GetServiceProviderByID(logCtx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "Service provider not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Service provider not found"})
		} else {
			slog.ErrorContext(logCtx, "Failed to get service provider from DB", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve service provider"})
		}
		return
	}

	// Map DB result to response DTO
	resp := dto.ServiceProviderResponse{
		ID:                  sp.ID,
		Name:                sp.Name,
		Email:               sp.Email,
		Status:              sp.Status,
		DefaultCurrencyCode: sp.DefaultCurrencyCode,
		CreatedAt:           sp.CreatedAt.Time,
		UpdatedAt:           sp.UpdatedAt.Time,
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateServiceProvider handles PUT /service-providers/:id
func (h *ServiceProviderHandler) UpdateServiceProvider(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateServiceProvider")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil { /* ... handle invalid ID format ... */
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, int32(id))

	var req dto.UpdateServiceProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil { /* ... handle bind error ... */
		return
	}

	// Prepare params using sqlc.narg helper for optional fields
	// Note: Need to map pointers from DTO to sqlc null types correctly
	params := database.UpdateServiceProviderParams{
		ID:                  int32(id),
		Name:                req.Name,
		Email:               req.Email,
		Status:              req.Status,
		DefaultCurrencyCode: req.DefaultCurrencyCode,
	}

	// Check if default currency change requires wallet handling? Maybe disallow for now.
	if params.DefaultCurrencyCode != nil {
		slog.WarnContext(logCtx, "Attempting to change default currency code - ensure wallets are handled if necessary")
	}

	updatedSP, err := h.dbQueries.UpdateServiceProvider(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "Service provider not found for update")
			c.JSON(http.StatusNotFound, gin.H{"error": "Service provider not found"})
		} else {
			// Handle potential unique constraint violation on email update
			slog.ErrorContext(logCtx, "Failed to update service provider", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service provider"})
		}
		return
	}

	slog.InfoContext(logCtx, "Service provider updated successfully")
	c.JSON(http.StatusOK, updatedSP)
}

// DeleteServiceProvider handles DELETE /service-providers/:id
func (h *ServiceProviderHandler) DeleteServiceProvider(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteServiceProvider")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, int32(id))

	// Consider implications: Deleting SP cascades to credentials, sender IDs, wallets etc.
	// Maybe better to just set status to 'inactive' or 'deleted'?
	// For now, implementing actual delete. Add confirmation step if needed.

	err = h.dbQueries.DeleteServiceProvider(logCtx, int32(id))
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete service provider", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service provider"})
		return
	}

	slog.InfoContext(logCtx, "Service provider deleted successfully")
	c.Status(http.StatusNoContent) // 204 No Content on successful DELETE
}
