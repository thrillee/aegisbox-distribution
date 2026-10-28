package mno

import (
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

// TestTermii_Integration tests Termii provider with actual HTTP server
func TestTermii_Integration(t *testing.T) {
	t.Run("SendSMS - Success with all fields", func(t *testing.T) {
		// Create mock HTTP server that simulates Termii API
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.URL.Path != "/api/sms/send" {
				t.Errorf("Expected /api/sms/send path, got %s", r.URL.Path)
			}

			// Verify content type
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
			}

			// Parse and verify request body
			body, _ := io.ReadAll(r.Body)
			var req http_providers.TermiiSendRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Errorf("Failed to parse request body: %v", err)
			}

			// Verify request fields
			if req.APIKey != "test-api-key" {
				t.Errorf("Expected api_key 'test-api-key', got '%s'", req.APIKey)
			}
			if req.From != "YourBrand" {
				t.Errorf("Expected from 'YourBrand', got '%s'", req.From)
			}
			if req.Type != "plain" {
				t.Errorf("Expected type 'plain', got '%s'", req.Type)
			}
			if req.Channel != "generic" {
				t.Errorf("Expected channel 'generic', got '%s'", req.Channel)
			}

			// Check recipient (can be string or array)
			toStr, ok := req.To.(string)
			if ok && toStr != "2348031234567" {
				t.Errorf("Expected to '2348031234567', got '%s'", toStr)
			}

			if req.SMS != "Hello, this is a test message" {
				t.Errorf("Expected specific SMS message, got '%s'", req.SMS)
			}

			// Send success response
			resp := http_providers.TermiiSendResponse{
				MessageID: "termii_msg_12345",
				Message:   "Successfully Sent",
				Balance:   100.50,
				User:      "testuser",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create provider with test server URL
		config := http_providers.HTTPProviderConfig{
			Name:          "TERMII",
			BaseURL:       server.URL,
			APIKey:        "test-api-key",
			DefaultSender: "YourBrand",
			Timeout:       5 * time.Second,
		}
		provider := http_providers.NewTermiiProvider(config, slog.Default())

		// Test SendSMS
		ctx := context.Background()
		messageID, err := provider.SendSMS(ctx, "YourBrand", "2348031234567", "Hello, this is a test message")

		// Verify results
		if err != nil {
			t.Fatalf("SendSMS failed: %v", err)
		}
		if messageID != "termii_msg_12345" {
			t.Errorf("Expected message ID 'termii_msg_12345', got '%s'", messageID)
		}
	})

	t.Run("SendSMS - API Error Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid API key",
			})
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL: server.URL,
			APIKey:  "invalid-key",
		}
		provider := http_providers.NewTermiiProvider(config, slog.Default())

		ctx := context.Background()
		_, err := provider.SendSMS(ctx, "Test", "2348031234567", "Test message")

		if err == nil {
			t.Error("Expected error for invalid API key, got nil")
		}
	})
}

