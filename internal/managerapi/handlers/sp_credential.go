package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/auth"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/managerapi/handlers/dto"
)

type SPCredentialHandler struct {
	dbQueries database.Querier
	dbPool    *pgxpool.Pool // Needed if Tx required
}

func NewSPCredentialHandler(q database.Querier, pool *pgxpool.Pool) *SPCredentialHandler {
	return &SPCredentialHandler{dbQueries: q, dbPool: pool}
}

// CreateSPCredential handles POST /sp-credentials
func (h *SPCredentialHandler) CreateSPCredential(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "CreateSPCredential")
	var req dto.CreateSPCredentialRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, req.ServiceProviderID) // Add SPID if available early

	// --- Validate SP Exists ---
	_, err := h.dbQueries.GetServiceProviderByID(logCtx, req.ServiceProviderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: service_provider_id %d not found", req.ServiceProviderID)})
		} else {
			slog.ErrorContext(logCtx, "Failed to validate service provider existence", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate service provider"})
		}
		return
	}

	// --- Prepare Params & Hash Secrets ---
	params := database.CreateSPCredentialParams{
		ServiceProviderID: req.ServiceProviderID,
		Protocol:          req.Protocol,
		Status:            "active", // Default to active
	}
	if req.Status != nil {
		params.Status = *req.Status
	}

	if req.Protocol == "smpp" {
		if req.SystemID == nil || req.Password == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required SMPP fields: system_id, password"})
			return
		}
		if req.BindType == nil { // Default bind type if not provided
			defaultBindType := "trx"
			params.BindType = &defaultBindType
		} else {
			params.BindType = req.BindType
		}
		params.SystemID = req.SystemID
		hashedPassword, hashErr := auth.HashPassword(*req.Password)
		if hashErr != nil {
			slog.ErrorContext(logCtx, "Failed to hash password", slog.Any("error", hashErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process credential"})
			return
		}
		params.PasswordHash = &hashedPassword
		logCtx = logging.ContextWithSystemID(logCtx, *req.SystemID)

	} else if req.Protocol == "http" {
		if req.APIKeyIdentifier == nil || req.APIKeySecret == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required HTTP fields: api_key_identifier, api_key_secret"})
			return
		}
		params.ApiKeyIdentifier = req.APIKeyIdentifier
		hashedAPIKey, hashErr := auth.HashAPIKey(*req.APIKeySecret)
		if hashErr != nil {
			slog.ErrorContext(logCtx, "Failed to hash API key", slog.Any("error", hashErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process credential"})
			return
		}
		params.ApiKeyHash = &hashedAPIKey
		// Handle http_config JSONB
		if req.HTTPConfig != nil {
			// Validate if it's valid JSON? json.Valid(req.HTTPConfig)
			if !json.Valid(req.HTTPConfig) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format for http_config"})
				return
			}
			params.HttpConfig = req.HTTPConfig // Assign []byte
		}
		logCtx = logging.ContextWithAPIKeyIdentifier(logCtx, *req.APIKeyIdentifier) // Add helper if needed

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol specified"})
		return
	}

	// --- Create Credential ---
	createdCred, err := h.dbQueries.CreateSPCredential(logCtx, params)
	if err != nil {
		if isUniqueViolationError(err) {
			// Determine which field caused the violation (system_id or api_key_identifier)
			fieldName := "identifier"
			fieldValue := ""
			if req.Protocol == "smpp" {
				fieldName = "system_id"
				fieldValue = *req.SystemID
			}
			if req.Protocol == "http" {
				fieldName = "api_key_identifier"
				fieldValue = *req.APIKeyIdentifier
			}
			slog.WarnContext(logCtx, "SP Credential creation failed: Duplicate identifier", slog.String("field", fieldName), slog.String("value", fieldValue))
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Credential %s '%s' already exists", fieldName, fieldValue)})
			return
		}
		slog.ErrorContext(logCtx, "Failed to create SP credential", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credential"})
		return
	}

	resp := mapDBCredToResponse(createdCred)
	slog.InfoContext(logging.ContextWithCredentialID(logCtx, createdCred.ID), "SP Credential created successfully")
	c.JSON(http.StatusCreated, resp)
}

// ListSPCredentials handles GET /sp-credentials
func (h *SPCredentialHandler) ListSPCredentials(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "ListSPCredentials")

	// --- Filter by Service Provider ID (Optional but recommended) ---
	spIDFilter := 0
	spIDStr := c.Query("sp_id")
	if spIDStr != "" {
		id, err := strconv.ParseInt(spIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sp_id format"})
			return
		}
		spIDFilter = int(id)
		logCtx = logging.ContextWithSPID(logCtx, int32(id))
	}

	spID := int32(spIDFilter)

	limit := int32(10)
	offset := int32(0)

	// --- Fetch Data ---
	listParams := database.ListSPCredentialsParams{
		Column1: spID, // Use Column1 for nullable filter $1
		Limit:   limit,
		Offset:  offset,
	}

	total, err := h.dbQueries.CountSPCredentials(logCtx, spID)
	if err != nil { /* ... handle count error ... */
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK,
			dto.PaginatedListResponse{
				Data:       []dto.SPCredentialResponse{},
				Pagination: dto.PaginationResponse{Total: 0, Limit: limit, Offset: offset},
			})
		return
	}

	creds, err := h.dbQueries.ListSPCredentials(logCtx, listParams)
	if err != nil { /* ... handle list error ... */
		return
	}

	// Map results
	respData := make([]dto.SPCredentialResponse, len(creds))
	for i, cred := range creds {
		respData[i] = mapDBCredToResponse(cred)
	}
	paginatedResp := dto.PaginatedListResponse{
		Data:       respData,
		Pagination: dto.PaginationResponse{Total: total, Limit: 0, Offset: 10},
	}
	c.JSON(http.StatusOK, paginatedResp)
}

