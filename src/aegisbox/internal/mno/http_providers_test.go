package mno

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/thrillee/aegisbox/internal/mno/http_providers"
)

// Mock HTTP server for testing
func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// Test logger
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- Termii Provider Tests ---

func TestTermii_SendSMS_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json")
		}

		// Parse and verify request body
		var req http_providers.TermiiSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.From != "TestSender" {
			t.Errorf("Expected from='TestSender', got '%s'", req.From)
		}
		if req.SMS != "Test message" {
			t.Errorf("Expected message='Test message', got '%s'", req.SMS)
		}

		// Send successful response
		resp := http_providers.TermiiSendResponse{
			MessageID: "12345",
			Message:   "Successfully Sent",
			Balance:   100.50,
			User:      "test@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		DefaultSender: "DefaultSender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewTermiiProvider(config, testLogger())

	ctx := context.Background()
	messageID, err := provider.SendSMS(ctx, "TestSender", "2348012345678", "Test message")

	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if messageID != "12345" {
		t.Errorf("Expected messageID='12345', got '%s'", messageID)
	}
}

func TestTermii_SendSMS_APIError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid API key"}`))
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "invalid-key",
		DefaultSender: "Sender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewTermiiProvider(config, testLogger())

	ctx := context.Background()
	_, err := provider.SendSMS(ctx, "", "2348012345678", "Test")

	if err == nil {
		t.Fatal("Expected error for invalid API key, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error to contain status 400, got: %v", err)
	}
}