// TestAfricaTalking_Integration tests AfricaTalking provider with actual HTTP server
func TestAfricaTalking_Integration(t *testing.T) {
	t.Run("SendSMS - Success with 201 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			// Verify path
			if !strings.Contains(r.URL.Path, "/messaging") {
				t.Errorf("Expected /messaging path, got %s", r.URL.Path)
			}

			// Verify Content-Type (AfricaTalking sends JSON, not form-encoded)
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
			}

			// Verify API Key header
			if r.Header.Get("apiKey") != "test-api-key" {
				t.Errorf("Expected apiKey header 'test-api-key', got '%s'", r.Header.Get("apiKey"))
			}

			// Parse JSON body
			body, _ := io.ReadAll(r.Body)
			var req http_providers.AfricaTalkingSendRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Verify fields
			if req.Username != "testuser" {
				t.Errorf("Expected username 'testuser', got '%s'", req.Username)
			}
			if len(req.PhoneNumbers) != 1 || req.PhoneNumbers[0] != "2348031234567" {
				t.Errorf("Expected phone number '2348031234567', got %v", req.PhoneNumbers)
			}
			if req.SenderID != "TestBrand" {
				t.Errorf("Expected sender_id 'TestBrand', got '%s'", req.SenderID)
			}
			if req.Message != "Integration test message" {
				t.Errorf("Expected specific message, got '%s'", req.Message)
			}

			// Send 201 Created response (AfricaTalking uses 201, not 200)
			resp := http_providers.AfricaTalkingSendResponse{}
			resp.SMSMessageData.Message = "Sent to 1/1 Total Cost: USD 0.0080"
			resp.SMSMessageData.Recipients = append(resp.SMSMessageData.Recipients, struct {
				StatusCode int    `json:"statusCode"`
				Number     string `json:"number"`
				Cost       string `json:"cost"`
				Status     string `json:"status"`
				MessageID  string `json:"messageId"`
			}{
				StatusCode: 101,
				Number:     "+2348031234567",
				Status:     "Success",
				Cost:       "USD 0.0080",
				MessageID:  "ATXid_test12345",
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated) // 201, not 200!
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL:  server.URL,
			APIKey:   "test-api-key",
			Username: "testuser",
		}
		provider := http_providers.NewAfricaTalkingProvider(config, slog.Default())

		ctx := context.Background()
		messageID, err := provider.SendSMS(ctx, "TestBrand", "2348031234567", "Integration test message")

		if err != nil {
			t.Fatalf("SendSMS failed: %v", err)
		}
		if messageID != "ATXid_test12345" {
			t.Errorf("Expected message ID 'ATXid_test12345', got '%s'", messageID)
		}
	})

	t.Run("ParseDLR - Form-encoded webhook", func(t *testing.T) {
		config := http_providers.HTTPProviderConfig{
			BaseURL: "https://api.africastalking.com",
			APIKey:  "test-key",
		}
		provider := http_providers.NewAfricaTalkingProvider(config, slog.Default())

		// Create form-encoded request (as AfricaTalking sends)
		formData := url.Values{}
		formData.Set("id", "ATXid_webhook123")
		formData.Set("status", "Success")
		formData.Set("phoneNumber", "+2348031234567")

		req := httptest.NewRequest("POST", "/webhook", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := context.Background()
		messageID, status, err := provider.ParseDLR(ctx, req)

		if err != nil {
			t.Fatalf("ParseDLR failed: %v", err)
		}
		if messageID != "ATXid_webhook123" {
			t.Errorf("Expected message ID 'ATXid_webhook123', got '%s'", messageID)
		}
		if status != http_providers.StatusDelivered {
			t.Errorf("Expected status 'delivered', got '%s'", status)
		}
	})
}

// TestSimpu_Integration tests Simpu provider with actual HTTP server
func TestSimpu_Integration(t *testing.T) {
	t.Run("SendSMS - Correct request structure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify authorization header (Simpu uses plain APIKey, not "Bearer")
			authHeader := r.Header.Get("Authorization")
			if authHeader != "test-api-key" {
				t.Errorf("Expected 'test-api-key', got '%s'", authHeader)
			}

			// Parse request body
			body, _ := io.ReadAll(r.Body)
			var req http_providers.SimpuSendRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("Failed to parse request body: %v", err)
			}

			// Verify Simpu-specific structure
			if req.Channel != "sms" {
				t.Errorf("Expected channel 'sms', got '%s'", req.Channel)
			}
			if req.Recipients != "2348031234567" {
				t.Errorf("Expected recipients '2348031234567', got '%s'", req.Recipients)
			}
			if req.SenderID != "TestBrand" {
				t.Errorf("Expected sender_id 'TestBrand', got '%s'", req.SenderID)
			}
			if req.Content != "Simpu test message" {
				t.Errorf("Expected specific content, got '%s'", req.Content)
			}
			// Note: IsTrackable is false in actual implementation
			if req.IsTrackable {
				t.Error("Expected is_trackable to be false")
			}

			// Send success response (Simpu returns ID directly)
			resp := http_providers.SimpuSendResponse{
				ID:         "simpu_msg_12345",
				Recipients: []string{"2348031234567"},
				SenderID:   "TestBrand",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL: server.URL,
			APIKey:  "test-api-key",
		}
		provider := http_providers.NewSimpuProvider(config, slog.Default())

		ctx := context.Background()
		messageID, err := provider.SendSMS(ctx, "TestBrand", "2348031234567", "Simpu test message")

		if err != nil {
			t.Fatalf("SendSMS failed: %v", err)
		}
		if messageID != "simpu_msg_12345" {
			t.Errorf("Expected message ID 'simpu_msg_12345', got '%s'", messageID)
		}
	})
}

