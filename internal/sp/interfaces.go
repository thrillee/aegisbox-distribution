package sp

import (
	"context"
	"time"
)

// IncomingSPMessage holds details received from an SP (Client).
type IncomingSPMessage struct {
	ServiceProviderID int32  // Identified SP
	CredentialID      int32  // Identified credential (e.g., smpp_credentials.id)
	Protocol          string // "smpp" or "http"
	ClientMessageRef  string // Optional reference ID provided by the client
	SenderID          string
	DestinationMSISDN string
	MessageContent    string // Full message content (might need reassembly if SP sends multipart)
	TotalSegments     int32  // If SP indicated multipart via UDH (needs detection)
	SegmentSeqn       int32  // If SP indicated multipart via UDH
	ConcatRef         int32  // Concatenation reference from UDH
	IsFlash           bool
	RequestDLR        bool
	ReceivedAt        time.Time
	// Add other fields: DataCoding, ESMClass if available from SP request
}

// Acknowledgement is sent back to the SP after initial message acceptance.
type Acknowledgement struct {
	InternalMessageID int64  // Our messages.id (Null if initial acceptance failed)
	Status            string // e.g., "accepted", "rejected"
	Error             string // Error description if rejected
}

// IncomingMessageHandler defines the interface for the core logic that handles validated incoming messages.
// Both the SMPP Server and HTTP Server will delegate to an implementation of this.
type IncomingMessageHandler interface {
	HandleIncomingMessage(ctx context.Context, msg IncomingSPMessage) (Acknowledgement, error)
}

// --- DLR Forwarding ---

// SPDetails identifies the specific SP connection for forwarding.
type SPDetails struct {
	ServiceProviderID int32
	Protocol          string // "smpp" or "http"
	SMPPSystemID      string // Required if Protocol is "smpp"
	HTTPCallbackURL   string // Required if Protocol is "http"
	HTTPAuthConfig    string // Auth details for HTTP callback (e.g., Bearer token, basic auth) - Store securely
}

// ForwardedDLRInfo contains the data needed to construct a DLR to be sent *to* the SP.
type ForwardedDLRInfo struct {
	InternalMessageID int64     // Our messages.id
	ClientMessageRef  string    // The SP's original reference ID (messages.client_ref)
	SourceAddr        string    // Original Destination MSISDN (of the MT message)
	DestAddr          string    // Original Sender ID (used by the SP)
	SubmitDate        time.Time // When the message was submitted by SP (messages.submitted_at)
	DoneDate          time.Time // When the message reached terminal state (messages.completed_at)
	Status            string    // Final status mapped for SP (e.g., "DELIVRD", "UNDELIV", "REJECTD")
	ErrorCode         string    // Error code (mapped from our internal codes or MNO codes)
	NetworkCode       string    // Optional MCC/MNC
	TotalSegments     int32     // Total segments for the message
	// Potentially add segment sequence if forwarding per-segment DLRs? Usually aggregate.
}

// DLRForwarder defines the interface for sending a DLR back to an SP.
type DLRForwarder interface {
	ForwardDLR(ctx context.Context, spDetails SPDetails, dlr ForwardedDLRInfo) error
}
