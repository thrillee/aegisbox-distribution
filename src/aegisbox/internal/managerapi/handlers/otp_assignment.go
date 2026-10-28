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

type OtpSenderTemplateAssignmentHandler struct {
	dbQueries database.Querier
}

func NewOtpSenderTemplateAssignmentHandler(q database.Querier) *OtpSenderTemplateAssignmentHandler {
	return &OtpSenderTemplateAssignmentHandler{dbQueries: q}
}

// Create an OTP Sender-Template Assignment
func (h *OtpSenderTemplateAssignmentHandler) Create(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateOtpSenderTemplateAssignment")
	var req dto.CreateOtpSenderTemplateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	params := database.CreateOtpSenderTemplateAssignmentParams{
		OtpAlternativeSenderID: req.OtpAlternativeSenderID,
		OtpMessageTemplateID:   req.OtpMessageTemplateID,
		Priority:               req.Priority, // SQLC will use default from DB if not set / not narg
	}
	if req.MnoID != nil {
		params.MnoID = req.MnoID
	}
	if req.Status != "" {
		params.Status = req.Status
	} else {
		params.Status = "active" // Default
	}

	// --- Optional: Validate Foreign Keys exist ---
	_, errAltSender := h.dbQueries.GetOtpAlternativeSenderByID(logCtx, req.OtpAlternativeSenderID)
	if errAltSender != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid otp_alternative_sender_id: %d", req.OtpAlternativeSenderID)})
		return
	}
	_, errTemplate := h.dbQueries.GetOtpMessageTemplateByID(logCtx, req.OtpMessageTemplateID)
	if errTemplate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid otp_message_template_id: %d", req.OtpMessageTemplateID)})
		return
	}
	if params.MnoID != nil {
		_, errMno := h.dbQueries.GetMNOByID(logCtx, *params.MnoID)
		if errMno != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid mno_id: %d", params.MnoID)})
			return
		}
	}
	// --- End Validation ---

	assignment, err := h.dbQueries.CreateOtpSenderTemplateAssignment(logCtx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			c.JSON(http.StatusConflict, gin.H{"error": "This sender and template assignment (with this MNO, if specified) already exists."})
			return
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sender_id, template_id, or mno_id provided."})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create assignment"})
		return
	}

	// Refetch with JOINs for better response
	detailedAssignment, _ := h.dbQueries.GetOtpSenderTemplateAssignmentByID(logCtx, assignment.ID) // Error handling omitted for brevity
	c.JSON(http.StatusCreated, detailedAssignment)
}

// GetByID retrieves an OTP Sender-Template Assignment by its ID
func (h *OtpSenderTemplateAssignmentHandler) GetByID(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetOtpSenderTemplateAssignmentByID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID."})
		return
	}
	// Assuming GetOtpSenderTemplateAssignmentByIDRow includes JOINs for names
	assignment, err := h.dbQueries.GetOtpSenderTemplateAssignmentByID(logCtx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found."})
		} else {
			slog.ErrorContext(logCtx, "Error fetching assignment by ID", slog.Any("error", err), slog.Int("id", int(id)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve assignment."})
		}
		return
	}
	c.JSON(http.StatusOK, assignment)
}

// List OTP Sender-Template Assignments with pagination and filtering
func (h *OtpSenderTemplateAssignmentHandler) List(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListOtpSenderTemplateAssignments")
	// limit, offset := utils.ParsePagination(c)
	limitParam, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 32)
	offsetParam, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	limit := int32(limitParam)
	offset := int32(offsetParam)

	listParams := database.ListOtpSenderTemplateAssignmentsParams{
		LimitVal:  limit,
		OffsetVal: offset,
	}

	countFilterParams := database.CountOtpSenderTemplateAssignmentsParams{}

	if senderIDStr := c.Query("sender_id"); senderIDStr != "" {
		if senderIDVal, err := strconv.ParseInt(senderIDStr, 10, 32); err == nil {
			senderId := int32(senderIDVal)
			listParams.OtpAlternativeSenderID = &senderId
			countFilterParams.OtpAlternativeSenderID = &senderId
		}
	}

	if templateIDStr := c.Query("template_id"); templateIDStr != "" {
		if templateIDVal, err := strconv.ParseInt(templateIDStr, 10, 32); err == nil {
			templateId := int32(templateIDVal)
			listParams.OtpMessageTemplateID = &templateId
			countFilterParams.OtpMessageTemplateID = &templateId
		}
	}

	if mnoIDStr := c.Query("mno_id"); mnoIDStr != "" {
		if mnoIDVal, err := strconv.ParseInt(mnoIDStr, 10, 32); err == nil {
			mnoId := int32(mnoIDVal)
			listParams.MnoID = &mnoId
		}
	}

	if status := c.Query("status"); status != "" {
		listParams.Status = &status
	}

	// Add other filters similarly

	// Assuming ListOtpSenderTemplateAssignmentsRow includes JOINs for names
	assignments, err := h.dbQueries.ListOtpSenderTemplateAssignments(logCtx, listParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list assignments", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve assignments"})
		return
	}
	// Assuming CountOtpSenderTemplateAssignments with similar filter params
	total, _ := h.dbQueries.CountOtpSenderTemplateAssignments(logCtx, countFilterParams)

	c.JSON(http.StatusOK, dto.PaginatedListResponse{
		Data:       assignments,
		Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
	})
}

// Update an OTP Sender-Template Assignment
func (h *OtpSenderTemplateAssignmentHandler) Update(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateOtpSenderTemplateAssignment")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID."})
		return
	}
	var req dto.UpdateOtpSenderTemplateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	params := database.UpdateOtpSenderTemplateAssignmentParams{
		ID: int32(id),
	}

	if req.OtpAlternativeSenderID != nil {
		params.OtpAlternativeSenderID = req.OtpAlternativeSenderID
	}
	if req.OtpMessageTemplateID != nil {
		params.OtpMessageTemplateID = req.OtpMessageTemplateID
	}
	if req.MnoID != nil {
		params.MnoID = req.MnoID
	} else if req.ClearMnoID != nil && *req.ClearMnoID {
		params.MnoID = nil
	}
	if req.Priority != nil {
		params.Priority = req.Priority
	}
	if req.Status != nil {
		params.Status = req.Status
	}

	updatedAssignment, err := h.dbQueries.UpdateOtpSenderTemplateAssignment(logCtx, params)
	if err != nil {
		// Handle pgx.ErrNoRows, unique constraint violations, FK violations
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found."})
			return
		}
		slog.ErrorContext(logCtx, "Failed to update assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update assignment."})
		return
	}
	detailedAssignment, _ := h.dbQueries.GetOtpSenderTemplateAssignmentByID(logCtx, updatedAssignment.ID)
	c.JSON(http.StatusOK, detailedAssignment)
}

// Delete an OTP Sender-Template Assignment
func (h *OtpSenderTemplateAssignmentHandler) Delete(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteOtpSenderTemplateAssignment")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID."})
		return
	}
	err = h.dbQueries.DeleteOtpSenderTemplateAssignment(logCtx, int32(id))
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to delete assignment", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete assignment."})
		return
	}
	c.Status(http.StatusNoContent)
}
