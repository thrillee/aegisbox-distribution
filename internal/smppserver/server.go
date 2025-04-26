package smppserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/thrillee/aegisbox/internal/auth"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
)

// Compile-time check for ServerSessionManager implementation
var _ sp.ServerSessionManager = (*Server)(nil)

// sessionState holds information about a currently bound client connection.
type sessionState1 struct {
	credentialID      int32
	serviceProviderID int32
	systemID          string
	bindType          string
	boundAt           time.Time
	connection        *gosmpp.Connection
}

// Server implements the SMPP server logic.
type Server1 struct {
	config         config.ServerConfig
	dbQueries      database.Querier
	messageHandler sp.IncomingMessageHandler // Core logic handler
	gosmppServer   net.Listener
	sessions       sync.Map // map[string]*sessionState - Key: SystemID
	stopOnce       sync.Once
	wg             sync.WaitGroup
	ctx            context.Context
}

// NewServer creates a new SMPP server instance.
func NewServer1(cfg config.ServerConfig, q database.Querier, handler sp.IncomingMessageHandler) *Server {
	if handler == nil {
		panic("Message handler cannot be nil for SMPP Server")
	}
	return &Server{
		config:         cfg,
		dbQueries:      q,
		messageHandler: handler,
	}
}

// Implement SMPP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.gosmppServer = ln
	s.wg.Add(1)
	go s.serve()
	return nil
}

func (s *Server) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.gosmppServer.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				slog.Error("Accept failed", "error", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	// Create gosmpp connection with settings
	connection := gosmpp.NewConnection(
		conn,
	)
	defer connection.Close()

	for {
		pdu, err := s.connection.Read()
		if err != nil {
			return
		}
		switch pdu.Header().CommandID {
		case data.BIND_TRANSCEIVER:
			s.handleBind(s, pdu)
		case *pdu.SubmitSM:
			s.handleSubmitSM(ss, pdu)
		case pdu.BindingType:
			s.handleUnbind2(ss, pdu)
		case gosmpp.CmdEnquireLink:
			// respond immediately
			s.connection.Write(pdu.Resp())
		default:
			ss.connection.Write(pdu.Resp().SetCommandStatus(gosmpp.ESME_RINVCMDID))
		}
	}

	session := &sessionState{connection: connection}
	ctx := context.Background()

	for {
		p, err := connection.ReadPDU()
		if err != nil {
			slog.InfoContext(ctx, "Connection closed", "error", err)
			s.removeSession(session)
			return
		}

		switch pd := p.(type) {
		case *pdu.BindTransmitter:
			s.handleBindTransmitter(ctx, pd, session)
		case *pdu.BindReceiver:
			s.handleBindReceiver(ctx, pd, session)
		case *pdu.BindTransceiver:
			s.handleBindTransceiver(ctx, pd, session)
		case *pdu.SubmitSM:
			s.handleSubmitSM(ctx, pd, session)
		case *pdu.EnquireLink:
			s.handleEnquireLink(ctx, pd, session)
		case *pdu.Unbind:
			s.handleUnbind(ctx, pd, session)
		default:
			slog.WarnContext(ctx, "Unhandled PDU type", "type", p.GetHeader().ID)
		}
	}
}

// Separate handlers for each bind type
func (s *Server) handleBindTransmitter(ctx context.Context, p *pdu.BindTransmitter, session *sessionState) {
	s.handleBind(ctx, p, session, "transmitter")
}

func (s *Server) handleBindReceiver(ctx context.Context, p *pdu.BindReceiver, session *sessionState) {
	s.handleBind(ctx, p, session, "receiver")
}

func (s *Server) handleBindTransceiver(ctx context.Context, p *pdu.BindTransceiver, session *sessionState) {
	s.handleBind(ctx, p, session, "transceiver")
}

