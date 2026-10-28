package handlers

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
	"github.com/thrillee/aegisbox/internal/sms"
)

// RoutingConfigHandler handles routing configuration endpoints. It reads and writes the
// routing_config columns the gateway uses (see GetRoutingConfigForGroup and
// GetGlobalRoutingConfig), so saved settings take effect.
type RoutingConfigHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool
}

// NewRoutingConfigHandler creates a new routing config handler
func NewRoutingConfigHandler(q database.Querier, pool *pgxpool.Pool) *RoutingConfigHandler {
	return &RoutingConfigHandler{
		dbQueries: q,
		dbPool:    pool,
	}
}

// Defaults used when there is no global routing_config row (same as migration 00003).
var defaultGlobalRoutingConfig = dto.UpdateGlobalRoutingConfigRequest{
	SelectionStrategy:      "priority_weighted",
	MaxRouteRetries:        2,
	RouteFailureThreshold:  5,
	RouteRecoveryThreshold: 3,
	RouteHealthWindowSize:  1000,
	SenderBlockThreshold:   50,
	SenderUnblockThreshold: 70,
	SenderHealthWindowSize: 1000,
	DLRTimeoutSeconds:      sms.DefaultDLRTimeoutSeconds,
}

// ListConfigs handles GET /routing/config: every routing group override.
func (h *RoutingConfigHandler) ListConfigs(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListRoutingConfigs")

	rows, err := h.dbQueries.ListGroupRoutingConfigs(logCtx)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list routing configs", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch routing configurations"})
		return
	}

	configs := make([]dto.RoutingGroupConfigResponse, 0, len(rows))
	for _, r := range rows {
		configs = append(configs, toRoutingGroupConfigResponse(database.RoutingConfig{
			ID: r.ID, RoutingGroupID: r.RoutingGroupID, MaxRouteRetries: r.MaxRouteRetries,
			SelectionStrategy: r.SelectionStrategy, RouteFailureThreshold: r.RouteFailureThreshold,
			CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}, r.RoutingGroupName))
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"count":   len(configs),
	})
}

// GetConfig handles GET /routing/config/:group_id.
func (h *RoutingConfigHandler) GetConfig(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetRoutingConfig")

	groupID, ok := parseGroupIDParam(c)
	if !ok {
		return
	}

	r, err := h.dbQueries.GetRoutingConfigByGroupID(logCtx, &groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{
				"config":  nil,
				"message": "No configuration found for this routing group, using the global config",
			})
			return
		}
		slog.ErrorContext(logCtx, "Failed to get routing config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch routing configuration"})
		return
	}

	c.JSON(http.StatusOK, toRoutingGroupConfigResponse(database.RoutingConfig{
		ID: r.ID, RoutingGroupID: r.RoutingGroupID, MaxRouteRetries: r.MaxRouteRetries,
		SelectionStrategy: r.SelectionStrategy, RouteFailureThreshold: r.RouteFailureThreshold,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}, r.RoutingGroupName))
}

// CreateConfig handles POST /routing/config.
func (h *RoutingConfigHandler) CreateConfig(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateRoutingConfig")

	var req dto.CreateRoutingGroupConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if _, err := h.dbQueries.GetRoutingConfigByGroupID(logCtx, &req.RoutingGroupID); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Configuration for this routing group already exists"})
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(logCtx, "Failed to check routing config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create routing configuration"})
		return
	}

	h.saveGroupConfig(c, logCtx, req.RoutingGroupID, req.UpdateRoutingGroupConfigRequest, http.StatusCreated)
}

// UpdateConfig handles PUT /routing/config/:group_id. It creates the override if it does not exist.
func (h *RoutingConfigHandler) UpdateConfig(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateRoutingConfig")

	groupID, ok := parseGroupIDParam(c)
	if !ok {
		return
	}

	var req dto.UpdateRoutingGroupConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	h.saveGroupConfig(c, logCtx, groupID, req, http.StatusOK)
}