func TestTermii_ParseDLR_Success(t *testing.T) {
	dlrPayload := http_providers.TermiiDLR{
		Type:      "dlr",
		MessageID: "msg-123",
		Status:    "DELIVERED",
		Receiver:  "2348012345678",
	}

	body, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest("POST", "/dlr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	config := http_providers.HTTPProviderConfig{}
	provider := http_providers.NewTermiiProvider(config, testLogger())

	ctx := context.Background()
	messageID, status, err := provider.ParseDLR(ctx, req)

	if err != nil {
		t.Fatalf("ParseDLR failed: %v", err)
	}
	if messageID != "msg-123" {
		t.Errorf("Expected messageID='msg-123', got '%s'", messageID)
	}
	if status != http_providers.StatusDelivered {
		t.Errorf("Expected status='%s', got '%s'", http_providers.StatusDelivered, status)
	}
}

// --- AfricaTalking Provider Tests ---

func TestAfricaTalking_SendSMS_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("apiKey") == "" {
			t.Error("Expected apiKey header")
		}

		// Verify request
		var req http_providers.AfricaTalkingSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if len(req.PhoneNumbers) != 1 {
			t.Errorf("Expected 1 phone number, got %d", len(req.PhoneNumbers))
		}

		// Send successful response
		resp := http_providers.AfricaTalkingSendResponse{}
		resp.SMSMessageData.Message = "Sent"
		resp.SMSMessageData.Recipients = []struct {
			StatusCode int    `json:"statusCode"`
			Number     string `json:"number"`
			Cost       string `json:"cost"`
			Status     string `json:"status"`
			MessageID  string `json:"messageId"`
		}{
			{
				StatusCode: 101,
				Number:     "2348012345678",
				Cost:       "KES 0.80",
				Status:     "Success",
				MessageID:  "ATXid_123456",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // AfricaTalking returns 201
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		Username:      "sandbox",
		DefaultSender: "TESTAPP",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewAfricaTalkingProvider(config, testLogger())

	ctx := context.Background()
	messageID, err := provider.SendSMS(ctx, "TESTAPP", "2348012345678", "Test message")

	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if messageID != "ATXid_123456" {
		t.Errorf("Expected messageID='ATXid_123456', got '%s'", messageID)
	}
}

func TestAfricaTalking_ParseDLR_Success(t *testing.T) {
	// AfricaTalking sends DLRs as form data
	formData := url.Values{}
	formData.Set("id", "ATXid_123456")
	formData.Set("status", "Success")
	formData.Set("phoneNumber", "+2348012345678")
	formData.Set("networkCode", "62130")

	req := httptest.NewRequest("POST", "/dlr", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	config := http_providers.HTTPProviderConfig{}
	provider := http_providers.NewAfricaTalkingProvider(config, testLogger())

	ctx := context.Background()
	messageID, status, err := provider.ParseDLR(ctx, req)

	if err != nil {
		t.Fatalf("ParseDLR failed: %v", err)
	}
	if messageID != "ATXid_123456" {
		t.Errorf("Expected messageID='ATXid_123456', got '%s'", messageID)
	}
	if status != http_providers.StatusDelivered {
		t.Errorf("Expected status='%s', got '%s'", http_providers.StatusDelivered, status)
	}
}

// --- Simpu Provider Tests ---

func TestSimpu_SendSMS_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req http_providers.SimpuSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Channel != "sms" {
			t.Errorf("Expected channel='sms', got '%s'", req.Channel)
		}
		if req.Content != "Test message" {
			t.Errorf("Expected content='Test message', got '%s'", req.Content)
		}

		resp := http_providers.SimpuSendResponse{
			ID:         "simpu-msg-789",
			Recipients: []string{"2348012345678"},
			SenderID:   "TestSender",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		DefaultSender: "TestSender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewSimpuProvider(config, testLogger())

	ctx := context.Background()
	messageID, err := provider.SendSMS(ctx, "TestSender", "2348012345678", "Test message")

	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if messageID != "simpu-msg-789" {
		t.Errorf("Expected messageID='simpu-msg-789', got '%s'", messageID)
	}
}

// --- Simpu OTP Provider Tests ---

func TestSimpuOTP_SendSMS_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req http_providers.SimpuOTPSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Channel != "sms" {
			t.Errorf("Expected channel='sms', got '%s'", req.Channel)
		}
		if req.Token == "" {
			t.Error("Expected token to be set")
		}
		if req.Length < 4 || req.Length > 6 {
			t.Errorf("Expected length between 4-6, got %d", req.Length)
		}

		resp := http_providers.SimpuOTPSendResponse{}
		resp.Data.OTP.ID = "otp-id-12345"
		resp.Data.OTP.StatusDesc = "SENT"
		resp.Data.OTP.Recipient = 2348012345678
		resp.Message = "Successful"
		resp.Status = 200
		resp.Type = "success"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		DefaultSender: "TestSender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewSimpuOTPProvider(config, testLogger())

	ctx := context.Background()
	// Message contains OTP code
	messageID, err := provider.SendSMS(ctx, "TestSender", "2348012345678", "Your OTP is 123456")

	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if messageID != "otp-id-12345" {
		t.Errorf("Expected messageID='otp-id-12345', got '%s'", messageID)
	}
}

func TestSimpuOTP_ExtractOTP(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedToken string
		expectedLen   int
	}{
		{
			name:          "4-digit OTP",
			message:       "Your OTP is 1234",
			expectedToken: "1234",
			expectedLen:   4,
		},
		{
			name:          "6-digit OTP",
			message:       "Your verification code: 123456",
			expectedToken: "123456",
			expectedLen:   6,
		},
		{
			name:          "5-digit OTP",
			message:       "Code 54321 expires in 10 minutes",
			expectedToken: "54321",
			expectedLen:   5,
		},
		{
			name:          "No OTP",
			message:       "Hello there",
			expectedToken: "",
			expectedLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the internal extractOTPFromMessage function
			// For now, we'll just test the provider behavior
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				var req http_providers.SimpuOTPSendRequest
				json.NewDecoder(r.Body).Decode(&req)

				if tt.expectedToken != "" {
					if req.Token != tt.expectedToken {
						t.Errorf("Expected token='%s', got '%s'", tt.expectedToken, req.Token)
					}
					if req.Length != tt.expectedLen {
						t.Errorf("Expected length=%d, got %d", tt.expectedLen, req.Length)
					}
				}

				resp := http_providers.SimpuOTPSendResponse{}
				resp.Data.OTP.ID = "test-otp-id"
				resp.Data.OTP.StatusDesc = "SENT"
				resp.Status = 200
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			})
			defer server.Close()

			config := http_providers.HTTPProviderConfig{
				BaseURL: server.URL,
				APIKey:  "test-key",
				Timeout: 10 * time.Second,
			}
			provider := http_providers.NewSimpuOTPProvider(config, testLogger())
			_, err := provider.SendSMS(context.Background(), "", "1234567890", tt.message)
			if err != nil {
				t.Fatalf("SendSMS failed: %v", err)
			}
		})
	}
}

// --- eBulkSMS Provider Tests ---

