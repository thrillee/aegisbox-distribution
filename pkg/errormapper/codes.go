package errormapper

const (
	// Routing & Validation Failures
	ErrorCodeNoRoute           = "NO_ROUTE"
	ErrorCodeInvalidSenderID   = "INVALID_SENDER"
	ErrorCodeTemplateMismatch  = "TEMPLATE_MISMATCH" // Example if using templates
	ErrorCodeValidationFailure = "VALIDATION_FAIL"   // Generic validation failure
	ErrorCodeInvalidMSISDN     = "INVALID_MSISDN"

	// Billing Failures
	ErrorCodeInsufficientFunds = "INSUF_FUNDS"
	ErrorCodePricingNotFound   = "NO_PRICE"
	ErrorCodeWalletNotFound    = "NO_WALLET"

	// Submission/MNO Failures (Can also be MNO-specific codes passed through)
	ErrorCodeMnoUnavailable = "MNO_UNAVAILABLE" // Cannot connect or get client
	ErrorCodeMnoSubmitFail  = "MNO_SUBMIT_FAIL" // Generic MNO rejection during submit_sm
	ErrorCodeMnoFailure     = "MNO_FAILURE"     // Generic MNO failure post-submission (from DLR)

	// System Errors
	ErrorCodeSystemError      = "SYS_ERR" // General internal error
	ErrorCodeDatabaseError    = "DB_ERR"
	ErrorCodeQueueError       = "QUEUE_ERR"
	ErrorCodeConfigError      = "CONFIG_ERR"
	ErrorCodeReversalFailed   = "REVERSAL_FAILED" // Specific error for reversal failure
	ErrorCodeCommitFailed     = "COMMIT_FAILED"   // DB Tx Commit failed
	ErrorCodeFetchDetailsFail = "FETCH_FAIL"      // Failed to fetch necessary data

	// Success/Info (Not errors, but might be used in status mapping)
	StatusCodeDelivered     = "DELIVRD"
	StatusCodeAccepted      = "ACCEPTD" // e.g., Accepted by MNO
	StatusCodeUnknown       = "UNKNOWN"
	StatusCodeRejected      = "REJECTD" // Rejected by us or MNO
	StatusCodeExpired       = "EXPIRED"
	StatusCodeUndeliverable = "UNDELIV"
)
