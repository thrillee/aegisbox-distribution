package mno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/codes"
	"github.com/thrillee/aegisbox/pkg/segmenter"
)

// Compile-time check
var _ Connector = (*SMPPMNOConnector)(nil)

// SMPPConnectorConfig holds specific configuration for an SMPP connection derived from mno_connections.
type SMPPConnectorConfig struct {
	ConnectionID      int32         `json:"connection_id"` // From mno_connections.id
	MNOID             int32         `json:"mno_id"`        // From mno_connections.mno_id
	SystemID          string        `json:"system_id"`
	Password          string        `json:"-"` // Avoid logging password
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	UseTLS            bool          `json:"use_tls"`
	BindType          string        `json:"bind_type"` // "trx", "tx", "rx"
	SystemType        string        `json:"system_type"`
	EnquireLink       time.Duration `json:"enquire_link"`        // Calculated from DB secs
	RequestTimeout    time.Duration `json:"request_timeout"`     // Calculated from DB secs
	ConnectRetryDelay time.Duration `json:"connect_retry_delay"` // Calculated from DB secs
	MaxWindowSize     uint          `json:"max_window_size"`     // Max outstanding requests
	DefaultDataCoding data.Encoding `json:"default_data_coding"` // Use gosmpp type
	SourceAddrTON     byte          `json:"source_addr_ton"`     // Default TON/NPI values
	SourceAddrNPI     byte          `json:"source_addr_npi"`
	DestAddrTON       byte          `json:"dest_addr_ton"`
	DestAddrNPI       byte          `json:"dest_addr_npi"`
}