// GetSPCredential handles GET /sp-credentials/:id
func (h *SPCredentialHandler) GetSPCredential(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetSPCredential")
	id, err := parseIDParam(c) // Use common helper for /:id
	if err != nil {
		return
	}
	logCtx = logging.ContextWithCredentialID(logCtx, id)

	cred, err := h.dbQueries.GetSPCredentialByID(logCtx, id)
	if err != nil {
		handleGetError(c, logCtx, "SP Credential", err) // Use common helper
		return
	}

	resp := mapDBCredToResponse(cred)
	c.JSON(http.StatusOK, resp)
}

// UpdateSPCredential handles PUT /sp-credentials/:id
func (h *SPCredentialHandler) UpdateSPCredential(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "UpdateSPCredential")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithCredentialID(logCtx, id)

	// Fetch existing first to know the protocol and prevent changing identifiers
	existingCred, err := h.dbQueries.GetSPCredentialByID(logCtx, id)
	if err != nil {
		handleGetError(c, logCtx, "SP Credential", err)
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, existingCred.ServiceProviderID) // Add SPID

	var req dto.UpdateSPCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Prepare params for update
	params := database.UpdateSPCredentialParams{
		ID:         id,
		Scope:      req.Scope,
		Status:     req.Status,
		HttpConfig: req.HTTPConfig, // Directly assign []byte (if req.HTTPConfig is json.RawMessage)
	}

	// Handle password update only if protocol is SMPP and password is provided
	if existingCred.Protocol == "smpp" && req.Password != nil {
		hashedPassword, hashErr := auth.HashPassword(*req.Password)
		if hashErr != nil {
			slog.ErrorContext(logCtx, "Failed to hash new password during update", slog.Any("error", hashErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password update"})
			return
		}
		params.PasswordHash = &hashedPassword
		slog.InfoContext(logCtx, "Password will be updated")
	} else if req.Password != nil {
		slog.WarnContext(logCtx,
			"Password provided in update request, but credential protocol is not SMPP. Ignoring password.", slog.String("protocol", existingCred.Protocol))
	}

	// Handle http_config update only if protocol is HTTP
	if existingCred.Protocol == "http" && req.HTTPConfig != nil {
		if !json.Valid(req.HTTPConfig) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format for http_config"})
			return
		}
		params.HttpConfig = req.HTTPConfig // Assign []byte
		slog.InfoContext(logCtx, "HTTP Config will be updated")
	} else if req.HTTPConfig != nil {
		slog.WarnContext(logCtx,
			"HTTP Config provided in update request, but credential protocol is not HTTP. Ignoring http_config.", slog.String("protocol", existingCred.Protocol))
		params.HttpConfig = nil // Ensure we don't try to update it
	}

	// Update Credential
	updatedCred, err := h.dbQueries.UpdateSPCredential(logCtx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { // Should not happen if GetByID succeeded
			c.JSON(http.StatusNotFound, gin.H{"error": "SP Credential not found"})
		} else {
			slog.ErrorContext(logCtx, "Failed to update SP credential", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update credential"})
		}
		return
	}

	resp := mapDBCredToResponse(updatedCred)
	slog.InfoContext(logCtx, "SP Credential updated successfully")
	c.JSON(http.StatusOK, resp)
}

// DeleteSPCredential handles DELETE /sp-credentials/:id
func (h *SPCredentialHandler) DeleteSPCredential(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "DeleteSPCredential")
	id, err := parseIDParam(c)
	if err != nil {
		return
	}
	logCtx = logging.ContextWithCredentialID(logCtx, id)

	// Fetch existing first maybe to log SP ID? Optional.
	// existingCred, err := h.dbQueries.GetSPCredentialByID(logCtx, id)
	// if err != nil { handleGetError(c, logCtx, "SP Credential", err); return }
	// logCtx = logging.ContextWithSPID(logCtx, existingCred.ServiceProviderID)

	err = h.dbQueries.DeleteSPCredential(logCtx, id)
	if err != nil {
		// Check for pgx.ErrNoRows? Unlikely on DELETE.
		slog.ErrorContext(logCtx, "Failed to delete SP credential", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credential"})
		return
	}

	slog.InfoContext(logCtx, "SP Credential deleted successfully")
	c.Status(http.StatusNoContent)
}

// --- Helper ---

// mapDBCredToResponse converts DB model to API DTO, omitting sensitive data.
func mapDBCredToResponse(cred database.SpCredential) dto.SPCredentialResponse {
	resp := dto.SPCredentialResponse{
		BaseSPCredential: dto.BaseSPCredential{
			ID:                cred.ID,
			ServiceProviderID: cred.ServiceProviderID,
			Protocol:          cred.Protocol,
			Status:            cred.Status,
			CreatedAt:         cred.CreatedAt.Time,
			UpdatedAt:         cred.UpdatedAt.Time,
		},
		SystemID:         cred.SystemID, // sql.NullString maps ok to *string via omitempty? Check marshal. Assign pointers explicitly if needed.
		BindType:         cred.BindType,
		APIKeyIdentifier: cred.ApiKeyIdentifier,
		HTTPConfig:       cred.HttpConfig, // Assign []byte to json.RawMessage
	}
	return resp
}
