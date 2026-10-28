package dto

import "time"

// CreateRouteSenderRequest adds a sender ID to a connection's route sender pool.
type CreateRouteSenderRequest struct {
	SenderID string `json:"sender_id" binding:"required,min=1,max=16"`
	Status   string `json:"status"`
	Position int32  `json:"position"`
}

// UpdateRouteSenderRequest changes a pool sender's status or position.
type UpdateRouteSenderRequest struct {
	Status   string `json:"status" binding:"required"`
	Position int32  `json:"position"`
}

// UpdateSenderPoolStrategyRequest sets how a connection picks from its pool.
type UpdateSenderPoolStrategyRequest struct {
	Strategy string `json:"strategy" binding:"required"`
}

// UpdateRouteSenderPoolRequest turns the route sender pool on or off for a service provider.
type UpdateRouteSenderPoolRequest struct {
	Enabled bool `json:"enabled"`
}

// RouteSenderResponse is one sender in a connection's pool.
type RouteSenderResponse struct {
	ID              int32      `json:"id"`
	MnoConnectionID int32      `json:"mno_connection_id"`
	SenderID        string     `json:"sender_id"`
	Status          string     `json:"status"`
	Position        int32      `json:"position"`
	UsageCount      int64      `json:"usage_count"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// RouteSenderPoolResponse is a connection's pool and the strategy it picks with.
type RouteSenderPoolResponse struct {
	MnoConnectionID int32                 `json:"mno_connection_id"`
	MnoID           int32                 `json:"mno_id"`
	Strategy        string                `json:"strategy"`
	Senders         []RouteSenderResponse `json:"senders"`
}

// RouteSenderPoolSettingResponse is whether a service provider uses the route sender pool.
type RouteSenderPoolSettingResponse struct {
	ServiceProviderID int32 `json:"service_provider_id"`
	Enabled           bool  `json:"enabled"`
}
