package http_providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type SimpuOTPProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	senderID string
	logger   *slog.Logger
}

func NewSimpuOTPProvider(config HTTPProviderConfig, logger *slog.Logger) *SimpuOTPProvider {
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

	return &SimpuOTPProvider{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		apiKey:   config.APIKey,
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		senderID: config.DefaultSender,
		logger:   logger,
	}
}

type SimpuOTPSendRequest struct {
	Recipient string `json:"recipient"`
	Length    int    `json:"length"`
	ExpiresIn int    `json:"expires_in"`
	Token     string `json:"token"`
	Channel   string `json:"channel"`
}

type SimpuOTPSendResponse struct {
	Data struct {
		OTP struct {
			CreatedDatetime string `json:"created_datetime"`
			ExpiresIn       int    `json:"expires_in"`
			ID              string `json:"id"`
			Length          int    `json:"length"`
			OrganisationID  string `json:"organisation_id"`
			Recipient       int64  `json:"recipient"`
			Status          int    `json:"status"`
			StatusDesc      string `json:"status_desc"`
			ValidUntil      string `json:"valid_until"`
		} `json:"otp"`
	} `json:"data"`
	Message string `json:"message"`
	Meta    struct {
		Charge struct {
			Amount           float64 `json:"amount"`
			AvailableBalance float64 `json:"available_balance"`
			BalanceBefore    float64 `json:"balance_before"`
			Encoding         string  `json:"encoding"`
		} `json:"charge"`
		Method   string `json:"method"`
		Resource string `json:"resource"`
	} `json:"meta"`
	Status int    `json:"status"`
	Type   string `json:"type"`
}

func (s *SimpuOTPProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	s.logger.InfoContext(ctx, "Sending OTP via Simpu OTP API", "from", from, "to", to)

	// Extract OTP from message using regex (looks for 4-6 digit numbers)
	otpToken, otpLength := extractOTPFromMessage(message)
	if otpToken == "" {
		s.logger.WarnContext(ctx, "No OTP found in message, using default 4-digit", "message", message)
		otpToken = "1234" // fallback
		otpLength = 4
	}

	// Clean phone number (remove any non-digits)
	recipient := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, to)

	reqBody := SimpuOTPSendRequest{
		Recipient: recipient,
		Length:    otpLength,
		ExpiresIn: 15, // Default 15 minutes
		Token:     otpToken,
		Channel:   "sms",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/otp/request_token", bytes.NewBuffer(jsonBody))
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", s.apiKey)
	}

	s.logger.DebugContext(ctx, "Sending HTTP request to Simpu OTP API", "url", req.URL.String())

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.ErrorContext(ctx, "HTTP request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to read response", "error", err)
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	s.logger.DebugContext(ctx, "Simpu OTP Response", "status", resp.StatusCode, "response", string(body))

	if resp.StatusCode != http.StatusOK {
		s.logger.WarnContext(ctx, "Simpu OTP API returned non-200 status", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response SimpuOTPSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Data.OTP.ID == "" {
		s.logger.ErrorContext(ctx, "Send failed: no OTP ID in response")
		return "", fmt.Errorf("send failed: no OTP ID in response")
	}

	s.logger.InfoContext(ctx, "OTP sent successfully via Simpu OTP API",
		"otp_id", response.Data.OTP.ID,
		"status", response.Data.OTP.StatusDesc,
		"recipient", response.Data.OTP.Recipient)

	return response.Data.OTP.ID, nil
}

func (s *SimpuOTPProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

func (s *SimpuOTPProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	// Simpu OTP webhooks follow similar structure to regular Simpu webhooks
	var webhook struct {
		Data struct {
			Object struct {
				Callback             string `json:"callback"`
				Status               string `json:"status"`
				Retries              int    `json:"retries"`
				Object               string `json:"object"`
				ExternalAggregateRef string `json:"external_aggregate_ref"`
				OrganisationID       string `json:"organisation_id"`
				CreatedDatetime      string `json:"created_datetime"`
				TransactionID        string `json:"transaction_id"`
				OTPID                string `json:"otp_id"` // OTP-specific field
			} `json:"object"`
		} `json:"data"`
		ID        string `json:"id"`
		Object    string `json:"object"`
		Type      string `json:"type"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(req.Body).Decode(&webhook); err != nil {
		s.logger.ErrorContext(ctx, "Failed to decode Simpu OTP DLR webhook", "error", err)
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	// Try OTP ID first, then TransactionID
	messageID := webhook.Data.Object.OTPID
	if messageID == "" {
		messageID = webhook.Data.Object.TransactionID
	}
	if messageID == "" {
		// Fallback to extracting from callback URL
		messageID = extractMessageIDFromCallback(webhook.Data.Object.Callback)
	}
	if messageID == "" {
		// Last resort: use event ID
		messageID = webhook.ID
	}

	status := webhook.Data.Object.Status

	s.logger.InfoContext(ctx, "Simpu OTP DLR webhook received",
		"message_id", messageID,
		"otp_id", webhook.Data.Object.OTPID,
		"transaction_id", webhook.Data.Object.TransactionID,
		"status", status,
		"event_id", webhook.ID,
		"type", webhook.Type)

	return messageID, NormalizeStatus("SIMPU_OTP", status), nil
}

func (s *SimpuOTPProvider) Name() string {
	return "SIMPU_OTP"
}

func (s *SimpuOTPProvider) SupportsDLR() bool {
	return true
}

func (s *SimpuOTPProvider) SupportsAutoDLR() bool {
	return true
}

func (s *SimpuOTPProvider) GetDefaultSenderID() string {
	return s.senderID
}

// Helper function to extract OTP token from message
// Looks for 4-6 digit numbers in the message
func extractOTPFromMessage(message string) (string, int) {
	// Look for patterns like "Your OTP is 1234" or "Code: 123456"
	re := regexp.MustCompile(`\b(\d{4,6})\b`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		token := matches[1]
		return token, len(token)
	}
	return "", 0
}
