package dto

import "time"

// MNOResponse defines the structure for MNO data returned by the API.
type MNOResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	CountryCode string    `json:"country_code"`
	NetworkCode *string   `json:"network_code,omitempty"` // Use pointer for nullable field
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateMNORequest defines the expected JSON body for creating an MNO.
type CreateMNORequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	CountryCode string  `json:"country_code" binding:"required,max=5"`
	NetworkCode *string `json:"network_code" binding:"omitempty,max=10"`          // Optional
	Status      *string `json:"status" binding:"omitempty,oneof=active inactive"` // Optional, defaults usually handled in DB
}

// UpdateMNORequest defines the expected JSON body for updating an MNO.
type UpdateMNORequest struct {
	// All fields optional for update
	Name        *string `json:"name" binding:"omitempty,max=100"`
	CountryCode *string `json:"country_code" binding:"omitempty,max=5"`
	NetworkCode *string `json:"network_code" binding:"omitempty,max=10"`
	Status      *string `json:"status" binding:"omitempty,oneof=active inactive"`
}
