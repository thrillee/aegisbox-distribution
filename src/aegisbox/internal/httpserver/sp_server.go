package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/pkg/codes"
)

// Config holds HTTP Server specific configuration.

// Server implements the HTTP server for SPs.
type Server struct {
	config         config.HttpConfig
	dbQueries      database.Querier
	messageHandler sp.IncomingMessageHandler
	httpServer     *http.Server
	stopOnce       sync.Once
}

// NewServer creates a new HTTP server instance.
func NewServer(cfg config.HttpConfig, q database.Querier, handler sp.IncomingMessageHandler) *Server {
	if handler == nil {
		panic("Message handler cannot be nil for HTTP Server")
	}
	return &Server{
		config:         cfg,
		dbQueries:      q,
		messageHandler: handler,
	}
}

// ListenAndServe starts the HTTP server and blocks until shutdown.
func (s *Server) ListenAndServe() error {
	if s.httpServer != nil {
		return errors.New("http server already started")
	}

	mux := http.NewServeMux()
	// Apply Recovery middleware THEN Auth middleware
	recoveryHandler := s.recoveryMiddleware(s.handleIncomingSMS)
	authHandler := s.authMiddleware(recoveryHandler)
	mux.HandleFunc("POST /sms/incoming", authHandler)

	s.httpServer = &http.Server{
		Addr:         s.config.Addr,
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn), // Use slog for server errors
	}

	slog.Info("Starting SP HTTP Server", slog.String("address", s.config.Addr))
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("SP HTTP server ListenAndServe error", slog.Any("error", err))
		return err
	}
	slog.Info("SP HTTP Server stopped.")
	return nil
}

// --- Recovery Middleware ---

func (s *Server) recoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(r.Context(), "Panic recovered in HTTP handler",
					slog.Any("panic", rec),
					slog.String("stack", fmt.Sprintf("%+v", rec)))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
}

// --- Authentication Middleware ---

type contextKey string

const spAuthInfoKey contextKey = "spAuthInfo"