// Helper function to create config from DB model
// Assumes sqlc model `database.MnoConnection` has fields matching the ALTER TABLE additions
func NewSMPPConfigFromDB(dbConn database.MnoConnection) (SMPPConnectorConfig, error) {
	cfg := SMPPConnectorConfig{
		ConnectionID:      dbConn.ID,
		MNOID:             dbConn.MnoID,
		SystemID:          *dbConn.SystemID,
		Password:          *dbConn.Password,
		Host:              *dbConn.Host,
		Port:              int(*dbConn.Port),
		UseTLS:            *dbConn.UseTls,
		BindType:          *dbConn.BindType,
		SystemType:        *dbConn.SystemType,
		EnquireLink:       time.Duration(*dbConn.EnquireLinkIntervalSecs) * time.Second,
		RequestTimeout:    time.Duration(*dbConn.RequestTimeoutSecs) * time.Second,
		ConnectRetryDelay: time.Duration(*dbConn.ConnectRetryDelaySecs) * time.Second,
		MaxWindowSize:     uint(*dbConn.MaxWindowSize),              // Handle NullInt32 - Provide default?
		DefaultDataCoding: data.Encoding(*dbConn.DefaultDataCoding), // Handle NullInt32/byte mapping
		SourceAddrTON:     byte(*dbConn.SourceAddrTon),              // Handle NullInt32/byte mapping
		SourceAddrNPI:     byte(*dbConn.SourceAddrNpi),              // Handle NullInt32/byte mapping
		DestAddrTON:       byte(*dbConn.DestAddrTon),                // Handle NullInt32/byte mapping
		DestAddrNPI:       byte(*dbConn.DestAddrNpi),                // Handle NullInt32/byte mapping
	}

	// Add validation and default values if DB values are missing/invalid
	if cfg.EnquireLink <= 0 {
		cfg.EnquireLink = 30 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.ConnectRetryDelay <= 0 {
		cfg.ConnectRetryDelay = 5 * time.Second
	}
	if cfg.MaxWindowSize <= 0 {
		cfg.MaxWindowSize = 10
	}
	if cfg.Host == "" || cfg.Port == 0 || cfg.SystemID == "" {
		return cfg, errors.New("missing required SMPP config fields (Host, Port, SystemID)")
	}
	// Validate BindType? TON/NPI values?

	return cfg, nil
}

// submitJob holds info needed to correlate SubmitSM with its response.
type submitJob struct {
	dbSegmentID int64
	segmentSeqn int32
	pdu         *pdu.SubmitSM // Keep reference to original PDU? Or just needed fields?
}

// SMPPMNOConnector implements the mno.Connector interface for SMPP.
type SMPPMNOConnector struct {
	config         SMPPConnectorConfig
	session        *gosmpp.Session
	segmenter      segmenter.Segmenter // Injected segmenter
	dbQueries      database.Querier    // For updating segments
	dlrHandler     DLRHandlerFunc      // Callback for incoming DLRs
	status         atomic.Value        // Current status string (thread-safe)
	connMu         sync.Mutex          // Mutex for session creation/closing
	handlerMu      sync.RWMutex        // Mutex for dlrHandler access/update
	pendingSubmits sync.Map            // Map[uint32]*submitJob - Correlate sequence numbers with jobs
	stopSignal     chan struct{}       // Signals background tasks to stop
	wg             sync.WaitGroup      // Waits for background tasks to finish
}

// NewSMPPMNOConnector creates a new, unconnected SMPP connector.
func NewSMPPMNOConnector(cfg SMPPConnectorConfig, seg segmenter.Segmenter, q database.Querier) (*SMPPMNOConnector, error) {
	if cfg.Host == "" || cfg.Port == 0 || cfg.SystemID == "" {
		return nil, errors.New("missing required SMPP config fields (Host, Port, SystemID)")
	}
	// Set defaults if not provided
	if cfg.EnquireLink <= 0 {
		cfg.EnquireLink = 30 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.ConnectRetryDelay <= 0 {
		cfg.ConnectRetryDelay = 5 * time.Second
	}
	if cfg.MaxWindowSize <= 0 {
		cfg.MaxWindowSize = 10
	} // Default window size

	c := &SMPPMNOConnector{
		config:     cfg,
		segmenter:  seg,
		dbQueries:  q,
		stopSignal: make(chan struct{}),
	}
	c.status.Store(codes.StatusDisconnected) // Initial status
	return c, nil
}

// =============================================================================
// Connection Management & Lifecycle
// =============================================================================

// ConnectAndBind attempts to establish and bind the SMPP session.
// It starts background goroutines for handling PDUs and keep-alives.
// Should be called by the MNO Manager.
func (c *SMPPMNOConnector) ConnectAndBind(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.session != nil && c.Status() != codes.StatusDisconnected {
		slog.WarnContext(ctx, "ConnectAndBind called on already active/connecting session", slog.String("status", c.Status()))
		return errors.New("session already active or connecting")
	}

	c.stopSignal = make(chan struct{}) // Recreate stop channel if previously closed
	c.status.Store(codes.StatusConnecting)
	slog.InfoContext(ctx, "Attempting to connect and bind SMPP session",
		slog.String("host", c.config.Host),
		slog.Int("port", c.config.Port),
		slog.String("system_id", c.config.SystemID),
		slog.Int("mno_id", int(c.config.MNOID)),
		slog.Int("conn_id", int(c.config.ConnectionID)),
	)

	auth := gosmpp.Auth{
		SMSC:       fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		SystemID:   c.config.SystemID,
		Password:   c.config.Password,
		SystemType: c.config.SystemType,
	}

	// TODO: Select TLS or NonTLS Dialer based on c.config.UseTLS
	dialer := gosmpp.NonTLSDialer

	var connector gosmpp.Connector
	switch strings.ToLower(c.config.BindType) {
	case "trx", "transceiver":
		connector = gosmpp.TRXConnector(dialer, auth)
	case "tx", "transmitter":
		connector = gosmpp.TXConnector(dialer, auth)
	case "rx", "receiver":
		connector = gosmpp.RXConnector(dialer, auth)
	default:
		c.status.Store(codes.StatusDisconnected)
		return fmt.Errorf("unsupported bind type: %s", c.config.BindType)
	}

	settings := gosmpp.Settings{
		EnquireLink:  c.config.EnquireLink,
		ReadTimeout:  c.config.RequestTimeout + 5*time.Second, // Read timeout slightly longer
		WriteTimeout: c.config.RequestTimeout,

		// Use Windowed Tracking for async submit correlation
		WindowedRequestTracking: &gosmpp.WindowedRequestTracking{
			MaxWindowSize:         uint8(c.config.MaxWindowSize),
			PduExpireTimeOut:      c.config.RequestTimeout,       // Timeout for expecting a response PDU
			ExpireCheckTimer:      5 * time.Second,               // How often to check for expired PDUs
			EnableAutoRespond:     false,                         // We handle responses manually where needed
			OnReceivedPduRequest:  c.handleReceivedPduRequest(),  // Handler for incoming PDUs (DeliverSM, EnquireLink...)
			OnExpectedPduResponse: c.handleExpectedPduResponse(), // Handler for expected responses (SubmitSMResp...)
			OnExpiredPduRequest:   c.handleExpirePduRequest(),    // Handler for requests that timed out
			OnClosePduRequest:     c.handleOnClosePduRequest(),   // Handler for requests closed before response
		},

		// Other callbacks
		OnSubmitError:    c.onSubmitError,    // Errors during Submit() call itself
		OnReceivingError: c.onReceivingError, // Network errors during PDU reading
		OnRebindingError: c.onRebindingError, // Errors during automatic rebind attempts
		OnClosed:         c.onClosed,         // Called when session truly closes
	}

	// Attempt to establish session
	// Use a background context associated with the connector's lifecycle
	sessionCtx, cancelSessionCtx := context.WithCancel(context.Background())

	sess, err := gosmpp.NewSession(connector, settings, 5*time.Second /* connect timeout */)
	if err != nil {
		c.status.Store(codes.StatusDisconnected)
		cancelSessionCtx() // Cancel context if session creation failed
		slog.ErrorContext(ctx, "SMPP session creation failed", slog.Any("error", err))
		return fmt.Errorf("gosmpp.NewSession failed: %w", err)
	}

	c.session = sess
	c.status.Store(codes.StatusBound) // Assume bound immediately after NewSession succeeds (binding happens internally)
	slog.InfoContext(ctx, "SMPP Session established and bound successfully", slog.String("status", c.Status()))

	// Start goroutine to handle cancellation and session closing
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		select {
		case <-c.stopSignal:
			slog.InfoContext(ctx, "Stop signal received, closing session.", slog.Int("conn_id", int(c.config.ConnectionID)))
			c.closeSession(ctx)
			cancelSessionCtx()
		case <-sessionCtx.Done():
			slog.InfoContext(ctx, "Session context done.", slog.Int("conn_id", int(c.config.ConnectionID)))
			c.closeSession(ctx) // Ensure session is closed if context ends
		}
	}()

	return nil
}

// closeSession handles the actual unbinding and closing.
func (c *SMPPMNOConnector) closeSession(ctx context.Context) {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.session == nil {
		return // Already closed or never opened
	}

	currentStatus := c.Status()
	if currentStatus == codes.StatusUnbinding || currentStatus == codes.StatusDisconnected {
		slog.DebugContext(ctx, "Session already closing or closed", slog.String("status", currentStatus))
		return // Avoid duplicate closing
	}

	slog.InfoContext(ctx, "Closing SMPP session...", slog.Int("conn_id", int(c.config.ConnectionID)), slog.String("status", c.Status()))
	c.status.Store(codes.StatusUnbinding)

	// Attempt graceful unbind first (optional, Close might handle it)
	// if c.Status() == codes.StatusBound {
	//  unbind PDU might need to be created and sent via Windowed tracking?
	//  Or gosmpp Close handles unbind internally? Assume Close handles it.
	// }

	err := c.session.Close()
	if err != nil {
		slog.WarnContext(ctx, "Error during SMPP session close", slog.Any("error", err), slog.Int("conn_id", int(c.config.ConnectionID)))
	} else {
		slog.InfoContext(ctx, "SMPP session closed.", slog.Int("conn_id", int(c.config.ConnectionID)))
	}

	c.session = nil
	c.status.Store(codes.StatusDisconnected)
	// Clear pending submits? Windowed tracker's OnClosePduRequest should handle this.
}

// Shutdown signals the connector to close gracefully.
func (c *SMPPMNOConnector) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Shutdown requested for SMPP connector", slog.Int("conn_id", int(c.config.ConnectionID)))
	close(c.stopSignal) // Signal the background goroutine to initiate close
	c.wg.Wait()         // Wait for background tasks (like the closer goroutine) to finish
	slog.InfoContext(ctx, "Shutdown completed for SMPP connector", slog.Int("conn_id", int(c.config.ConnectionID)))
	return nil
}

