package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
)

// Compile-time check
var _ Pricer = (*DefaultPricer)(nil)

// DefaultPricer holds the pool for transaction management.
type DefaultPricer struct {
	dbPool *pgxpool.Pool
}

func NewDefaultPricer(pool *pgxpool.Pool) *DefaultPricer {
	return &DefaultPricer{dbPool: pool}
}

// PriceAndDebit handles pricing and debiting within a transaction using decimal types.
func (p *DefaultPricer) PriceAndDebit(ctx context.Context, msgID int64) (result *PricingResult, err error) {
	// Add message ID to context for logging
	logCtx := logging.ContextWithMessageID(ctx, msgID)
	slog.DebugContext(logCtx, "Starting price and debit process")

	var totalCost decimal.Decimal
	result = &PricingResult{Debited: false}
	finalStatus := "failed_pricing" // Pessimistic default
	var logMsg string               // Message for critical status update log

	// --- Start Transaction ---
	tx, err := p.dbPool.BeginTx(logCtx, pgx.TxOptions{}) // Use logCtx
	if err != nil {
		result.Error = fmt.Errorf("failed to begin transaction: %w", err)
		slog.ErrorContext(logCtx, "Failed to begin transaction", slog.Any("error", err))
		return result, err // Return internal DB error
	}

	// --- Get Querier operating within the transaction ---
	qtx := database.New(tx) // Create sqlc querier attached to the transaction

	// Defer rollback/commit logic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(logCtx) // Use logCtx for rollback logging if needed
			panic(p)                // Re-panic after rollback attempt
		}

		// Decide whether to commit or rollback based on the 'err' variable
		if err != nil {
			slog.WarnContext(logCtx, "Rolling back pricing transaction", slog.Any("error", err))
			if rbErr := tx.Rollback(logCtx); rbErr != nil {
				slog.ErrorContext(logCtx, "Error rolling back pricing transaction", slog.Any("rollback_error", rbErr), slog.Any("original_error", err))
				// Store original error in result if not already set
				if result.Error == nil {
					result.Error = err
				}
			}
			// Ensure the result error reflects the actual issue if rollback succeeded
			if result.Error == nil {
				result.Error = err
			}
		} else {
			// Commit if err is nil (meaning all steps succeeded)
			if cmErr := tx.Commit(logCtx); cmErr != nil {
				slog.ErrorContext(logCtx, "Error committing pricing transaction", slog.Any("error", cmErr))
				result.Error = cmErr          // Surface the commit error
				err = cmErr                   // Set outer 'err' to prevent incorrect status update
				result.Debited = false        // Commit failed, so debit effectively failed
				finalStatus = "failed_commit" // Specific status for commit failure
			} else {
				// --- SUCCESS ---
				slog.InfoContext(logCtx, "Successfully priced and debited message")
				result.Debited = true
				finalStatus = "queued_for_send" // Set final status only on successful commit
			}
		}

		// --- Update status AFTER commit/rollback attempt ---
		// This runs regardless of transaction outcome to record the final state.
		dbQueriesFromPool := database.New(p.dbPool) // Use main pool querier for this update

		var errMsgForDB string
		if result.Error != nil {
			errMsgForDB = result.Error.Error()
			logMsg = fmt.Sprintf("Updating message status to %s after transaction failure", finalStatus)
		} else if result.Debited {
			logMsg = fmt.Sprintf("Updating message status to %s after successful transaction", finalStatus)
		} else {
			// Should not happen often - Commit failed?
			logMsg = fmt.Sprintf("Updating message status to %s after commit failure", finalStatus)
		}

		statusUpdateErr := dbQueriesFromPool.UpdateMessagePriced(logCtx, database.UpdateMessagePricedParams{
			Column2:          totalCost,
			Column5:          finalStatus,
			ProcessingStatus: finalStatus,
			ErrorDescription: &errMsgForDB,
			ID:               msgID,
		})
		if statusUpdateErr != nil {
			// This is bad - failed to record the outcome. Log critically.
			slog.ErrorContext(logging.ContextWithMessageID(context.Background(), msgID), // Use fresh context with ID for critical log
				"CRITICAL: Failed to update message status AFTER pricing transaction",
				slog.String("attempted_status", finalStatus),
				slog.Any("update_error", statusUpdateErr),
				slog.Any("totalCost", totalCost),
				slog.Any("original_tx_error", result.Error), // Log original error for context
			)
			// Potentially add to a retry queue or alert.
		} else {
			slog.InfoContext(logCtx, logMsg, slog.Any("Cost", totalCost)) // Log the outcome using original context
		}
	}() // End of defer func

	// --- Transactional Logic ---

	// 0. Fetch message details including currency and segments within Tx
	// Assuming you created GetMessageDetailsForPricing:
	// SELECT service_provider_id, routed_mno_id, currency_code, total_segments FROM messages WHERE id = $1 LIMIT 1;
	pricingDetails, err := qtx.GetMessageDetailsForPricing(logCtx, msgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("message with ID %d not found for pricing", msgID)
		} else {
			err = fmt.Errorf("failed to get pricing details: %w", err)
		}
		result.Error = err
		return result, nil // Let defer handle rollback & status update
	}

	// Add SP ID to logging context
	logCtx = logging.ContextWithSPID(logCtx, pricingDetails.ServiceProviderID)

	spID := pricingDetails.ServiceProviderID
	mnoID := pricingDetails.RoutedMnoID
	currency := pricingDetails.CurrencyCode
	totalSegments := pricingDetails.TotalSegments

	logCtx = logging.ContextWithSPID(logCtx, spID)

	if currency == "" {
		err = fmt.Errorf("message currency code is missing for message ID %d", msgID)
		result.Error = err
		return result, nil
	}
	if totalSegments <= 0 { // Should be at least 1
		totalSegments = 1
		slog.WarnContext(logCtx, "Message total_segments is invalid, defaulting to 1", slog.Int("original_value", int(pricingDetails.TotalSegments)))
	}

	// 1. Get Price (per segment)
	priceRule, err := qtx.GetApplicablePrice(logCtx, database.GetApplicablePriceParams{
		ServiceProviderID: spID,
		CurrencyCode:      currency,
		MnoID:             mnoID, // Pass sql.NullInt32 directly
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("no applicable pricing rule found for SP %d, Currency %s, MNO %v", spID, currency, mnoID)
		} else {
			err = fmt.Errorf("failed to get price: %w", err)
		}
		result.Error = err
		return result, nil
	}

	pricePerSegment := priceRule.PricePerSms
	if !pricePerSegment.IsPositive() {
		err = fmt.Errorf("invalid price per segment (zero or negative: %s)", pricePerSegment)
		result.Error = err
		return result, nil
	}

	// 2. Get Wallet (FOR UPDATE locks the row until tx ends)
	wallet, err := qtx.GetWalletForUpdate(logCtx, database.GetWalletForUpdateParams{
		ServiceProviderID: spID,
		CurrencyCode:      currency, // Use the derived currency
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("wallet not found for SP %d, Currency %s", spID, currency)
		} else {
			err = fmt.Errorf("failed to get wallet: %w", err)
		}
		result.Error = err
		return result, nil
	}

	// wallet.Balance should now be decimal.Decimal
	balance := wallet.Balance

	// 3. Calculate Total Cost & Check Balance
	totalCost = pricePerSegment.Mul(decimal.NewFromInt32(totalSegments))

	slog.DebugContext(logCtx, "Pricing check",
		slog.String("balance", fmt.Sprintf("%v", balance)),
		slog.String("price_per_segment", fmt.Sprintf("%v", pricePerSegment)),
		slog.Int("total_segments", int(totalSegments)),
		slog.String("total_cost", totalCost.String()),
	)

	if balance.LessThan(totalCost) {
		err = fmt.Errorf("insufficient balance for %d segments (%d < %d)", totalSegments, balance, totalCost)
		result.Error = err
		return result, nil // Logical failure, defer handles rollback
	}

	// 4. Debit Wallet & Record Transaction
	newBalance := balance.Sub(totalCost)

	// Update Wallet Balance
	_, err = qtx.UpdateWalletBalance(logCtx, database.UpdateWalletBalanceParams{
		Balance: newBalance, // Pass decimal.Decimal directly
		ID:      wallet.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to update wallet balance: %w", err)
		result.Error = err
		return result, nil
	}

	// Create Wallet Transaction
	txDesc := fmt.Sprintf("SMS charge (%d segments) MNO %v", totalSegments, *mnoID)
	_, err = qtx.CreateWalletTransaction(logCtx, database.CreateWalletTransactionParams{
		WalletID:        wallet.ID,
		MessageID:       &msgID,
		TransactionType: "debit",
		Amount:          totalCost,
		BalanceBefore:   balance,
		BalanceAfter:    newBalance,
		Description:     &txDesc,
	})
	if err != nil {
		err = fmt.Errorf("failed to create wallet transaction: %w", err)
		result.Error = err
		return result, nil
	}

	slog.DebugContext(logCtx, "Wallet debited successfully", slog.String("new_balance", newBalance.String()))

	// If we reach here without errors, the outer 'err' is nil.
	// Defer will attempt to commit the transaction.
	return result, nil
}
