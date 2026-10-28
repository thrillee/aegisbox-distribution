package dto

import "time"

// RouteHealthResponse represents route health metrics
type RouteHealthResponse struct {
	AssignmentID        int32      `json:"assignment_id"`
	RoutingGroupName    string     `json:"routing_group_name"`
	PrefixGroupName     string     `json:"prefix_group_name"`
	MNOName             string     `json:"mno_name"`
	HealthScore         string     `json:"health_score"`
	SuccessCount        int32      `json:"success_count"`
	FailCount           int32      `json:"fail_count"`
	TotalAttempts       int32      `json:"total_attempts"`
	ConsecutiveFailures int32      `json:"consecutive_failures"`
	IsAvailable         bool       `json:"is_available"`
	LastAttemptAt       *time.Time `json:"last_attempt_at"`
	Priority            int32      `json:"priority"`
	Weight              int32      `json:"weight"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// SenderHealthResponse represents sender ID health metrics
type SenderHealthResponse struct {
	ID              int32      `json:"id"`
	SenderID        string     `json:"sender_id"`
	MNOName         string     `json:"mno_name"`
	DeliveryRate    string     `json:"delivery_rate"`
	DLRResponseRate string     `json:"dlr_response_rate"`
	HealthScore     string     `json:"health_score"`
	TotalMessages   int32      `json:"total_messages"`
	IsBlocked       bool       `json:"is_blocked"`
	BlockedAt       *time.Time `json:"blocked_at,omitempty"`
	BlockedReason   *string    `json:"blocked_reason,omitempty"`
	LastMessageAt   *time.Time `json:"last_message_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// HealthStatsResponse represents aggregated health statistics
type HealthStatsResponse struct {
	TotalRoutes       int32   `json:"total_routes"`
	HealthyRoutes     int32   `json:"healthy_routes"`
	WarningRoutes     int32   `json:"warning_routes"`
	CriticalRoutes    int32   `json:"critical_routes"`
	UnavailableRoutes int32   `json:"unavailable_routes"`
	TotalSenders      int32   `json:"total_senders"`
	BlockedSenders    int32   `json:"blocked_senders"`
	AvgDeliveryRate   float64 `json:"avg_delivery_rate"`
}

// MarkRouteAvailableRequest for manual route availability control
type MarkRouteAvailableRequest struct {
	// Pointer so an explicit false is accepted ("required" rejects a false bool)
	Available *bool  `json:"available" binding:"required"`
	Reason    string `json:"reason"`
}

// UnblockSenderRequest for manually unblocking a sender
type UnblockSenderRequest struct {
	Reason string `json:"reason"`
}

// HealthEventType represents the type of health event
type HealthEventType string

const (
	EventRouteMarkedUnavailable HealthEventType = "route-marked-unavailable"
	EventRouteMarkedAvailable   HealthEventType = "route-marked-available"
	EventRouteAutoUnavailable   HealthEventType = "route-auto-unavailable"
	EventRouteRecovered         HealthEventType = "route-recovered"
	EventSenderBlocked          HealthEventType = "sender-blocked"
	EventSenderUnblocked        HealthEventType = "sender-unblocked"
	EventHealthUpdated          HealthEventType = "health-updated"
)

// HealthEvent represents a health event for SSE streaming
type HealthEvent struct {
	Type      HealthEventType `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      interface{}     `json:"data"`
}

// RouteHealthEvent data for route-related events
type RouteHealthEvent struct {
	AssignmentID        int32  `json:"assignment_id"`
	RoutingGroupName    string `json:"routing_group_name"`
	MNOName             string `json:"mno_name"`
	HealthScore         string `json:"health_score"`
	ConsecutiveFailures int32  `json:"consecutive_failures"`
	IsAvailable         bool   `json:"is_available"`
}

// SenderHealthEvent data for sender-related events
type SenderHealthEvent struct {
	SenderID      string `json:"sender_id"`
	MNOName       string `json:"mno_name"`
	HealthScore   string `json:"health_score"`
	DeliveryRate  string `json:"delivery_rate"`
	IsBlocked     bool   `json:"is_blocked"`
	BlockedReason string `json:"blocked_reason,omitempty"`
}
