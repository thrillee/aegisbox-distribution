package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thrillee/aegisbox/internal/database"
)

// MockQuerier implements a mock database querier for testing
type MockQuerier struct {
	database.Querier
	segments           map[string]database.FindSegmentByMnoMessageIDRow
	updatedSegments    map[int64]database.UpdateSegmentDLRParams
	routingAssignments map[int32]database.GetRoutingAssignmentByIDRow
	findSegmentError   error
	updateSegmentError error
	getAssignmentError error
}

func NewMockQuerier() *MockQuerier {
	return &MockQuerier{
		segments:           make(map[string]database.FindSegmentByMnoMessageIDRow),
		updatedSegments:    make(map[int64]database.UpdateSegmentDLRParams),
		routingAssignments: make(map[int32]database.GetRoutingAssignmentByIDRow),
	}
}

func (m *MockQuerier) FindSegmentByMnoMessageID(ctx context.Context, mnoMessageID *string) (database.FindSegmentByMnoMessageIDRow, error) {
	if m.findSegmentError != nil {
		return database.FindSegmentByMnoMessageIDRow{}, m.findSegmentError
	}
	if seg, ok := m.segments[*mnoMessageID]; ok {
		return seg, nil
	}
	// Return error when segment not found (similar to pgx.ErrNoRows)
	return database.FindSegmentByMnoMessageIDRow{}, fmt.Errorf("segment not found")
}

func (m *MockQuerier) UpdateSegmentDLR(ctx context.Context, params database.UpdateSegmentDLRParams) error {
	if m.updateSegmentError != nil {
		return m.updateSegmentError
	}
	m.updatedSegments[params.ID] = params
	return nil
}

func (m *MockQuerier) GetRoutingAssignmentByID(ctx context.Context, id int32) (database.GetRoutingAssignmentByIDRow, error) {
	if m.getAssignmentError != nil {
		return database.GetRoutingAssignmentByIDRow{}, m.getAssignmentError
	}
	if assignment, ok := m.routingAssignments[id]; ok {
		return assignment, nil
	}
	return database.GetRoutingAssignmentByIDRow{}, nil
}

func setupTestRouter(handler *WebhookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	webhookGroup := router.Group("/api/v1/webhooks/dlr")
	{
		webhookGroup.POST("/generic", handler.HandleProviderWebhook)
		webhookGroup.POST("/termii", handler.HandleTermiiWebhook)
		webhookGroup.POST("/africatalking", handler.HandleAfricaTalkingWebhook)
		webhookGroup.POST("/simpu", handler.HandleSimpuWebhook)
		webhookGroup.POST("/ebulksms", handler.HandleEBulkSMSWebhook)
	}

	return router
}

func TestTermiiWebhook_Success(t *testing.T) {
	// Setup
	mockDB := NewMockQuerier()
	segmentID := int64(123)
	messageID := "termii_msg_123"
	routingAssignmentID := int32(1)
	mnoID := int32(10)

	mockDB.segments[messageID] = database.FindSegmentByMnoMessageIDRow{
		SegmentID:           segmentID,
		MessageID:           456,
		SegmentSeqn:         1,
		RoutingAssignmentID: &routingAssignmentID,
		RoutedMnoID:         &mnoID,
	}

	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request
	dlrPayload := map[string]interface{}{
		"type":       "dlr",
		"message_id": messageID,
		"status":     "Delivered",
		"receiver":   "2348031234567",
		"sender":     "TestBrand",
	}

	jsonBody, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/termii?conn_id=1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the segment was updated
	updatedParams, ok := mockDB.updatedSegments[segmentID]
	require.True(t, ok, "Segment should have been updated")
	assert.NotNil(t, updatedParams.DlrStatus)
	assert.Equal(t, "delivered", *updatedParams.DlrStatus)
}

