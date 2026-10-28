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

type EBulkSMSProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	senderID string
	logger   *slog.Logger
}

func NewEBulkSMSProvider(config HTTPProviderConfig, logger *slog.Logger) *EBulkSMSProvider {
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

	return &EBulkSMSProvider{
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

type EBulkSMSRequest struct {
	JSON     string `json:"json"`
	SMS      string `json:"sms"`
	Sender   string `json:"sender"`
	To       string `json:"to"`
	Flash    bool   `json:"flash"`
	Append   string `json:"append"`
	URLRoute string `json:"urlRoute"`
}

type EBulkSMSResponse struct {
	MainError     string `json:"main_error"`
	MainCode      int    `json:"main_code"`
	ResponseCount int    `json:"response_count"`
	Response      []struct {
		Status     string `json:"Status"`
		StatusCode int    `json:"Status_Code"`
		MessageID  string `json:"Message_Id"`
	} `json:"Response"`
}

func (e *EBulkSMSProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	e.logger.InfoContext(ctx, "Sending SMS via eBulkSMS", "from", from, "to", to)

	if from == "" {
		from = e.senderID
	}

	reqBody := EBulkSMSRequest{
		SMS:    message,
		Sender: from,
		To:     to,
		Flash:  false,
		Append: "2",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		e.logger.ErrorContext(ctx, "Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/sms/format/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		e.logger.ErrorContext(ctx, "Failed to create request", "error", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("APIKey", e.apiKey)

	e.logger.DebugContext(ctx, "Sending HTTP request to eBulkSMS", "url", req.URL.String())

	resp, err := e.client.Do(req)
	if err != nil {
		e.logger.ErrorContext(ctx, "HTTP request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.logger.ErrorContext(ctx, "Failed to read response", "error", err)
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	e.logger.DebugContext(ctx, "eBulkSMS Response", "status", resp.StatusCode, "response", string(body))

	if resp.StatusCode != http.StatusOK {
		e.logger.WarnContext(ctx, "eBulkSMS API returned non-200 status", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response EBulkSMSResponse
	if err := json.Unmarshal(body, &response); err != nil {
		e.logger.ErrorContext(ctx, "Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.Response) > 0 && response.Response[0].MessageID != "" {
		messageID := response.Response[0].MessageID
		e.logger.InfoContext(ctx, "SMS sent successfully via eBulkSMS", "message_id", messageID)
		return messageID, nil
	}

	e.logger.ErrorContext(ctx, "No message ID returned in response")
	return "", fmt.Errorf("no message ID returned")
}

func (e *EBulkSMSProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

func (e *EBulkSMSProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	var dlr struct {
		MessageID string `json:"Message_Id"`
		Status    string `json:"Status"`
	}
	if err := json.NewDecoder(req.Body).Decode(&dlr); err != nil {
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	e.logger.DebugContext(ctx, "EBulkSMS DLR", "message_id", dlr.MessageID, "status", dlr.Status)
	return dlr.MessageID, NormalizeStatus("EBULKSMS", dlr.Status), nil
}

func (e *EBulkSMSProvider) Name() string {
	return "EBULKSMS"
}

func (e *EBulkSMSProvider) SupportsDLR() bool {
	return true
}

func (e *EBulkSMSProvider) SupportsAutoDLR() bool {
	return false
}

func (e *EBulkSMSProvider) GetDefaultSenderID() string {
	return e.senderID
}