// saveGroupConfig upserts a group override. Columns the router only reads from the global
// config are copied from it, so the row stays consistent with the global settings.
func (h *RoutingConfigHandler) saveGroupConfig(c *gin.Context, logCtx context.Context, groupID int32, req dto.UpdateRoutingGroupConfigRequest, status int) {
	group, err := h.dbQueries.GetRoutingGroupByID(logCtx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routing group not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to validate routing group", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate routing group"})
		return
	}

	global, err := h.dbQueries.GetGlobalRoutingConfig(logCtx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(logCtx, "Failed to get global routing config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save routing configuration"})
		return
	}

	saved, err := h.dbQueries.UpsertRoutingConfig(logCtx, database.UpsertRoutingConfigParams{
		RoutingGroupID:             &groupID,
		SenderIDBlockThreshold:     global.SenderIDBlockThreshold,
		SenderIDUnblockThreshold:   global.SenderIDUnblockThreshold,
		SenderIDDlrTimeoutSecs:     global.SenderIDDlrTimeoutSecs,
		SenderIDHealthWindowSize:   global.SenderIDHealthWindowSize,
		RouteAvailabilityThreshold: global.RouteAvailabilityThreshold,
		RouteFailureThreshold:      &req.RouteFailureThreshold,
		RouteRecoveryThreshold:     global.RouteRecoveryThreshold,
		RouteHealthWindowSize:      global.RouteHealthWindowSize,
		EnableAutomaticFailover:    global.EnableAutomaticFailover,
		MaxRouteRetries:            &req.MaxRouteRetries,
		SelectionStrategy:          &req.SelectionStrategy,
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to save routing config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save routing configuration"})
		return
	}

	slog.InfoContext(logCtx, "Routing group config saved", slog.Int("routing_group_id", int(groupID)))
	c.JSON(status, toRoutingGroupConfigResponse(saved, group.Name))
}

// GetGlobalConfig handles GET /routing/config/global.
func (h *RoutingConfigHandler) GetGlobalConfig(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetGlobalRoutingConfig")

	global, err := h.dbQueries.GetGlobalRoutingConfig(logCtx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, toGlobalRoutingConfigResponse(defaultsAsRoutingConfig()))
			return
		}
		slog.ErrorContext(logCtx, "Failed to get global routing config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global configuration"})
		return
	}
	c.JSON(http.StatusOK, toGlobalRoutingConfigResponse(global))
}

// UpdateGlobalConfig handles PUT /routing/config/global.
func (h *RoutingConfigHandler) UpdateGlobalConfig(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateGlobalRoutingConfig")

	var req dto.UpdateGlobalRoutingConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	saved, err := h.dbQueries.UpdateGlobalRoutingConfig(logCtx, database.UpdateGlobalRoutingConfigParams{
		SenderIDBlockThreshold:   percentToNumeric(req.SenderBlockThreshold),
		SenderIDUnblockThreshold: percentToNumeric(req.SenderUnblockThreshold),
		SenderIDDlrTimeoutSecs:   &req.DLRTimeoutSeconds,
		SenderIDHealthWindowSize: &req.SenderHealthWindowSize,
		RouteFailureThreshold:    &req.RouteFailureThreshold,
		RouteRecoveryThreshold:   &req.RouteRecoveryThreshold,
		RouteHealthWindowSize:    &req.RouteHealthWindowSize,
		MaxRouteRetries:          &req.MaxRouteRetries,
		SelectionStrategy:        &req.SelectionStrategy,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// No global row yet (e.g. deleted by hand): create it.
		saved, err = h.dbQueries.UpsertRoutingConfig(logCtx, database.UpsertRoutingConfigParams{
			SenderIDBlockThreshold:     percentToNumeric(req.SenderBlockThreshold),
			SenderIDUnblockThreshold:   percentToNumeric(req.SenderUnblockThreshold),
			SenderIDDlrTimeoutSecs:     &req.DLRTimeoutSeconds,
			SenderIDHealthWindowSize:   &req.SenderHealthWindowSize,
			RouteAvailabilityThreshold: percentToNumeric(60),
			RouteFailureThreshold:      &req.RouteFailureThreshold,
			RouteRecoveryThreshold:     &req.RouteRecoveryThreshold,
			RouteHealthWindowSize:      &req.RouteHealthWindowSize,
			EnableAutomaticFailover:    boolPtr(true),
			MaxRouteRetries:            &req.MaxRouteRetries,
			SelectionStrategy:          &req.SelectionStrategy,
		})
	}
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to update global routing config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update global configuration"})
		return
	}

	slog.InfoContext(logCtx, "Global routing config updated")
	c.JSON(http.StatusOK, toGlobalRoutingConfigResponse(saved))
}

func parseGroupIDParam(c *gin.Context) (int32, bool) {
	groupID, err := strconv.ParseInt(c.Param("group_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id"})
		return 0, false
	}
	return int32(groupID), true
}

