package smppserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/thrillee/aegisbox/internal/auth"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
)

const (
	BindTransceiver     uint32 = 0x00000009
	BindTransceiverResp        = 0x80000009
	SubmitSM            uint32 = 0x00000004
	SubmitSMResp               = 0x80000004
	DeliverSM           uint32 = 0x00000005
	Unbind              uint32 = 0x00000006
	UnbindResp                 = 0x80000006
	EnquireLink         uint32 = 0x00000015
	EnquireLinkResp            = 0x80000015
)

// PDUHeader is the SMPP PDU header (16 bytes).
type PDUHeader struct {
	Length         uint32
	CommandID      uint32
	CommandStatus  uint32
	SequenceNumber uint32
}

// sessionState holds bound client info for the raw server.
type sessionState struct {
	credentialID      int32
	serviceProviderID int32
	systemID          string
	conn              net.Conn      // Raw TCP connection
	writer            *bufio.Writer // Buffered writer for efficiency
	readMu            sync.Mutex    // Mutex for reading PDUs
	writeMu           sync.Mutex    // Mutex for writing PDUs
	boundAt           time.Time
	lastActivity      time.Time // Track last PDU read/write for idle timeout
}

// Compile-time checks for interface implementation
var (
	_ RawPDUForwarder = (*Server)(nil) // Interface for the forwarder to use
)

// Server implements the raw SMPP TCP server.
type Server struct {
	config         config.ServerConfig       // Server config (Addr, timeouts)
	dbQueries      database.Querier          // For auth
	messageHandler sp.IncomingMessageHandler // Core logic handler
	sessions       map[string]*sessionState  // map[systemID]*sessionState
	sessionsMu     sync.RWMutex              // Mutex for the sessions map
	listener       net.Listener
	shutdown       chan struct{}  // Signal for graceful shutdown
	wg             sync.WaitGroup // Waitgroup for active connections
}

// NewServer creates a new raw SMPP server instance.
func NewServer(ctx context.Context, cfg config.ServerConfig, q database.Querier, handler sp.IncomingMessageHandler) *Server {
	if handler == nil {
		panic("Message handler cannot be nil for SMPP Server")
	}
	return &Server{
		config:         cfg,
		dbQueries:      q,
		messageHandler: handler,
		sessions:       make(map[string]*sessionState),
		shutdown:       make(chan struct{}),
	}
}

// ListenAndServe accepts raw TCP connections and handles SMPP sessions.
func (s *Server) ListenAndServe() error {
	slog.Info("Starting Raw SMPP Server", slog.String("address", s.config.Addr))
	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		slog.Error("Failed to listen on address", slog.String("address", s.config.Addr), slog.Any("error", err))
		return fmt.Errorf("net.Listen failed: %w", err)
	}
	s.listener = ln // Store listener to allow closing

	// Goroutine to handle shutdown signal
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		<-s.shutdown // Wait for shutdown signal
		slog.Info("Shutdown signal received by listener goroutine.")
		_ = s.listener.Close() // Close listener to stop accepting new connections
	}()

	// Accept loop
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.shutdown: // Check if shutdown was requested
				slog.Info("SMPP Listener closed gracefully.")
				return nil // Normal exit after listener closed
			default:
				slog.Error("Failed to accept connection", slog.Any("error", err))
				// Maybe add a small delay before retrying?
				time.Sleep(100 * time.Millisecond)
				continue // Continue accepting? Or return error?
			}
		}

		logCtx := logging.ContextWithRemoteAddr(context.Background(), conn.RemoteAddr().String()) // Add helper
		slog.InfoContext(logCtx, "Accepted new SMPP connection")
		ss := &sessionState{
			conn:         conn,
			writer:       bufio.NewWriter(conn),
			lastActivity: time.Now(),
		}

		s.wg.Add(1) // Add to waitgroup for this connection handler
		go s.handleSession(logCtx, ss)
	}
}

