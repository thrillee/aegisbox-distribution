package smppserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5" // For pgx.ErrNoRows
	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
)

// Compile-time check for ServerSessionManager implementation
var _ sp.ServerSessionManager = (*Server)(nil)

// sessionState holds information about a currently bound client connection.
type sessionState struct {
	credentialID      int32
	serviceProviderID int32
	systemID          string
	bindType          string
	boundAt           time.Time
	connection        *gosmpp.Connection
}

// Server implements the SMPP server logic.
type Server struct {
	config         config.ServerConfig
	dbQueries      database.Querier
	messageHandler sp.IncomingMessageHandler // Core logic handler
	gosmppServer   *gosmpp.Server
	sessions       sync.Map // map[string]*sessionState - Key: SystemID
	stopOnce       sync.Once
	wg             sync.WaitGroup
}

// NewServer creates a new SMPP server instance.
func NewServer(cfg config.ServerConfig, q database.Querier, handler sp.IncomingMessageHandler) *Server {
	if handler == nil {
		panic("Message handler cannot be nil for SMPP Server")
	}
	return &Server{
		config:         cfg,
		dbQueries:      q,
		messageHandler: handler,
	}
}

// ListenAndServe starts the SMPP server and blocks until shutdown.
func (s *Server) ListenAndServe() error {
	if s.gosmppServer != nil {
		return errors.New("server already started")
	}

	bindAuthHandler := func(ctx context.Context, systemID, password string, conn gosmpp.Connection) (pdu.Status, error) {
		logCtx := logging.ContextWithSystemID(ctx, systemID)
		slog.InfoContext(logCtx, "Received bind request")

		// Use a timeout for the auth DB lookup
		authCtx, cancel := context.WithTimeout(logCtx, 5*time.Second)
		defer cancel()

		cred, err := s.dbQueries.GetSPCredentialBySystemID(authCtx, &systemID) // Use updated query
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.WarnContext(logCtx, "Bind failed: SystemID not found")
				return pdu.StatusInvSysID, nil // Invalid System ID
			}
			slog.ErrorContext(logCtx, "Bind auth DB error", slog.Any("error", err))
			return pdu.StatusSystemError, err // Indicate temporary failure
		}

		// Check if it's an SMPP credential
		if cred.Protocol != "smpp" {
			slog.WarnContext(logCtx, "Bind failed: Credential found but not for SMPP protocol", slog.String("protocol", cred.Protocol))
			return pdu.StatusBindFailed, nil
		}

		// Check password hash
		if cred.PasswordHash != nil || !auth.CheckPasswordHash(password, cred.PasswordHash) {
			slog.WarnContext(logCtx, "Bind failed: Invalid password")
			return pdu.StatusInvPasswd, nil // Invalid Password
		}

		// --- Authentication Successful ---

		// Store session state - This needs access to the underlying writer.
		// gosmpp.Connection likely provides access.
		session := &sessionState{
			credentialID:      cred.ID,
			serviceProviderID: cred.ServiceProviderID,
			systemID:          systemID,
			connection:        &conn, // Store the connection itself as the writer
			bindType:          *cred.BindType,
			boundAt:           time.Now(),
		}
		s.sessions.Store(systemID, session) // Store session keyed by SystemID

		slog.InfoContext(logCtx, "Bind successful", slog.Int("sp_id", int(session.serviceProviderID)))

		// Associate SystemID with context logger for subsequent PDU handlers for this conn? Complex.

		return pdu.StatusOk, nil
	}

	// Define handler for SubmitSM within the generic request handler
	genericHandler := gosmpp.DefaultGenericRequestHandler{
		HandlerFunc: func(ctx context.Context, req pdu.PDU, resp pdu.Responsable, conn gosmpp.Connection) (pdu.Status, error) {
			logCtx := ctx                   // Start with base context
			systemID, ok := conn.SystemID() // Attempt to get bound system ID
			if ok {
				logCtx = logging.ContextWithSystemID(logCtx, systemID)
			}
			logCtx = logging.ContextWithPDUInfo(logCtx, req.CommandID().String(), req.GetSequenceNumber())

			switch p := req.(type) {
			case *pdu.SubmitSM:
				return s.handleSubmitSM(logCtx, p, resp, systemID) // Pass systemID if available
			case *pdu.EnquireLink:
				slog.DebugContext(logCtx, "Received EnquireLink")
				return pdu.StatusOk, nil // Default handler sends response
			case *pdu.Unbind:
				slog.InfoContext(logCtx, "Received Unbind request")
				// Remove session on unbind request
				if sysID, ok := conn.SystemID(); ok {
					slog.InfoContext(logCtx, "Closing session due to unbind request")
					s.removeSession(sysID) // Clean up session map
				}
				// Default handler sends UnbindResp and closes connection
				return pdu.StatusOk, nil
			// Add other PDU types if needed
			default:
				slog.DebugContext(logCtx, "Passing PDU to default handler")
				// Use default for others (like DataSM, responses to server requests if any)
				return gosmpp.DefaultPDUHandlerFunc(ctx, req, resp, conn)
			}
		},
	}.HandlerFunc // Get the handler function

	s.gosmppServer = &gosmpp.Server{
		Addr:           s.config.Addr,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		BindTimeout:    s.config.BindTimeout,
		EnquireLink:    s.config.EnquireLink,
		MaxConnections: s.config.MaxConnections,
		TLS:            nil, // Add TLS config later if needed

		// Use BindAuthController for more control over bind and session storage
		BindAuthController: bindAuthHandler,

		// Use our customized generic handler
		GenericRequestHandler: genericHandler,

		OnClosed: func(conn gosmpp.Connection, state gosmpp.State) {
			systemID, ok := conn.SystemID()
			if ok {
				logCtx := logging.ContextWithSystemID(context.Background(), systemID)
				slog.InfoContext(logCtx, "SMPP client connection closed", slog.String("state", state.String()))
				s.removeSession(systemID) // Clean up session on close
			} else {
				slog.InfoContext(context.Background(), "Unauthenticated SMPP client connection closed", slog.String("state", state.String()), slog.String("remote_addr", conn.RemoteAddr().String()))
			}
		},
		// Add other event handlers like OnError if needed
	}

	slog.Info("Starting SMPP Server", slog.String("address", s.config.Addr))
	err := s.gosmppServer.ListenAndServe() // This blocks
	if err != nil && !errors.Is(err, gosmpp.ErrServerClosed) {
		slog.Error("SMPP server ListenAndServe error", slog.Any("error", err))
		return err
	}
	slog.Info("SMPP Server stopped.")
	return nil
}

