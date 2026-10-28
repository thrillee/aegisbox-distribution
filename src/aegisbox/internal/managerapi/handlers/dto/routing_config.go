package dto

import "time"

// RoutingGroupConfigResponse is a routing group's override of the global routing config.
// The router reads selection_strategy, max_route_retries and route_failure_threshold from it;
// every other setting comes from the global config.
type RoutingGroupConfigResponse struct {
	ID                    int32     `json:"id"`
	RoutingGroupID        int32     `json:"routing_group_id"`
	RoutingGroupName      string    `json:"routing_group_name"`
	SelectionStrategy     string    `json:"selection_strategy"`
	MaxRouteRetries       int32     `json:"max_route_retries"`
	RouteFailureThreshold int32     `json:"route_failure_threshold"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// CreateRoutingGroupConfigRequest creates a routing group override.
type CreateRoutingGroupConfigRequest struct {
	RoutingGroupID int32 `json:"routing_group_id" binding:"required"`
	UpdateRoutingGroupConfigRequest
}

// UpdateRoutingGroupConfigRequest creates or replaces a routing group override.
type UpdateRoutingGroupConfigRequest struct {
	SelectionStrategy     string `json:"selection_strategy" binding:"required,oneof=priority_weighted health_based round_robin"`
	MaxRouteRetries       int32  `json:"max_route_retries" binding:"min=0,max=5"`
	RouteFailureThreshold int32  `json:"route_failure_threshold" binding:"required,min=1,max=20"`
}

// GlobalRoutingConfigResponse is the global routing config used by the router, the sender and
// route health trackers and the DLR timeout worker. Thresholds are percentages (0 to 100).
type GlobalRoutingConfigResponse struct {
	ID                     int32   `json:"id"`
	SelectionStrategy      string  `json:"selection_strategy"`
	MaxRouteRetries        int32   `json:"max_route_retries"`
	RouteFailureThreshold  int32   `json:"route_failure_threshold"`
	RouteRecoveryThreshold int32   `json:"route_recovery_threshold"`
	RouteHealthWindowSize  int32   `json:"route_health_window_size"`
	SenderBlockThreshold   float64 `json:"sender_block_threshold"`
	SenderUnblockThreshold float64 `json:"sender_unblock_threshold"`
	SenderHealthWindowSize int32   `json:"sender_health_window_size"`
	DLRTimeoutSeconds      int32   `json:"dlr_timeout_seconds"`
	// Fixed by the gateway, read only
	Schedule  WorkerScheduleResponse `json:"schedule"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// WorkerScheduleResponse reports the gateway's fixed worker schedules.
type WorkerScheduleResponse struct {
	DLRTimeoutCheckIntervalSeconds     int32 `json:"dlr_timeout_check_interval_seconds"`
	SenderAutoUnblockCheckIntervalMins int32 `json:"sender_auto_unblock_check_interval_minutes"`
	SenderAutoUnblockAfterInactiveDays int32 `json:"sender_auto_unblock_after_inactive_days"`
	RingBufferCleanupIntervalHours     int32 `json:"ring_buffer_cleanup_interval_hours"`
	RingBufferRetentionDays            int32 `json:"ring_buffer_retention_days"`
}

// UpdateGlobalRoutingConfigRequest replaces the global routing config.
type UpdateGlobalRoutingConfigRequest struct {
	SelectionStrategy      string  `json:"selection_strategy" binding:"required,oneof=priority_weighted health_based round_robin"`
	MaxRouteRetries        int32   `json:"max_route_retries" binding:"min=0,max=5"`
	RouteFailureThreshold  int32   `json:"route_failure_threshold" binding:"required,min=1,max=20"`
	RouteRecoveryThreshold int32   `json:"route_recovery_threshold" binding:"required,min=1,max=20"`
	RouteHealthWindowSize  int32   `json:"route_health_window_size" binding:"required,min=100,max=10000"`
	SenderBlockThreshold   float64 `json:"sender_block_threshold" binding:"min=0,max=100"`
	SenderUnblockThreshold float64 `json:"sender_unblock_threshold" binding:"min=0,max=100,gtefield=SenderBlockThreshold"`
	SenderHealthWindowSize int32   `json:"sender_health_window_size" binding:"required,min=100,max=10000"`
	DLRTimeoutSeconds      int32   `json:"dlr_timeout_seconds" binding:"required,min=60,max=3600"`
}