// handleSession reads and processes PDUs for a single connection.
func (s *Server) handleSession(ctx context.Context, ss *sessionState) {
	defer func() {
		// Cleanup on exit
		if ss.systemID != "" { // Only remove if session was bound
			s.removeSession(ss.systemID)
		}
		_ = ss.conn.Close()
		slog.InfoContext(ctx, "Closed SMPP client connection")
		s.wg.Done() // Decrement waitgroup counter
	}()

	r := bufio.NewReader(ss.conn)
	isBound := false // Track bind status for this session

	for {
		// Set read deadline (for idle timeout and graceful shutdown)
		// TODO: Make idle timeout configurable
		_ = ss.conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout + 30*time.Second)) // Generous read timeout

		hdr, body, err := s.readPDU(r, ss) // Use read method with mutex
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.InfoContext(ctx, "Client closed connection (EOF).")
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				slog.InfoContext(ctx, "Client connection read timeout/idle.")
				// Optional: Send EnquireLink here if idle? Or just close.
				// Let's close for simplicity now.
			} else if errors.Is(err, net.ErrClosed) {
				slog.InfoContext(ctx, "Connection closed (likely during shutdown).")
			} else {
				slog.WarnContext(ctx, "Error reading PDU", slog.Any("error", err))
			}
			return // Exit handler on any read error/EOF/timeout
		}
		ss.lastActivity = time.Now() // Update activity timestamp

		// Enrich context with PDU info for logging this request
		logCtx := logging.ContextWithPDUInfo(ctx, commandIDToString(hdr.CommandID), int32(hdr.SequenceNumber))
		if ss.systemID != "" {
			logCtx = logging.ContextWithSystemID(logCtx, ss.systemID)
		}

		// --- PDU Handling ---
		// Require BIND before most other commands
		if !isBound && hdr.CommandID != CommandBindTransceiver && hdr.CommandID != CommandBindReceiver && hdr.CommandID != CommandBindTransmitter {
			slog.WarnContext(logCtx, "Received command before BIND", slog.String("command", commandIDToString(hdr.CommandID)))
			s.writeErrorResponse(ss, hdr, StatusInvBndSts) // Invalid Bind Status
			continue                                       // Don't process further
		}

		switch hdr.CommandID {
		case CommandBindTransceiver, CommandBindReceiver, CommandBindTransmitter:
			if isBound {
				slog.WarnContext(logCtx, "Received BIND request on already bound session.")
				s.writeErrorResponse(ss, hdr, StatusInvBndSts) // Already bound
			} else {
				isBound = s.handleBind(logCtx, ss, hdr, body) // handleBind returns true on successful bind
			}
		// Unified entry point for SubmitSM and SubmitMultiSM based on initial header
		case CommandSubmitSM, CommandSubmitMultiSM:
			s.handleGenericSubmit(logCtx, ss, hdr, body)
		case CommandUnbind:
			s.handleUnbind(logCtx, ss, hdr)
			return // Close connection after unbind response sent
		case CommandEnquireLink:
			slog.DebugContext(logCtx, "Received EnquireLink")
			// Send EnquireLinkResp
			respHdr := makeHeader(CommandEnquireLinkResp, hdr.SequenceNumber, StatusOk)
			s.writePDU(ss, respHdr, nil) // Response has no body
		default:
			slog.WarnContext(logCtx, "Received unknown/unhandled Command ID", slog.Any("COMMAND ID", hdr.CommandID), slog.Any("COMMAND STATUS", hdr.CommandStatus))
			// Unknown => Generic Nack with Invalid Command ID status
			s.writeErrorResponse(ss, hdr, StatusInvCmdID)
		}

		// Flush buffer after handling PDU and sending response (if any)
		ss.writeMu.Lock()
		flushErr := ss.writer.Flush()
		ss.writeMu.Unlock()
		if flushErr != nil {
			slog.WarnContext(logCtx, "Error flushing writer buffer", slog.Any("error", flushErr))
			return // Connection likely broken
		}
	}
}

// readPDU reads the header and body, handling potential concurrent reads.
func (s *Server) readPDU(r *bufio.Reader, ss *sessionState) (PDUHeader, []byte, error) {
	ss.readMu.Lock() // Protect reading from the connection
	defer ss.readMu.Unlock()

	var hdr PDUHeader
	hdrBytes := make([]byte, 16)

	// Read exactly 16 bytes for the header
	_, err := io.ReadFull(r, hdrBytes)
	if err != nil {
		// Don't return hdr if read failed, it's unpopulated/invalid
		return PDUHeader{}, nil, err // Propagate EOF or other read errors
	}

	// Parse header fields from bytes
	hdr.Length = binary.BigEndian.Uint32(hdrBytes[0:4])
	hdr.CommandID = binary.BigEndian.Uint32(hdrBytes[4:8])
	hdr.CommandStatus = binary.BigEndian.Uint32(hdrBytes[8:12])
	hdr.SequenceNumber = binary.BigEndian.Uint32(hdrBytes[12:16])

	// Basic validation
	if hdr.Length < 16 {
		return hdr, nil, fmt.Errorf("invalid PDU length: %d (must be >= 16)", hdr.Length)
	}

	// Calculate body length (handle potential overflow if Length is huge?)
	bodyLen := int(hdr.Length) - 16

	// Read body if it exists
	body := make([]byte, bodyLen)
	if bodyLen > 0 {
		_, err = io.ReadFull(r, body)
		if err != nil {
			return hdr, nil, fmt.Errorf("error reading PDU body (expected %d bytes): %w", bodyLen, err)
		}
	}

	return hdr, body, nil
}

