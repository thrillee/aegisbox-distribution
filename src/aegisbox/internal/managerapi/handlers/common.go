package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	DefaultLimit  = 20
	MaxLimit      = 100
	DefaultOffset = 0
)

// parsePagination extracts limit and offset from query params with validation and defaults.
func parsePagination(c *gin.Context) (limit, offset int32) {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultLimit))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(DefaultOffset))

	limit64, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit64 <= 0 {
		limit = DefaultLimit
	} else if limit64 > MaxLimit {
		slog.WarnContext(c.Request.Context(), "Requested limit exceeds maximum, capping.", slog.Int64("requested", limit64), slog.Int("max", MaxLimit))
		limit = MaxLimit
	} else {
		limit = int32(limit64)
	}

	offset64, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset64 < 0 {
		offset = DefaultOffset
	} else {
		offset = int32(offset64)
	}

	return limit, offset
}

// Helper to check for unique violation errors (adapt based on actual pgx error type)
func isUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError // Use correct type from pgx driver
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Helper to check for foreign key violation errors
func isForeignKeyViolationError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// handleGetError standardizes error responses for GET /resource/:id endpoints.
func handleGetError(c *gin.Context, logCtx context.Context, resourceName string, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		slog.InfoContext(logCtx, resourceName+" not found")
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", resourceName)})
	} else {
		slog.ErrorContext(logCtx, fmt.Sprintf("Failed to get %s from DB", resourceName), slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve %s", resourceName)})
	}
}

// parseIDParam extracts and validates an ID path parameter. Writes error response if invalid.
func parseIDParam(c *gin.Context) (int32, error) {
	idStr := c.Param("id") // Assumes path parameter is named "id"
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "Invalid ID parameter", slog.String("id", idStr), slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format in URL path"})
		return 0, err // Return error to stop handler execution
	}
	return int32(id), nil
}

// parseIDParam specifically for connection ID (conn_id)
func parseConnIDParam(c *gin.Context) (int32, error) {
	idStr := c.Param("conn_id") // Assumes path parameter is named "conn_id"
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "Invalid Connection ID parameter", slog.String("conn_id", idStr), slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Connection ID format in URL path"})
		return 0, err
	}
	return int32(id), nil
}

// parseIDParam specifically for mno ID (mno_id)
func parseMnoIDParam(c *gin.Context) (int32, error) {
	idStr := c.Param("mno_id") // Assumes path parameter is named "mno_id"
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "Invalid MNO ID parameter", slog.String("mno_id", idStr), slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MNO ID format in URL path"})
		return 0, err
	}
	return int32(id), nil
}