// Status returns the current connection status.
func (c *SMPPMNOConnector) Status() string {
	return c.status.Load().(string)
}

// ConnectionID returns the configured connection ID.
func (c *SMPPMNOConnector) ConnectionID() int32 {
	return c.config.ConnectionID
}

// MnoID returns the configured MNO ID.
func (c *SMPPMNOConnector) MnoID() int32 {
	return c.config.MNOID
}

// =============================================================================
// Message Submission
// =============================================================================

// SubmitMessage handles segmentation and submits segments via the SMPP session.
func (c *SMPPMNOConnector) SubmitMessage(ctx context.Context, msg PreparedMessage) (SubmitResult, error) {
	logCtx := logging.ContextWithMessageID(ctx, msg.InternalMessageID)
	logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
	slog.InfoContext(logCtx, "Submitting message via SMPP")

	result := SubmitResult{WasSubmitted: false}

	if c.Status() != codes.StatusBound {
		err := fmt.Errorf("cannot submit message, session not bound (status: %s)", c.Status())
		slog.ErrorContext(logCtx, err.Error())
		result.Error = err
		return result, nil // Return logical error
	}

	// 1. Segmentation (using injected segmenter)
	// Determine encoding based on content or PreparedMessage field if added
	segments, requiresUCS2, err := c.segmenter.GetSegments(msg.MessageContent)
	if err != nil {
		err = fmt.Errorf("message segmentation failed: %w", err)
		slog.ErrorContext(logCtx, err.Error())
		result.Error = err
		return result, nil // Return logical error
	}
	totalSegments := len(segments)
	if totalSegments == 0 {
		err = errors.New("segmentation resulted in zero segments")
		slog.ErrorContext(logCtx, err.Error())
		result.Error = err
		return result, nil
	}

	// 2. Create DB Segment Records (required BEFORE submitting)
	segmentIDs := make([]int64, totalSegments)
	dbCtx, cancel := context.WithTimeout(logCtx, 5*time.Second) // Timeout for DB operations
	defer cancel()
	for i := 0; i < totalSegments; i++ {
		segID, dbErr := c.dbQueries.CreateMessageSegment(dbCtx, database.CreateMessageSegmentParams{
			MessageID:   msg.InternalMessageID,
			SegmentSeqn: int32(i + 1),
		})
		if dbErr != nil {
			err = fmt.Errorf("failed to create DB record for segment %d/%d: %w", i+1, totalSegments, dbErr)
			slog.ErrorContext(logCtx, err.Error())
			result.Error = err // Mark overall error
			// Don't attempt to send if we can't create DB records
			return result, err // Return internal error
		}
		segmentIDs[i] = segID
	}
	slog.DebugContext(logCtx, "Created segment DB records", slog.Int("count", totalSegments))

	// 3. Submit Segments
	result.Segments = make([]SegmentSubmitInfo, totalSegments)
	var submitErrorsExist bool

	for i, content := range segments {
		dbSegmentID := segmentIDs[i]
		segmentSeqn := int32(i + 1)
		segLogCtx := logging.ContextWithSegmentID(logCtx, dbSegmentID)

		// Create SubmitSM PDU
		p, err := c.buildSubmitSM(segLogCtx, msg, content, totalSegments, segmentSeqn, requiresUCS2)
		if err != nil {
			slog.ErrorContext(segLogCtx, "Failed to build SubmitSM PDU", slog.Any("error", err), slog.Int("seqn", int(segmentSeqn)))
			errCode := codes.ErrorCodeSystemError
			result.Segments[i] = SegmentSubmitInfo{Seqn: segmentSeqn, IsSuccess: false, Error: err, ErrorCode: &errCode}
			submitErrorsExist = true
			continue // Try next segment? Or stop? Let's stop if PDU build fails.
			// result.Error = fmt.Errorf("failed to build PDU for segment %d: %w", segmentSeqn, err)
			// return result, nil // Stop entire submission
		}

		// Store job info for correlation *before* submitting
		job := &submitJob{
			dbSegmentID: dbSegmentID,
			segmentSeqn: segmentSeqn,
			pdu:         p, // Store reference if needed in callbacks
		}
		// Use sequence number from PDU as key
		seq := p.GetSequenceNumber()
		c.pendingSubmits.Store(seq, job)

		// Submit asynchronously
		err = c.session.Transceiver().Submit(p) // Use Transceiver for TRX bind type

		if err != nil {
			// Remove pending job if submit fails immediately
			c.pendingSubmits.Delete(seq)

			// Handle submission errors (e.g., window full, encoding issues)
			slog.WarnContext(segLogCtx, "Failed to submit segment PDU", slog.Any("error", err), slog.Int("seqn", int(segmentSeqn)))
			result.Segments[i] = SegmentSubmitInfo{Seqn: segmentSeqn, IsSuccess: false, Error: err}
			// Map specific gosmpp errors to codes if possible
			if errors.Is(err, gosmpp.ErrWindowsFull) {
				tmpErr := string(codes.ErrorCodeMnoWindowFull)
				result.Segments[i].ErrorCode = &tmpErr // Define this code
				// TODO: Implement retry/wait logic for window full?
			} else {
				tmpErr := string(codes.MsgStatusSendFailed)
				result.Segments[i].ErrorCode = &tmpErr
			}
			submitErrorsExist = true
			// Update DB immediately for submit failure? Let OnSubmitError handle maybe?
			// For now, record in result. DB update happens via async callbacks if submit succeeds.
			errDesc := err.Error()
			_ = c.dbQueries.UpdateSegmentSendFailed(segLogCtx, database.UpdateSegmentSendFailedParams{
				ID: dbSegmentID, ErrorDescription: &errDesc, ErrorCode: result.Segments[i].ErrorCode,
			})

			// Continue to next segment even if one fails? Yes.
		} else {
			// Submit call succeeded, response will come via callback
			slog.DebugContext(segLogCtx, "Segment PDU submitted, awaiting response", slog.Uint64("pdu_seq", uint64(seq)))
			result.Segments[i] = SegmentSubmitInfo{Seqn: segmentSeqn, IsSuccess: false} // Mark as submitted, success determined by response
			result.WasSubmitted = true                                                  // Mark that at least one segment submit call succeeded
		}
	} // End segment loop

	// Update overall result error if any segment submission failed immediately
	if submitErrorsExist && result.Error == nil {
		// Provide a generic error indicating some segments failed submission
		result.Error = errors.New("one or more segments failed immediate submission")
	}

	slog.InfoContext(logCtx, "Finished submitting segments", slog.Bool("was_submitted_attempted", result.WasSubmitted), slog.Bool("immediate_errors", submitErrorsExist))
	// Return result reflecting submission *attempts*. DB is updated asynchronously.
	return result, nil
}

