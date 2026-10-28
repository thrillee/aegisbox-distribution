package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
	"github.com/thrillee/aegisbox/internal/sms"
)

// HealthEventBroadcaster manages SSE connections and broadcasts health events
type HealthEventBroadcaster struct {
	clients map[chan dto.HealthEvent]struct{}
	mu      sync.RWMutex
}

// NewHealthEventBroadcaster creates a new event broadcaster
func NewHealthEventBroadcaster() *HealthEventBroadcaster {
	return &HealthEventBroadcaster{
		clients: make(map[chan dto.HealthEvent]struct{}),
	}
}

// Subscribe adds a new client channel
func (b *HealthEventBroadcaster) Subscribe(ch chan dto.HealthEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[ch] = struct{}{}
	slog.Info("SSE client subscribed", slog.Int("total_clients", len(b.clients)))
}

// Unsubscribe removes a client channel
func (b *HealthEventBroadcaster) Unsubscribe(ch chan dto.HealthEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, ch)
	close(ch)
	slog.Info("SSE client unsubscribed", slog.Int("total_clients", len(b.clients)))
}

// Broadcast sends an event to all connected clients
func (b *HealthEventBroadcaster) Broadcast(event dto.HealthEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- event:
		default:
			// Client channel is full or slow, skip
			slog.Warn("Failed to send event to SSE client (channel full)")
		}
	}
}

// HealthHandler handles health monitoring endpoints
type HealthHandler struct {
	dbQueries   database.Querier
	dbPool      *pgxpool.Pool
	broadcaster *HealthEventBroadcaster
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(q database.Querier, pool *pgxpool.Pool, broadcaster *HealthEventBroadcaster) *HealthHandler {
	return &HealthHandler{
		dbQueries:   q,
		dbPool:      pool,
		broadcaster: broadcaster,
	}
}

// MarkRouteAvailable manually marks a route as available or unavailable
func (h *HealthHandler) MarkRouteAvailable(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "MarkRouteAvailable")

	assignmentIDStr := c.Param("assignment_id")
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment_id"})
		return
	}

	var req dto.MarkRouteAvailableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	available := *req.Available
	if _, err := h.dbQueries.GetRoutingAssignmentByID(logCtx, int32(assignmentID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routing assignment not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to get routing assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update route availability"})
		return
	}

	// Same route-wide health record (sender NULL) the router checks when picking routes
	tracker := sms.NewRouteHealthTracker(h.dbQueries, slog.Default())
	if available {
		err = tracker.MarkRouteAvailable(logCtx, int32(assignmentID), nil)
	} else {
		reason := req.Reason
		if reason == "" {
			reason = "Manually marked unavailable via admin UI"
		}
		err = tracker.MarkRouteUnavailable(logCtx, int32(assignmentID), nil, reason)
	}
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to update route availability", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update route availability"})
		return
	}

	// Broadcast event
	eventType := dto.EventRouteMarkedAvailable
	if !available {
		eventType = dto.EventRouteMarkedUnavailable
	}

	h.broadcaster.Broadcast(dto.HealthEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"assignment_id": int32(assignmentID),
			"is_available":  available,
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Route marked as %s successfully", map[bool]string{true: "available", false: "unavailable"}[available]),
		"assignment_id": int32(assignmentID),
		"is_available":  available,
	})
}

// UnblockSender manually unblocks a sender ID for a specific MNO
func (h *HealthHandler) UnblockSender(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UnblockSender")

	senderID := c.Param("sender_id")
	mnoIDStr := c.Param("mno_id")
	mnoID, err := strconv.ParseInt(mnoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mno_id"})
		return
	}

	var req dto.UnblockSenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Reason is optional, so allow empty body
		req.Reason = "Manually unblocked via admin UI"
	}

	// Update sender_id_health table
	query := `
		UPDATE sender_id_health 
		SET is_blocked = false, 
		    blocked_at = NULL,
		    blocked_reason = NULL,
		    updated_at = NOW()
		WHERE sender_id = $1 AND mno_id = $2 AND is_blocked = true
		RETURNING id
	`
	var updatedID int32
	err = h.dbPool.QueryRow(logCtx, query, senderID, int32(mnoID)).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sender is not blocked or not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to unblock sender", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock sender"})
		return
	}

	// Broadcast event
	h.broadcaster.Broadcast(dto.HealthEvent{
		Type:      dto.EventSenderUnblocked,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sender_id": senderID,
			"mno_id":    int32(mnoID),
			"reason":    req.Reason,
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Sender unblocked successfully",
		"sender_id": senderID,
		"mno_id":    int32(mnoID),
	})
}

