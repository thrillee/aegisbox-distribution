package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
	"github.com/thrillee/aegisbox/internal/sms"
	"github.com/thrillee/aegisbox/internal/wallet"
)

// DLRHandler handles manual DLRs and per service provider auto DLR config.
type DLRHandler struct {
	dbQueries  database.Querier
	dlrService *sms.DLRService
}

// NewDLRHandler creates a new DLRHandler.
func NewDLRHandler(q database.Querier, pool *pgxpool.Pool) *DLRHandler {
	return &DLRHandler{
		dbQueries: q,
		dlrService: sms.NewDLRService(sms.DLRServiceDependencies{
			DBQueries:     q,
			WalletService: wallet.NewService(pool, q),
		}),
	}
}

// SendManualDLR handles POST /messages/:id/dlr
// Completes the message's segments that never got a DLR and forwards the DLR to the SP.
func (h *DLRHandler) SendManualDLR(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "SendManualDLR")

	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}
	logCtx = logging.ContextWithMessageID(logCtx, messageID)

	var req dto.ManualDLRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	status, err := sms.NormalizeManualDLRStatus(req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.dbQueries.GetSPMessageInfoForDLR(logCtx, messageID); err != nil {
		handleGetError(c, logCtx, "Message", err)
		return
	}

	result, err := h.dlrService.ApplyManualDLR(logCtx, sms.ManualDLRRequest{
		MessageID: messageID,
		Status:    status,
		ErrorCode: req.ErrorCode,
		Source:    sms.DLRSourceManual,
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to apply manual DLR", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply manual DLR; nothing was changed"})
		return
	}
	if result.SegmentsUpdated == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Message has no segments waiting for a final DLR"})
		return
	}

	slog.InfoContext(logCtx, "Manual DLR sent", slog.String("dlr_status", status), slog.Int("segments", result.SegmentsUpdated))
	c.JSON(http.StatusOK, toManualDLRResult(messageID, result))
}

// SendBulkManualDLR handles POST /messages/dlr/bulk
func (h *DLRHandler) SendBulkManualDLR(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "SendBulkManualDLR")

	var req dto.BulkManualDLRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	status, err := sms.NormalizeManualDLRStatus(req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := dto.BulkManualDLRResponse{Status: status, Results: make([]dto.ManualDLRResult, 0, len(req.MessageIDs))}
	seen := make(map[int64]bool, len(req.MessageIDs))
	for _, messageID := range req.MessageIDs {
		if seen[messageID] {
			continue
		}
		seen[messageID] = true

		applied, err := h.dlrService.ApplyManualDLR(logCtx, sms.ManualDLRRequest{
			MessageID:          messageID,
			Status:             status,
			ErrorCode:          req.ErrorCode,
			Source:             sms.DLRSourceManual,
			IncludeUnconfirmed: req.IncludeUnconfirmed,
		})
		result := toManualDLRResult(messageID, applied)
		switch {
		case err != nil:
			slog.ErrorContext(logCtx, "Failed to apply manual DLR", slog.Int64("message_id", messageID), slog.Any("error", err))
			result.Error = "failed to apply DLR; nothing was changed"
			resp.Failed++
		case applied.SegmentsUpdated == 0:
			result.Error = "no segments waiting for a final DLR"
			resp.Skipped++
		default:
			resp.Forwarded++
			if result.ForwardWarning != "" {
				resp.Warnings++
			}
		}
		resp.Results = append(resp.Results, result)
	}

	slog.InfoContext(logCtx, "Bulk manual DLR completed",
		slog.String("dlr_status", status),
		slog.Int("forwarded", resp.Forwarded),
		slog.Int("skipped", resp.Skipped),
		slog.Int("failed", resp.Failed))
	c.JSON(http.StatusOK, resp)
}

// GetAutoDLR handles GET /service-providers/:id/auto-dlr
func (h *DLRHandler) GetAutoDLR(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetAutoDLR")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, id)

	sp, err := h.dbQueries.GetServiceProviderByID(logCtx, id)
	if err != nil {
		handleGetError(c, logCtx, "Service provider", err)
		return
	}
	c.JSON(http.StatusOK, toAutoDLRResponse(sp))
}

// UpdateAutoDLR handles PUT /service-providers/:id/auto-dlr
func (h *DLRHandler) UpdateAutoDLR(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateAutoDLR")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, id)

	var req dto.UpdateAutoDLRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	status, err := sms.NormalizeManualDLRStatus(req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sp, err := h.dbQueries.UpdateServiceProviderAutoDLR(logCtx, database.UpdateServiceProviderAutoDLRParams{
		AutoDlrEnabled:      req.Enabled,
		AutoDlrStatus:       status,
		AutoDlrDelaySeconds: req.DelaySeconds,
		ID:                  id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service provider not found"})
			return
		}
		slog.ErrorContext(logCtx, "Failed to update auto DLR config", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auto DLR config"})
		return
	}

	slog.InfoContext(logCtx, "Auto DLR config updated",
		slog.Bool("enabled", sp.AutoDlrEnabled),
		slog.String("status", sp.AutoDlrStatus),
		slog.Int("delay_seconds", int(sp.AutoDlrDelaySeconds)))
	c.JSON(http.StatusOK, toAutoDLRResponse(sp))
}

func toManualDLRResult(messageID int64, result sms.ManualDLRResult) dto.ManualDLRResult {
	return dto.ManualDLRResult{
		MessageID:       messageID,
		SegmentsUpdated: result.SegmentsUpdated,
		Forwarded:       result.QueueID != 0,
		QueueID:         result.QueueID,
		ForwardWarning:  result.ForwardWarning,
	}
}

func toAutoDLRResponse(sp database.ServiceProvider) dto.AutoDLRResponse {
	resp := dto.AutoDLRResponse{
		ServiceProviderID: sp.ID,
		Enabled:           sp.AutoDlrEnabled,
		Status:            sp.AutoDlrStatus,
		DelaySeconds:      sp.AutoDlrDelaySeconds,
	}
	if sp.AutoDlrEnabledAt.Valid {
		resp.EnabledAt = &sp.AutoDlrEnabledAt.Time
	}
	return resp
}

// RetryDLRForwarding handles POST /dlr-forwarding/retry
// Puts failed DLR forwarding jobs back in the queue so the gateway forwards them to the SP again.
func (h *DLRHandler) RetryDLRForwarding(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "RetryDLRForwarding")

	var req dto.RetryDLRForwardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	requeued, err := h.dbQueries.RequeueFailedDLRForwarding(logCtx, req.IDs)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to requeue DLR forwarding jobs", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to requeue DLR forwarding jobs"})
		return
	}
	if requeued == nil {
		requeued = []int64{}
	}

	slog.InfoContext(logCtx, "Requeued failed DLR forwarding jobs", slog.Int("requeued", len(requeued)))
	c.JSON(http.StatusOK, dto.RetryDLRForwardingResponse{
		Requeued: requeued,
		Skipped:  len(req.IDs) - len(requeued),
	})
}