// writePDU constructs and writes a PDU, handling potential concurrent writes.
func (s *Server) writePDU(ss *sessionState, hdr PDUHeader, body []byte) error {
	ss.writeMu.Lock() // Protect writing to the connection/buffer
	defer ss.writeMu.Unlock()

	hdr.Length = uint32(16 + len(body)) // Ensure length is correct

	buf := make([]byte, hdr.Length)
	binary.BigEndian.PutUint32(buf[0:], hdr.Length)
	binary.BigEndian.PutUint32(buf[4:], hdr.CommandID)
	binary.BigEndian.PutUint32(buf[8:], hdr.CommandStatus)
	binary.BigEndian.PutUint32(buf[12:], hdr.SequenceNumber)
	if len(body) > 0 {
		copy(buf[16:], body)
	}

	// Write the entire PDU to the buffered writer
	n, err := ss.writer.Write(buf)
	if err == nil && n != len(buf) {
		err = io.ErrShortWrite
	}
	if err != nil {
		slog.ErrorContext(context.TODO(), "Failed to write PDU to buffer", slog.Any("error", err), slog.String("system_id", ss.systemID))
		return err
	}

	return nil
}

// writeErrorResponse simplifies sending a response PDU with only a header and status.
func (s *Server) writeErrorResponse(ss *sessionState, reqHdr PDUHeader, status uint32) {
	respHdr := makeHeader(reqHdr.CommandID|0x80000000, reqHdr.SequenceNumber, status)
	if err := s.writePDU(ss, respHdr, nil); err != nil {
		slog.WarnContext(context.TODO(), "Failed to write error response PDU", slog.Any("error", err), slog.String("system_id", ss.systemID))
	}
}

// makeHeader creates a basic PDU header. Status defaults to OK (0).
func makeHeader(cmdID, seq, status uint32) PDUHeader {
	return PDUHeader{
		Length:         16, // Base length, body adds later
		CommandID:      cmdID,
		CommandStatus:  status,
		SequenceNumber: seq,
	}
}

// helper to read C-style NUL-terminated string from a byte slice.
// Returns the string, the number of bytes read (including NUL), and ok status.
func readCString(b []byte) (string, int, bool) {
	idx := bytes.IndexByte(b, 0x00)
	if idx == -1 {
		// NUL terminator not found
		return "", 0, false
	}
	return string(b[:idx]), idx + 1, true
}

