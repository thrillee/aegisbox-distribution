package mno

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

type HTTPProvider interface {
	SendSMS(ctx context.Context, from, to, message string) (string, error)
	CheckStatus(ctx context.Context, messageID string) (string, error)
	ParseDLR(ctx context.Context, req *http.Request) (messageID, status string, err error)
	Name() string
	SupportsDLR() bool
	GetDefaultSenderID() string
}

type HTTPProviderConfig struct {
	Name           string
	BaseURL        string
	APIKey         string
	Username       string
	Password       string
	SecretKey      string
	DefaultSender  string
	RateLimit      int
	Timeout        time.Duration
	ProviderConfig map[string]interface{}
}

const (
	StatusDelivered   = "delivered"
	StatusSent        = "sent"
	StatusPending     = "pending"
	StatusFailed      = "failed"
	StatusExpired     = "expired"
	StatusUnreachable = "unreachable"
	StatusDNDBlocked  = "dnd_blocked"
	StatusUnknown     = "unknown"
)

type TermiiProvider struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	senderID  string
	secretKey string
	logger    *slog.Logger
}

func NewTermiiProvider(config HTTPProviderConfig, logger *slog.Logger) *TermiiProvider {
	return &TermiiProvider{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey:    config.APIKey,
		baseURL:   strings.TrimRight(config.BaseURL, "/"),
		senderID:  config.DefaultSender,
		secretKey: config.SecretKey,
		logger:    logger,
	}
}

type TermiiSendRequest struct {
	APIKey  string `json:"api_key"`
	To      any    `json:"to"`
	From    string `json:"from"`
	SMS     string `json:"sms"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

type TermiiSendResponse struct {
	MessageID string  `json:"message_id"`
	Message   string  `json:"message"`
	Balance   float64 `json:"balance"`
	User      string  `json:"user"`
}

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

func (t *TermiiProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	if from == "" {
		from = t.senderID
	}

	var recipients any = to
	if strings.Contains(to, ",") {
		recipients = strings.Split(to, ",")
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
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/api/sms/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response TermiiSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.MessageID == "" {
		return "", fmt.Errorf("send failed: %s", response.Message)
	}

	return response.MessageID, nil
}

func (t *TermiiProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

func (t *TermiiProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	var dlr TermiiDLR
	if err := json.NewDecoder(req.Body).Decode(&dlr); err != nil {
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	t.logger.InfoContext(ctx, "Termii Received Webhook", "dlr", dlr)

	status := normalizeStatus("TERMII", dlr.Status)
	return dlr.MessageID, status, nil
}

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

func (t *TermiiProvider) Name() string {
	return "TERMII"
}

func (t *TermiiProvider) SupportsDLR() bool {
	return true
}

func (t *TermiiProvider) GetDefaultSenderID() string {
	return t.senderID
}

type AfricaTalkingProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	username string
	senderID string
	logger   *slog.Logger
}

func NewAfricaTalkingProvider(config HTTPProviderConfig, logger *slog.Logger) *AfricaTalkingProvider {
	return &AfricaTalkingProvider{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey:   config.APIKey,
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		username: config.Username,
		senderID: config.DefaultSender,
		logger:   logger,
	}
}

type AfricaTalkingSendRequest struct {
	Username     string   `json:"username"`
	Message      string   `json:"message"`
	SenderID     string   `json:"senderId"`
	PhoneNumbers []string `json:"phoneNumbers"`
}

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

func (a *AfricaTalkingProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	if from == "" {
		from = a.senderID
	}

	recipients := strings.Split(to, ",")

	reqBody := AfricaTalkingSendRequest{
		Username:     a.username,
		Message:      message,
		SenderID:     from,
		PhoneNumbers: recipients,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/version1/messaging/bulk", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apiKey", a.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	a.logger.DebugContext(ctx, "AfricaTalking Response", "response", string(body))

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response AfricaTalkingSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.SMSMessageData.Recipients) == 0 {
		return "", fmt.Errorf("no recipients processed")
	}

	return response.SMSMessageData.Recipients[0].MessageID, nil
}

func (a *AfricaTalkingProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

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

	return messageID, normalizeStatus("AFRICATALKING", status), nil
}

func (a *AfricaTalkingProvider) Name() string {
	return "AFRICATALKING"
}

func (a *AfricaTalkingProvider) SupportsDLR() bool {
	return true
}

func (a *AfricaTalkingProvider) GetDefaultSenderID() string {
	return a.senderID
}

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
	return &MultiTexterProvider{
		client: &http.Client{
			Timeout: config.Timeout,
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
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/sendsms", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	m.logger.DebugContext(ctx, "MultiTexter Response", "response", string(body))

	var response MultiTexterSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Status != 1 {
		return "", fmt.Errorf("send failed: %s (code %d)", response.Message, response.Status)
	}

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

func (m *MultiTexterProvider) GetDefaultSenderID() string {
	return m.senderID
}

type SimpuProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	senderID string
	logger   *slog.Logger
}

func NewSimpuProvider(config HTTPProviderConfig, logger *slog.Logger) *SimpuProvider {
	return &SimpuProvider{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey:   config.APIKey,
		baseURL:  strings.TrimRight(config.BaseURL, "/"),
		senderID: config.DefaultSender,
		logger:   logger,
	}
}

type SimpuSendRequest struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Message string `json:"message"`
}

type SimpuSendResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

func (s *SimpuProvider) SendSMS(ctx context.Context, from, to, message string) (string, error) {
	if from == "" {
		from = s.senderID
	}

	reqBody := SimpuSendRequest{
		To:      to,
		From:    from,
		Message: message,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/sms/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response SimpuSendResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return response.MessageID, nil
}

func (s *SimpuProvider) CheckStatus(ctx context.Context, messageID string) (string, error) {
	return "pending", nil
}

func (s *SimpuProvider) ParseDLR(ctx context.Context, req *http.Request) (string, string, error) {
	var dlr struct {
		MessageID string `json:"message_id"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(req.Body).Decode(&dlr); err != nil {
		return "", "", fmt.Errorf("failed to decode DLR: %v", err)
	}

	s.logger.DebugContext(ctx, "Simpu DLR", "message_id", dlr.MessageID, "status", dlr.Status)
	return dlr.MessageID, normalizeStatus("SIMPU", dlr.Status), nil
}

