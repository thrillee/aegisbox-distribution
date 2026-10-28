package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno"
)

// WebhookHandler handles DLR webhooks from HTTP MNO providers
type WebhookHandler struct {
	dbQueries    database.Querier
	configLoader *mno.ProviderConfigLoader
	logger       *slog.Logger
}

// NewWebhookHandler creates a new webhook handler instance
func NewWebhookHandler(q database.Querier) *WebhookHandler {
	logger := slog.Default()
	return &WebhookHandler{
		dbQueries:    q,
		configLoader: mno.NewProviderConfigLoader(q, logger),
		logger:       logger,
	}
}

// HandleProviderWebhook is a provider-agnostic webhook handler
// It determines the provider from query parameters and delegates parsing
func (h *WebhookHandler) HandleProviderWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// Get provider name from query parameter
	providerName := c.Query("provider")
	if providerName == "" {
		h.logger.WarnContext(ctx, "Webhook received without provider parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider parameter is required"})
		return
	}

	// Get connection ID from query parameter (optional, for routing)
	connIDStr := c.Query("conn_id")

	h.logger.InfoContext(ctx, "Received DLR webhook",
		"provider", providerName,
		"conn_id", connIDStr,
		"remote_addr", c.ClientIP())

	// Parse DLR based on provider
	messageID, status, err := h.parseDLRByProvider(ctx, providerName, c.Request)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to parse DLR webhook",
			"provider", providerName,
			"error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse webhook", "details": err.Error()})
		return
	}

	if messageID == "" {
		h.logger.WarnContext(ctx, "DLR webhook contained no message ID",
			"provider", providerName)
		c.JSON(http.StatusBadRequest, gin.H{"error": "no message ID in webhook"})
		return
	}

	// Process the DLR using common logic
	h.processDLR(c, messageID, status)
}

// parseDLRByProvider parses the webhook based on the provider type
// It loads the provider configuration from the database including API keys
func (h *WebhookHandler) parseDLRByProvider(ctx context.Context, providerName string, req *http.Request) (messageID, status string, err error) {
	// Load provider from database with full configuration (including API keys)
	provider, err := h.configLoader.GetProviderByName(ctx, providerName)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to load provider from database",
			"provider", providerName,
			"error", err)
		return "", "", err
	}

	// Clone the request body for parsing since providers read it
	bodyBytes, err := cloneRequestBody(req)
	if err != nil {
		return "", "", err
	}
	req.Body = bodyBytes

	return provider.ParseDLR(ctx, req)
}

// UnknownProviderError represents an error for unknown providers
type UnknownProviderError struct {
	Name string
}

func (e *UnknownProviderError) Error() string {
	return "unknown provider: " + e.Name
}

// HandleTermiiWebhook is a dedicated handler for Termii webhooks
func (h *WebhookHandler) HandleTermiiWebhook(c *gin.Context) {
	h.handleProviderSpecificWebhook(c, "TERMII")
}

// HandleAfricaTalkingWebhook is a dedicated handler for AfricaTalking webhooks
func (h *WebhookHandler) HandleAfricaTalkingWebhook(c *gin.Context) {
	h.handleProviderSpecificWebhook(c, "AFRICATALKING")
}

// HandleSimpuWebhook is a dedicated handler for Simpu webhooks
func (h *WebhookHandler) HandleSimpuWebhook(c *gin.Context) {
	h.handleProviderSpecificWebhook(c, "SIMPU")
}

// HandleSimpuOTPWebhook is a dedicated handler for Simpu OTP webhooks
func (h *WebhookHandler) HandleSimpuOTPWebhook(c *gin.Context) {
	h.handleProviderSpecificWebhook(c, "SIMPU_OTP")
}

// HandleEBulkSMSWebhook is a dedicated handler for eBulkSMS webhooks
func (h *WebhookHandler) HandleEBulkSMSWebhook(c *gin.Context) {
	h.handleProviderSpecificWebhook(c, "EBULKSMS")
}

