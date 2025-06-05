package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv" // For parsing query params

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn" // For PgError
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database" // Your SQLC package
	"github.com/thrillee/aegisbox/internal/logging"  // Your logging package
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
	// "github.com/thrillee/aegisbox/internal/utils" // Your utils for parseIDParam, parsePagination etc.
)

type OtpAlternativeSenderHandler struct {
	dbQueries database.Querier
	// dbPool *pgxpool.Pool // If transactions are needed directly in handler
}

func NewOtpAlternativeSenderHandler(q database.Querier) *OtpAlternativeSenderHandler {
	return &OtpAlternativeSenderHandler{dbQueries: q}
}

// CreateOtpAlternativeSender godoc
// @Summary Create a new OTP alternative sender
// @Description Adds a new OTP alternative sender to the system.
// @Tags otp-alternative-senders
// @Accept json
// @Produce json
// @Param sender body dto.CreateOtpAlternativeSenderRequest true "OTP Alternative Sender to create"
// @Success 201 {object} dto.OtpAlternativeSenderResponse
// @Failure 400 {object} gin.H{"error": "message"} "Invalid request body"
// @Failure 409 {object} gin.H{"error": "message"} "Conflict, e.g., sender_id_string already exists"
// @Failure 500 {object} gin.H{"error": "message"} "Internal server error"
// @Router /otp-alternative-senders [post]
func (h *OtpAlternativeSenderHandler) Create(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateOtpAlternativeSender")
	var req dto.CreateOtpAlternativeSenderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	params := database.CreateOtpAlternativeSenderParams{
		SenderIDString: req.SenderIDString,
		MaxUsageCount:  req.MaxUsageCount,
	}
	if req.MnoID != nil {
		params.MnoID = req.MnoID
	}
	if req.Status != "" {
		params.Status = req.Status
	} else {
		params.Status = "active" // Default status
	}
	if req.ResetIntervalHours != nil {
		params.ResetIntervalHours = req.ResetIntervalHours
	}
	if req.Notes != nil {
		params.Notes = req.Notes
	}
	if req.ServiceProviderID != nil {
		params.ServiceProviderID = req.ServiceProviderID
	}

	// --- Optional: Validate MNO ID and SP ID exist if provided ---
	if params.MnoID != nil {
		_, err := h.dbQueries.GetMNOByID(logCtx, *params.MnoID) // Assuming GetMNOByID exists
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid mno_id: MNO %d not found", params.MnoID)})
			} else {
				slog.ErrorContext(logCtx, "Failed to validate MNO ID", slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate MNO ID"})
			}
			return
		}
	}
	if params.ServiceProviderID != nil {
		_, err := h.dbQueries.GetServiceProviderByID(logCtx, *params.ServiceProviderID) // Assuming GetServiceProviderByID exists
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid service_provider_id: Service Provider %d not found", params.ServiceProviderID)})
			} else {
				slog.ErrorContext(logCtx, "Failed to validate Service Provider ID", slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Service Provider ID"})
			}
			return
		}
	}
	// --- End Validation ---

	sender, err := h.dbQueries.CreateOtpAlternativeSender(logCtx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			slog.WarnContext(logCtx, "Failed to create OTP alternative sender due to unique constraint", slog.String("detail", pgErr.Detail))
			c.JSON(http.StatusConflict, gin.H{"error": "OTP alternative sender with this sender_id_string already exists."})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create OTP alternative sender", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OTP alternative sender"})
		return
	}

	// Refetch to include potential JOINed data like MNO name if GetOtpAlternativeSenderByIDRow includes it
	detailedSender, _ := h.dbQueries.GetOtpAlternativeSenderByID(logCtx, sender.ID) // Error handling omitted for brevity
	c.JSON(http.StatusCreated, detailedSender)
}