// buildSubmitSM constructs the PDU for a single segment.
func (c *SMPPMNOConnector) buildSubmitSM(ctx context.Context, msg PreparedMessage, segmentContent string, totalSegments int, segmentSeqn int32, requiresUCS2 bool) (*pdu.SubmitSM, error) {
	p := pdu.NewSubmitSM().(*pdu.SubmitSM)

	// Set Addresses
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(c.config.SourceAddrTON) // Use configured defaults
	srcAddr.SetNpi(c.config.SourceAddrNPI)
	if err := srcAddr.SetAddress(msg.SenderID); err != nil {
		return nil, fmt.Errorf("invalid source address (SenderID) '%s': %w", msg.SenderID, err)
	}
	p.SourceAddr = srcAddr

	destAddr := pdu.NewAddress()
	destAddr.SetTon(c.config.DestAddrTON)
	destAddr.SetNpi(c.config.DestAddrNPI)
	if err := destAddr.SetAddress(msg.DestinationMSISDN); err != nil {
		return nil, fmt.Errorf("invalid destination address '%s': %w", msg.DestinationMSISDN, err)
	}
	p.DestAddr = destAddr

	// Set Message Content and Encoding
	var coding data.Encoding
	if requiresUCS2 {
		coding = data.UCS2
	} else {
		coding = c.config.DefaultDataCoding // Use configured default (e.g., data.GSM7BIT)
	}

	err := p.Message.SetMessageWithEncoding(segmentContent, coding)
	if err != nil {
		return nil, fmt.Errorf("failed to set message content (encoding: %d): %w", coding, err)
	}

	// Set Flags and Options
	p.ProtocolID = 0 // Usually 0
	if msg.RequestDLR {
		p.RegisteredDelivery = 1 // Request DLR (receipt requested)
	} else {
		p.RegisteredDelivery = 0 // No DLR requested
	}
	p.ReplaceIfPresentFlag = 0 // Standard behavior

	// ESM Class: Handle UDH and Message Type (e.g., Flash)
	p.EsmClass = 0 // smpp.MSG_MODE_DEFAULT | smpp.MSG_TYPE_DEFAULT // Default
	if totalSegments > 1 {
		p.EsmClass |= 0
	}

	return p, nil
}

