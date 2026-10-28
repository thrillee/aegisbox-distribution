package mno

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// DLRServer implements the HTTP server for handling DLR callbacks from MNO HTTP APIs.
type DLRServer struct {
	config        DLRServerConfig
	handleDLRFunc DLRHandlerFunc
	httpServer    *http.Server
	stopOnce      sync.Once
}

// DLRServerConfig holds configuration for the DLR callback server
type DLRServerConfig struct {
	Addr         string        // Server address (e.g., ":8080")
	ReadTimeout  time.Duration // HTTP read timeout
	WriteTimeout time.Duration // HTTP write timeout
	IdleTimeout  time.Duration // HTTP idle timeout
	APIKey       string        // Optional: API key for authenticating incoming requests
}

// NewDLRServer creates a new DLR callback server instance.
func NewDLRServer(cfg DLRServerConfig, handleDLRFunc DLRHandlerFunc) *DLRServer {
	if handleDLRFunc == nil {
		panic("ConnectorManager cannot be nil for DLR Server")
	}

	// Set default timeouts if not provided
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 60 * time.Second
	}

	return &DLRServer{
		config:        cfg,
		handleDLRFunc: handleDLRFunc,
	}
}

// ListenAndServe starts the DLR callback server and blocks until shutdown.
func (s *DLRServer) ListenAndServe() error {
	if s.httpServer != nil {
		return errors.New("DLR server already started")
	}

	mux := http.NewServeMux()

	// Apply auth middleware if API key is configured
	var dlrHandler http.HandlerFunc = s.handleDLRCallback
	if s.config.APIKey != "" {
		dlrHandler = s.authMiddleware(s.handleDLRCallback)
	}

	// Register DLR callback endpoints
	mux.HandleFunc("POST /aegis-dlr/callback", dlrHandler)
	mux.HandleFunc("POST /aegis-dlr/callback/", dlrHandler) // Handle trailing slash

	// Health check endpoint
	mux.HandleFunc("GET /health", s.handleHealthCheck)

	s.httpServer = &http.Server{
		Addr:         s.config.Addr,
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
	}

	slog.Info("Starting DLR Callback Server", slog.String("address", s.config.Addr))
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("DLR server ListenAndServe error", slog.Any("error", err))
		return err
	}
	slog.Info("DLR Callback Server stopped.")
	return nil
}

// Shutdown gracefully shuts down the DLR callback server.
func (s *DLRServer) Shutdown(ctx context.Context) error {
	var err error
	s.stopOnce.Do(func() {
		if s.httpServer != nil {
			slog.Info("Shutting down DLR Callback Server...")
			err = s.httpServer.Shutdown(ctx)
		}
	})
	return err
}

// authMiddleware provides API key authentication for incoming requests
func (s *DLRServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" {
			apiKey = r.Header.Get("Authorization")
			if apiKey != "" && len(apiKey) > 7 && apiKey[:7] == "Bearer " {
				apiKey = apiKey[7:]
			}
		}

		if apiKey != s.config.APIKey {
			slog.Warn("Unauthorized DLR callback attempt",
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// handleDLRCallback processes incoming DLR callbacks
func (s *DLRServer) handleDLRCallback(w http.ResponseWriter, r *http.Request) {
	var dlrData DLRData
	if err := json.NewDecoder(r.Body).Decode(&dlrData); err != nil {
		slog.Error("Failed to decode DLR callback", slog.Any("error", err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	slog.Debug("Received HTTP DLR callback",
		slog.String("message_id", dlrData.MessageID),
		slog.String("status", dlrData.Status),
		slog.String("callback_url", dlrData.CallbackURL),
	)

	connID, _ := strconv.ParseInt(r.URL.Query().Get("connID"), 0, 64)
	internalMessageID, _ := strconv.ParseInt(r.URL.Query().Get("msgID"), 0, 64)

	// Process the DLR callback
	dlrInfo := DLRInfo{
		MnoMessageID:  dlrData.MessageID,
		OriginalMsgID: internalMessageID, // Now we can include the internal message ID
		SegmentSeqn:   1,                 // HTTP APIs typically send full messages, so sequence is 1
		Status:        dlrData.Status,
		Timestamp:     time.Now(), // HTTP API doesn't provide timestamp, use current time
		RawDLRData: map[string]any{
			"callback_url": dlrData.CallbackURL,
		},
	}
	s.handleDLRFunc(context.Background(), int32(connID), dlrInfo)

	slog.Info("Successfully processed HTTP DLR callback",
		slog.Int64("internal_message_id", internalMessageID),
		slog.String("message_id", dlrData.MessageID),
		slog.String("status", dlrData.Status),
	)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleHealthCheck provides a simple health check endpoint
func (s *DLRServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "dlr-callback-server",
	})
}

func StartDLRServer(dlrHandleFunc DLRHandlerFunc, port string) (*DLRServer, error) {
	config := DLRServerConfig{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		APIKey:       "", // Set this if you want to require authentication
	}

	server := NewDLRServer(config, dlrHandleFunc)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			slog.Error("DLR server failed", slog.Any("error", err))
		}
	}()

	return server, nil
}