// handleSubmitSM handles incoming SubmitSM requests.
func (s *Server) handleSubmitSM(ctx context.Context, p *pdu.SubmitSM, resp pdu.Responsable, systemID string) (pdu.Status, error) {
	slog.InfoContext(ctx, "Received SubmitSM request")

	// 1. Get Session State (Authentication happened during bind)
	session, ok := s.getSession(systemID)
	if !ok {
		slog.WarnContext(ctx, "SubmitSM rejected: Client session not found or not bound", slog.String("system_id", systemID))
		// Can't proceed without valid session
		// ESME_RINVBNDSTS (Invalid Bind Status) might be appropriate
		return pdu.StatusBindFailed, errors.New("client session not bound")
	}
	logCtx := logging.ContextWithSPID(ctx, session.serviceProviderID) // Add SP ID

	// 2. Extract Data & Prepare Internal Message
	// TODO: Detect incoming UDH and handle reassembly if required before calling handler.
	// This example assumes single part messages or pre-reassembled content.
	shortMessage, err := p.Message.GetMessage() // Get message content
	if err != nil {
		slog.WarnContext(logCtx, "Failed to get message content from SubmitSM", slog.Any("error", err))
		return pdu.StatusInvMsgLen, err // Invalid message length? Or system error?
	}

	// Get client message reference if provided (e.g., message_payload TLV or other means)
	clientRef := ""
	// Example: Check for message_payload TLV (0x0424)
	if tlv := p.GetTLV(pdutlv.TagMessagePayload); tlv != nil {
		// Use payload as message content? Check spec/use case.
		// Or maybe another TLV for client ref? For now, leave clientRef empty.
	}

	incomingMsg := sp.IncomingSPMessage{
		ServiceProviderID: session.serviceProviderID,
		CredentialID:      session.credentialID,
		Protocol:          "smpp",
		ClientMessageRef:  clientRef, // Populate if client sends a reference
		SenderID:          p.SourceAddr.Address(),
		DestinationMSISDN: p.DestAddr.Address(),
		MessageContent:    shortMessage,
		TotalSegments:     1, // Assume 1 unless UDH detected and handled
		SegmentSeqn:       1,
		IsFlash:           false, // TODO: Detect flash indication (e.g., DataCoding)
		RequestDLR:        p.RegisteredDelivery == 1,
		ReceivedAt:        time.Now(),
	}

	// 3. Call Core Message Handler
	ack, err := s.messageHandler.HandleIncomingMessage(logCtx, incomingMsg)
	if err != nil {
		// Core handler failed (e.g., DB insert error)
		slog.ErrorContext(logCtx, "Incoming message handler failed", slog.Any("error", err))
		// Map internal error to SMPP status code?
		return pdu.StatusSystemError, err // Return temporary system error
	}

	// 4. Process Acknowledgement
	if ack.Status != "accepted" || !ack.InternalMessageID.Valid {
		slog.WarnContext(logCtx, "Message rejected by core handler", slog.String("reason", ack.Error))
		// Map rejection reason to SMPP status code?
		// Example mapping:
		status := pdu.StatusMessageQFull // Default rejection
		if strings.Contains(ack.Error, "invalid sender") {
			status = pdu.StatusInvSrcAddr
		}
		if strings.Contains(ack.Error, "insufficient funds") {
			status = pdu.StatusEsmeRThrottled
		} // Abuse slightly
		// ... more mappings ...
		return status, errors.New(ack.Error)
	}

	// 5. Send SubmitSMResp (Success)
	// Set the message ID in the response PDU - SMPP expects string
	resp.SetField(pdufield.MessageID, fmt.Sprintf("%d", ack.InternalMessageID))
	slog.InfoContext(logCtx, "Message accepted by core handler", slog.Int64("internal_msg_id", ack.InternalMessageID))

	return pdu.StatusOk, nil // Success
}