// StreamHealthEvents streams health events via Server-Sent Events
func (h *HealthHandler) StreamHealthEvents(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "StreamHealthEvents")
	slog.InfoContext(logCtx, "New SSE connection established")

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create client channel
	clientChan := make(chan dto.HealthEvent, 10)
	h.broadcaster.Subscribe(clientChan)
	defer h.broadcaster.Unsubscribe(clientChan)

	// Send initial connection event
	c.SSEvent("connected", map[string]string{"status": "connected"})
	c.Writer.Flush()

	// Stream events
	ctx := c.Request.Context()
	for {
		select {
		case event := <-clientChan:
			eventData, err := json.Marshal(event.Data)
			if err != nil {
				slog.ErrorContext(logCtx, "Failed to marshal event data", slog.Any("error", err))
				continue
			}
			c.SSEvent(string(event.Type), string(eventData))
			c.Writer.Flush()

		case <-ctx.Done():
			slog.InfoContext(logCtx, "SSE connection closed by client")
			return
		}
	}
}

// GetHealthStats returns aggregated health statistics
func (h *HealthHandler) GetHealthStats(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetHealthStats")

	// Query route health statistics
	routeStatsQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE health_score > 70) as healthy, -- health_score is 0-100
			COUNT(*) FILTER (WHERE health_score BETWEEN 50 AND 70) as warning,
			COUNT(*) FILTER (WHERE health_score < 50) as critical,
			COUNT(*) FILTER (WHERE is_available = false) as unavailable
		FROM route_health
	`
	var routeStats struct {
		Total       int32
		Healthy     int32
		Warning     int32
		Critical    int32
		Unavailable int32
	}

	err := h.dbPool.QueryRow(logCtx, routeStatsQuery).Scan(
		&routeStats.Total,
		&routeStats.Healthy,
		&routeStats.Warning,
		&routeStats.Critical,
		&routeStats.Unavailable,
	)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to query route stats", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch health statistics"})
		return
	}

	// Query sender health statistics
	senderStatsQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_blocked = true) as blocked,
			COALESCE(AVG(delivery_rate::numeric), 0) as avg_delivery_rate
		FROM sender_id_health
	`
	var senderStats struct {
		Total           int32
		Blocked         int32
		AvgDeliveryRate float64
	}

	err = h.dbPool.QueryRow(logCtx, senderStatsQuery).Scan(
		&senderStats.Total,
		&senderStats.Blocked,
		&senderStats.AvgDeliveryRate,
	)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to query sender stats", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch health statistics"})
		return
	}

	response := dto.HealthStatsResponse{
		TotalRoutes:       routeStats.Total,
		HealthyRoutes:     routeStats.Healthy,
		WarningRoutes:     routeStats.Warning,
		CriticalRoutes:    routeStats.Critical,
		UnavailableRoutes: routeStats.Unavailable,
		TotalSenders:      senderStats.Total,
		BlockedSenders:    senderStats.Blocked,
		AvgDeliveryRate:   senderStats.AvgDeliveryRate,
	}

	c.JSON(http.StatusOK, response)
}

// GetRecentHealthEvents returns recent health events from the ring buffer
func (h *HealthHandler) GetRecentHealthEvents(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetRecentHealthEvents")

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 500 {
		limit = 50
	}

	// Query recent route delivery attempts
	query := `
		SELECT 
			rda.routing_assignment_id,
			rda.success,
			rda.attempted_at,
			rg.name as routing_group_name,
			m.name as mno_name
		FROM route_delivery_attempts rda
		LEFT JOIN routing_assignments ra ON rda.routing_assignment_id = ra.id
		LEFT JOIN routing_groups rg ON ra.routing_group_id = rg.id
		LEFT JOIN mnos m ON ra.mno_id = m.id
		ORDER BY rda.attempted_at DESC
		LIMIT $1
	`

	rows, err := h.dbPool.Query(logCtx, query, limit)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to query recent events", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent events"})
		return
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	for rows.Next() {
		var (
			assignmentID     int32
			success          bool
			attemptedAt      time.Time
			routingGroupName *string
			mnoName          *string
		)
		if err := rows.Scan(&assignmentID, &success, &attemptedAt, &routingGroupName, &mnoName); err != nil {
			slog.ErrorContext(logCtx, "Failed to scan event row", slog.Any("error", err))
			continue
		}

		events = append(events, map[string]interface{}{
			"assignment_id":      assignmentID,
			"success":            success,
			"attempted_at":       attemptedAt,
			"routing_group_name": routingGroupName,
			"mno_name":           mnoName,
		})
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(logCtx, "Error iterating event rows", slog.Any("error", err))
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}
