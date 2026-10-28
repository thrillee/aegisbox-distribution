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

type SimpuProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	senderID string
	logger   *slog.Logger
}

func NewSimpuProvider(config HTTPProviderConfig, logger *slog.Logger) *SimpuProvider {
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

	return &SimpuProvider{
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

type SimpuSendRequest struct {
	Channel     string `json:"channel"`
	Recipients  string `json:"recipients"`
	SenderID    string `json:"sender_id"`
	Content     string `json:"content"`
	IsTrackable bool   `json:"is_trackable"`
}

type SimpuSendResponse struct {
	ID         string   `json:"id"`
	Recipients []string `json:"recipients"`
	SenderID   string   `json:"sender_id"`
}

func (s *SimpuProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	s.logger.InfoContext(ctx, "Sending SMS via Simpu", "from", from, "to", to)

	if from == "" {
		from = s.senderID
	}

	// Prepare recipients
	recipients := strings.Split(to, ",")
	for i, num := range recipients {
		recipients[i] = strings.TrimSpace(num)
	}

	reqBody := SimpuSendRequest{
		Channel:     "sms",
		Recipients:  strings.Join(recipients, ", "),
		SenderID:    from,
		Content:     message,
		IsTrackable: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/sms/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", s.apiKey)
	}

	s.logger.DebugContext(ctx, "Sending HTTP request to Simpu", "url", req.URL.String())

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

	s.logger.DebugContext(ctx, "Simpu Response", "status", resp.StatusCode, "response", string(body))

	if resp.StatusCode != http.StatusOK {
		s.logger.WarnContext(ctx, "Simpu API returned non-200 status", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response SimpuSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.ID == "" {
		s.logger.ErrorContext(ctx, "Send failed: no message ID in response")
		return "", fmt.Errorf("send failed: no message ID in response")
	}

	s.logger.InfoContext(ctx, "SMS sent successfully via Simpu", "message_id", response.ID)
	return response.ID, nil
}

func (s *SimpuProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

func (s *SimpuProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	// Simpu sends webhooks in this format
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
			} `json:"object"`
		} `json:"data"`
		ID        string `json:"id"`
		Object    string `json:"object"`
		Type      string `json:"type"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(req.Body).Decode(&webhook); err != nil {
		s.logger.ErrorContext(ctx, "Failed to decode Simpu DLR webhook", "error", err)
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	// Use TransactionID as the message ID (this is what we sent as the message ID)
	messageID := webhook.Data.Object.TransactionID
	if messageID == "" {
		// Fallback to extracting from callback URL
		messageID = extractMessageIDFromCallback(webhook.Data.Object.Callback)
	}
	if messageID == "" {
		// Last resort: use event ID
		messageID = webhook.ID
	}

	status := webhook.Data.Object.Status

	s.logger.InfoContext(ctx, "Simpu DLR webhook received",
		"message_id", messageID,
		"transaction_id", webhook.Data.Object.TransactionID,
		"status", status,
		"event_id", webhook.ID,
		"type", webhook.Type)

	return messageID, NormalizeStatus("SIMPU", status), nil
}

// Helper to extract message ID from Simpu callback URL
func extractMessageIDFromCallback(callback string) string {
	if callback == "" {
		return ""
	}
	parts := strings.Split(callback, "=")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func (s *SimpuProvider) Name() string {
	return "SIMPU"
}

func (s *SimpuProvider) SupportsDLR() bool {
	return true
}

func (s *SimpuProvider) SupportsAutoDLR() bool {
	return true
}

func (s *SimpuProvider) GetDefaultSenderID() string {
	return s.senderID
}