// handleBind parses bind PDU, performs authentication, stores session, and sends response.
// Returns true if bind was successful, false otherwise.
func (s *Server) handleBind(ctx context.Context, ss *sessionState, hdr PDUHeader, body []byte) bool {
	slog.DebugContext(ctx, "Handling Bind request")
	var systemID, password, systemType string
	var ok bool
	var n int
	offset := 0

	// Parse required fields carefully
	systemID, n, ok = readCString(body[offset:])
	if !ok || systemID == "" {
		slog.WarnContext(ctx, "Bind failed: Cannot parse system_id or is empty")
		s.writeErrorResponse(ss, hdr, StatusInvSysID)
		return false
	}
	offset += n
	logCtx := logging.ContextWithSystemID(ctx, systemID) // Add systemID to context now

	password, n, ok = readCString(body[offset:])
	if !ok { // Password can be empty, but must be terminated
		slog.WarnContext(logCtx, "Bind failed: Cannot parse password field")
		s.writeErrorResponse(ss, hdr, StatusInvPasswd) // Treat parsing error as invalid password?
		return false
	}
	offset += n

	systemType, n, ok = readCString(body[offset:])
	if !ok {
		slog.WarnContext(logCtx, "Bind failed: Cannot parse system_type field")
		s.writeErrorResponse(ss, hdr, StatusSystemError) // Or a more specific param error if available
		return false
	}
	slog.InfoContext(ctx, "Binding:Sytem Type", slog.String("sytemType", systemType))
	offset += n

	// TODO: Parse interface_version, addr_ton, addr_npi, address_range if needed

	// --- Authentication ---
	authCtx, cancel := context.WithTimeout(logCtx, 5*time.Second)
	defer cancel()
	cred, err := s.dbQueries.GetSPCredentialBySystemID(authCtx, &systemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "Bind failed: SystemID not found in DB")
			s.writeErrorResponse(ss, hdr, StatusInvSysID)
			return false
		}
		slog.ErrorContext(logCtx, "Bind auth DB error", slog.Any("error", err))
		s.writeErrorResponse(ss, hdr, StatusSystemError)
		return false
	}

	// Check protocol and password
	if cred.Protocol != "smpp" {
		slog.WarnContext(logCtx, "Bind failed: Credential protocol mismatch", slog.String("expected", "smpp"), slog.String("actual", cred.Protocol))
		s.writeErrorResponse(ss, hdr, StatusBindFailed)
		return false
	}
	if cred.PasswordHash != nil && !auth.CheckPasswordHash(password, *cred.PasswordHash) {
		slog.WarnContext(logCtx, "Bind failed: Invalid password")
		s.writeErrorResponse(ss, hdr, StatusInvPasswd)
		return false
	}
	// --- Auth Success ---

	// Update session state
	ss.systemID = systemID
	ss.credentialID = cred.ID
	ss.serviceProviderID = cred.ServiceProviderID
	ss.boundAt = time.Now()
	ss.lastActivity = ss.boundAt

	logCtx = logging.ContextWithSPID(logCtx, ss.serviceProviderID) // Add SP ID

	// Store session
	s.sessionsMu.Lock()
	// Check if another session exists for this systemID (shouldn't happen with unique constraint)
	if existing, exists := s.sessions[systemID]; exists {
		slog.WarnContext(logCtx, "Duplicate bind attempt for already bound system ID. Closing old connection.", slog.String("old_remote_addr", existing.conn.RemoteAddr().String()))
		_ = existing.conn.Close() // Force close old connection
		// delete(s.sessions, systemID) // removeSession will be called by old handler exiting
	}
	s.sessions[systemID] = ss
	s.sessionsMu.Unlock()

	// Send successful BindResp
	respHdr := makeHeader(CommandBindResp, hdr.SequenceNumber, StatusOk)
	// Body: system_id<NUL>[Optional TLVs]
	respBody := append([]byte(systemID), 0x00) // Respond with *our* system ID (make configurable)
	respHdr.Length += uint32(len(respBody))
	if err := s.writePDU(ss, respHdr, respBody); err != nil {
		slog.ErrorContext(logCtx, "Failed to write successful BindResp PDU", slog.Any("error", err))
		// If write fails, connection is likely dead. Cleanup happens when handleSession exits.
		return false // Indicate bind failed due to write error
	}
	slog.InfoContext(logCtx, "Bind successful")
	return true // Bind succeeded
}

// handleUnbind sends response and marks session for removal.
func (s *Server) handleUnbind(ctx context.Context, ss *sessionState, hdr PDUHeader) {
	slog.InfoContext(ctx, "Handling Unbind request")
	respHdr := makeHeader(CommandUnbindResp, hdr.SequenceNumber, StatusOk)
	if err := s.writePDU(ss, respHdr, nil); err != nil {
		slog.WarnContext(ctx, "Failed to write UnbindResp PDU", slog.Any("error", err))
		// Session will be cleaned up when handleSession exits anyway
	}
}