// =============================================================================
// Asynchronous Callback Handlers (passed to gosmpp.Settings)
// =============================================================================

func (c *SMPPMNOConnector) onSubmitError(p pdu.PDU, err error) {
	// Error during the Submit() call itself (e.g., window full)
	logCtx := context.Background() // No request context here
	logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
	slog.WarnContext(logCtx, "gosmpp OnSubmitError callback triggered", slog.Any("error", err), slog.String("pdu_cmd", p.CommandID().String()), slog.Uint32("pdu_seq", p.GetSequenceNumber()))

	// We already handled error returned by Submit() call in SubmitMessage.
	// This callback might be redundant or provide extra info.
	// If ErrWindowsFull, we might implement waiting/retry logic here or in SubmitMessage.
}

func (c *SMPPMNOConnector) onReceivingError(err error) {
	// Network or PDU parsing errors when reading from SMSC
	logCtx := context.Background()
	logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
	slog.ErrorContext(logCtx, "gosmpp OnReceivingError callback triggered", slog.Any("error", err))
	// This might indicate a need to reconnect. The library might handle auto-rebind.
}

func (c *SMPPMNOConnector) onRebindingError(err error) {
	// Error during automatic rebind attempts
	logCtx := context.Background()
	logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
	slog.ErrorContext(logCtx, "gosmpp OnRebindingError callback triggered", slog.Any("error", err))
	// Maybe update status to disconnected if rebind fails persistently?
}

