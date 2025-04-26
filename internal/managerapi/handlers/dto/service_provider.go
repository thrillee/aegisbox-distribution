package dto

import "time"

// CreateServiceProviderRequest defines the expected JSON body for creating an SP.
type CreateServiceProviderRequest struct {
	Name                string `json:"name" binding:"required"`
	Email               string `json:"email" binding:"required,email"`
	DefaultCurrencyCode string `json:"default_currency_code" binding:"required,iso4217"` // Use validator tag
	// Add initial status if desired, otherwise defaults to 'pending'
}

// ServiceProviderResponse defines the structure for SP data returned by the API.
type ServiceProviderResponse struct {
	ID                  int32     `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	Status              string    `json:"status"`
	DefaultCurrencyCode string    `json:"default_currency_code"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UpdateServiceProviderRequest defines the expected JSON body for updating an SP.
type UpdateServiceProviderRequest struct {
	Name                *string `json:"name"` // Use pointers for optional updates
	Email               *string `json:"email" binding:"omitempty,email"`
	Status              *string `json:"status" binding:"omitempty,oneof=pending active inactive suspended"` // Example status validation
	DefaultCurrencyCode *string `json:"default_currency_code" binding:"omitempty,iso4217"`
}