// handleGenericSubmit parses SubmitSM or SubmitMulti PDU, routes to appropriate handler logic,
// calls the core message handler, and sends the corresponding response.
// It uses the parsed PDU's CommandID for routing.
func (s *Server) handleGenericSubmit(ctx context.Context, ss *sessionState, originalHdr PDUHeader, bodyBytes []byte) {
	slog.InfoContext(ctx, "Handling generic submit request",
		slog.String("original_header_command_id_str", commandIDToString(originalHdr.CommandID)),
		slog.Uint64("original_header_command_id_val", uint64(originalHdr.CommandID)),
		slog.Uint64("original_sequence_number", uint64(originalHdr.SequenceNumber)))

	// Reconstruct the full PDU bytes as received for gosmpp parsing
	// The bodyBytes already excludes the 16-byte header, originalHdr contains that header.
	fullPduBytesForParsing := make([]byte, 16+len(bodyBytes))
	binary.BigEndian.PutUint32(fullPduBytesForParsing[0:4], originalHdr.Length)
	binary.BigEndian.PutUint32(fullPduBytesForParsing[4:8], originalHdr.CommandID)
	binary.BigEndian.PutUint32(fullPduBytesForParsing[8:12], originalHdr.CommandStatus)
	binary.BigEndian.PutUint32(fullPduBytesForParsing[12:16], originalHdr.SequenceNumber)
	copy(fullPduBytesForParsing[16:], bodyBytes)

	// Create a gosmpp buffer
	smppBuf := pdu.NewBuffer(fullPduBytesForParsing)

	// Parse the PDU using gosmpp
	parsedPdu, err := pdu.Parse(smppBuf)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse PDU in handleGenericSubmit",
			slog.Any("error", err),
			slog.String("original_header_command_id", commandIDToString(originalHdr.CommandID)))
		s.writeErrorResponse(ss, originalHdr, StatusInvMsgLen) // Invalid Message Length or other suitable error
		return
	}

	parsedCommandID := uint32(parsedPdu.GetHeader().CommandID)
	slog.InfoContext(ctx, "PDU parsed in handleGenericSubmit",
		slog.String("original_header_command_id", commandIDToString(originalHdr.CommandID)),
		slog.String("parsed_pdu_command_id", commandIDToString(parsedCommandID)))

	// Explicitly log if there's a mismatch, as this is key to the problem described
	if originalHdr.CommandID != parsedCommandID {
		slog.WarnContext(ctx, "MISMATCH between received header CommandID and parsed PDU CommandID",
			slog.String("received_command", commandIDToString(originalHdr.CommandID)),
			slog.String("parsed_command", commandIDToString(parsedCommandID)))
		// We will proceed based on the parsedCommandID as it reflects the PDU body's content.
	}

	// Switch on the PARSED PDU's command ID
	switch parsedCommandID {
	case CommandSubmitSM: // Corresponds to CommandSubmitSM (0x00000004)
		submitSM, ok := parsedPdu.(*pdu.SubmitSM)
		if !ok {
			slog.WarnContext(ctx, "Parsed PDU as SubmitSMID, but type assertion to *pdu.SubmitSM failed",
				slog.String("pdu_type", fmt.Sprintf("%T", parsedPdu)),
				slog.String("original_header_command_id", commandIDToString(originalHdr.CommandID)))
			s.writeErrorResponse(ss, originalHdr, StatusInvCmdID) // Use originalHdr for seq num
			return
		}
		slog.InfoContext(ctx, "Processing as SubmitSM based on parsed PDU body")
		s.processSubmitSM(ctx, ss, originalHdr, submitSM)

	case CommandSubmitMultiSM: // Corresponds to CommandSubmitMultiSM (0x00000021)
		submitMulti, ok := parsedPdu.(*pdu.SubmitMulti)
		if !ok {
			slog.WarnContext(ctx, "Parsed PDU as SubmitMultiID, but type assertion to *pdu.SubmitMulti failed",
				slog.String("pdu_type", fmt.Sprintf("%T", parsedPdu)),
				slog.String("original_header_command_id", commandIDToString(originalHdr.CommandID)))
			s.writeErrorResponse(ss, originalHdr, StatusInvCmdID) // Use originalHdr for seq num
			return
		}
		slog.InfoContext(ctx, "Processing as SubmitMulti based on parsed PDU body")
		s.processSubmitMulti(ctx, ss, originalHdr, submitMulti)

	default:
		slog.WarnContext(ctx, "Received PDU with unhandled parsed Command ID in handleGenericSubmit",
			slog.String("parsed_command_id_str", commandIDToString(parsedCommandID)),
			slog.Uint64("parsed_command_id_val", uint64(parsedCommandID)),
			slog.String("original_header_command_id", commandIDToString(originalHdr.CommandID)))
		s.writeErrorResponse(ss, originalHdr, StatusInvCmdID) // Invalid Command ID
	}
}

