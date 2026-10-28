package dto

import (
	"time"

	"github.com/shopspring/decimal" // Use decimal for price
)

// PricingRuleResponse defines the structure for pricing rule data returned by the API.
type PricingRuleResponse struct {
	ID                int32           `json:"id"`
	ServiceProviderID int32           `json:"service_provider_id"`
	MnoID             *int32          `json:"mno_id,omitempty"`   // Use sql.NullInt32
	MnoName           *string         `json:"mno_name,omitempty"` // Join MNO Name
	CurrencyCode      string          `json:"currency_code"`
	PricePerSms       decimal.Decimal `json:"price_per_sms"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// CreatePricingRuleRequest defines the expected JSON body for creating a rule.
type CreatePricingRuleRequest struct {
	ServiceProviderID int32           `json:"service_provider_id" binding:"required"`
	MnoID             *int32          `json:"mno_id"` // Optional: If nil, it's a default rule for the currency
	CurrencyCode      string          `json:"currency_code" binding:"required,iso4217"`
	PricePerSms       decimal.Decimal `json:"price_per_sms" binding:"required"` // Gin needs custom validator for decimal? Or validate in handler.
}
