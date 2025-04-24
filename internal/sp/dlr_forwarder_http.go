package sp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/errormapper"
)

// HTTPForwarderConfig holds configuration for the HTTP client.
type HTTPForwarderConfig struct {
	Timeout time.Duration
}

// HTTPSPForwarder implements DLRForwarder for HTTP clients.
type HTTPSPForwarder struct {
	config HTTPForwarderConfig
	client *http.Client
}

// NewHTTPSPForwarder creates a new HTTP DLR forwarder.
func NewHTTPSPForwarder(cfg HTTPForwarderConfig) *HTTPSPForwarder {
	// Consider allowing injection of http.Client for testing/customization
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}
	return &HTTPSPForwarder{
		config: cfg,
		client: httpClient,
	}
}

// ForwardDLR implements the DLRForwarder interface for HTTP.
func (f *HTTPSPForwarder) ForwardDLR(ctx context.Context, spDetails SPDetails, dlr ForwardedDLRInfo) error {
	if spDetails.HTTPCallbackURL == "" {
		return errors.New("cannot forward DLR via HTTP: missing CallbackURL in SPDetails")
	}
	callbackURL := spDetails.HTTPCallbackURL
	logCtx := logging.ContextWithMessageID(ctx, dlr.InternalMessageID)
	logCtx = logging.ContextWithSPID(logCtx, spDetails.ServiceProviderID)
	logCtx = logging.ContextWithCallbackURL(logCtx, callbackURL) // Add helper if needed

	slog.InfoContext(logCtx, "Attempting to forward DLR via HTTP", slog.String("dlr_status", dlr.Status))

	// 1. Map DLR info to the expected payload format (e.g., JSON)
	// Define a struct for the HTTP payload
	type httpDLRPayload struct {
		MessageID     string    `json:"message_id"` // Use SP's ref ideally
		InternalID    int64     `json:"internal_id"`
		SourceAddr    string    `json:"source_address"`
		DestAddr      string    `json:"destination_address"`
		SubmitDate    time.Time `json:"submit_date_utc"`
		DoneDate      time.Time `json:"done_date_utc"`
		Status        string    `json:"status"`     // e.g., DELIVERED, FAILED, REJECTED
		ErrorCode     string    `json:"error_code"` // Mapped error code
		NetworkCode   *string   `json:"network_code,omitempty"`
		TotalSegments int32     `json:"total_segments"`
	}

	payloadID := fmt.Sprintf("%d", dlr.InternalMessageID)
	if dlr.ClientMessageRef != nil {
		payloadID = *dlr.ClientMessageRef
	}

	// Map internal status/code to HTTP status/code
	httpStatusCode := "0" // Default success code for HTTP callback? Or use dlr.Status?
	if dlr.ErrorCode != nil {
		httpStatusCode = errormapper.MapErrorCode(*dlr.ErrorCode, "http")
	}

	payload := httpDLRPayload{
		MessageID:     payloadID,
		InternalID:    dlr.InternalMessageID,
		SourceAddr:    dlr.SourceAddr,
		DestAddr:      dlr.DestAddr,
		SubmitDate:    dlr.SubmitDate.UTC(),
		DoneDate:      dlr.DoneDate.UTC(),
		Status:        strings.ToUpper(dlr.Status),
		ErrorCode:     httpStatusCode,
		TotalSegments: dlr.TotalSegments,
	}
	if dlr.NetworkCode != nil {
		payload.NetworkCode = dlr.NetworkCode
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to marshal HTTP DLR payload", slog.Any("error", err))
		return fmt.Errorf("failed to marshal DLR payload: %w", err)
	}

	// 2. Create HTTP Request
	req, err := http.NewRequestWithContext(logCtx, http.MethodPost, callbackURL, bytes.NewReader(payloadBytes))
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to create HTTP DLR request", slog.Any("error", err))
		return fmt.Errorf("failed to create HTTP DLR request: %w", err)
	}

	// Set Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AegisBox-DLR-Forwarder/1.0") // Identify our service

	// 3. Add Authentication (Example: Bearer Token)
	// Authentication details should come from spDetails.HTTPAuthConfig
	// This needs secure parsing and handling.
	if spDetails.HTTPAuthConfig == "" {
		authHeader := fmt.Sprintf("Bearer %s", spDetails.HTTPAuthConfig)
		req.Header.Set("Authorization", authHeader)
		slog.DebugContext(logCtx, "Added Authorization header for HTTP DLR forwarding")
	} else {
		slog.WarnContext(logCtx, "No HTTP Authentication configured for SP callback")
	}

	// 4. Send Request
	slog.DebugContext(logCtx, "Sending HTTP DLR request", slog.String("url", callbackURL), slog.String("payload", string(payloadBytes)))
	resp, err := f.client.Do(req)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to send HTTP DLR request", slog.Any("error", err))
		// This could be a temporary network issue, maybe retry later if using queue worker
		return fmt.Errorf("failed to send HTTP DLR request to %s: %w", callbackURL, err)
	}
	defer resp.Body.Close()

	// 5. Check Response Status
	// SP should ideally return 2xx on successful receipt.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.InfoContext(logCtx, "Successfully forwarded DLR via HTTP", slog.Int("http_status", resp.StatusCode))
		return nil
	} else {
		// SP endpoint rejected the DLR or encountered an error
		err = fmt.Errorf("received non-2xx status code (%d) from SP HTTP callback %s", resp.StatusCode, callbackURL)
		slog.WarnContext(logCtx, "SP HTTP endpoint returned error status for DLR",
			slog.Int("http_status", resp.StatusCode),
			slog.Any("error", err),
		)
		// Treat as failure, maybe retry later if using queue worker
		return err
	}
}
