package mno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/thrillee/aegisbox/internal/mno/http_providers"
	"github.com/thrillee/aegisbox/pkg/codes"
)

// HTTPProviderConnector wraps an HTTPProvider (TERMII, AfricaTalking, etc.)
// and implements the Connector interface so it can be used by the Manager
type HTTPProviderConnector struct {
	connectionID int32
	mnoID        int32
	provider     http_providers.HTTPProvider
	dlrHandler   DLRHandlerFunc
	logger       *slog.Logger
	mu           sync.RWMutex
	status       string
}

// NewHTTPProviderConnector creates a new connector from an HTTPProvider
func NewHTTPProviderConnector(connectionID, mnoID int32, provider http_providers.HTTPProvider, logger *slog.Logger) *HTTPProviderConnector {
	return &HTTPProviderConnector{
		connectionID: connectionID,
		mnoID:        mnoID,
		provider:     provider,
		logger:       logger,
		status:       codes.StatusHttpOk,
	}
}

// SubmitMessage implements the Connector interface
func (c *HTTPProviderConnector) SubmitMessage(ctx context.Context, msgDetails PreparedMessage) (SubmitResult, error) {
	c.mu.RLock()
	status := c.status
	c.mu.RUnlock()

	if status != codes.StatusHttpOk {
		return SubmitResult{
			WasSubmitted: false,
			Error:        fmt.Errorf("connector is not in ready state: %s", status),
		}, nil
	}

	c.logger.InfoContext(ctx, "Submitting message via HTTP provider",
		slog.String("provider", c.provider.Name()),
		slog.Int64("message_id", msgDetails.InternalMessageID),
		slog.String("destination", msgDetails.DestinationMSISDN))

	// Use sender ID from message, fallback to provider's default
	senderID := msgDetails.SenderID
	if senderID == "" {
		senderID = c.provider.GetDefaultSenderID()
	}

	// Call the provider's SendSMS method
	messageID, err := c.provider.SendSMS(ctx, senderID, msgDetails.DestinationMSISDN, msgDetails.MessageContent)

	// Initialize segments array
	segments := make([]SegmentSubmitInfo, len(msgDetails.DBSeqs))

	if err != nil {
		// All segments failed
		c.logger.ErrorContext(ctx, "Provider send failed",
			slog.String("provider", c.provider.Name()),
			slog.Any("error", err))

		errCode := "PROV_ERR" // fits message_segments.error_code VARCHAR(10)
		for i := range segments {
			dbSegID := msgDetails.DBSeqs[i]
			segments[i] = SegmentSubmitInfo{
				Seqn:      int32(i + 1),
				IsSuccess: false,
				Error:     err,
				ErrorCode: &errCode,
				DBSegID:   &dbSegID,
			}
		}

		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
			Error:        err,
		}, nil
	}

	// Provider returned message ID successfully
	// Most HTTP providers treat the message as a single unit
	// but we need to match our database segment structure
	if messageID == "" {
		// No message ID returned - treat as failure
		errCode := "NO_MSG_ID"
		errMsg := "Provider did not return message ID"
		for i := range segments {
			dbSegID := msgDetails.DBSeqs[i]
			segments[i] = SegmentSubmitInfo{
				Seqn:      int32(i + 1),
				IsSuccess: false,
				Error:     errors.New(errMsg),
				ErrorCode: &errCode,
				DBSegID:   &dbSegID,
			}
		}
		return SubmitResult{
			Segments:     segments,
			WasSubmitted: false,
		}, nil
	}

	// Map provider message ID to all segments
	// HTTP providers handle segmentation internally, so we use the same
	// message ID for all our database segments
	wasSubmitted := true
	for i := range segments {
		dbSegID := msgDetails.DBSeqs[i]
		segments[i] = SegmentSubmitInfo{
			Seqn:         int32(i + 1),
			MnoMessageID: &messageID,
			IsSuccess:    true,
			DBSegID:      &dbSegID,
		}
	}

	c.logger.InfoContext(ctx, "Message submitted successfully via HTTP provider",
		slog.String("provider", c.provider.Name()),
		slog.Int("segment_count", len(segments)),
		slog.String("message_id", messageID))
	slog.Bool("support_auto_dlr", c.provider.SupportsAutoDLR())

	// Auto-DLR: If provider supports auto-DLR (e.g., Akuuk), immediately send "delivered" status
	if c.provider.SupportsAutoDLR() {
		c.logger.InfoContext(ctx, "Provider supports auto-DLR, sending delivered status immediately",
			slog.String("provider", c.provider.Name()),
			slog.String("message_id", messageID))

		// Send the auto-DLR asynchronously to avoid blocking the submit response
		go func() {
			// Use a background context with timeout to avoid cancellation issues
			dlrCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := c.ProcessDLRWebhook(dlrCtx, messageID, "delivered", "", time.Now()); err != nil {
				c.logger.ErrorContext(dlrCtx, "Failed to process auto-DLR",
					slog.String("provider", c.provider.Name()),
					slog.String("message_id", messageID),
					slog.Any("error", err))
			}
		}()
	}

	return SubmitResult{
		Segments:     segments,
		WasSubmitted: wasSubmitted,
	}, nil
}

