package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/linkscan"
	"github.com/thrillee/aegisbox/internal/sms"
	"github.com/thrillee/aegisbox/internal/wallet"
)

// LinkScanHandler serves link scan screenshots to the portal and takes admin decisions on
// messages held by the link check (Phase 2).
type LinkScanHandler struct {
	q             database.Querier
	screenshotDir string
	gate          *linkscan.Gate
}

func NewLinkScanHandler(q database.Querier, pool *pgxpool.Pool, screenshotDir string) *LinkScanHandler {
	dlr := sms.NewDLRService(sms.DLRServiceDependencies{DBQueries: q, WalletService: wallet.NewService(pool, q)})
	rejecter := sms.NewLinkHoldRejecter(q, dlr, wallet.NewService(pool, q))
	return &LinkScanHandler{
		q:             q,
		screenshotDir: screenshotDir,
		gate:          linkscan.NewGate(q, nil, nil, rejecter),
	}
}

// SetupLinkScanRoutes adds the link scan routes.
func SetupLinkScanRoutes(router gin.IRouter, q database.Querier, pool *pgxpool.Pool, screenshotDir string) {
	h := NewLinkScanHandler(q, pool, screenshotDir)
	router.GET("/link-scans/:id/screenshot", h.Screenshot)
	router.POST("/link-scan-holds/:message_id/release", h.Release)
	router.POST("/link-scan-holds/:message_id/reject", h.Reject)
}

// Screenshot handles GET /link-scans/:id/screenshot (JPEG).
func (h *LinkScanHandler) Screenshot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scan ID"})
		return
	}
	path, err := h.q.GetLinkScanScreenshot(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read scan"})
		return
	}
	if path == nil || *path == "" || h.screenshotDir == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No screenshot (none taken, or deleted after the retention period)"})
		return
	}
	// The stored path is relative; cleaning it under "/" keeps it inside the screenshot folder.
	full := filepath.Join(h.screenshotDir, filepath.Clean("/"+*path))
	if !strings.HasSuffix(full, ".jpg") {
		c.JSON(http.StatusNotFound, gin.H{"error": "No screenshot"})
		return
	}
	data, err := os.ReadFile(full)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Screenshot file not found"})
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "image/jpeg", data)
}

type holdDecisionRequest struct {
	By string `json:"by"`
}

// Release handles POST /link-scan-holds/:message_id/release.
func (h *LinkScanHandler) Release(c *gin.Context) { h.decide(c, true) }

// Reject handles POST /link-scan-holds/:message_id/reject.
func (h *LinkScanHandler) Reject(c *gin.Context) { h.decide(c, false) }

func (h *LinkScanHandler) decide(c *gin.Context, release bool) {
	id, err := strconv.ParseInt(c.Param("message_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}
	var req holdDecisionRequest
	_ = c.ShouldBindJSON(&req)
	by := strings.TrimSpace(req.By)
	if by == "" {
		by = "admin"
	}
	if len(by) > 200 {
		by = by[:200]
	}
	if err := h.gate.AdminDecide(c.Request.Context(), id, release, by); err != nil {
		if errors.Is(err, linkscan.ErrNotInReview) {
			c.JSON(http.StatusConflict, gin.H{"error": "Message is not held (already sent or rejected)"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "Link hold decision failed", slog.Int64("message_id", id), slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply the decision"})
		return
	}
	action := "rejected"
	if release {
		action = "released"
	}
	c.JSON(http.StatusOK, gin.H{"message_id": id, "status": action})
}
