package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

type OtpMessageTemplateHandler struct {
	dbQueries database.Querier
}

func NewOtpMessageTemplateHandler(q database.Querier) *OtpMessageTemplateHandler {
	return &OtpMessageTemplateHandler{dbQueries: q}
}

// Create an OTP Message Template
func (h *OtpMessageTemplateHandler) Create(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateOtpMessageTemplate")
	var req dto.CreateOtpMessageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	params := database.CreateOtpMessageTemplateParams{
		Name:            req.Name,
		ContentTemplate: req.ContentTemplate,
	}
	if req.DefaultBrandName != nil {
		params.DefaultBrandName = req.DefaultBrandName
	}
	if req.Status != "" {
		params.Status = req.Status
	} else {
		params.Status = "active" // Default
	}

	template, err := h.dbQueries.CreateOtpMessageTemplate(logCtx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation for name
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Template with name '%s' already exists.", req.Name)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create OTP message template", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create template"})
		return
	}
	c.JSON(http.StatusCreated, mapDbOtpMessageTemplateToResponse(template))
}

// GetByID retrieves an OTP Message Template by its ID
func (h *OtpMessageTemplateHandler) GetByID(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetOtpMessageTemplateByID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID."})
		return
	}
	template, err := h.dbQueries.GetOtpMessageTemplateByID(logCtx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found."})
		} else {
			slog.ErrorContext(logCtx, "Error fetching template by ID", slog.Any("error", err), slog.Int("id", int(id)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve template."})
		}
		return
	}
	c.JSON(http.StatusOK, mapDbOtpMessageTemplateToResponse(template))
}

// List OTP Message Templates with pagination
func (h *OtpMessageTemplateHandler) List(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListOtpMessageTemplates")
	// limit, offset := utils.ParsePagination(c)
	limitParam, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 32)
	offsetParam, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	limit := int32(limitParam)
	offset := int32(offsetParam)

	listParams := database.ListOtpMessageTemplatesParams{
		LimitVal:  limit,
		OffsetVal: offset,
	}

	var statusVal *string

	// Add filters if your ListOtpMessageTemplatesParams supports them (e.g., by status)
	if statusStr := c.Query("status"); statusStr != "" {
		statusVal = &statusStr
		listParams.Status = statusVal
	}

	templates, err := h.dbQueries.ListOtpMessageTemplates(logCtx, listParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list OTP message templates", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
		return
	}

	total, err := h.dbQueries.CountOtpMessageTemplates(logCtx, statusVal) // Assuming CountOtpMessageTemplates exists
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count OTP message templates", slog.Any("error", err))
		// Handle error or proceed with count 0 / partial data
	}

	respData := make([]dto.OtpMessageTemplateResponse, len(templates))
	for i, t := range templates {
		respData[i] = mapDbOtpMessageTemplateToResponse(t)
	}
	c.JSON(http.StatusOK, dto.PaginatedListResponse{
		Data:       respData,
		Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
	})
}

// Update an OTP Message Template
func (h *OtpMessageTemplateHandler) Update(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateOtpMessageTemplate")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID."})
		return
	}

	var req dto.UpdateOtpMessageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	params := database.UpdateOtpMessageTemplateParams{
		ID: int32(id),
	}
	if req.Name != nil {
		params.Name = req.Name
	}
	if req.ContentTemplate != nil {
		params.ContentTemplate = req.ContentTemplate
	}
	if req.DefaultBrandName != nil {
		params.DefaultBrandName = req.DefaultBrandName
	}
	if req.Status != nil {
		params.Status = req.Status
	}

	updatedTemplate, err := h.dbQueries.UpdateOtpMessageTemplate(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found."})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation for name
			c.JSON(http.StatusConflict, gin.H{"error": "Template with the new name already exists."})
			return
		}
		slog.ErrorContext(logCtx, "Failed to update template", slog.Any("error", err), slog.Int("id", int(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update template."})
		return
	}
	c.JSON(http.StatusOK, mapDbOtpMessageTemplateToResponse(updatedTemplate))
}

// Delete an OTP Message Template
func (h *OtpMessageTemplateHandler) Delete(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteOtpMessageTemplate")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID."})
		return
	}
	err = h.dbQueries.DeleteOtpMessageTemplate(logCtx, int32(id))
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete template", slog.Any("error", err), slog.Int("id", int(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete template."})
		return
	}
	c.Status(http.StatusNoContent)
}

func mapDbOtpMessageTemplateToResponse(t database.OtpMessageTemplate) dto.OtpMessageTemplateResponse {
	resp := dto.OtpMessageTemplateResponse{
		ID:              t.ID,
		Name:            t.Name,
		ContentTemplate: t.ContentTemplate,
		Status:          t.Status,
		CreatedAt:       t.CreatedAt.Time,
		UpdatedAt:       t.UpdatedAt.Time,
	}
	if t.DefaultBrandName != nil {
		resp.DefaultBrandName = t.DefaultBrandName
	}
	return resp
}
