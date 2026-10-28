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

type MultiTexterProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	senderID string
	email    string
	password string
	logger   *slog.Logger
}

func NewMultiTexterProvider(config HTTPProviderConfig, logger *slog.Logger) *MultiTexterProvider {
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

	return &MultiTexterProvider{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		apiKey:   config.APIKey,
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		senderID: config.DefaultSender,
		email:    config.Username,
		password: config.Password,
		logger:   logger,
	}
}

type MultiTexterSendRequest struct {
	Message    string `json:"message"`
	SenderName string `json:"sender_name"`
	Recipients string `json:"recipients"`
}

type MultiTexterSendResponse struct {
	Status  int    `json:"status"`
	Message string `json:"msg"`
	MsgID   string `json:"msgid,omitempty"`
}

func (m *MultiTexterProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	m.logger.InfoContext(ctx, "Sending SMS via MultiTexter", "from", from, "to", to)

	if from == "" {
		from = m.senderID
	}

	reqBody := MultiTexterSendRequest{
		Message:    message,
		SenderName: from,
		Recipients: to,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/sendsms", bytes.NewBuffer(jsonBody))
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	m.logger.DebugContext(ctx, "Sending HTTP request to MultiTexter", "url", req.URL.String())

	resp, err := m.client.Do(req)
	if err != nil {
		m.logger.ErrorContext(ctx, "HTTP request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to read response", "error", err)
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	m.logger.DebugContext(ctx, "MultiTexter Response", "status", resp.StatusCode, "response", string(body))

	var response MultiTexterSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		m.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Status != 1 {
		m.logger.WarnContext(ctx, "MultiTexter send failed", "status", response.Status, "message", response.Message)
		return "", fmt.Errorf("send failed: %s (code %d)", response.Message, response.Status)
	}

	m.logger.InfoContext(ctx, "SMS sent successfully via MultiTexter", "msg_id", response.MsgID)
	return response.MsgID, nil
}

func (m *MultiTexterProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

func (m *MultiTexterProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	return "", "", fmt.Errorf("MultiTexter does not support webhook DLR, use polling")
}

func (m *MultiTexterProvider) Name() string {
	return "MULTITEXTER"
}

func (m *MultiTexterProvider) SupportsDLR() bool {
	return false
}

func (m *MultiTexterProvider) SupportsAutoDLR() bool {
	return false
}

func (m *MultiTexterProvider) GetDefaultSenderID() string {
	return m.senderID
}
