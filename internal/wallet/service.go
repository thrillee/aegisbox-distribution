package wallet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
)

// Service defines wallet operations including reversals.
type Service interface {
	ReverseDebit(ctx context.Context, messageID int64, reason string) error
}

// service implements the Wallet Service.
type service struct {
	dbPool    *pgxpool.Pool
	dbQueries database.Querier // Use main querier for finding initial transaction
}

// NewService creates a new wallet service.
func NewService(pool *pgxpool.Pool, queries database.Querier) Service {
	return &service{
		dbPool:    pool,
		dbQueries: queries,
	}
}

// ReverseDebit reverses a specific debit transaction linked to a message ID.
func (s *service) ReverseDebit(ctx context.Context, messageID int64, reason string) (err error) {
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	slog.InfoContext(logCtx, "Attempting debit reversal", slog.String("reason", reason))

	// --- Start Transaction ---
	tx, err := s.dbPool.BeginTx(logCtx, pgx.TxOptions{})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to begin reversal transaction", slog.Any("error", err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Get transactional querier
	qtx := database.New(tx)

	// Defer rollback/commit
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(logCtx)
			panic(p)
		} else if err != nil {
			slog.ErrorContext(logCtx, "Rolling back reversal transaction", slog.Any("error", err))
			if rbErr := tx.Rollback(logCtx); rbErr != nil {
				slog.ErrorContext(logCtx, "Error rolling back reversal transaction", slog.Any("rollback_error", rbErr), slog.Any("original_error", err))
			}
		} else {
			if cmErr := tx.Commit(logCtx); cmErr != nil {
				slog.ErrorContext(logCtx, "Error committing reversal transaction", slog.Any("error", cmErr))
				err = cmErr // Surface commit error
			} else {
				slog.InfoContext(logCtx, "Reversal transaction committed successfully")
			}
		}
	}() // End defer

	// 1. Find the original Debit Transaction using non-transactional querier? No, must be Tx.
	// Need query: FindDebitTransactionForMessage(message_id) -> id, wallet_id, amount
	originalDebit, err := qtx.FindDebitTransactionForMessage(logCtx, &messageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("original debit transaction not found for message ID %d", messageID)
			slog.WarnContext(logCtx, err.Error())
			// Cannot proceed without original debit, but don't rollback, commit Tx (maybe log failure elsewhere?)
			// Or should we rollback? Let's rollback as the state didn't change.
			return err // Defer will rollback because err is set
		}
		err = fmt.Errorf("failed to find original debit transaction: %w", err)
		slog.ErrorContext(logCtx, err.Error())
		return err // Defer will rollback
	}

	// Ensure Amount is positive for reversal calculation
	// originalDebit.Amount is decimal.Decimal from sqlc override
	if !originalDebit.Amount.IsPositive() {
		err = fmt.Errorf("original debit transaction amount is not positive (%s) for Tx ID %d", originalDebit.Amount.String(), originalDebit.ID)
		slog.ErrorContext(logCtx, err.Error())
		return err // Defer will rollback
	}
	amountToReverse := originalDebit.Amount

	// 2. Get Wallet (lock the row within the transaction)
	wallet, err := qtx.GetWalletForUpdateByID(logCtx, originalDebit.WalletID) // Need query GetWalletForUpdateByID(wallet_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("wallet not found (ID: %d) for reversal", originalDebit.WalletID)
		} else {
			err = fmt.Errorf("failed to get wallet (ID: %d) for update: %w", originalDebit.WalletID, err)
		}
		slog.ErrorContext(logCtx, err.Error())
		return err // Defer will rollback
	}

	// 3. Calculate New Balance (Credit the amount back)
	balanceBefore := wallet.Balance
	newBalance := balanceBefore.Add(amountToReverse)
	logCtx = logging.ContextWithWalletID(logCtx, wallet.ID) // Add wallet ID if helper exists

	slog.InfoContext(logCtx, "Calculating reversal",
		slog.String("balance_before", balanceBefore.String()),
		slog.String("amount_to_reverse", amountToReverse.String()),
		slog.String("new_balance", newBalance.String()),
	)

	// 4. Update Wallet Balance
	_, err = qtx.UpdateWalletBalance(logCtx, database.UpdateWalletBalanceParams{
		Balance: newBalance,
		ID:      wallet.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to update wallet balance during reversal: %w", err)
		slog.ErrorContext(logCtx, err.Error())
		return err // Defer will rollback
	}

	// 5. Create Reversal (Credit) Transaction
	desc := fmt.Sprintf("Reversal: %s", reason)
	_, err = qtx.CreateWalletTransaction(logCtx, database.CreateWalletTransactionParams{
		WalletID:        wallet.ID,
		MessageID:       &messageID,
		TransactionType: "reversal", // Use specific type
		Amount:          amountToReverse,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    newBalance,
		Description:     &desc,
		// ReferenceTransactionID: originalDebit.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to create reversal wallet transaction: %w", err)
		slog.ErrorContext(logCtx, err.Error())
		return err // Defer will rollback
	}

	slog.InfoContext(logCtx, "Reversal transaction recorded in DB")

	// If execution reaches here, err is nil, defer will commit.
	return nil
}
