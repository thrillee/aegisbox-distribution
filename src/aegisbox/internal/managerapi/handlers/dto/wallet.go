package dto

import (
	"time"

	"github.com/shopspring/decimal" // Use decimal for money
)

// WalletResponse defines the structure for wallet data returned by the API.
type WalletResponse struct {
	ID                   int32           `json:"id"`
	ServiceProviderID    int32           `json:"service_provider_id"`
	CurrencyCode         string          `json:"currency_code"`
	Balance              decimal.Decimal `json:"balance"`                           // Use decimal
	LowBalanceThreshold  decimal.Decimal `json:"low_balance_threshold"`             // Use decimal
	LowBalanceNotifiedAt *time.Time      `json:"low_balance_notified_at,omitempty"` // Use sql.NullTime for nullable timestamp
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// CreditWalletRequest defines the body for POST /wallets/:id/credit
type CreditWalletRequest struct {
	// Using string for amount allows for more flexible input, parse to decimal in handler
	Amount      string  `json:"amount" binding:"required"`
	Description *string `json:"description"` // Optional description for the transaction
}
