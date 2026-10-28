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

// AkuukProvider implements HTTP SMS provider for Akuuk API
type AkuukProvider struct {
	client    *http.Client
	baseURL   string
	accountID string // Account ID (stored in Username field)
	apiKey    string // API Key
	senderID  string
	logger    *slog.Logger
}

// NewAkuukProvider creates a new Akuuk provider instance
func NewAkuukProvider(config HTTPProviderConfig, logger *slog.Logger) *AkuukProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// For Akuuk, the API key can be in either APIKey or Password field
	// (depending on how it's configured in the database)
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = config.Password
	}

	return &AkuukProvider{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		baseURL:   strings.TrimRight(config.BaseURL, "/"),
		accountID: config.Username, // Account ID stored in Username
		apiKey:    apiKey,          // API Key from either APIKey or Password field
		senderID:  config.DefaultSender,
		logger:    logger,
	}
}

// AkuukSendRequest represents the request payload for Akuuk SMS API
type AkuukSendRequest struct {
	MRCReference string `json:"mrcReference"` // Our message ID
	CountryCode  string `json:"countryCode"`  // e.g., "234"
	Number       string `json:"number"`       // Full phone with country code
	Sender       string `json:"sender"`       // Sender ID
	Type         string `json:"type"`         // "text" or "flash"
	Message      string `json:"message"`      // SMS content
}

// AkuukSendResponse represents the response from Akuuk SMS API
type AkuukSendResponse struct {
	Code         int    `json:"code"`         // HTTP status code
	Status       string `json:"status"`       // "Success" or "Failed"
	Message      string `json:"message"`      // Response message
	MRCReference string `json:"mrcReference"` // Our mrcReference echoed
	TxnReference int64  `json:"txnReference"` // Akuuk's transaction reference
	TxnStatus    string `json:"txnStatus"`    // Transaction status
	TxnDate      string `json:"txnDate"`      // Transaction date
	GtwStatus    string `json:"gtwStatus"`    // Gateway status
	Number       string `json:"number"`       // Phone number
	Network      string `json:"network"`      // Network name
	Sender       string `json:"sender"`       // Sender ID used
	Type         string `json:"type"`         // Message type
	Content      string `json:"content"`      // Message content
	Cost         string `json:"cost"`         // Message cost
	Currency     string `json:"currency"`     // Currency (e.g., "NGN")
}

// AkuukDLR represents a delivery receipt from Akuuk
type AkuukDLR struct {
	MessageID     string `json:"messageId"`
	Reference     string `json:"reference"`
	Status        string `json:"status"`
	Recipient     string `json:"recipient"`
	Sender        string `json:"sender"`
	DeliveredAt   string `json:"deliveredAt"`
	FailureCode   string `json:"failureCode"`
	FailureReason string `json:"failureReason"`
}

// SendSMS sends an SMS through Akuuk API
func (a *AkuukProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	a.logger.InfoContext(ctx, "Sending SMS via Akuuk", "from", from, "to", to)

	if from == "" {
		from = a.senderID
	}

	// Extract country code and number
	phoneNumber := strings.TrimPrefix(to, "+")

	// Extract country code (first 3 digits for most countries)
	countryCode := ""
	number := phoneNumber
	if len(phoneNumber) >= 10 {
		if phoneNumber[0] == '1' && len(phoneNumber) == 11 {
			countryCode = "1"
			number = phoneNumber
		} else if len(phoneNumber) >= 12 {
			countryCode = phoneNumber[:3]
			number = phoneNumber
		}
	}

	reqBody := AkuukSendRequest{
		MRCReference: fmt.Sprintf("MSG_%d", time.Now().UnixNano()),
		CountryCode:  countryCode,
		Number:       number,
		Sender:       from,
		Type:         "text",
		Message:      message,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messaging/sms", bytes.NewBuffer(jsonBody))
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.SetBasicAuth(a.accountID, a.apiKey)

	fmt.Println("Akuuk Send Request:", string(jsonBody))

	a.logger.DebugContext(ctx, "Sending HTTP request to Akuuk", "url", req.URL.String())

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

	a.logger.InfoContext(ctx, "Akuuk Response", "status", resp.StatusCode, "response", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a.logger.WarnContext(ctx, "Akuuk API returned non-2xx status", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response AkuukSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		a.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if the status is "Success" (Akuuk returns capitalized "Success")
	if !strings.EqualFold(response.Status, "success") {
		a.logger.WarnContext(ctx, "Akuuk send failed", "message", response.Message, "status", response.Status)
		return "", fmt.Errorf("send failed: %s", response.Message)
	}

	// Use TxnReference as the message ID
	messageID := fmt.Sprintf("%d", response.TxnReference)
	if messageID == "0" || messageID == "" {
		// Fallback to MRCReference if TxnReference is not available
		messageID = response.MRCReference
	}

	a.logger.InfoContext(ctx, "SMS sent successfully via Akuuk",
		"message_id", messageID,
		"txn_reference", response.TxnReference,
		"network", response.Network,
		"cost", response.Cost,
		"currency", response.Currency)

	return messageID, nil
}

// CheckStatus checks the delivery status of a message
func (a *AkuukProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return StatusPending, nil
}

// ParseDLR parses a delivery receipt from an HTTP request
func (a *AkuukProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	var dlr AkuukDLR
	if err := json.NewDecoder(req.Body).Decode(&dlr); err != nil {
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	a.logger.InfoContext(ctx, "Akuuk DLR received",
		"message_id", dlr.MessageID,
		"reference", dlr.Reference,
		"status", dlr.Status,
		"recipient", dlr.Recipient)

	messageID := dlr.Reference
	if messageID == "" {
		messageID = dlr.MessageID
	}

	status := NormalizeStatus("AKUUK", dlr.Status)
	return messageID, status, nil
}

// Name returns the provider name
func (a *AkuukProvider) Name() string {
	return "AKUUK"
}

// SupportsDLR returns whether the provider supports delivery receipts
func (a *AkuukProvider) SupportsDLR() bool {
	return true
}

// SupportsAutoDLR returns true since Akuuk auto-sends DLR on 200 OK
func (a *AkuukProvider) SupportsAutoDLR() bool {
	return true
}

// GetDefaultSenderID returns the default sender ID
func (a *AkuukProvider) GetDefaultSenderID() string {
	return a.senderID
}