// RegisterDLRHandler implements the Connector interface
func (c *HTTPProviderConnector) RegisterDLRHandler(handler DLRHandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dlrHandler = handler
}

// ProcessDLRWebhook handles incoming DLR webhooks from HTTP providers
// This is called by the webhook handler after parsing the provider-specific DLR format
func (c *HTTPProviderConnector) ProcessDLRWebhook(ctx context.Context, messageID, status, errorCode string, timestamp time.Time) error {
	c.mu.RLock()
	handler := c.dlrHandler
	c.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("no DLR handler registered")
	}

	c.logger.InfoContext(ctx, "Processing DLR webhook",
		slog.String("provider", c.provider.Name()),
		slog.String("message_id", messageID),
		slog.String("status", status))

	// Convert to DLRInfo
	dlrInfo := DLRInfo{
		MnoMessageID: messageID,
		Status:       status,
		ErrorCode:    errorCode,
		Timestamp:    timestamp,
		NetworkCode:  sql.NullString{}, // HTTP providers typically don't provide this
		RawDLRData:   make(map[string]any),
	}

	// Call the registered handler
	if err := handler(ctx, c.connectionID, dlrInfo); err != nil {
		c.logger.ErrorContext(ctx, "DLR handler failed",
			slog.String("provider", c.provider.Name()),
			slog.Any("error", err))
		return err
	}

	return nil
}

// Shutdown implements the Connector interface
func (c *HTTPProviderConnector) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.InfoContext(ctx, "Shutting down HTTP provider connector",
		slog.String("provider", c.provider.Name()),
		slog.Int("connection_id", int(c.connectionID)))

	// HTTP providers don't need explicit shutdown
	// Just mark as disconnected
	c.status = codes.StatusDisconnected
	return nil
}

// Status implements the Connector interface
func (c *HTTPProviderConnector) Status() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// ConnectionID implements the Connector interface
func (c *HTTPProviderConnector) ConnectionID() int32 {
	return c.connectionID
}

// MnoID implements the Connector interface
func (c *HTTPProviderConnector) MnoID() int32 {
	return c.mnoID
}

// Provider returns the underlying HTTPProvider (for testing/debugging)
func (c *HTTPProviderConnector) Provider() http_providers.HTTPProvider {
	return c.provider
}

// HealthCheck performs a health check on the provider
func (c *HTTPProviderConnector) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// HTTP providers are considered healthy if they're in StatusHttpOk
	if c.status == codes.StatusHttpOk {
		return nil
	}

	return fmt.Errorf("provider not in healthy state: %s", c.status)
}
