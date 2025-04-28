package dto

import (
	"encoding/json"
	"time"
)

// BaseSPCredential contains common fields (used for response).
type BaseSPCredential struct {
	ID                int32     `json:"id"`
	ServiceProviderID int32     `json:"service_provider_id"`
	Protocol          string    `json:"protocol"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SPCredentialResponse combines base with protocol-specific fields (excluding sensitive data).
type SPCredentialResponse struct {
	BaseSPCredential
	SystemID         *string         `json:"system_id,omitempty"`          // SMPP
	BindType         *string         `json:"bind_type,omitempty"`          // SMPP
	APIKeyIdentifier *string         `json:"api_key_identifier,omitempty"` // HTTP
	HTTPConfig       json.RawMessage `json:"http_config,omitempty"`        // HTTP (Raw JSON)
}

// CreateSPCredentialRequest defines body for POST /sp-credentials.
type CreateSPCredentialRequest struct {
	ServiceProviderID int32   `json:"service_provider_id" binding:"required"`
	Protocol          string  `json:"protocol" binding:"required,oneof=smpp http"`
	Status            *string `json:"status" binding:"omitempty,oneof=active inactive"`
	// SMPP
	SystemID *string `json:"system_id" binding:"required_if=Protocol smpp"`
	Password *string `json:"password" binding:"required_if=Protocol smpp"`
	BindType *string `json:"bind_type" binding:"required_if=Protocol smpp,omitempty,oneof=trx tx rx"`
	// HTTP
	APIKeyIdentifier *string         `json:"api_key_identifier" binding:"required_if=Protocol http"`
	APIKeySecret     *string         `json:"api_key_secret" binding:"required_if=Protocol http"` // Sent only on create
	HTTPConfig       json.RawMessage `json:"http_config"`                                        // Optional JSON config
}

// UpdateSPCredentialRequest defines body for PUT /sp-credentials/:id.
// Cannot change protocol or SP ID. Can update status, password/apikey, http_config.
type UpdateSPCredentialRequest struct {
	Status     *string         `json:"status" binding:"omitempty,oneof=active inactive"`
	Password   *string         `json:"password"`    // Optional: To reset password (for SMPP)
	HTTPConfig json.RawMessage `json:"http_config"` // Optional: To update HTTP settings
}