// processSubmitSM handles the logic for a parsed SubmitSM PDU.
func (s *Server) processSubmitSM(ctx context.Context, ss *sessionState, originalHdr PDUHeader, submitSM *pdu.SubmitSM) {
	// Extract message details from parsed PDU
	sourceAddr := submitSM.SourceAddr.Address()
	destAddr := submitSM.DestAddr.Address()
	messageContent, err := submitSM.Message.GetMessage()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to Get Message from SubmitSM", slog.Any("error", err))
		s.writeErrorResponse(ss, originalHdr, StatusInvMsgLen)
		return
	}
	registeredDelivery := submitSM.RegisteredDelivery

	totalParts := 1
	mref := 0 // Changed from 1 to 0 for non-concatenated default, check SMPP spec or gosmpp UDH behavior
	currentPartNum := 1

	if udhBytes := submitSM.Message.UDH(); len(udhBytes) > 0 {
		udh := pdu.UDH(udhBytes)
		tpByte, cpByte, mrByte, found := udh.GetConcatInfo() // GetConcatInfo often returns 8-bit ref
		if found {
			totalParts = int(tpByte)
			currentPartNum = int(cpByte) // Make sure to get current part number
			mref = int(mrByte)
		}
	}

	// --- Call Core Handler ---
	inMsg := sp.IncomingSPMessage{
		ServiceProviderID: ss.serviceProviderID,
		CredentialID:      ss.credentialID,
		Protocol:          "smpp",
		ClientMessageRef:  "", // Populate from optional params if needed
		SenderID:          sourceAddr,
		DestinationMSISDN: destAddr,
		MessageContent:    messageContent,
		TotalSegments:     int32(totalParts),
		SegmentSeqn:       int32(currentPartNum), // Use current part number from UDH
		ConcatRef:         int32(mref),
		IsFlash:           isFlashMessage(submitSM.Message.Encoding().DataCoding()), // Ensure isFlashMessage exists
		RequestDLR:        (registeredDelivery & 0x01) != 0,
		ReceivedAt:        time.Now().UTC(),
	}

	ack, coreErr := s.messageHandler.HandleIncomingMessage(ctx, inMsg)
	if coreErr != nil {
		slog.ErrorContext(ctx, "Incoming message handler failed for SubmitSM", slog.Any("error", coreErr))
		s.writeErrorResponse(ss, originalHdr, StatusSystemError)
		return
	}

	if ack.Status != "accepted" || ack.InternalMessageID == "" {
		slog.WarnContext(ctx, "Message rejected by core handler for SubmitSM", slog.String("reason", ack.Error))
		status := mapErrorToSMPPStatus(ack.Error) // Ensure mapErrorToSMPPStatus exists
		s.writeErrorResponse(ss, originalHdr, status)
		return
	}

	// Create and send SubmitSMResp
	resp := submitSM.GetResponse().(*pdu.SubmitSMResp)
	resp.MessageID = ack.InternalMessageID
	// Note: resp.Status() is already ESME_ROK by default from GetResponse()

	// Marshal the gosmpp response PDU to get its bytes
	respBuf := pdu.NewBuffer(nil)
	resp.Marshal(respBuf)
	marshaledRespPduBytes := respBuf.Bytes() // Full PDU bytes (header + body)

	if len(marshaledRespPduBytes) < 16 {
		slog.ErrorContext(ctx, "Marshaled SubmitSMResp is too short", slog.Int("length", len(marshaledRespPduBytes)))
		return
	}

	// Write response
	respHdr := PDUHeader{
		Length:         uint32(len(marshaledRespPduBytes)),
		CommandID:      SubmitSMResp,
		CommandStatus:  uint32(resp.GetHeader().CommandStatus),
		SequenceNumber: originalHdr.SequenceNumber,
	}

	if err := s.writePDU(ss, respHdr, marshaledRespPduBytes[16:]); err != nil {
		slog.ErrorContext(ctx, "Failed to write SubmitSMResp", slog.Any("error", err))
	} else {
		slog.InfoContext(ctx, "SubmitSM processed successfully",
			slog.String("internal_msg_id", ack.InternalMessageID))
	}
}