// GetByID godoc
// @Summary Get an OTP alternative sender by ID
// @Description Retrieves details of a specific OTP alternative sender.
// @Tags otp-alternative-senders
// @Produce json
// @Param id path int true "OTP Alternative Sender ID"
// @Success 200 {object} dto.OtpAlternativeSenderResponse
// @Failure 404 {object} gin.H{"error": "message"} "Sender not found"
// @Failure 500 {object} gin.H{"error": "message"} "Internal server error"
// @Router /otp-alternative-senders/{id} [get]
func (h *OtpAlternativeSenderHandler) GetByID(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetOtpAlternativeSenderByID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Assuming GetOtpAlternativeSenderByIDRow includes JOINs for MNO name, SP name
	sender, err := h.dbQueries.GetOtpAlternativeSenderByID(logCtx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "OTP alternative sender not found"})
		} else {
			slog.ErrorContext(logCtx, "Failed to get OTP alternative sender by ID", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP alternative sender"})
		}
		return
	}
	c.JSON(http.StatusOK, sender)
}

// List godoc
// @Summary List OTP alternative senders
// @Description Retrieves a paginated list of OTP alternative senders. Supports filtering by status, mno_id, service_provider_id.
// @Tags otp-alternative-senders
// @Produce json
// @Param limit query int false "Limit number of results" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Param status query string false "Filter by status (active, inactive, depleted_awaiting_reset)"
// @Param mno_id query int false "Filter by MNO ID"
// @Param service_provider_id query int false "Filter by Service Provider ID"
// @Success 200 {object} dto.PaginatedListResponse{Data=[]dto.OtpAlternativeSenderResponse}
// @Failure 500 {object} gin.H{"error": "message"} "Internal server error"
// @Router /otp-alternative-senders [get]
func (h *OtpAlternativeSenderHandler) List(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListOtpAlternativeSenders")
	// limit, offset := utils.ParsePagination(c)
	limitParam, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 32)
	offsetParam, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	limit := int32(limitParam)
	offset := int32(offsetParam)

	listParams := database.ListOtpAlternativeSendersParams{
		LimitVal:  limit,
		OffsetVal: offset,
	}
	// Example for filter params, your ListOtpAlternativeSendersParams might differ
	if statusVal := c.Query("status"); statusVal != "" {
		listParams.Status = &statusVal
	}
	if mnoIDStr := c.Query("mno_id"); mnoIDStr != "" {
		if mnoIDVal, err := strconv.ParseInt(mnoIDStr, 10, 32); err == nil {
			mnoID := int32(mnoIDVal)
			listParams.MnoID = &mnoID
		}
	}
	if spIDStr := c.Query("sp_id"); spIDStr != "" {
		if spIDVal, err := strconv.ParseInt(spIDStr, 10, 32); err == nil {
			spID := int32(spIDVal)
			listParams.ServiceProviderID = &spID
		}
	}

	senders, err := h.dbQueries.ListOtpAlternativeSenders(logCtx, listParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list OTP alternative senders", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP alternative senders"})
		return
	}

	countParams := database.CountOtpAlternativeSendersParams{
		Status:            listParams.Status,
		MnoID:             listParams.MnoID,
		ServiceProviderID: listParams.ServiceProviderID,
	} // Assuming a similar params struct for count query
	total, err := h.dbQueries.CountOtpAlternativeSenders(logCtx, countParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to count OTP alternative senders", slog.Any("error", err))
		// Continue with an estimated total or handle error
	}

	c.JSON(http.StatusOK, dto.PaginatedListResponse{
		Data:       senders,
		Pagination: dto.PaginationResponse{Total: total, Limit: limit, Offset: offset},
	})
}

