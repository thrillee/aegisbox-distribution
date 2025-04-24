package codes

// Connection Status Codes
const (
	StatusDisconnected  = "disconnected"
	StatusConnecting    = "connecting"
	StatusBinding       = "binding" // Specifically for SMPP
	StatusBound         = "bound"   // Specifically for SMPP
	StatusUnbinding     = "unbinding"
	StatusBindingFailed = "binding_failed"
	StatusHttpOk        = "http_ok"    // For active HTTP connections/endpoints
	StatusHttpError     = "http_error" // Indicate issues with HTTP endpoint
	StatusDisabled      = "disabled"   // Manually disabled in config
	StatusActive        = "active"     // General active state (can be bound or http_ok)
	StatusError         = "error"      // General error state
)

// Message Processing Status Codes (already used, but good to centralize)
const (
	MsgStatusReceived         = "received"
	MsgStatusValidated        = "validated"
	MsgStatusFailedValidation = "failed_validation"
	MsgStatusRouted           = "routed"
	MsgStatusFailedRouting    = "failed_routing"
	MsgStatusPricingChecked   = "pricing_checked" // Or maybe skipped if combined
	MsgStatusFailedPricing    = "failed_pricing"
	MsgStatusQueuedForSend    = "queued_for_send"
	MsgStatusSendAttempted    = "send_attempted"
	MsgStatusSendFailed       = "send_failed"   // Catastrophic send fail or MNO reject on submit
	MsgStatusFailedCommit     = "failed_commit" // Specific status for DB commit issues
)

// Message Final Status Codes (already used, but good to centralize)
const (
	MsgFinalStatusPending    = "pending"
	MsgFinalStatusProcessing = "processing" // Still waiting on DLRs
	MsgFinalStatusDelivered  = "delivered"
	MsgFinalStatusFailed     = "failed"   // Failed after MNO accepted (from DLR or timeout)
	MsgFinalStatusRejected   = "rejected" // Failed validation/routing or MNO rejected on submit
	MsgFinalStatusExpired    = "expired"
	MsgFinalStatusUnknown    = "unknown"
)

// MNO Submission Error Codes (Examples)
const (
	ErrorCodeMnoWindowFull = "MNO_WINDOW_FULL"
	ErrorCodeMnoTimeout    = "MNO_TIMEOUT"
	ErrorCodeSystemError   = "SYS_ERR"
)