// processSubmitMulti handles the logic for a parsed SubmitMulti PDU.
func (s *Server) processSubmitMulti(ctx context.Context, ss *sessionState, originalHdr PDUHeader, submitMulti *pdu.SubmitMulti) {
	sourceAddr := submitMulti.SourceAddr.Address()
	messageContentBytes, err := submitMulti.Message.GetMessage() // This is []byte
	if err != nil {
		slog.ErrorContext(ctx, "Failed to Get Message from SubmitMulti", slog.Any("error", err))
		s.writeErrorResponse(ss, originalHdr, StatusInvMsgLen)
		return
	}
	messageContentStr := string(messageContentBytes) // Convert to string for IncomingSPMessage
	registeredDelivery := submitMulti.RegisteredDelivery

	totalParts := 1
	currentPartNum := 1
	mref := 0
	if udhBytes := submitMulti.Message.UDH(); len(udhBytes) > 0 {
		udh := pdu.UDH(udhBytes)
		tpByte, cpByte, mrByte, found := udh.GetConcatInfo()
		if found {
			totalParts = int(tpByte)
			currentPartNum = int(cpByte)
			mref = int(mrByte)
		}
	}

	destAddressesInfo := submitMulti.DestAddrs.Get() // This is []DestinationAddress
	var unsuccessSMEs pdu.UnsuccessSMEs
	var firstSuccessMessageID string

	for _, destInfo := range destAddressesInfo {
		destAddrStr := destInfo.Address().Address()
		destAddrTON := destInfo.Address().Ton()
		destAddrNPI := destInfo.Address().Npi()

		inMsg := sp.IncomingSPMessage{
			ServiceProviderID: ss.serviceProviderID,
			CredentialID:      ss.credentialID,
			Protocol:          "smpp",
			// ClientMessageRef: // SubmitMulti doesn't typically have a per-destination client ref.
			SenderID:          sourceAddr,
			DestinationMSISDN: destAddrStr,
			MessageContent:    messageContentStr,
			TotalSegments:     int32(totalParts),
			SegmentSeqn:       int32(currentPartNum),
			ConcatRef:         int32(mref),
			IsFlash:           isFlashMessage(submitMulti.Message.Encoding().DataCoding()),
			RequestDLR:        (registeredDelivery & 0x01) != 0,
			ReceivedAt:        time.Now().UTC(),
			// DataCoding:     dataCoding,
			// ESMClass:       esmClass,
		}

		ack, coreErr := s.messageHandler.HandleIncomingMessage(ctx, inMsg)
		currentDestAddressForResp := pdu.NewAddress() // For UnsuccessSME
		currentDestAddressForResp.SetTon(destAddrTON)
		currentDestAddressForResp.SetNpi(destAddrNPI)
		currentDestAddressForResp.SetAddress(destAddrStr)

		if coreErr != nil {
			slog.ErrorContext(ctx, "Core handler failed for destination in SubmitMulti",
				slog.Any("error", coreErr), slog.String("destination", destAddrStr))
			unsuccessSME := pdu.UnsuccessSME{
				Address: currentDestAddressForResp,
			}
			unsuccessSME.SetErrorStatusCode(data.CommandStatusType(StatusSystemError))
			unsuccessSMEs.Add(unsuccessSME)

			continue
		}

		if ack.Status != "accepted" || ack.InternalMessageID == "" {
			slog.WarnContext(ctx, "Message rejected by core for destination in SubmitMulti",
				slog.String("reason", ack.Error), slog.String("destination", destAddrStr))

			unsuccessSME := pdu.UnsuccessSME{
				Address: currentDestAddressForResp,
			}
			unsuccessSME.SetErrorStatusCode(data.CommandStatusType(uint32(mapErrorToSMPPStatus(ack.Error))))
			unsuccessSMEs.Add(unsuccessSME)

		} else {
			if firstSuccessMessageID == "" {
				firstSuccessMessageID = ack.InternalMessageID
			}
			slog.InfoContext(ctx, "Message for destination in SubmitMulti accepted",
				slog.String("internal_msg_id", ack.InternalMessageID), slog.String("destination", destAddrStr))
		}
	}
	// --- Send SubmitMultiResp ---
	respPDU := submitMulti.GetResponse() // This is pdu.PDU
	submitMultiResp, castOk := respPDU.(*pdu.SubmitMultiResp)
	if !castOk {
		slog.ErrorContext(ctx, "Failed to cast GetResponse to *pdu.SubmitMultiResp",
			slog.Any("response_pdu_type", fmt.Sprintf("%T", respPDU)))
		s.writeErrorResponse(ss, originalHdr, StatusSystemError) // Generic error
		return
	}

	submitMultiResp.MessageID = firstSuccessMessageID
	if len(unsuccessSMEs.Get()) > 0 {
		submitMultiResp.UnsuccessSMEs = unsuccessSMEs
		if len(unsuccessSMEs.Get()) == len(destAddressesInfo) { // All failed
			submitMultiResp.MessageID = "" // As per spec, MessageID can be NULL if all fail.
			// Consider setting a specific CommandStatus on the response header if all fail.
			// submitMultiResp.Header().SetCommandStatus(data.ESME_RSUBMITFAIL) // Example
		}
	}
	// submitMultiResp.Header().SetCommandStatus(data.ESME_ROK) // gosmpp default is ROK. Adjust if needed.

	respBuf := pdu.NewBuffer(nil)
	submitMultiResp.Marshal(respBuf)
	marshaledRespPduBytes := respBuf.Bytes()

	if len(marshaledRespPduBytes) < 16 {
		slog.ErrorContext(ctx, "Marshaled SubmitMultiResp is too short", slog.Int("length", len(marshaledRespPduBytes)))
		return
	}

	responseHeaderToSend := PDUHeader{
		CommandID:      SubmitSMResp,
		SequenceNumber: originalHdr.SequenceNumber,
		Length:         uint32(len(marshaledRespPduBytes)),
		CommandStatus:  uint32(submitMultiResp.GetHeader().CommandStatus),
	}
	responseBodyBytesToSend := marshaledRespPduBytes[16:]

	if err := s.writePDU(ss, responseHeaderToSend, responseBodyBytesToSend); err != nil {
		slog.ErrorContext(ctx, "Failed to write SubmitMultiResp", slog.Any("error", err))
	} else {
		slog.InfoContext(ctx, "SubmitMulti processed",
			slog.String("batch_msg_id_sent_in_resp", submitMultiResp.MessageID),
			slog.Int("unsuccessful_smes", len(submitMultiResp.UnsuccessSMEs.Get())),
			slog.Uint64("sequence_number", uint64(originalHdr.SequenceNumber)))
	}
}