// Update godoc
// @Summary Update an OTP alternative sender
// @Description Updates details of an existing OTP alternative sender.
// @Tags otp-alternative-senders
// @Accept json
// @Produce json
// @Param id path int true "OTP Alternative Sender ID"
// @Param sender body dto.UpdateOtpAlternativeSenderRequest true "OTP Alternative Sender data to update"
// @Success 200 {object} dto.OtpAlternativeSenderResponse
// @Failure 400 {object} gin.H{"error": "message"} "Invalid request body or ID"
// @Failure 404 {object} gin.H{"error": "message"} "Sender not found"
// @Failure 409 {object} gin.H{"error": "message"} "Conflict, e.g., sender_id_string already exists"
// @Failure 500 {object} gin.H{"error": "message"} "Internal server error"
// @Router /otp-alternative-senders/{id} [put]
func (h *OtpAlternativeSenderHandler) Update(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateOtpAlternativeSender")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req dto.UpdateOtpAlternativeSenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	params := database.UpdateOtpAlternativeSenderParams{
		ID: int32(id),
		// Initialize all pgtype fields to Valid:false, then set if value present in request
	}

	if req.SenderIDString != nil {
		params.SenderIDString = req.SenderIDString
	}
	if req.MnoID != nil {
		params.MnoID = req.MnoID
	} else if req.ClearMnoID != nil && *req.ClearMnoID {
		params.MnoID = nil
	}
	if req.Status != nil {
		params.Status = req.Status
	}
	if req.MaxUsageCount != nil {
		params.MaxUsageCount = req.MaxUsageCount
	}
	if req.CurrentUsageCount != nil {
		params.CurrentUsageCount = req.CurrentUsageCount
	}
	if req.ResetIntervalHours != nil {
		params.ResetIntervalHours = req.ResetIntervalHours
	}
	if req.LastResetAt != nil {
		params.LastResetAt = pgtype.Timestamptz{Time: *req.LastResetAt, Valid: true}
	}
	if req.Notes != nil {
		params.Notes = req.Notes
	}
	if req.ServiceProviderID != nil {
		params.ServiceProviderID = req.ServiceProviderID
	} else if req.ClearServiceProviderID != nil && *req.ClearServiceProviderID {
		params.ServiceProviderID = nil
	}

	// --- Optional: Validate MNO ID and SP ID if being changed ---
	if params.MnoID != nil && *params.MnoID != 0 { // Check if MNO ID is being set to a non-zero value
		_, err := h.dbQueries.GetMNOByID(logCtx, *params.MnoID)
		if err != nil { /* ... handle MNO not found ... */
		}
	}
	if params.ServiceProviderID != nil && *params.ServiceProviderID != 0 {
		_, err := h.dbQueries.GetServiceProviderByID(logCtx, *params.ServiceProviderID)
		if err != nil { /* ... handle SP not found ... */
		}
	}
	// --- End Validation ---

	updatedSender, err := h.dbQueries.UpdateOtpAlternativeSender(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "OTP alternative sender not found"})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "Update failed: sender_id_string may already exist for another sender."})
			return
		}
		slog.ErrorContext(logCtx, "Failed to update OTP alternative sender", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update OTP alternative sender"})
		return
	}
	detailedSender, _ := h.dbQueries.GetOtpAlternativeSenderByID(logCtx, updatedSender.ID)
	c.JSON(http.StatusOK, detailedSender)
}

