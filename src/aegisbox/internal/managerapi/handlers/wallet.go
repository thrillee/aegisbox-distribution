package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal" // Import decimal
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

type WalletHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool // Pool is required for CreditWallet transaction
}

func NewWalletHandler(q database.Querier, pool *pgxpool.Pool) *WalletHandler {
	if pool == nil {
		panic("DB Pool cannot be nil for WalletHandler")
	}
	return &WalletHandler{dbQueries: q, dbPool: pool}
}

// ListWallets handles GET /wallets (filtered by sp_id)
func (h *WalletHandler) ListWallets(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListWallets")

	// --- Filter by Service Provider ID (REQUIRED) ---
	spIDFilter := parseNullInt32Filter(c, "sp_id") // Use helper from common.go
	if spIDFilter == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid required query parameter: sp_id"})
		return
	}
	spID := spIDFilter
	fmt.Println("SpID: ", spID)
	logCtx = logging.ContextWithSPID(logCtx, *spID)

	// --- Pagination ---
	limit, offset := parsePagination(c) // Use helper from common.go

	// --- Fetch Data ---
	listParams := database.ListWalletsBySPParams{
		ServiceProviderID: *spID,
		Limit:             limit,
		Offset:            offset,
	}

	total, err := h.dbQueries.CountWalletsBySP(logCtx, *spID)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count wallets for SP", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve wallet count"})
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK, dto.PaginatedListResponse{Data: []dto.WalletResponse{}, Pagination: dto.PaginationResponse{Total: 0, Limit: limit, Offset: offset}})
		return
	}

	wallets, err := h.dbQueries.ListWalletsBySP(logCtx, listParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list wallets for SP", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve wallets"})
		return
	}

	// Map results
	respData := make([]dto.WalletResponse, len(wallets))
	for i, w := range wallets {
		respData[i] = mapDBWalletToResponse(w)
	}
	paginatedResp := dto.PaginatedListResponse{
		Data:       respData,
		Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
	}
	c.JSON(http.StatusOK, paginatedResp)
}

// CreditWallet handles POST /wallets/:id/credit
func (h *WalletHandler) CreditWallet(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreditWallet")
	walletID, err := parseIDParam(c) // Use helper for /:id
	if err != nil {
		return
	}
	logCtx = logging.ContextWithWalletID(logCtx, walletID) // Add helper if needed

	var req dto.CreditWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate and parse amount string to decimal
	creditAmount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format: must be a valid number"})
		return
	}
	if !creditAmount.IsPositive() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount: must be positive"})
		return
	}
	// Ensure amount precision matches DB? decimal library handles large precision.
	// Truncate if necessary: creditAmount = creditAmount.Truncate(6)

	slog.InfoContext(logCtx, "Processing manual wallet credit", slog.String("amount", creditAmount.String()))

	// --- Transaction: Get Wallet, Update Balance, Create Transaction ---
	tx, err := h.dbPool.BeginTx(logCtx, pgx.TxOptions{})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to begin transaction for credit", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback(context.Background()) // Rollback if commit doesn't happen

	qtx := database.New(tx) // Transactional querier

	// 1. Get Wallet for Update (Locks row)
	wallet, err := qtx.GetWalletForUpdateByID(logCtx, walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "Wallet not found for credit operation")
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		} else {
			slog.ErrorContext(logCtx, "Failed to get wallet for update", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve wallet"})
		}
		return // Rollback happens in defer
	}
	logCtx = logging.ContextWithSPID(logCtx, wallet.ServiceProviderID) // Add SPID from wallet

	// 2. Calculate new balance
	balanceBefore := wallet.Balance
	newBalance := balanceBefore.Add(creditAmount)

	// 3. Update Wallet Balance
	_, err = qtx.UpdateWalletBalance(logCtx, database.UpdateWalletBalanceParams{
		Balance: newBalance, // decimal.Decimal
		ID:      walletID,
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to update wallet balance during credit", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return // Rollback happens in defer
	}

	// 4. Create Credit Transaction
	desc := "Manual credit via API"
	if req.Description != nil && *req.Description != "" {
		desc = *req.Description
	}
	_, err = qtx.CreateWalletTransaction(logCtx, database.CreateWalletTransactionParams{
		WalletID:        walletID,
		MessageID:       nil, // No message associated with manual credit
		TransactionType: "credit",
		Amount:          creditAmount,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    newBalance,
		Description:     &desc,
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to create credit wallet transaction", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return // Rollback happens in defer
	}

	// 5. Commit Transaction
	if err := tx.Commit(logCtx); err != nil {
		slog.ErrorContext(logCtx, "Failed to commit credit transaction", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database commit error"})
		return
	}

	// 6. Fetch updated wallet state (optional, or just return success)
	updatedWallet, err := h.dbQueries.GetWalletByID(logCtx, walletID) // Use non-transactional querier
	if err != nil {
		slog.ErrorContext(logCtx, "Wallet credited successfully, but failed to fetch updated state", slog.Any("error", err))
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Wallet credited, but failed to fetch updated details"})
		return
	}

	resp := mapDBWalletToResponse(updatedWallet)
	slog.InfoContext(logCtx, "Wallet credited successfully")
	c.JSON(http.StatusOK, resp) // Return updated wallet
}

// --- Helper ---

// mapDBWalletToResponse converts the database model to the API response DTO.
func mapDBWalletToResponse(w database.Wallet) dto.WalletResponse {
	return dto.WalletResponse{
		ID:                   w.ID,
		ServiceProviderID:    w.ServiceProviderID,
		CurrencyCode:         w.CurrencyCode,
		Balance:              w.Balance,                    // Should be decimal.Decimal
		LowBalanceThreshold:  w.LowBalanceThreshold,        // Should be decimal.Decimal
		LowBalanceNotifiedAt: &w.LowBalanceNotifiedAt.Time, // sql.NullTime
		CreatedAt:            w.CreatedAt.Time,             // Adjust if pgtype
		UpdatedAt:            w.UpdatedAt.Time,             // Adjust if pgtype
	}
}