func isFlashMessage(dataCoding byte) bool {
	return (dataCoding & 0xF0) == 0x10
}

func mapErrorToSMPPStatus(err string) uint32 {
	switch err {
	case "throttled":
		return StatusThrottled
	case "invalid sender":
		return StatusInvSrcAddr
	case "invalid receiver":
		return StatusInvDstAddr
	default:
		return StatusSystemError
	}
}

// ForwardRawPDU implements RawPDUForwarder interface for sending DLRs etc.
func (s *Server) ForwardRawPDU(ctx context.Context, systemID string, pduBytes []byte) error {
	logCtx := logging.ContextWithSystemID(ctx, systemID)
	slog.DebugContext(logCtx, "Attempting to forward raw PDU", slog.Int("pdu_len", len(pduBytes)))

	s.sessionsMu.RLock() // Read lock to find session
	ss, ok := s.sessions[systemID]
	s.sessionsMu.RUnlock()

	if !ok || ss == nil {
		slog.WarnContext(logCtx, "Cannot forward PDU: Session not found")
		return fmt.Errorf("session not found for systemID: %s", systemID)
	}

	// Write using session's mutex
	ss.writeMu.Lock()
	defer ss.writeMu.Unlock()

	n, err := ss.writer.Write(pduBytes)
	if err == nil && n != len(pduBytes) {
		err = io.ErrShortWrite
	}
	// Flush immediately after writing DLR/forwarded PDU
	if err == nil {
		err = ss.writer.Flush()
	}

	if err != nil {
		slog.ErrorContext(logCtx, "Failed to write/flush raw PDU to client", slog.Any("error", err))
		// Consider removing session if write fails repeatedly?
		// s.removeSession(systemID)
		return fmt.Errorf("write/flush error for systemID %s: %w", systemID, err)
	}

	slog.InfoContext(logCtx, "Successfully forwarded raw PDU")
	return nil
}

// Shutdown gracefully stops the SMPP server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Shutdown requested for raw SMPP server...")
	// Signal the accept loop to stop
	close(s.shutdown)

	// Close the listener immediately
	if s.listener != nil {
		_ = s.listener.Close()
	}

	// Close all active client connections
	slog.InfoContext(ctx, "Closing active client connections...")
	s.sessionsMu.Lock() // Lock map for iteration
	sessionsToClose := make([]*sessionState, 0, len(s.sessions))
	for _, ss := range s.sessions {
		sessionsToClose = append(sessionsToClose, ss)
	}
	// Clear map immediately while holding write lock
	s.sessions = make(map[string]*sessionState)
	s.sessionsMu.Unlock()

	var closeWg sync.WaitGroup
	for _, ss := range sessionsToClose {
		closeWg.Add(1)
		go func(connToClose net.Conn, sysID string) {
			defer closeWg.Done()
			slog.DebugContext(ctx, "Closing connection", slog.String("system_id", sysID))
			_ = connToClose.Close()
		}(ss.conn, ss.systemID)
	}
	closeWg.Wait() // Wait for close calls to initiate
	slog.InfoContext(ctx, "All active client connections closed.")

	// Wait for handleSession goroutines to exit
	slog.InfoContext(ctx, "Waiting for connection handlers to finish...")
	s.wg.Wait()
	slog.InfoContext(ctx, "Raw SMPP Server shutdown complete.")
	return nil
}

// removeSession removes session (called when handler exits or duplicate bind)
func (s *Server) removeSession(systemID string) {
	if systemID == "" {
		return
	}
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	if _, exists := s.sessions[systemID]; exists {
		logCtx := logging.ContextWithSystemID(context.Background(), systemID)
		slog.InfoContext(logCtx, "Removing client session from map")
		delete(s.sessions, systemID)
	}
}