// Delete godoc
// @Summary Delete an OTP alternative sender
// @Description Deletes a specific OTP alternative sender by ID.
// @Tags otp-alternative-senders
// @Produce json
// @Param id path int true "OTP Alternative Sender ID"
// @Success 204 "Successfully deleted"
// @Failure 400 {object} gin.H{"error": "message"} "Invalid ID format"
// @Failure 404 {object} gin.H{"error": "message"} "Sender not found (already deleted or never existed)"
// @Failure 500 {object} gin.H{"error": "message"} "Internal server error"
// @Router /otp-alternative-senders/{id} [delete]
func (h *OtpAlternativeSenderHandler) Delete(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteOtpAlternativeSender")
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Optional: Check if sender exists before deleting to return 404 if not found
	// _, err = h.dbQueries.GetOtpAlternativeSenderByID(logCtx, int32(id))
	// if errors.Is(err, pgx.ErrNoRows) {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "OTP alternative sender not found"})
	// 	return
	// }

	err = h.dbQueries.DeleteOtpAlternativeSender(logCtx, int32(id))
	if err != nil {
		// If err is pgx.ErrNoRows here, it means it was already deleted or never existed.
		// Depending on desired idempotency, this might not be a 500 error.
		// For now, any error from delete is treated as 500.
		slog.ErrorContext(logCtx, "Failed to delete OTP alternative sender", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete OTP alternative sender"})
		return
	}
	c.Status(http.StatusNoContent)
}

// In a suitable handler, e.g., OtpConfigHandler or OtpAlternativeSenderHandler
// ListSendersWithTemplates godoc
// @Summary List alternative senders with their assigned templates
// @Description Retrieves all active alternative senders and their actively assigned templates.
// @Tags otp-config
// @Produce json
// @Success 200 {array} dto.SenderWithTemplatesResponse
// @Failure 500 {object} gin.H{"error": "message"} "Internal server error"
// @Router /otp-config/senders-with-templates [get]
func (h *OtpAlternativeSenderHandler) ListSendersWithTemplates(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListSendersWithTemplates")

	params := database.ListAlternativeSendersWithAssignedTemplatesParams{}

	if spIDStr := c.Query("sp_id"); spIDStr != "" {
		if spIDVal, err := strconv.ParseInt(spIDStr, 10, 32); err == nil {
			spId := int32(spIDVal)
			params.FilterServiceProviderID = &spId
		}
	}

	if senderIDstr := c.Query("sender_id"); senderIDstr != "" {
		if senderIDVal, err := strconv.ParseInt(senderIDstr, 10, 32); err == nil {
			sId := int32(senderIDVal)
			params.FilterServiceProviderID = &sId
		}
	}

	rows, err := h.dbQueries.ListAlternativeSendersWithAssignedTemplates(logCtx, params)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to list alternative senders with templates", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sender and template configurations"})
		return
	}

	// Group results by alternative_sender_id
	groupedResults := make(map[int32]*dto.SenderWithTemplatesResponse)
	order := []int32{} // To maintain order of senders

	for _, row := range rows {
		senderID := row.AlternativeSenderID
		if _, exists := groupedResults[senderID]; !exists {
			order = append(order, senderID)
			groupedResults[senderID] = &dto.SenderWithTemplatesResponse{
				AlternativeSenderID:     senderID,
				SenderIDString:          row.SenderIDString,
				AlternativeSenderStatus: row.AlternativeSenderStatus,
				AssignedTemplates:       []dto.AssignedTemplateInfo{},
			}
			if row.SenderMnoID != nil {
				groupedResults[senderID].MnoID = row.SenderMnoID
			}
			if row.SenderMnoName != nil {
				groupedResults[senderID].MnoName = row.SenderMnoName
			}
			if row.SenderServiceProviderID != nil {
				groupedResults[senderID].ServiceProviderID = row.SenderServiceProviderID
			}
			if row.SenderServiceProviderName != nil {
				groupedResults[senderID].ServiceProviderName = row.SenderServiceProviderName
			}
		}

		// Add template info if assignment_id is valid (meaning there was a join match for assignment)
		if row.AssignmentID != nil {
			templateInfo := dto.AssignedTemplateInfo{
				TemplateID:         *row.TemplateID,   // Assuming TemplateID is NOT NULL if AssignmentID is valid
				TemplateName:       *row.TemplateName, // Assuming TemplateName is NOT NULL
				AssignmentID:       *row.AssignmentID,
				AssignmentStatus:   *row.AssignmentStatus,
				AssignmentPriority: *row.AssignmentPriority,
			}
			if row.AssignmentMnoID != nil {
				templateInfo.MnoID = row.AssignmentMnoID
			}
			if row.AssignmentMnoName != nil {
				templateInfo.MnoName = row.AssignmentMnoName
			}
			groupedResults[senderID].AssignedTemplates = append(groupedResults[senderID].AssignedTemplates, templateInfo)
		}
	}

	// Convert map to slice to maintain order
	finalResponse := make([]dto.SenderWithTemplatesResponse, 0, len(order))
	for _, id := range order {
		finalResponse = append(finalResponse, *groupedResults[id])
	}

	c.JSON(http.StatusOK, finalResponse)
}