// Unified bind handler
func (s *Server) handleBind2(ctx context.Context, bindPdu interface{}, session *sessionState, bindType string) {
	var systemID, password string

	switch p := bindPdu.(type) {
	case *data.BindTransmitter:
		systemID = p.SystemID
		password = p.Password
	case *pdu.BindReceiver:
		systemID = p.SystemID
		password = p.Password
	case *pdu.BindTransceiver:
		systemID = p.SystemID
		password = p.Password
	default:
		return
	}

	logCtx := logging.ContextWithSystemID(ctx, systemID)
	slog.InfoContext(logCtx, "Bind request received")

	cred, err := s.dbQueries.GetSPCredentialBySystemID(ctx, &systemID)
	if err != nil {
		slog.ErrorContext(logCtx, "Database error", "error", err)
		s.sendBindResponse(session, data.ESME_RBINDFAIL, systemID, bindType)
		return
	}

	if cred.Protocol != "smpp" {
		slog.WarnContext(logCtx, "Invalid protocol", "expected", "smpp", "got", cred.Protocol)
		s.sendBindResponse(session, data.ESME_RINVSYSID, systemID, bindType)
		return
	}

	if !auth.CheckPasswordHash(password, *cred.PasswordHash) {
		slog.WarnContext(logCtx, "Invalid password")
		s.sendBindResponse(session, data.ESME_RINVPASWD, systemID, bindType)
		return
	}

	session.systemID = systemID
	session.serviceProviderID = cred.ServiceProviderID
	session.credentialID = cred.ID
	session.bindType = bindType
	session.boundAt = time.Now()

	s.sessions.Store(systemID, session)
	slog.InfoContext(logCtx, "Bind successful", "sp_id", cred.ServiceProviderID)

	s.sendBindResponse(session, data.ESME_ROK, systemID, bindType)
}

func (s *Server) handleSubmitSM(ss *sessionState, pdu *pdu.PDU) {
	body := pdu.Body()
	dest := body.DestinationAddr()
	src := body.SourceAddr()
	ext := body.ShortMessageBytes()
	esm := body.ESMClass()
	dc := body.DataCoding()
	dlr := body.RegisteredDelivery()

	msg := sp.IncomingSPMessage{
		ServiceProviderID: ss.serviceProviderID,
		CredentialID:      ss.credentialID,
		Protocol:          "smpp",
		SenderID:          src.Address,
		DestinationMSISDN: dest.Address,
		MessageContent:    string(ext),
		TotalSegments:     1,
		SegmentSeqn:       1,
		ConcatRef:         0,
		IsFlash:           esm&0x10 != 0,
		RequestDLR:        dlr.Flags() != 0,
		ReceivedAt:        time.Now(),
	}

	// Delegate to core logic
	ack, err := s.messageHandler.HandleIncomingMessage(s.ctx, msg)
	resp := pdu.Resp()
	if err != nil || ack.Status != "accepted" {
		reason := data.ESME_RSUBMITFAIL
		if err != nil {
			// map custom errors here
		}
		resp.SetCommandStatus(reason)
		ss.connection.Write(resp)
		return
	}

	// success: set message ID
	resp.Body().SetMessageID(fmt.Sprintf("%d", ack.InternalMessageID))
	ss.connection.Write(resp)

	// schedule DLR if requested
	if msg.RequestDLR {
		go s.generateReceipt(ss, ack.InternalMessageID, msg)
	}
}

func (s *Server) sendBindResponse(session *sessionState, status pdu.CommandStatus, systemID, bindType string) {
	var resp pdu.Respondable

	switch bindType {
	case "transmitter":
		resp = pdu.NewBindTransmitterResp()
	case "receiver":
		resp = pdu.NewBindReceiverResp()
	case "transceiver":
		resp = pdu.NewBindTransceiverResp()
	default:
		return
	}

	resp.SetSequenceNumber(session.connection.SequenceNumber())
	resp.SetStatus(status)
	if r, ok := resp.(interface{ SetSystemID(string) }); ok {
		r.SetSystemID(systemID)
	}

	if _, err := session.connection.WritePDU(resp); err != nil {
		slog.Error("Failed to send bind response", "error", err)
	}
}

// SubmitSM handler
func (s *Server) handleSubmitSM(ctx context.Context, p *pdu.SubmitSM, session *sessionState) {
	logCtx := logging.ContextWithSystemID(ctx, session.systemID)
	logCtx = logging.ContextWithSPID(logCtx, session.serviceProviderID)

	// Create response
	resp := pdu.NewSubmitSMResp()
	resp.SetSequenceNumber(p.GetSequenceNumber())
	// resp.MessageID = generateMessageID() // Implement your message ID generation

	if _, err := session.connection.WritePDU(resp); err != nil {
		slog.ErrorContext(logCtx, "Failed to send SubmitSM response", "error", err)
		return
	}

	// Process message
	message := sp.IncomingMessage{
		SourceAddr:      p.SourceAddr.Address(),
		DestinationAddr: p.DestAddr.Address(),
		Text:            p.Message.GetMessage(),
		ClientRef:       p.EsmClass,
		// ... map other fields
	}

	if err := s.messageHandler.HandleIncomingMessage(logCtx, message); err != nil {
		slog.ErrorContext(logCtx, "Failed to process message", "error", err)
	}
}