func toRoutingGroupConfigResponse(rc database.RoutingConfig, groupName string) dto.RoutingGroupConfigResponse {
	resp := dto.RoutingGroupConfigResponse{
		ID:                    rc.ID,
		RoutingGroupName:      groupName,
		SelectionStrategy:     derefString(rc.SelectionStrategy, defaultGlobalRoutingConfig.SelectionStrategy),
		MaxRouteRetries:       derefInt32(rc.MaxRouteRetries, defaultGlobalRoutingConfig.MaxRouteRetries),
		RouteFailureThreshold: derefInt32(rc.RouteFailureThreshold, defaultGlobalRoutingConfig.RouteFailureThreshold),
		CreatedAt:             rc.CreatedAt.Time,
		UpdatedAt:             rc.UpdatedAt.Time,
	}
	if rc.RoutingGroupID != nil {
		resp.RoutingGroupID = *rc.RoutingGroupID
	}
	return resp
}

func toGlobalRoutingConfigResponse(rc database.RoutingConfig) dto.GlobalRoutingConfigResponse {
	d := defaultGlobalRoutingConfig
	return dto.GlobalRoutingConfigResponse{
		ID:                     rc.ID,
		SelectionStrategy:      derefString(rc.SelectionStrategy, d.SelectionStrategy),
		MaxRouteRetries:        derefInt32(rc.MaxRouteRetries, d.MaxRouteRetries),
		RouteFailureThreshold:  derefInt32(rc.RouteFailureThreshold, d.RouteFailureThreshold),
		RouteRecoveryThreshold: derefInt32(rc.RouteRecoveryThreshold, d.RouteRecoveryThreshold),
		RouteHealthWindowSize:  derefInt32(rc.RouteHealthWindowSize, d.RouteHealthWindowSize),
		SenderBlockThreshold:   numericToPercent(rc.SenderIDBlockThreshold, d.SenderBlockThreshold),
		SenderUnblockThreshold: numericToPercent(rc.SenderIDUnblockThreshold, d.SenderUnblockThreshold),
		SenderHealthWindowSize: derefInt32(rc.SenderIDHealthWindowSize, d.SenderHealthWindowSize),
		DLRTimeoutSeconds:      derefInt32(rc.SenderIDDlrTimeoutSecs, d.DLRTimeoutSeconds),
		Schedule: dto.WorkerScheduleResponse{
			DLRTimeoutCheckIntervalSeconds:     int32(sms.DLRTimeoutCheckInterval.Seconds()),
			SenderAutoUnblockCheckIntervalMins: int32(sms.SenderAutoUnblockCheckInterval.Minutes()),
			SenderAutoUnblockAfterInactiveDays: int32(sms.SenderAutoUnblockAfterInactive.Hours() / 24),
			RingBufferCleanupIntervalHours:     int32(sms.RingBufferCleanupInterval.Hours()),
			RingBufferRetentionDays:            sms.RingBufferRetentionDays,
		},
		UpdatedAt: rc.UpdatedAt.Time,
	}
}

func defaultsAsRoutingConfig() database.RoutingConfig {
	d := defaultGlobalRoutingConfig
	return database.RoutingConfig{
		SelectionStrategy:        &d.SelectionStrategy,
		MaxRouteRetries:          &d.MaxRouteRetries,
		RouteFailureThreshold:    &d.RouteFailureThreshold,
		RouteRecoveryThreshold:   &d.RouteRecoveryThreshold,
		RouteHealthWindowSize:    &d.RouteHealthWindowSize,
		SenderIDBlockThreshold:   percentToNumeric(d.SenderBlockThreshold),
		SenderIDUnblockThreshold: percentToNumeric(d.SenderUnblockThreshold),
		SenderIDHealthWindowSize: &d.SenderHealthWindowSize,
		SenderIDDlrTimeoutSecs:   &d.DLRTimeoutSeconds,
	}
}

// percentToNumeric stores a percentage with two decimals, as the NUMERIC(5,2) columns do.
func percentToNumeric(v float64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(int64(v*100 + 0.5)), Exp: -2, Valid: true}
}

func numericToPercent(n pgtype.Numeric, fallback float64) float64 {
	if !n.Valid {
		return fallback
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return fallback
	}
	return f.Float64
}

func derefInt32(v *int32, fallback int32) int32 {
	if v == nil {
		return fallback
	}
	return *v
}

func derefString(v *string, fallback string) string {
	if v == nil {
		return fallback
	}
	return *v
}

func boolPtr(b bool) *bool { return &b }
