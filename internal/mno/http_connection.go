package mno

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/pkg/codes"
)

// HTTPSendRequest represents the request payload for sending messages
type HTTPSendRequest struct {
	To          string `json:"to"`
	Message     string `json:"message"`
	SenderID    string `json:"sender_id"`
	CallbackURL string `json:"callback_url"`
}

// HTTPSendResponse represents the response from the HTTP API
type HTTPSendResponse struct {
	MessageID string `json:"message_id"`
}

// DLRData represents the DLR callback payload
type DLRData struct {
	MessageID   string `json:"message_id"`
	Status      string `json:"status"`
	CallbackURL string `json:"callback_url"`
}

// HTTPMNOConnector implements the Connector interface for HTTP-based MNO connections
type HTTPMNOConnector struct {
	config     *HTTPConfig
	httpClient *http.Client
	dlrHandler DLRHandlerFunc
	mu         sync.RWMutex
	status     string
}

// HTTPConfig holds configuration for HTTP MNO connection
type HTTPConfig struct {
	ID          int32  // mno_connections.id
	MnoID       int32  // mnos.id
	BaseURL     string // e.g., "http://52.10.34.217/aegisbox"
	APIKey      string // X-API-KEY header value
	CallbackURL string // URL for receiving DLRs
	Timeout     time.Duration
}

// NewHTTPConfigFromDB creates HTTPConfig from database connection
func NewHTTPConfigFromDB(dbConn database.GetActiveMNOConnectionsRow) (*HTTPConfig, error) {
	var httpConfigParams map[string]any

	if err := json.Unmarshal(dbConn.HttpConfig, &httpConfigParams); err != nil {
		return nil, err
	}

	slog.Info("HTTP Config", "httpConfigParams", httpConfigParams)

	baseURL := httpConfigParams["endpoint_url"].(string)
	callbackURL := httpConfigParams["callback_url"].(string)
	apiKey := httpConfigParams["auth_token"].(string)
	timeoutSecs := httpConfigParams["timeout_secs"].(float64)

	config := HTTPConfig{
		MnoID:       dbConn.MnoID,
		ID:          dbConn.ID,
		BaseURL:     baseURL,
		CallbackURL: callbackURL,
		APIKey:      apiKey,
		Timeout:     time.Duration(timeoutSecs),
	}
	var timeoutSeconds int

	config.Timeout = time.Duration(timeoutSeconds) * time.Second
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second // default timeout
	}

	return &config, nil
}

// NewHTTPMNOConnector creates a new HTTP MNO connector
func NewHTTPMNOConnector(config *HTTPConfig) *HTTPMNOConnector {
	return &HTTPMNOConnector{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		status: codes.StatusHttpOk,
	}
}