// HandleMultiTexterWebhook is a dedicated handler for MultiTexter webhooks
func (h *WebhookHandler) HandleMultiTexterWebhook(c *gin.Context) {
	h.handleProviderSpecificWebhook(c, "MULTITEXTER")
}

// handleProviderSpecificWebhook handles webhooks for a specific provider
func (h *WebhookHandler) handleProviderSpecificWebhook(c *gin.Context, providerName string) {
	ctx := c.Request.Context()

	// Use the explicitly provided provider name instead of query parameter
	h.logger.InfoContext(ctx, "Received DLR webhook",
		"provider", providerName,
		"conn_id", c.Query("conn_id"),
		"remote_addr", c.ClientIP())

	// Parse DLR based on provider
	messageID, status, err := h.parseDLRByProvider(ctx, providerName, c.Request)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to parse DLR webhook",
			"provider", providerName,
			"error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse webhook", "details": err.Error()})
		return
	}

	if messageID == "" {
		h.logger.WarnContext(ctx, "DLR webhook contained no message ID",
			"provider", providerName)
		c.JSON(http.StatusBadRequest, gin.H{"error": "no message ID in webhook"})
		return
	}

	// Process the DLR (same logic as HandleProviderWebhook)
	h.processDLR(c, messageID, status)
}

// processDLR contains the common DLR processing logic
func (h *WebhookHandler) processDLR(c *gin.Context, messageID, status string) {
	ctx := c.Request.Context()

	// Look up the message segment by MNO message ID
	segment, err := h.dbQueries.FindSegmentByMnoMessageID(ctx, &messageID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to find message segment",
			"mno_message_id", messageID,
			"error", err)
		// Still return 200 to acknowledge receipt
		c.JSON(http.StatusOK, gin.H{"status": "acknowledged", "warning": "message not found"})
		return
	}

	// Use the routed MNO ID from the segment query
	// For HTTP MNO connections, we'll use the first active connection
	// In practice, the connection ID is stored in mno_connection_id field of message_segments
	// but since FindSegmentByMnoMessageID doesn't return it, we need to work with what we have

	// For now, use conn_id from query params if available
	connIDStr := c.Query("conn_id")
	var connID int32
	if connIDStr != "" {
		// Parse the connection ID from query
		var parsed int64
		if _, err := fmt.Sscanf(connIDStr, "%d", &parsed); err == nil {
			connID = int32(parsed)
		}
	}

	// If no conn_id in query and we have a routing assignment, look it up
	if connID == 0 && segment.RoutingAssignmentID != nil {
		assignment, err := h.dbQueries.GetRoutingAssignmentByID(ctx, *segment.RoutingAssignmentID)
		if err == nil {
			// Use the MNO ID to find an active connection
			// This is a simplified approach - in production you might want to store the actual connection ID used
			if segment.RoutedMnoID != nil {
				connID = *segment.RoutedMnoID // Use MNO ID as fallback
			}
			_ = assignment // Silence unused warning
		}
	}

	if connID == 0 {
		h.logger.WarnContext(ctx, "Cannot determine connection ID for DLR",
			"segment_id", segment.SegmentID,
			"mno_message_id", messageID)
		c.JSON(http.StatusOK, gin.H{"status": "acknowledged", "warning": "no connection ID"})
		return
	}

	// Update the segment DLR status directly in the database
	err = h.dbQueries.UpdateSegmentDLR(ctx, database.UpdateSegmentDLRParams{
		DlrStatus: &status,
		ErrorCode: nil, // TODO: Extract error code from DLR if available
		ID:        segment.SegmentID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update segment DLR status",
			"segment_id", segment.SegmentID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update DLR status"})
		return
	}

	h.logger.InfoContext(ctx, "DLR webhook processed successfully",
		"mno_message_id", messageID,
		"status", status,
		"segment_id", segment.SegmentID)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// cloneRequestBody clones the request body to allow multiple reads
func cloneRequestBody(req *http.Request) (io.ReadCloser, error) {
	if req.Body == nil {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()

	return io.NopCloser(bytes.NewReader(bodyBytes)), nil
}