// Implement DLR forwarding
func (f *SMPSPForwarder) ForwardDLR(ctx context.Context, spDetails SPDetails, dlr ForwardedDLRInfo) error {
	conn, err := f.sessionMgr.GetBoundClientConnection(spDetails.SMPPSystemID)
	if err != nil {
		return fmt.Errorf("no active connection for systemID %s: %w", spDetails.SMPPSystemID, err)
	}

	dlrPDU := buildDLRPDU(dlr)
	if err := conn.WritePDU(dlrPDU); err != nil {
		return fmt.Errorf("failed to send DLR: %w", err)
	}

	return nil
}

func buildDLRPDU(dlr ForwardedDLRInfo) pdu.PDU {
	d := pdu.NewDeliverSM()

	// Source address (original sender)
	srcAddr := pdu.NewAddress()
	srcAddr.SetAddress(dlr.DestAddr)
	d.SourceAddr = srcAddr

	// Destination address (original receiver)
	destAddr := pdu.NewAddress()
	destAddr.SetAddress(dlr.SourceAddr)
	d.DestAddr = destAddr

	// Message text - format according to SMPP spec
	message := buildDLRMessage(dlr)
	_ = d.Message.SetMessage(message)

	// Set DLR specific parameters
	d.RegisteredDelivery = 1
	d.ReceiptedMessageID = dlr.ClientMessageRef
	d.MessageState = dlrStatusToMessageState(dlr.Status)
	d.ErrorCode = dlrErrorCode(dlr.ErrorCode)

	return d
}

func buildDLRMessage(dlr ForwardedDLRInfo) string {
	// Format according to SMPP Appendix B
	return fmt.Sprintf("id:%s submit date:%s done date:%s stat:%s err:%s",
		dlr.ClientMessageRef,
		dlr.SubmitDate.Format("0601021504"),
		dlr.DoneDate.Format("0601021504"),
		dlr.Status,
		dlr.ErrorCode,
	)
}

// Helper functions for DLR conversion
func dlrStatusToMessageState(status string) pdu.MessageState {
	switch status {
	case "DELIVRD":
		return data.DELI
	case "UNDELIV":
		return pdu.UNDELIVERABLE
	case "REJECTD":
		return pdu.REJECTED
	default:
		return pdu.UNKNOWN
	}
}

func (s *Server) generateReceipt(ss *sessionState, internalID int64, msg IncomingSPMessage) {
	delay := time.Duration(rand.Intn(5)+1) * time.Second
	time.Sleep(delay)

	// build DeliverSM receipt
	dlr := pdu.NewDeliverSM()
	dlr.SetSourceAddr(msg.DestinationMSISDN)
	dlr.Body().SetDestinationAddr(msg.SenderID)
	text := fmt.Sprintf("id:%d sub:001 dlvrd:001 submit date:%s done date:%s stat:DELIVRD err:000 text:%s",
		internalID,
		msg.ReceivedAt.Format("0601021504"),
		time.Now().Format("0601021504"),
		msg.MessageContent,
	)
	dlr.Body().SetShortMessage([]byte(text))

	ss.connection.Write(dlr)
}

func dlrErrorCode(err *string) byte {
	if err == nil {
		return 0
	}
	// Implement your error code mapping logic
	return 0
}

// Session management
func (s *Server) GetBoundClientConnection(systemID string) (*gosmpp.Connection, error) {
	if sess, ok := s.sessions.Load(systemID); ok {
		return sess.(*sessionState).connection, nil
	}
	return nil, fmt.Errorf("no active session for %s", systemID)
}

func (s *Server) removeSession(sess *sessionState) {
	if sess.systemID != "" {
		s.sessions.Delete(sess.systemID)
	}
}

// Handle other PDU types
func (s *Server) handleEnquireLink(ctx context.Context, p *pdu.EnquireLink, session *sessionState) {
	resp := pdu.NewEnquireLinkResp()
	resp.SetSequenceNumber(p.SequenceNumber)
	_, err := session.connection.WritePDU(resp)
	if err != nil {
		slog.ErrorContext(ctx, "Enquire Link Failed", "error", err)
	}
}

func (s *Server) handleUnbind(ctx context.Context, p *pdu.Unbind, session *sessionState) {
	resp := pdu.NewUnbindResp()
	resp.SetSequenceNumber(p.SequenceNumber)
	_, err := session.connection.WritePDU(resp)
	s.removeSession(session)

	if err != nil {
		slog.ErrorContext(ctx, "Unbind Failed", "error", err)
	}
}

// Shutdown implementation
func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		if s.gosmppServer != nil {
			_ = s.gosmppServer.Close()
		}
		s.wg.Wait()
	})
}
