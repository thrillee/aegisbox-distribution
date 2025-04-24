package mno

import (
	"context"
	"database/sql" // Import sql package
	"time"

	"github.com/thrillee/aegisbox/internal/database"
)

// DLRInfo holds information parsed from a Delivery Report received from an MNO.
type DLRInfo struct {
	MnoMessageID  string         // The Message ID assigned by the MNO
	OriginalMsgID int64          // Our internal logical message ID (messages.id) - Optional
	SegmentSeqn   int32          // Segment sequence number (1-based) - Optional
	Status        string         // Standard DLR status (e.g., "DELIVRD", "UNDELIV")
	ErrorCode     string         // Error code from MNO DLR (nullable)
	Timestamp     time.Time      // Timestamp when the DLR was generated or received
	NetworkCode   sql.NullString // Optional: MCC/MNC if provided (nullable)
	RawDLRData    map[string]any // Optional: Store extra non-standard fields
}

// DLRHandlerFunc defines the callback signature for processing incoming DLRs.
type DLRHandlerFunc func(ctx context.Context, connectorID int32, dlrInfo DLRInfo) error

// PreparedMessage contains all necessary details to submit a message to an MNO.
type PreparedMessage struct {
	InternalMessageID int64 // Our messages.id
	TotalSegments     int32
	SenderID          string
	DestinationMSISDN string
	MessageContent    string // Full message content (connector handles segmentation)
	IsFlash           bool
	DataCoding        byte
	ESMClass          byte
	RequestDLR        bool
	MNOInfo           database.Mno // Information about the target MNO (optional)
	// Add other fields: ValidityPeriod, Priority, TLVs etc. if needed
}

// SegmentSubmitInfo holds the result for submitting a single segment.
type SegmentSubmitInfo struct {
	Seqn         int32   // Segment sequence number (1-based)
	MnoMessageID *string // MNO-assigned Message ID (nullable if submit fails before ID)
	ErrorCode    *string // Protocol-specific error code on submission failure (nullable)
	Error        error   // Go error if submission failed
	IsSuccess    bool
}

// SubmitResult holds the outcome of submitting a logical message (potentially multiple segments).
type SubmitResult struct {
	Segments     []SegmentSubmitInfo // Result per segment
	WasSubmitted bool                // True if at least one segment was successfully submitted to MNO network layer
	Error        error               // Overall error if submission failed catastrophically (e.g., connection lost before any attempt)
}

// RoutingResult holds the outcome of the routing step.
type RoutingResult struct {
	MNOID     int32 // Use sql.NullInt32 for nullable FK
	ShouldUse bool
	Error     error
}

// Connector defines the interface for interacting with a specific MNO connection (SMPP or HTTP).
type Connector interface {
	// SubmitMessage sends a logical message, handling segmentation internally.
	SubmitMessage(ctx context.Context, msgDetails PreparedMessage) (SubmitResult, error)

	// RegisterDLRHandler provides the callback function for the connector to invoke when a DLR is received.
	RegisterDLRHandler(handler DLRHandlerFunc)

	// Shutdown gracefully closes the connection.
	Shutdown(ctx context.Context) error

	// Status returns the current connection status.
	Status() string

	// ConnectionID returns the ID from the mno_connections table.
	ConnectionID() int32

	// MnoID returns the ID from the mnos table.
	MnoID() int32
}

// ConnectorProvider defines an interface for getting an active MNO connector.
// This abstracts the MNO Manager for the Sender.
type ConnectorProvider interface {
	GetConnector(ctx context.Context, mnoID int32) (Connector, error)
}