func (s *SimpuProvider) Name() string {
	return "SIMPU"
}

func (s *SimpuProvider) SupportsDLR() bool {
	return true
}

func (s *SimpuProvider) GetDefaultSenderID() string {
	return s.senderID
}

type EBulkSMSProvider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	senderID string
	logger   *slog.Logger
}

func NewEBulkSMSProvider(config HTTPProviderConfig, logger *slog.Logger) *EBulkSMSProvider {
	return &EBulkSMSProvider{
		client: &http.Client{
			Timeout: config.Timeout,
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
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/sms/format/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("APIKey", e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response EBulkSMSResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.Response) > 0 {
		return response.Response[0].MessageID, nil
	}

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
	return dlr.MessageID, normalizeStatus("EBULKSMS", dlr.Status), nil
}

func (e *EBulkSMSProvider) Name() string {
	return "EBULKSMS"
}

func (e *EBulkSMSProvider) SupportsDLR() bool {
	return true
}

func (e *EBulkSMSProvider) GetDefaultSenderID() string {
	return e.senderID
}

func NewHTTPProvider(name string, config HTTPProviderConfig, logger *slog.Logger) HTTPProvider {
	switch strings.ToUpper(name) {
	case "TERMII":
		return NewTermiiProvider(config, logger)
	case "AFRICATALKING":
		return NewAfricaTalkingProvider(config, logger)
	case "MULTITEXTER":
		return NewMultiTexterProvider(config, logger)
	case "SIMPU":
		return NewSimpuProvider(config, logger)
	case "EBULKSMS":
		return NewEBulkSMSProvider(config, logger)
	default:
		return nil
	}
}

func normalizeStatus(provider, rawStatus string) string {
	normalized := strings.ToLower(strings.TrimSpace(rawStatus))

	switch provider {
	case "SIMPU":
		switch normalized {
		case "success", "delivered":
			return StatusDelivered
		case "failed", "undelivered":
			return StatusFailed
		case "pending", "sent":
			return StatusSent
		default:
			return StatusUnknown
		}
	case "TERMII":
		switch normalized {
		case "delivered":
			return StatusDelivered
		case "message sent":
			return StatusSent
		case "message failed", "rejected", "expired":
			return StatusFailed
		case "dnd active on phone number":
			return StatusDNDBlocked
		default:
			return StatusUnknown
		}
	case "MULTITEXTER":
		switch normalized {
		case "deliveredpp", "delivered":
			return StatusDelivered
		case "failed":
			return StatusFailed
		case "pending":
			return StatusPending
		case "sent":
			return StatusSent
		default:
			return StatusUnknown
		}
	case "AFRICATALKING":
		switch normalized {
		case "success":
			return StatusDelivered
		case "sent", "submitted", "buffered":
			return StatusSent
		case "rejected", "failed":
			return StatusFailed
		case "absentsubscriber":
			return StatusUnreachable
		case "expired":
			return StatusExpired
		default:
			return StatusUnknown
		}
	case "EBULKSMS":
		switch normalized {
		case "delivered":
			return StatusDelivered
		case "sent":
			return StatusSent
		case "failed", "rejected":
			return StatusFailed
		case "buffered":
			return StatusPending
		case "expired":
			return StatusExpired
		default:
			return StatusUnknown
		}
	default:
		return StatusUnknown
	}
}
