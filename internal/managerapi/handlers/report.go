package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"strconv" // For sp_id parsing

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
)

type ReportHandler struct {
	dbQueries database.Querier
}

func NewReportHandler(q database.Querier) *ReportHandler {
	return &ReportHandler{dbQueries: q}
}

// GetDailyReport handles GET /reports/daily
func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	logCtx := logging.ContextWithHandler(c.Request.Context(), "GetDailyReport")

	// 1. Parse Query Parameters (Date Range Required)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	spIDStr := c.Query("sp_id")
	spEmail := c.Query("sp_email")

	// Date range is mandatory for this report usually
	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameters: start_date, end_date (YYYY-MM-DD)"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
		return
	}
	// Basic validation: end date should not be before start date
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date cannot be before start_date"})
		return
	}

	spIDFilter, err := strconv.ParseInt(spIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sp_id format"})
		return
	}
	logCtx = logging.ContextWithSPID(logCtx, int32(spIDFilter))

	spEmailFilter := ""
	if spEmail != "" {
		spEmailFilter = spEmail
		logCtx = logging.ContextWithSPEmail(logCtx, spEmail) // Add helper
	}

	// 2. Call DB Query
	slog.InfoContext(logCtx, "Fetching daily report",
		slog.String("start_date", startDateStr),
		slog.String("end_date", endDateStr),
		slog.Any("sp_id", spIDStr),
		slog.Any("sp_email", spEmail),
	)

	// 2. Call DB Query
	slog.InfoContext(logCtx, "Fetching daily report", slog.Time("start_date", startDate), slog.Time("end_date", endDate), slog.Any("sp_id", spIDFilter), slog.Any("sp_email", spEmailFilter))
	reportRows, err := h.dbQueries.GetDailyReport(logCtx, database.GetDailyReportParams{
		StartDate: pgtype.Timestamptz{Time: startDate, Valid: true}, // Pass time.Time, query handles date part and timezone
		EndDate:   pgtype.Timestamptz{Time: endDate, Valid: true},
		Column1:   int32(spIDFilter), // Corresponds to $1 placeholder for sp_id filter
		Column2:   spEmailFilter,     // Corresponds to $2 placeholder for sp_email filter
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to fetch daily report from DB", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve daily report"})
		return
	}

	// 3. Map results to DTO
	// respData := make([]database.GetDailyReportRow, len(reportRows))

	c.JSON(http.StatusOK, reportRows) // Return array directly for report
}