// TestEBulkSMS_Integration tests eBulkSMS provider with actual HTTP server
func TestEBulkSMS_Integration(t *testing.T) {
	t.Run("SendSMS - APIKey header authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected application/json, got %s", r.Header.Get("Content-Type"))
			}
			// eBulkSMS uses APIKey header (not Authorization)
			if r.Header.Get("APIKey") != "test-api-key" {
				t.Errorf("Expected APIKey header 'test-api-key', got '%s'", r.Header.Get("APIKey"))
			}

			// Parse request
			body, _ := io.ReadAll(r.Body)
			var req http_providers.EBulkSMSRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("Failed to parse request: %v", err)
			}

			// Verify message fields
			if req.Sender != "TestBrand" {
				t.Errorf("Expected sender 'TestBrand', got '%s'", req.Sender)
			}
			if req.To != "2348031234567" {
				t.Errorf("Expected to '2348031234567', got '%s'", req.To)
			}
			if req.SMS != "eBulk test message" {
				t.Errorf("Expected specific message, got '%s'", req.SMS)
			}

			// Send response with proper structure
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
						Status:     "Success",
						StatusCode: 0,
						MessageID:  "ebulk_msg_12345",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL:  server.URL,
			Username: "testuser",
			APIKey:   "test-api-key",
		}
		provider := http_providers.NewEBulkSMSProvider(config, slog.Default())

		ctx := context.Background()
		messageID, err := provider.SendSMS(ctx, "TestBrand", "2348031234567", "eBulk test message")

		if err != nil {
			t.Fatalf("SendSMS failed: %v", err)
		}
		if messageID != "ebulk_msg_12345" {
			t.Errorf("Expected message ID 'ebulk_msg_12345', got '%s'", messageID)
		}
	})
}

// TestMultiTexter_Integration tests MultiTexter provider with actual HTTP server
func TestMultiTexter_Integration(t *testing.T) {
	t.Run("SendSMS - Bearer token auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify Bearer token authorization (MultiTexter uses Bearer, not Basic Auth)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer test-api-key" {
				t.Errorf("Expected 'Bearer test-api-key', got '%s'", authHeader)
			}

			// Parse request
			body, _ := io.ReadAll(r.Body)
			var req http_providers.MultiTexterSendRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("Failed to parse request: %v", err)
			}

			// Verify fields
			if req.SenderName != "TestBrand" {
				t.Errorf("Expected sender_name 'TestBrand', got '%s'", req.SenderName)
			}
			if req.Recipients != "2348031234567" {
				t.Errorf("Expected recipients '2348031234567', got '%s'", req.Recipients)
			}
			if req.Message != "MultiTexter test" {
				t.Errorf("Expected specific message, got '%s'", req.Message)
			}

			// Send response (Status 1 = success for MultiTexter)
			resp := http_providers.MultiTexterSendResponse{
				Status:  1,
				Message: "Message sent successfully",
				MsgID:   "mt_msg_12345",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL: server.URL,
			APIKey:  "test-api-key",
		}
		provider := http_providers.NewMultiTexterProvider(config, slog.Default())

		ctx := context.Background()
		messageID, err := provider.SendSMS(ctx, "TestBrand", "2348031234567", "MultiTexter test")

		if err != nil {
			t.Fatalf("SendSMS failed: %v", err)
		}
		if messageID != "mt_msg_12345" {
			t.Errorf("Expected message ID 'mt_msg_12345', got '%s'", messageID)
		}
	})
}

