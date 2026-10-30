package dto

import "time"

// ManualDLRRequest defines the JSON body for sending a manual DLR for one message.
type ManualDLRRequest struct {
	Status    string `json:"status" binding:"required"` // DELIVRD, UNDELIV, EXPIRED, REJECTD or DELETED
	ErrorCode string `json:"error_code" binding:"omitempty,max=10"`
}

// BulkManualDLRRequest defines the JSON body for sending a manual DLR for many messages.
type BulkManualDLRRequest struct {
	MessageIDs []int64 `json:"message_ids" binding:"required,min=1,max=200"`
	Status     string  `json:"status" binding:"required"`
	ErrorCode  string  `json:"error_code" binding:"omitempty,max=10"`
	// IncludeUnconfirmed also completes segments sent with no MNO message ID recorded.
	IncludeUnconfirmed bool `json:"include_unconfirmed"`
}

// ManualDLRResult reports the outcome of a manual DLR for one message.
// Forwarded means the DLR was queued for forwarding to the SP (queue_id); the
// forwarding worker retries delivery with backoff.
type ManualDLRResult struct {
	MessageID       int64  `json:"message_id"`
	SegmentsUpdated int    `json:"segments_updated"`
	Forwarded       bool   `json:"forwarded"`
	QueueID         int64  `json:"queue_id,omitempty"`
	ForwardWarning  string `json:"forward_warning,omitempty"`
	Error           string `json:"error,omitempty"`
}

// BulkManualDLRResponse summarises a bulk manual DLR.
type BulkManualDLRResponse struct {
	Status    string            `json:"status"`
	Forwarded int               `json:"forwarded"`
	Skipped   int               `json:"skipped"`
	Failed    int               `json:"failed"`
	Warnings  int               `json:"warnings"`
	Results   []ManualDLRResult `json:"results"`
}

// UpdateAutoDLRRequest defines the JSON body for a service provider's auto DLR config.
type UpdateAutoDLRRequest struct {
	Enabled      bool   `json:"enabled"`
	Status       string `json:"status" binding:"required"`
	DelaySeconds int32  `json:"delay_seconds" binding:"required,min=60,max=172800"`
}

// AutoDLRResponse is a service provider's auto DLR config.
type AutoDLRResponse struct {
	ServiceProviderID int32      `json:"service_provider_id"`
	Enabled           bool       `json:"enabled"`
	Status            string     `json:"status"`
	DelaySeconds      int32      `json:"delay_seconds"`
	EnabledAt         *time.Time `json:"enabled_at"`
}

// RetryDLRForwardingRequest defines the JSON body for requeueing failed DLR forwarding jobs.
type RetryDLRForwardingRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1,max=200"`
}

// RetryDLRForwardingResponse lists the jobs put back in the queue.
type RetryDLRForwardingResponse struct {
	Requeued []int64 `json:"requeued"`
	Skipped  int     `json:"skipped"`
}