func (c *SMPPMNOConnector) onClosed(state gosmpp.State) {
	// Session closed (either requested or unexpected)
	logCtx := context.Background()
	logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
	slog.WarnContext(logCtx, "gosmpp OnClosed callback triggered", slog.String("final_state", state.String()))
	c.status.Store(codes.StatusDisconnected)
	// Signal MNO Manager to potentially attempt reconnection after delay?
}

// --- WindowedRequestTracking Callbacks ---

func (c *SMPPMNOConnector) handleReceivedPduRequest() func(pdu.PDU) (pdu.PDU, bool) {
	return func(p pdu.PDU) (resp pdu.PDU, autoRespond bool) {
		// Handles PDUs initiated by SMSC (DeliverSM, EnquireLink, Unbind...)
		logCtx := context.Background()
		logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
		logCtx = logging.ContextWithPDUInfo(logCtx, p.GetHeader().CommandID.String(), p.GetSequenceNumber())

		switch pd := p.(type) {
		case *pdu.DeliverSM:
			slog.InfoContext(logCtx, "Received DeliverSM from MNO")
			c.processIncomingDeliverSM(logCtx, pd)
			return pd.GetResponse(), false // Send DeliverSMResp manually or let auto-respond handle if enabled

		case *pdu.EnquireLink:
			slog.DebugContext(logCtx, "Received EnquireLink from MNO")
			return pd.GetResponse(), false // Auto-respond handled by library usually, but return resp anyway

		case *pdu.Unbind:
			slog.InfoContext(logCtx, "Received Unbind request from MNO")
			// Need to close session gracefully
			c.status.Store(codes.StatusUnbinding)
			go c.closeSession(context.Background()) // Close session in background
			return pd.GetResponse(), false          // Acknowledge unbind

		case *pdu.DataSM:
			slog.InfoContext(logCtx, "Received DataSM from MNO")
			// TODO: Handle DataSM if needed (e.g., for USSD)
			return pd.GetResponse(), false // Send DataSMResp

		case *pdu.AlertNotification:
			slog.InfoContext(logCtx, "Received AlertNotification from MNO")
			// No response required for AlertNotification

		default:
			slog.WarnContext(logCtx, "Received unexpected PDU type from MNO")
		}
		return nil, false // No response for unhandled types or AlertNotification
	}
}

func (c *SMPPMNOConnector) handleExpectedPduResponse() func(response gosmpp.Response) {
	return func(response gosmpp.Response) {
		// Handles responses to our requests (SubmitSMResp, BindResp, EnquireLinkResp...)
		reqPDU := response.OriginalRequest.PDU
		respPDU := response.PDU

		logCtx := context.Background()
		logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
		logCtx = logging.ContextWithPDUInfo(logCtx, reqPDU.GetHeader().CommandID.String(), reqPDU.GetSequenceNumber()) // Original Seq No

		switch resp := respPDU.(type) {
		case *pdu.SubmitSMResp:
			slog.DebugContext(logCtx, "Received SubmitSMResp")
			c.processSubmitSMResp(logCtx, uint32(reqPDU.GetSequenceNumber()), resp)

		case *pdu.BindTransceiverResp, *pdu.BindTransmitterResp, *pdu.BindReceiverResp:
			status := respPDU.CommandStatus()
			slog.InfoContext(logCtx, "Received BindResp", slog.String("status", status.String()))
			if status != pdu.StatusOK {
				slog.ErrorContext(logCtx, "Bind failed", slog.String("status", status.String()))
				c.status.Store(codes.StatusBindingFailed) // Update status on bind failure
				// Close session? Library might auto-retry or close.
				go c.closeSession(context.Background()) // Close session on bind fail
			} else {
				c.status.Store(codes.StatusBound) // Confirm bound status
			}

		case *pdu.EnquireLinkResp:
			slog.DebugContext(logCtx, "Received EnquireLinkResp")
			// Keepalive successful

		case *pdu.UnbindResp:
			slog.InfoContext(logCtx, "Received UnbindResp")
			// Session should be closing

		default:
			slog.WarnContext(logCtx, "Received unexpected response PDU type", slog.String("resp_type", respPDU.GetHeader().CommandID.String()))
		}
	}
}