type spAuthInfo struct {
	ServiceProviderID int32
	CredentialID      int32
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logCtx := r.Context()
		apiKeyIdentifier := ""

		// 1. Extract Key (Example: Bearer <identifier>_<secret>)
		authHeader := r.Header.Get("Authorization")
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			apiKeyIdentifier = after
		}

		if apiKeyIdentifier == "" {
			slog.WarnContext(logCtx, "HTTP Auth failed: Missing or invalid API Key format (expected Identifier_Secret)")
			http.Error(w, "Unauthorized: Invalid API Key format", http.StatusUnauthorized)
			return
		}
		// Add identifier to context early for logging
		logCtx = logging.ContextWithAPIKeyIdentifier(logCtx, apiKeyIdentifier)

		// 2. Query DB using Identifier
		authCtx, cancel := context.WithTimeout(logCtx, 3*time.Second)
		defer cancel()

		slog.DebugContext(logCtx, "Querying database for API key")
		cred, err := s.dbQueries.GetSPCredentialByKeyIdentifier(authCtx, &apiKeyIdentifier)
		slog.DebugContext(logCtx, "Database query completed", slog.Any("error", err))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.WarnContext(logCtx, "HTTP Auth failed: Invalid API Key Identifier")
				http.Error(w, "Unauthorized: Invalid API Key", http.StatusUnauthorized) // Keep error generic
			} else {
				slog.ErrorContext(logCtx, "HTTP Auth DB error fetching credential", slog.Any("error", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// --- Authentication Successful ---
		authData := spAuthInfo{
			ServiceProviderID: cred.ServiceProviderID,
			CredentialID:      cred.ID,
		}
		reqCtx := context.WithValue(r.Context(), spAuthInfoKey, authData)
		reqCtx = logging.ContextWithSPID(reqCtx, authData.ServiceProviderID)
		reqCtx = logging.ContextWithCredentialID(reqCtx, authData.CredentialID)

		slog.InfoContext(reqCtx, "HTTP authentication successful")
		next.ServeHTTP(w, r.WithContext(reqCtx))
	}
}

// --- HTTP Handler ---

// Define the expected JSON request body structure
type submitHTTPRequest struct {
	From       string  `json:"from"`        // Sender ID
	To         string  `json:"to"`          // Destination MSISDN
	Text       string  `json:"text"`        // Message Content
	ClientRef  *string `json:"client_ref"`  // Optional client reference
	RequestDLR *bool   `json:"request_dlr"` // Optional DLR request flag
	IsFlash    *bool   `json:"is_flash"`    // Optional flash flag
	// Add other fields if needed (e.g., encoding, validity_period)
}

// handleIncomingSMS handles POST requests to /sms/incoming
func (s *Server) handleIncomingSMS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Get context potentially enriched by middleware

	// Retrieve authentication info from context
	authData, ok := ctx.Value(spAuthInfoKey).(spAuthInfo)
	if !ok {
		// Should not happen if auth middleware ran correctly
		slog.ErrorContext(ctx, "Authentication info missing from context in handler")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	logCtx := ctx // Use the context which already has SPID etc.

	slog.InfoContext(logCtx, "Received incoming HTTP SMS request")

	// 1. Decode JSON Request Body
	var reqPayload submitHTTPRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Prevent unexpected fields
	if err := decoder.Decode(&reqPayload); err != nil {
		slog.WarnContext(logCtx, "Failed to decode incoming HTTP request body", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Bad Request: Invalid JSON format - %v", err), http.StatusBadRequest)
		return
	}

	// Basic validation
	if reqPayload.From == "" || reqPayload.To == "" || reqPayload.Text == "" {
		slog.WarnContext(logCtx, "HTTP request validation failed: Missing required fields")
		http.Error(w, "Bad Request: Missing required fields (from, to, text)", http.StatusBadRequest)
		return
	}

	// 2. Prepare Internal Message
	// Assume single segment for HTTP submissions for now
	var clientRef string
	if reqPayload.ClientRef != nil {
		clientRef = *reqPayload.ClientRef
	}
	requestDLR := true // Default to true? Or false? Let's default true.
	if reqPayload.RequestDLR != nil {
		requestDLR = *reqPayload.RequestDLR
	}
	isFlash := false
	if reqPayload.IsFlash != nil {
		isFlash = *reqPayload.IsFlash
	}

	incomingMsg := sp.IncomingSPMessage{
		ServiceProviderID: authData.ServiceProviderID,
		CredentialID:      authData.CredentialID,
		Protocol:          "http",
		ClientMessageRef:  clientRef,
		SenderID:          reqPayload.From,
		DestinationMSISDN: reqPayload.To,
		MessageContent:    reqPayload.Text,
		TotalSegments:     1, // Assume 1 for HTTP
		SegmentSeqn:       1,
		IsFlash:           isFlash,
		RequestDLR:        requestDLR,
		ReceivedAt:        time.Now(),
		// DataCoding/ESMClass not typically provided via simple HTTP API
	}

	// 3. Call Core Message Handler
	slog.InfoContext(logCtx, "Calling message handler")
	ack, err := s.messageHandler.HandleIncomingMessage(logCtx, incomingMsg)
	slog.InfoContext(logCtx, "Message handler returned", slog.Any("error", err), slog.Any("ack_status", ack.Status))
	if err != nil {
		slog.ErrorContext(logCtx, "Incoming message handler failed", slog.Any("error", err))
		http.Error(w, "Internal Server Error processing message", http.StatusInternalServerError)
		return
	}

	// 4. Send HTTP Response based on Acknowledgement
	slog.InfoContext(logCtx, "About to write HTTP response")
	w.Header().Set("Content-Type", "application/json")
	slog.InfoContext(logCtx, "Header set, about to write status")
	if ack.Status == "accepted" && ack.InternalMessageID != "" {
		slog.InfoContext(logCtx, "Writing accepted response", slog.String("msg_id", ack.InternalMessageID))
		w.WriteHeader(http.StatusAccepted) // 202 Accepted is suitable
		respBody := map[string]interface{}{
			"status":     "accepted",
			"message_id": fmt.Sprintf("%s", ack.InternalMessageID), // Return our internal ID as string
		}
		slog.InfoContext(logCtx, "About to encode response body")
		if err := json.NewEncoder(w).Encode(respBody); err != nil {
			slog.ErrorContext(logCtx, "Failed to encode HTTP success response", slog.Any("error", err))
		}
		slog.InfoContext(logCtx, "HTTP message accepted", slog.String("internal_msg_id", ack.InternalMessageID))
	} else {
		// Determine appropriate HTTP status code for rejection
		statusCode := http.StatusBadRequest // Default for rejection
		if strings.Contains(ack.Error, string(codes.ErrorCodeInsufficientFunds)) {
			statusCode = http.StatusPaymentRequired
		}
		if strings.Contains(ack.Error, string(codes.ErrorCodeSystemError)) {
			statusCode = http.StatusInternalServerError
		}

		w.WriteHeader(statusCode)
		respBody := map[string]interface{}{
			"status": "rejected",
			"error":  ack.Error,
		}
		if err := json.NewEncoder(w).Encode(respBody); err != nil {
			slog.ErrorContext(logCtx, "Failed to encode HTTP rejection response", slog.Any("error", err))
		}
		slog.WarnContext(logCtx, "HTTP message rejected", slog.String("reason", ack.Error))
	}
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Shutdown requested for SP HTTP server...")
	var err error
	s.stopOnce.Do(func() {
		if s.httpServer != nil {
			// Disable keep-alives before shutting down
			s.httpServer.SetKeepAlivesEnabled(false)
			err = s.httpServer.Shutdown(ctx)
			s.httpServer = nil
		}
		slog.InfoContext(ctx, "SP HTTP Server Shutdown() called.")
	})
	return err
}