func TestEBulkSMS_SendSMS_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req http_providers.EBulkSMSRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Sender != "TestSender" {
			t.Errorf("Expected sender='TestSender', got '%s'", req.Sender)
		}

		resp := http_providers.EBulkSMSResponse{
			MainError:     "",
			MainCode:      200,
			ResponseCount: 1,
			Response: []struct {
				Status     string `json:"Status"`
				StatusCode int    `json:"Status_Code"`
				MessageID  string `json:"Message_Id"`
			}{
				{
					Status:     "SUCCESS",
					StatusCode: 200,
					MessageID:  "ebulk-msg-456",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		DefaultSender: "TestSender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewEBulkSMSProvider(config, testLogger())

	ctx := context.Background()
	messageID, err := provider.SendSMS(ctx, "TestSender", "2348012345678", "Test message")

	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if messageID != "ebulk-msg-456" {
		t.Errorf("Expected messageID='ebulk-msg-456', got '%s'", messageID)
	}
}

// --- MultiTexter Provider Tests ---

func TestMultiTexter_SendSMS_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Error("Expected Bearer token authorization")
		}

		var req http_providers.MultiTexterSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		resp := http_providers.MultiTexterSendResponse{
			Status:  1,
			Message: "Message sent successfully",
			MsgID:   "mt-msg-999",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		Username:      "test@example.com",
		Password:      "password123",
		DefaultSender: "TestSender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewMultiTexterProvider(config, testLogger())

	ctx := context.Background()
	messageID, err := provider.SendSMS(ctx, "TestSender", "2348012345678", "Test message")

	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if messageID != "mt-msg-999" {
		t.Errorf("Expected messageID='mt-msg-999', got '%s'", messageID)
	}
}

func TestMultiTexter_SendSMS_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := http_providers.MultiTexterSendResponse{
			Status:  0,
			Message: "Insufficient balance",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL:       server.URL,
		APIKey:        "test-api-key",
		DefaultSender: "TestSender",
		Timeout:       10 * time.Second,
	}

	provider := http_providers.NewMultiTexterProvider(config, testLogger())

	ctx := context.Background()
	_, err := provider.SendSMS(ctx, "TestSender", "2348012345678", "Test message")

	if err == nil {
		t.Fatal("Expected error for insufficient balance, got nil")
	}
	if !strings.Contains(err.Error(), "Insufficient balance") {
		t.Errorf("Expected error about insufficient balance, got: %v", err)
	}
}

// --- Status Normalization Tests ---

func TestStatusNormalization(t *testing.T) {
	tests := []struct {
		provider string
		input    string
		expected string
	}{
		{"TERMII", "DELIVERED", http_providers.StatusDelivered},
		{"TERMII", "MESSAGE SENT", http_providers.StatusSent},
		{"TERMII", "MESSAGE FAILED", http_providers.StatusFailed},
		{"TERMII", "DND ACTIVE ON PHONE NUMBER", http_providers.StatusDNDBlocked},

		{"AFRICATALKING", "Success", http_providers.StatusDelivered},
		{"AFRICATALKING", "Sent", http_providers.StatusSent},
		{"AFRICATALKING", "Failed", http_providers.StatusFailed},
		{"AFRICATALKING", "Expired", http_providers.StatusExpired},

		{"SIMPU", "delivered", http_providers.StatusDelivered},
		{"SIMPU", "failed", http_providers.StatusFailed},
		{"SIMPU", "sent", http_providers.StatusSent},

		{"EBULKSMS", "DELIVERED", http_providers.StatusDelivered},
		{"EBULKSMS", "SENT", http_providers.StatusSent},
		{"EBULKSMS", "FAILED", http_providers.StatusFailed},

		{"UNKNOWN_PROVIDER", "anything", http_providers.StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"_"+tt.input, func(t *testing.T) {
			result := http_providers.NormalizeStatus(tt.provider, tt.input)
			if result != tt.expected {
				t.Errorf("normalizeStatus(%s, %s) = %s, want %s",
					tt.provider, tt.input, result, tt.expected)
			}
		})
	}
}

// --- Provider Factory Tests ---

func TestNewHTTPProvider(t *testing.T) {
	config := http_providers.HTTPProviderConfig{
		BaseURL:       "http://test.com",
		APIKey:        "test-key",
		DefaultSender: "Test",
		Timeout:       10 * time.Second,
	}
	logger := testLogger()

	tests := []struct {
		name     string
		expected http_providers.HTTPProvider
	}{
		{"TERMII", http_providers.NewTermiiProvider(config, logger)},
		{"AFRICATALKING", http_providers.NewAfricaTalkingProvider(config, logger)},
		{"SIMPU", http_providers.NewSimpuProvider(config, logger)},
		{"EBULKSMS", http_providers.NewEBulkSMSProvider(config, logger)},
		{"MULTITEXTER", http_providers.NewMultiTexterProvider(config, logger)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := http_providers.NewHTTPProvider(tt.name, config, logger)
			if provider == nil {
				t.Fatalf("NewHTTPProvider(%s) returned nil", tt.name)
			}
			if provider.Name() != tt.name {
				t.Errorf("Expected provider name %s, got %s", tt.name, provider.Name())
			}
		})
	}

	// Test unknown provider
	unknown := http_providers.NewHTTPProvider("UNKNOWN", config, logger)
	if unknown != nil {
		t.Error("Expected nil for unknown provider, got non-nil")
	}
}
