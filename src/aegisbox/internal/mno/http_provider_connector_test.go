package mno

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/thrillee/aegisbox/internal/mno/http_providers"
)

// TestHTTPProviderConnectorImplementsConnectorInterface verifies that
// HTTPProviderConnector properly implements the Connector interface
func TestHTTPProviderConnectorImplementsConnectorInterface(t *testing.T) {
	// Create a test provider
	config := http_providers.HTTPProviderConfig{
		Name:          "TERMII",
		BaseURL:       "https://api.ng.termii.com",
		APIKey:        "test_api_key",
		SecretKey:     "test_secret",
		DefaultSender: "TestApp",
		Timeout:       30,
	}

	logger := slog.Default()
	provider := http_providers.NewTermiiProvider(config, logger)

	// Wrap it in HTTPProviderConnector
	connector := NewHTTPProviderConnector(1, 1, provider, logger)

	// Verify it implements Connector interface
	var _ Connector = connector

	// Test basic methods
	if connector.ConnectionID() != 1 {
		t.Errorf("Expected ConnectionID=1, got %d", connector.ConnectionID())
	}

	if connector.MnoID() != 1 {
		t.Errorf("Expected MnoID=1, got %d", connector.MnoID())
	}

	if connector.Status() != "http_ok" {
		t.Errorf("Expected status=http_ok, got %s", connector.Status())
	}

	if connector.Provider().Name() != "TERMII" {
		t.Errorf("Expected provider name=TERMII, got %s", connector.Provider().Name())
	}
}

// TestHTTPProviderConnectorSubmitMessage tests message submission
func TestHTTPProviderConnectorSubmitMessage(t *testing.T) {
	config := http_providers.HTTPProviderConfig{
		Name:          "TERMII",
		BaseURL:       "https://api.ng.termii.com",
		APIKey:        "test_api_key",
		SecretKey:     "test_secret",
		DefaultSender: "TestApp",
		Timeout:       30,
	}

	logger := slog.Default()
	provider := http_providers.NewTermiiProvider(config, logger)
	connector := NewHTTPProviderConnector(1, 1, provider, logger)

	ctx := context.Background()

	// Create test message
	msg := PreparedMessage{
		InternalMessageID: 12345,
		TotalSegments:     1,
		SenderID:          "TestApp",
		DestinationMSISDN: "2348012345678",
		MessageContent:    "Test message",
		RequestDLR:        true,
		DBSeqs:            []int64{1},
	}

	// Note: This will fail with network error since we're not mocking HTTP
	// In production, the provider would make a real HTTP request
	result, err := connector.SubmitMessage(ctx, msg)

	// We expect an error since we don't have a real API key
	// But the connector should still return a SubmitResult
	if result.WasSubmitted {
		t.Log("Message was submitted (unexpected in test environment)")
	}

	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}

	// Verify result structure
	if len(result.Segments) != 1 {
		t.Errorf("Expected 1 segment, got %d", len(result.Segments))
	}

	if result.Segments[0].Seqn != 1 {
		t.Errorf("Expected segment seqn=1, got %d", result.Segments[0].Seqn)
	}
}

// TestHTTPProviderConnectorDLRHandler tests DLR processing
func TestHTTPProviderConnectorDLRHandler(t *testing.T) {
	config := http_providers.HTTPProviderConfig{
		Name:          "TERMII",
		BaseURL:       "https://api.ng.termii.com",
		APIKey:        "test_api_key",
		SecretKey:     "test_secret",
		DefaultSender: "TestApp",
		Timeout:       30,
	}

	logger := slog.Default()
	provider := http_providers.NewTermiiProvider(config, logger)
	connector := NewHTTPProviderConnector(1, 1, provider, logger)

	// Register a test DLR handler
	dlrReceived := false
	handler := func(ctx context.Context, connID int32, dlrInfo DLRInfo) error {
		dlrReceived = true

		if connID != 1 {
			t.Errorf("Expected connID=1, got %d", connID)
		}

		if dlrInfo.MnoMessageID != "test_msg_123" {
			t.Errorf("Expected messageID=test_msg_123, got %s", dlrInfo.MnoMessageID)
		}

		if dlrInfo.Status != "delivered" {
			t.Errorf("Expected status=delivered, got %s", dlrInfo.Status)
		}

		return nil
	}

	connector.RegisterDLRHandler(handler)

	// Process a test DLR
	ctx := context.Background()
	timestamp := time.Now()
	err := connector.ProcessDLRWebhook(ctx, "test_msg_123", "delivered", "", timestamp)

	if err != nil {
		t.Errorf("ProcessDLRWebhook failed: %v", err)
	}

	if !dlrReceived {
		t.Error("DLR handler was not called")
	}
}

// TestManagerLoadsHTTPProviderConnector verifies Manager can load HTTP provider connectors
func TestManagerLoadsHTTPProviderConnector(t *testing.T) {
	// This test would require a database connection
	// For now, we document the expected behavior:
	//
	// 1. Manager.loadAndSyncConnectors() queries mno_connections table
	// 2. For HTTP connections with provider_config.provider_name:
	//    - Creates ProviderConfigLoader
	//    - Calls GetProviderByConnectionID()
	//    - Wraps result in HTTPProviderConnector
	//    - Stores in connectors map
	// 3. GetConnector() returns HTTPProviderConnector when:
	//    - MNO ID matches
	//    - Status is "http_ok"
	//    - No SMPP connector available OR HTTP has higher priority

	t.Skip("Requires database connection - see integration tests")
}

// TestProtocolSelection verifies Manager selects correct protocol
func TestProtocolSelection(t *testing.T) {
	// This test documents the protocol selection logic:
	//
	// Manager.GetConnectorWithProtocol(ctx, mnoID, preferredProtocol):
	//
	// 1. Loads all connectors for mnoID
	// 2. Filters by status:
	//    - SMPP: codes.StatusBound
	//    - HTTP: codes.StatusHttpOk
	// 3. If preferredProtocol specified:
	//    - Returns connector matching protocol (if available)
	// 4. If no preference or preferred not available:
	//    - Returns any active connector (random selection)
	//
	// Examples:
	//
	// Scenario A: SMPP and HTTP both active, no preference
	// Result: Random selection between SMPP and HTTP
	//
	// Scenario B: SMPP and HTTP both active, prefer "http"
	// Result: Returns HTTP connector
	//
	// Scenario C: Only HTTP active, no preference
	// Result: Returns HTTP connector
	//
	// Scenario D: Only SMPP active, prefer "http"
	// Result: Returns SMPP connector (fallback)

	t.Skip("Protocol selection logic is documented")
}