// SubmitMessage implements the Connector interface
func (h *HTTPMNOConnector) SubmitMessage(ctx context.Context, msgDetails PreparedMessage) (SubmitResult, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.status != codes.StatusHttpOk {
		return SubmitResult{
			WasSubmitted: false,
			Error:        fmt.Errorf("connector is not in connected state: %s", h.status),
		}, nil
	}

	// For HTTP API, we typically send the full message as one request
	// The segmentation is handled by the HTTP API provider
	segments := make([]SegmentSubmitInfo, 1)

	callbackURL := h.config.CallbackURL
	callbackURL = strings.ReplaceAll(callbackURL, "[connID]", fmt.Sprintf("%v", h.ConnectionID()))
	callbackURL = strings.ReplaceAll(callbackURL, "[msgID]", fmt.Sprintf("%v", msgDetails.InternalMessageID))

	dbSegmentID := msgDetails.DBSeqs[0]

	// Prepare the HTTP request
	sendReq := HTTPSendRequest{
		To:          msgDetails.DestinationMSISDN,
		Message:     msgDetails.MessageContent,
		SenderID:    "", // msgDetails.SenderID
		CallbackURL: callbackURL,
	}

	jsonData, err := json.Marshal(sendReq)
	if err != nil {
		segments[0] = SegmentSubmitInfo{
			Seqn:      1,
			IsSuccess: false,
			Error:     fmt.Errorf("failed to marshal request: %w", err),
			DBSegID:   &dbSegmentID,
		}
		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
			Error:        err,
		}, nil
	}

	// Create HTTP request
	url := h.config.BaseURL
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		segments[0] = SegmentSubmitInfo{
			Seqn:      1,
			IsSuccess: false,
			Error:     fmt.Errorf("failed to create HTTP request: %w", err),
			DBSegID:   &dbSegmentID,
		}
		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
			Error:        err,
		}, nil
	}

	slog.InfoContext(ctx, "Sending SMS Via API", "url", url, "callbackURL", callbackURL, "jsonData", sendReq)

	// Set headers
	req.Header.Set("X-API-KEY", h.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		segments[0] = SegmentSubmitInfo{
			Seqn:      1,
			IsSuccess: false,
			Error:     fmt.Errorf("HTTP request failed: %w", err),
			DBSegID:   &dbSegmentID,
		}
		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
			Error:        err,
		}, nil
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("HTTP API returned status %d [%s]", resp.StatusCode, url)
		segments[0] = SegmentSubmitInfo{
			Seqn:      1,
			IsSuccess: false,
			Error:     fmt.Errorf(errorMsg),
			ErrorCode: &errorMsg,
			DBSegID:   &dbSegmentID,
		}
		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
		}, nil
	}

	var sendResp HTTPSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		segments[0] = SegmentSubmitInfo{
			Seqn:      1,
			IsSuccess: false,
			Error:     fmt.Errorf("failed to decode response: %w", err),
			DBSegID:   &dbSegmentID,
		}
		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
			Error:        err,
		}, nil
	}

	// Success case
	segments[0] = SegmentSubmitInfo{
		Seqn:         1,
		MnoMessageID: &sendResp.MessageID,
		IsSuccess:    true,
		DBSegID:      &dbSegmentID,
	}

	return SubmitResult{
		Segments:     segments,
		WasSubmitted: true,
	}, nil
}

// RegisterDLRHandler implements the Connector interface
func (h *HTTPMNOConnector) RegisterDLRHandler(handler DLRHandlerFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dlrHandler = handler
}

// ProcessDLRCallback processes incoming DLR callbacks from the HTTP API
// This should be called by your HTTP server when it receives DLR callbacks
func (h *HTTPMNOConnector) ProcessDLRCallback(ctx context.Context, dlrData DLRData) error {
	h.mu.RLock()
	handler := h.dlrHandler
	h.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("no DLR handler registered")
	}

	// Convert DLRData to DLRInfo
	dlrInfo := DLRInfo{
		MnoMessageID: dlrData.MessageID,
		Status:       dlrData.Status,
		Timestamp:    time.Now(), // HTTP API doesn't provide timestamp, use current time
		RawDLRData: map[string]any{
			"callback_url": dlrData.CallbackURL,
		},
	}

	return handler(ctx, h.config.ID, dlrInfo)
}

// Shutdown implements the Connector interface
func (h *HTTPMNOConnector) Shutdown(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.status = "disconnected"

	// For HTTP connector, we just need to mark as disconnected
	// The http.Client doesn't need explicit shutdown

	return nil
}

// Status implements the Connector interface
func (h *HTTPMNOConnector) Status() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// ConnectionID implements the Connector interface
func (h *HTTPMNOConnector) ConnectionID() int32 {
	return h.config.ID
}

// MnoID implements the Connector interface
func (h *HTTPMNOConnector) MnoID() int32 {
	return h.config.MnoID
}

// UpdateConnectionStatus allows updating the connection status (e.g., for health checks)
func (h *HTTPMNOConnector) UpdateConnectionStatus(status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status = status
}

// HealthCheck performs a basic health check of the HTTP API
func (h *HTTPMNOConnector) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "HEAD", h.config.BaseURL, nil)
	if err != nil {
		h.UpdateConnectionStatus("error")
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("X-API-KEY", h.config.APIKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.UpdateConnectionStatus("error")
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		h.UpdateConnectionStatus("error")
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	h.UpdateConnectionStatus("connected")
	return nil
}

// Enhanced ProcessDLRCallback that includes internal message ID
func (h *HTTPMNOConnector) ProcessDLRCallbackWithInternalID(ctx context.Context, dlrData DLRData, internalMessageID int64) error {
	h.mu.RLock()
	handler := h.dlrHandler
	h.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("no DLR handler registered")
	}

	// Convert DLRData to DLRInfo with internal message ID
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

	return handler(ctx, h.config.ID, dlrInfo)
}
