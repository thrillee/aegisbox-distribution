package http_providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AfricaTalkingProvider implements HTTP SMS provider for AfricaTalking API
type AfricaTalkingProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	username string
	senderID string
	logger   *slog.Logger
}

// NewAfricaTalkingProvider creates a new AfricaTalking provider instance
func NewAfricaTalkingProvider(config HTTPProviderConfig, logger *slog.Logger) *AfricaTalkingProvider {
	// Set default timeout if not provided
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &AfricaTalkingProvider{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		apiKey:   config.APIKey,
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		username: config.Username,
		senderID: config.DefaultSender,
		logger:   logger,
	}
}

// AfricaTalkingSendRequest represents the request payload for AfricaTalking SMS API
type AfricaTalkingSendRequest struct {
	Username     string   `json:"username"`
	Message      string   `json:"message"`
	SenderID     string   `json:"senderId"`
	PhoneNumbers []string `json:"phoneNumbers"`
}

// AfricaTalkingSendResponse represents the response from AfricaTalking SMS API
type AfricaTalkingSendResponse struct {
	SMSMessageData struct {
		Message    string `json:"Message"`
		Recipients []struct {
			StatusCode int    `json:"statusCode"`
			Number     string `json:"number"`
			Cost       string `json:"cost"`
			Status     string `json:"status"`
			MessageID  string `json:"messageId"`
		} `json:"Recipients"`
	} `json:"SMSMessageData"`
}

// SendSMS sends an SMS through AfricaTalking API
func (a *AfricaTalkingProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	a.logger.InfoContext(ctx, "Sending SMS via AfricaTalking", "from", from, "to", to)

	if from == "" {
		from = a.senderID
	}

	recipients := strings.Split(to, ",")
	for i, r := range recipients {
		recipients[i] = strings.TrimSpace(r)
	}

	reqBody := AfricaTalkingSendRequest{
		Username:     a.username,
		Message:      message,
		SenderID:     from,
		PhoneNumbers: recipients,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/version1/messaging/bulk", bytes.NewBuffer(jsonBody))
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apiKey", a.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	a.logger.DebugContext(ctx, "Sending HTTP request to AfricaTalking", "url", req.URL.String())

	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.ErrorContext(ctx, "HTTP request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to read response", "error", err)
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	a.logger.DebugContext(ctx, "AfricaTalking Response", "status", resp.StatusCode, "response", string(body))

	// AfricaTalking returns 201 Created on success
	if resp.StatusCode != http.StatusCreated {
		a.logger.WarnContext(ctx, "AfricaTalking API returned non-201 status", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response AfricaTalkingSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		a.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.SMSMessageData.Recipients) == 0 {
		a.logger.ErrorContext(ctx, "No recipients processed")
		return "", fmt.Errorf("no recipients processed")
	}

	messageID := response.SMSMessageData.Recipients[0].MessageID
	a.logger.InfoContext(ctx, "SMS sent successfully via AfricaTalking", "message_id", messageID)
	return messageID, nil
}

// CheckStatus checks the delivery status of a message
func (a *AfricaTalkingProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

// ParseDLR parses a delivery receipt from an HTTP request
func (a *AfricaTalkingProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	if err := req.ParseForm(); err != nil {
		a.logger.ErrorContext(ctx, "Failed to parse form data", "error", err)
		return "", "", fmt.Errorf("failed to parse form data: %v", err)
	}

	messageID := req.FormValue("id")
	status := req.FormValue("status")

	a.logger.DebugContext(ctx, "AfricaTalking DLR",
		"id", messageID,
		"status", status)

	return messageID, NormalizeStatus("AFRICATALKING", status), nil
}

// Name returns the provider name
func (a *AfricaTalkingProvider) Name() string {
	return "AFRICATALKING"
}

// SupportsDLR returns whether the provider supports delivery receipts
func (a *AfricaTalkingProvider) SupportsDLR() bool {
	return true
}

// SupportsAutoDLR returns false for AfricaTalking
func (a *AfricaTalkingProvider) SupportsAutoDLR() bool {
	return false
}

// GetDefaultSenderID returns the default sender ID
func (a *AfricaTalkingProvider) GetDefaultSenderID() string {
	return a.senderID
}
