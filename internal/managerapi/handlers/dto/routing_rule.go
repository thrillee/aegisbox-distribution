package dto

import "time"

// RoutingRuleResponse defines the structure for routing rule data returned by the API.
type RoutingRuleResponse struct {
	ID        int32     `json:"id"`
	Prefix    string    `json:"prefix"`
	MnoID     int32     `json:"mno_id"`
	MnoName   *string   `json:"mno_name,omitempty"` // Include MNO name via JOIN in query
	Priority  int32     `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateRoutingRuleRequest defines the expected JSON body for creating a rule.
type CreateRoutingRuleRequest struct {
	Prefix   string `json:"prefix" binding:"required,max=20"` // Add validation for prefix format? (e.g., digits only)
	MnoID    int32  `json:"mno_id" binding:"required"`
	Priority *int32 `json:"priority"` // Optional, defaults to 0 in DB
}

// UpdateRoutingRuleRequest defines the expected JSON body for updating a rule.
type UpdateRoutingRuleRequest struct {
	Prefix   *string `json:"prefix" binding:"omitempty,max=20"`
	MnoID    *int32  `json:"mno_id"`
	Priority *int32  `json:"priority"`
}
