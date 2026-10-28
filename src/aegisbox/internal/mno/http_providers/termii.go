package http_providers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// TermiiProvider implements HTTP SMS provider for Termii API
type TermiiProvider struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	senderID  string
	secretKey string
	logger    *slog.Logger
}

// NewTermiiProvider creates a new Termii provider instance
func NewTermiiProvider(config HTTPProviderConfig, logger *slog.Logger) *TermiiProvider {
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

	return &TermiiProvider{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		apiKey:    config.APIKey,
		baseURL:   strings.TrimRight(config.BaseURL, "/"),
		senderID:  config.DefaultSender,
		secretKey: config.SecretKey,
		logger:    logger,
	}
}

// TermiiSendRequest represents the request payload for Termii SMS API
type TermiiSendRequest struct {
	APIKey  string `json:"api_key"`
	To      any    `json:"to"`
	From    string `json:"from"`
	SMS     string `json:"sms"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

// TermiiSendResponse represents the response from Termii SMS API
type TermiiSendResponse struct {
	Code         string  `json:"code"`           // "ok" for success
	MessageID    string  `json:"message_id"`     // Message ID
	MessageIDStr string  `json:"message_id_str"` // Message ID as string
	Message      string  `json:"message"`        // Response message (e.g., "Successfully Sent")
	Balance      float64 `json:"balance"`        // Account balance
	User         string  `json:"user"`           // Account user name
}

// TermiiDLR represents a delivery receipt from Termii
type TermiiDLR struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	MessageID string `json:"message_id"`
	Receiver  string `json:"receiver"`
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	SentAt    string `json:"sent_at"`
	Cost      string `json:"cost"`
	Status    string `json:"status"`
	Channel   string `json:"channel"`
}

// SendSMS sends an SMS through Termii API
func (t *TermiiProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	t.logger.InfoContext(ctx, "Sending SMS via Termii", "from", from, "to", to)

	if from == "" {
		from = t.senderID
	}

	var recipients any = to
	if strings.Contains(to, ",") {
		recipientList := strings.Split(to, ",")
		for i, r := range recipientList {
			recipientList[i] = strings.TrimSpace(r)
		}
		recipients = recipientList
		t.logger.DebugContext(ctx, "Multiple recipients parsed", "count", len(recipientList))
	}

	reqBody := TermiiSendRequest{
		APIKey:  t.apiKey,
		To:      recipients,
		From:    from,
		SMS:     message,
		Type:    "plain",
		Channel: "generic",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/api/sms/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Println("Termii Send Request:", string(jsonBody))

	t.logger.DebugContext(ctx, "Sending HTTP request to Termii", "url", req.URL.String())

	resp, err := t.client.Do(req)
	if err != nil {
		t.logger.ErrorContext(ctx, "HTTP request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to read response", "error", err)
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	t.logger.InfoContext(ctx, "Termii Response", "status", resp.StatusCode, "response", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.logger.WarnContext(ctx, "Termii API returned non-2xx status", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response TermiiSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if the response code is "ok" (Termii returns "ok" for success)
	if !strings.EqualFold(response.Code, "ok") {
		t.logger.ErrorContext(ctx, "Send failed: invalid response code", "code", response.Code, "message", response.Message)
		return "", fmt.Errorf("send failed: %s (code: %s)", response.Message, response.Code)
	}

	// Get message ID (prefer message_id_str, fallback to message_id)
	messageID := response.MessageIDStr
	if messageID == "" {
		messageID = response.MessageID
	}

	if messageID == "" {
		t.logger.ErrorContext(ctx, "Send failed: no message ID in response", "message", response.Message)
		return "", fmt.Errorf("send failed: %s", response.Message)
	}

	t.logger.InfoContext(ctx, "SMS sent successfully via Termii",
		"message_id", messageID,
		"balance", response.Balance,
		"user", response.User)
	return messageID, nil
}

// CheckStatus checks the delivery status of a message
func (t *TermiiProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

// ParseDLR parses a delivery receipt from an HTTP request
func (t *TermiiProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	var dlr TermiiDLR
	if err := json.NewDecoder(req.Body).Decode(&dlr); err != nil {
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	t.logger.InfoContext(ctx, "Termii Received Webhook", "dlr", dlr)

	status := NormalizeStatus("TERMII", dlr.Status)
	return dlr.MessageID, status, nil
}

// verifySignature verifies the webhook signature from Termii
func (t *TermiiProvider) verifySignature(req *http.Request) bool {
	if t.secretKey == "" {
		return true
	}

	signature := req.Header.Get("X-Termii-Signature")
	if signature == "" {
		return false
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return false
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	mac := hmac.New(sha512.New, []byte(t.secretKey))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// Name returns the provider name
func (t *TermiiProvider) Name() string {
	return "TERMII"
}

// SupportsDLR returns whether the provider supports delivery receipts
func (t *TermiiProvider) SupportsDLR() bool {
	return true
}

// SupportsAutoDLR returns false for Termii
func (t *TermiiProvider) SupportsAutoDLR() bool {
	return false
}

// GetDefaultSenderID returns the default sender ID
func (t *TermiiProvider) GetDefaultSenderID() string {
	return t.senderID
}