func (c *SMPPMNOConnector) handleExpirePduRequest() func(pdu.PDU) bool {
	return func(p pdu.PDU) bool {
		// Handles requests we sent for which no response was received within timeout
		logCtx := context.Background()
		logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
		logCtx = logging.ContextWithPDUInfo(logCtx, p.GetHeader().CommandID.String(), p.GetSequenceNumber())

		slog.WarnContext(logCtx, "PDU request expired (no response)", slog.String("status", c.Status()))

		switch p.(type) {
		case *pdu.SubmitSM:
			// A submit timed out - MNO might have processed it or not. Ambiguous.
			// Mark segment as failed? Or unknown? Treat as failure for now.
			c.processSubmitSMTimeout(logCtx, uint32(p.GetSequenceNumber()))
			return false // Don't close connection for submit timeout

		case *pdu.EnquireLink:
			// EnquireLink timed out - connection is likely dead.
			slog.ErrorContext(logCtx, "EnquireLink expired, connection likely stale")
			go c.closeSession(context.Background()) // Initiate close
			return true                             // Indicate connection should be considered stale

		default:
			slog.WarnContext(logCtx, "Unhandled expired PDU type")
		}
		return false // Default: don't assume connection is stale
	}
}

func (c *SMPPMNOConnector) handleOnClosePduRequest() func(pdu.PDU) {
	return func(p pdu.PDU) {
		// Handles requests that were pending when the session closed.
		logCtx := context.Background()
		logCtx = logging.ContextWithMNOConnID(logCtx, c.ConnectionID())
		logCtx = logging.ContextWithPDUInfo(logCtx, p.GetHeader().CommandID.String(), p.GetSequenceNumber())
		slog.WarnContext(logCtx, "PDU request closed before response/expiry", slog.String("status", c.Status()))

		switch p.(type) {
		case *pdu.SubmitSM:
			// SubmitSM closed - Treat as failure.
			c.processSubmitSMTimeout(logCtx, uint32(p.GetSequenceNumber())) // Use same logic as timeout

		default:
			// Log others but usually no action needed
		}
		// Clean up from pending map
		c.pendingSubmits.Delete(p.GetSequenceNumber())
	}
}

// =============================================================================
// PDU Processing Logic
// =============================================================================

// processSubmitSMResp handles the asynchronous response to a SubmitSM request.
func (c *SMPPMNOConnector) processSubmitSMResp(ctx context.Context, sequence uint32, resp *pdu.SubmitSMResp) {
	// Retrieve the job details associated with this sequence number
	val, loaded := c.pendingSubmits.LoadAndDelete(sequence)
	if !loaded {
		slog.WarnContext(ctx, "Received SubmitSMResp for unknown or already processed sequence number")
		return
	}
	job := val.(*submitJob)
	logCtx := logging.ContextWithSegmentID(ctx, job.dbSegmentID) // Add segment ID

	status := resp.CommandStatus

	if status == pdu.StatusOK {
		mnoMsgID := resp.MessageID
		slog.InfoContext(logCtx, "SubmitSM successful according to Resp", slog.String("mno_msg_id", mnoMsgID))
		// Update segment as successfully submitted (MNO accepted)
		connID := c.ConnectionID()
		err := c.dbQueries.UpdateSegmentSent(logCtx, database.UpdateSegmentSentParams{
			ID:              job.dbSegmentID,
			MnoMessageID:    &mnoMsgID,
			MnoConnectionID: &connID,
		})
		if err != nil {
			slog.ErrorContext(logCtx, "Failed to update DB after successful SubmitSMResp", slog.Any("error", err))
			// Critical: MNO accepted but we couldn't record it. Alert maybe?
		}
	} else {
		// Submission rejected by MNO
		errMsg := fmt.Sprintf("SubmitSM rejected by MNO: %s (0x%X)", status.Desc(), status)
		slog.WarnContext(logCtx, "SubmitSM rejected by MNO", slog.String("status", status.String()), slog.Uint64("status_code", uint64(status)))
		errCode := fmt.Sprintf("%03X", uint32(status))
		err := c.dbQueries.UpdateSegmentSendFailed(logCtx, database.UpdateSegmentSendFailedParams{
			ID:               job.dbSegmentID,
			ErrorDescription: &errMsg,
			ErrorCode:        &errCode,
		})
		if err != nil {
			slog.ErrorContext(logCtx, "Failed to update DB after failed SubmitSMResp", slog.Any("error", err))
		}
		// TODO: Trigger Reversal here? No, reversal trigger is in Processor if Send() fails overall.
		// TODO: Generate failure DLR here? No, Processor handles that based on final status.
	}
}

