package handlers

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
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