// getSession retrieves the session state for a given systemID.
func (s *Server) getSession(systemID string) (*sessionState, bool) {
	val, ok := s.sessions.Load(systemID)
	if !ok {
		return nil, false
	}
	session, ok := val.(*sessionState)
	return session, ok
}

// removeSession removes a session from the map.
func (s *Server) removeSession(systemID string) {
	if _, ok := s.sessions.LoadAndDelete(systemID); ok {
		logCtx := logging.ContextWithSystemID(context.Background(), systemID)
		slog.InfoContext(logCtx, "Removed client session from map")
	}
}

func (s *Server) GetBoundClientConnection(systemID string) (*gosmpp.Connection, error) {
	session, ok := s.getSession(systemID) // getSession remains the same
	if !ok {
		return nil, fmt.Errorf("no active session found for systemID: %s", systemID)
	}
	if session.connection == nil {
		s.removeSession(systemID) // Clean up broken session state
		return nil, fmt.Errorf("session found for systemID %s but connection is nil", systemID)
	}

	// TODO: Check if connection is still valid? gosmpp might handle this internally.
	// If not, remove session and return error.

	return session.connection, nil
}

// Shutdown gracefully stops the SMPP server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Shutdown requested for SMPP server...")
	var err error
	s.stopOnce.Do(func() { // Ensure shutdown logic runs only once
		if s.gosmppServer != nil {
			err = s.gosmppServer.Close() // Close the listener, should unblock ListenAndServe
			s.gosmppServer = nil
		}
		// Close all active client connections? gosmppServer.Close() should trigger OnClosed callbacks.
		slog.InfoContext(ctx, "SMPP Server Close() called.")
	})
	// Wait for active requests? gosmpp doesn't expose http.Server style shutdown yet.
	return err
}