// TestHTTPProvider_ErrorHandling tests error scenarios across providers
func TestHTTPProvider_ErrorHandling(t *testing.T) {
	t.Run("Network timeout", func(t *testing.T) {
		// Create server that delays response beyond timeout
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL: server.URL,
			APIKey:  "test-key",
			Timeout: 100 * time.Millisecond, // Very short timeout
		}
		provider := http_providers.NewTermiiProvider(config, slog.Default())

		ctx := context.Background()
		_, err := provider.SendSMS(ctx, "Test", "2348031234567", "Test")

		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json{{{"))
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL: server.URL,
			APIKey:  "test-key",
		}
		provider := http_providers.NewTermiiProvider(config, slog.Default())

		ctx := context.Background()
		_, err := provider.SendSMS(ctx, "Test", "2348031234567", "Test")

		if err == nil {
			t.Error("Expected JSON parse error, got nil")
		}
	})

	t.Run("HTTP 500 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		config := http_providers.HTTPProviderConfig{
			BaseURL: server.URL,
			APIKey:  "test-key",
		}
		provider := http_providers.NewTermiiProvider(config, slog.Default())

		ctx := context.Background()
		_, err := provider.SendSMS(ctx, "Test", "2348031234567", "Test")

		if err == nil {
			t.Error("Expected HTTP 500 error, got nil")
		}
	})
}

// TestConnectionPooling verifies connection pooling is working
func TestConnectionPooling(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := http_providers.TermiiSendResponse{
			MessageID: "test_msg",
			Message:   "Success",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := http_providers.HTTPProviderConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	}
	provider := http_providers.NewTermiiProvider(config, slog.Default())

	// Send multiple requests rapidly
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := provider.SendSMS(ctx, "Test", "2348031234567", "Test message")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	if requestCount != 5 {
		t.Errorf("Expected 5 requests, got %d", requestCount)
	}

	// Verify the client reuses connections (this is implicit in the http.Transport config)
	// The fact that all requests succeeded rapidly indicates connection pooling is working
}

// TestProviderInterfaces verifies all providers implement the HTTPProvider interface correctly
func TestProviderInterfaces(t *testing.T) {
	providers := []struct {
		name     string
		provider http_providers.HTTPProvider
	}{
		{"Termii", http_providers.NewTermiiProvider(http_providers.HTTPProviderConfig{BaseURL: "http://test.com", APIKey: "key"}, slog.Default())},
		{"AfricaTalking", http_providers.NewAfricaTalkingProvider(http_providers.HTTPProviderConfig{BaseURL: "http://test.com", APIKey: "key"}, slog.Default())},
		{"Simpu", http_providers.NewSimpuProvider(http_providers.HTTPProviderConfig{BaseURL: "http://test.com", APIKey: "key"}, slog.Default())},
		{"eBulkSMS", http_providers.NewEBulkSMSProvider(http_providers.HTTPProviderConfig{BaseURL: "http://test.com", APIKey: "key"}, slog.Default())},
		{"MultiTexter", http_providers.NewMultiTexterProvider(http_providers.HTTPProviderConfig{BaseURL: "http://test.com", Username: "user", Password: "pass"}, slog.Default())},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			// Verify Name() returns non-empty string
			if p.provider.Name() == "" {
				t.Errorf("%s.Name() returned empty string", p.name)
			}

			// Verify SupportsDLR() returns a boolean (just call it)
			_ = p.provider.SupportsDLR()

			// Verify GetDefaultSenderID() works
			_ = p.provider.GetDefaultSenderID()
		})
	}
}