// processSubmitSMTimeout handles when a SubmitSM request expires without a response.
func (c *SMPPMNOConnector) processSubmitSMTimeout(ctx context.Context, sequence uint32) {
	val, loaded := c.pendingSubmits.LoadAndDelete(sequence)
	if !loaded {
		slog.WarnContext(ctx, "SubmitSM timed out for unknown or already processed sequence number")
		return
	}
	job := val.(*submitJob)
	logCtx := logging.ContextWithSegmentID(ctx, job.dbSegmentID)

	errMsg := "SubmitSM timed out waiting for response from MNO"
	slog.WarnContext(logCtx, errMsg)

	// Treat timeout as failure, but status is ambiguous. Mark with specific error.
	errCode := codes.ErrorCodeMnoTimeout
	err := c.dbQueries.UpdateSegmentSendFailed(logCtx, database.UpdateSegmentSendFailedParams{
		ID:               job.dbSegmentID,
		ErrorDescription: &errMsg,
		ErrorCode:        &errCode, // Define this code
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to update DB after SubmitSM timeout", slog.Any("error", err))
	}
	// TODO: Trigger Reversal here? Maybe, as MNO state is unknown. Risky. Best left to higher level logic based on DLRs or final status.
	// TODO: Generate failure DLR? Status is unknown, maybe wait or send specific 'TIMEOUT' status DLR?
}

// processIncomingDeliverSM handles received DeliverSM PDUs (potential DLRs or MO messages).
func (c *SMPPMNOConnector) processIncomingDeliverSM(ctx context.Context, p *pdu.DeliverSM) {
	c.handlerMu.RLock()
	handler := c.dlrHandler // Get current handler under read lock
	c.handlerMu.RUnlock()

	if handler == nil {
		slog.WarnContext(ctx, "Received DeliverSM but no DLR handler is registered", slog.Int("conn_id", int(c.config.ConnectionID)))
		return
	}

	if p.IsReceipt() { // Check ESM class bit for DLR
		slog.InfoContext(ctx, "DeliverSM is a Delivery Receipt")
		// Attempt to parse standard receipt format
		receipt, err := p.Receipt() // Use gosmpp helper
		if err != nil {
			slog.WarnContext(ctx, "Failed to parse standard DLR format from DeliverSM short_message", slog.Any("error", err))
			// TODO: Implement manual parsing for non-standard formats if needed
			// Fallback: Try to extract MNO ID if possible even if format is bad?
			// If cannot parse ID, cannot process DLR.
			return
		}

		dlrInfo := DLRInfo{
			MnoMessageID: receipt.MessageID,
			Status:       receipt.Stat,
			ErrorCode:    receipt.Err,      // Map error code if present
			Timestamp:    receipt.DoneDate, // Use DoneDate from DLR
			// SegmentSeqn: // Can we get this reliably from standard DLR? Usually no. Assume DLR is for specific MNO ID -> segment.
			// NetworkCode: // Extract if needed and possible
			RawDLRData: nil, // Add raw fields if needed
		}

		// Call the registered handler (e.g., sms.Processor.UpdateSegmentDLRStatus)
		err = handler(ctx, c.ConnectionID(), dlrInfo)
		if err != nil {
			slog.ErrorContext(ctx, "DLR Handler failed processing", slog.String("mno_msg_id", receipt.MessageID), slog.Any("error", err))
			// Should we NACK the DeliverSM? Usually no, just log the processing error.
		}

	} else {
		// This is a Mobile Originated (MO) message
		slog.InfoContext(ctx, "Received Mobile Originated (MO) message")
		// TODO: Implement MO message handling:
		// 1. Extract details: source_addr (MSISDN), dest_addr (Shortcode), short_message, data_coding etc.
		// 2. Insert into a dedicated 'mo_messages' table or route to an application logic handler.
		// Example: moHandler.HandleMO(ctx, source, dest, content, coding)
	}
}

// =============================================================================
// Interface Method Implementations
// =============================================================================

// RegisterDLRHandler sets the callback for incoming DLRs.
func (c *SMPPMNOConnector) RegisterDLRHandler(handler DLRHandlerFunc) {
	c.handlerMu.Lock()
	defer c.handlerMu.Unlock()
	slog.Info("Registering DLR Handler", slog.Int("conn_id", int(c.config.ConnectionID)))
	c.dlrHandler = handler
}