func TestAfricaTalkingWebhook_Success(t *testing.T) {
	// Setup
	mockDB := NewMockQuerier()
	segmentID := int64(124)
	messageID := "ATXid_test123"
	routingAssignmentID := int32(1)
	mnoID := int32(10)

	mockDB.segments[messageID] = database.FindSegmentByMnoMessageIDRow{
		SegmentID:           segmentID,
		MessageID:           457,
		SegmentSeqn:         1,
		RoutingAssignmentID: &routingAssignmentID,
		RoutedMnoID:         &mnoID,
	}

	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request (form-encoded)
	formData := url.Values{}
	formData.Set("id", messageID)
	formData.Set("status", "Success")
	formData.Set("phoneNumber", "+2348031234567")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/africatalking?conn_id=1", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the segment was updated
	updatedParams, ok := mockDB.updatedSegments[segmentID]
	require.True(t, ok, "Segment should have been updated")
	assert.NotNil(t, updatedParams.DlrStatus)
	assert.Equal(t, "delivered", *updatedParams.DlrStatus)
}

func TestSimpuWebhook_Success(t *testing.T) {
	// Setup
	mockDB := NewMockQuerier()
	segmentID := int64(125)
	messageID := "simpu_msg_123"
	routingAssignmentID := int32(1)
	mnoID := int32(10)

	mockDB.segments[messageID] = database.FindSegmentByMnoMessageIDRow{
		SegmentID:           segmentID,
		MessageID:           458,
		SegmentSeqn:         1,
		RoutingAssignmentID: &routingAssignmentID,
		RoutedMnoID:         &mnoID,
	}

	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request
	dlrPayload := map[string]interface{}{
		"event": "message.status.update",
		"data": map[string]interface{}{
			"message_id": messageID,
			"status":     "delivered",
			"timestamp":  "2024-01-15T10:30:00Z",
		},
	}

	jsonBody, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/simpu?conn_id=1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the segment was updated
	updatedParams, ok := mockDB.updatedSegments[segmentID]
	require.True(t, ok, "Segment should have been updated")
	assert.NotNil(t, updatedParams.DlrStatus)
	assert.Equal(t, "delivered", *updatedParams.DlrStatus)
}

func TestEBulkSMSWebhook_Success(t *testing.T) {
	// Setup
	mockDB := NewMockQuerier()
	segmentID := int64(126)
	messageID := "ebulk_msg_123"
	routingAssignmentID := int32(1)
	mnoID := int32(10)

	mockDB.segments[messageID] = database.FindSegmentByMnoMessageIDRow{
		SegmentID:           segmentID,
		MessageID:           459,
		SegmentSeqn:         1,
		RoutingAssignmentID: &routingAssignmentID,
		RoutedMnoID:         &mnoID,
	}

	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request
	dlrPayload := map[string]interface{}{
		"message_id":    messageID,
		"status":        "delivered",
		"recipient":     "2348031234567",
		"delivery_time": "2024-01-15 10:30:00",
	}

	jsonBody, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/ebulksms?conn_id=1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the segment was updated
	updatedParams, ok := mockDB.updatedSegments[segmentID]
	require.True(t, ok, "Segment should have been updated")
	assert.NotNil(t, updatedParams.DlrStatus)
	assert.Equal(t, "delivered", *updatedParams.DlrStatus)
}

func TestWebhook_MessageNotFound(t *testing.T) {
	// Setup - no segments in database
	mockDB := NewMockQuerier()
	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request for non-existent message
	dlrPayload := map[string]interface{}{
		"type":       "dlr",
		"message_id": "non_existent_msg",
		"status":     "Delivered",
	}

	jsonBody, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/termii?conn_id=1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should still return 200 OK to prevent retries
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "acknowledged", response["status"])
	assert.Contains(t, response, "warning")
}

func TestWebhook_GenericEndpoint(t *testing.T) {
	// Setup
	mockDB := NewMockQuerier()
	segmentID := int64(127)
	messageID := "generic_test_msg"
	routingAssignmentID := int32(1)
	mnoID := int32(10)

	mockDB.segments[messageID] = database.FindSegmentByMnoMessageIDRow{
		SegmentID:           segmentID,
		MessageID:           460,
		SegmentSeqn:         1,
		RoutingAssignmentID: &routingAssignmentID,
		RoutedMnoID:         &mnoID,
	}

	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request with provider in query string
	dlrPayload := map[string]interface{}{
		"type":       "dlr",
		"message_id": messageID,
		"status":     "Delivered",
	}

	jsonBody, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/generic?provider=TERMII&conn_id=1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the segment was updated
	updatedParams, ok := mockDB.updatedSegments[segmentID]
	require.True(t, ok, "Segment should have been updated")
	assert.NotNil(t, updatedParams.DlrStatus)
	assert.Equal(t, "delivered", *updatedParams.DlrStatus)
}

func TestWebhook_MissingProvider(t *testing.T) {
	// Setup
	mockDB := NewMockQuerier()
	handler := NewWebhookHandler(mockDB)
	router := setupTestRouter(handler)

	// Prepare request without provider parameter
	dlrPayload := map[string]interface{}{
		"message_id": "test_msg",
		"status":     "Delivered",
	}

	jsonBody, _ := json.Marshal(dlrPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/generic", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "provider")
}

func TestWebhook_StatusMapping(t *testing.T) {
	testCases := []struct {
		name           string
		provider       string
		inputStatus    interface{}
		expectedStatus string
	}{
		{"Termii Delivered", "termii", "Delivered", "delivered"},
		{"Termii Failed", "termii", "Failed", "failed"},
		{"Termii Sent", "termii", "Sent", "sent"},
		{"AfricaTalking Success", "africatalking", "Success", "delivered"},
		{"AfricaTalking Failed", "africatalking", "Failed", "failed"},
		{"Simpu delivered", "simpu", "delivered", "delivered"},
		{"Simpu failed", "simpu", "failed", "failed"},
		{"eBulkSMS delivered", "ebulksms", "delivered", "delivered"},
		{"MultiTexter 1", "multitexter", "1", "delivered"},
		{"MultiTexter 2", "multitexter", "2", "failed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockDB := NewMockQuerier()
			segmentID := int64(200 + len(mockDB.segments))
			messageID := "status_test_" + tc.name
			routingAssignmentID := int32(1)
			mnoID := int32(10)

			mockDB.segments[messageID] = database.FindSegmentByMnoMessageIDRow{
				SegmentID:           segmentID,
				MessageID:           500,
				SegmentSeqn:         1,
				RoutingAssignmentID: &routingAssignmentID,
				RoutedMnoID:         &mnoID,
			}

			handler := NewWebhookHandler(mockDB)
			router := setupTestRouter(handler)

			// Prepare request based on provider
			var req *http.Request
			if tc.provider == "africatalking" {
				formData := url.Values{}
				formData.Set("id", messageID)
				formData.Set("status", tc.inputStatus.(string))
				req = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/"+tc.provider+"?conn_id=1", strings.NewReader(formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				dlrPayload := map[string]interface{}{
					"message_id": messageID,
					"status":     tc.inputStatus,
				}
				if tc.provider == "termii" {
					dlrPayload["type"] = "dlr"
				}
				if tc.provider == "simpu" {
					dlrPayload = map[string]interface{}{
						"event": "message.status.update",
						"data": map[string]interface{}{
							"message_id": messageID,
							"status":     tc.inputStatus,
						},
					}
				}
				jsonBody, _ := json.Marshal(dlrPayload)
				req = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/dlr/"+tc.provider+"?conn_id=1", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			}

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)

			// Verify status was correctly mapped
			updatedParams, ok := mockDB.updatedSegments[segmentID]
			require.True(t, ok, "Segment should have been updated")
			assert.NotNil(t, updatedParams.DlrStatus)
			assert.Equal(t, tc.expectedStatus, *updatedParams.DlrStatus, "Status should be correctly mapped")
		})
	}
}
