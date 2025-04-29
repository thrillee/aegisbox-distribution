package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog" // Use slog
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
)

// Compile-time check
var _ Router = (*DefaultRouter)(nil)

type DefaultRouter struct {
	dbQueries database.Querier
}

func NewDefaultRouter(q database.Querier) *DefaultRouter {
	return &DefaultRouter{dbQueries: q}
}

func (r *DefaultRouter) Route(ctx context.Context, msisdn string) (*mno.RoutingResult, error) {
	logCtx := ctx                                  // Use incoming context which might have trace IDs etc.
	result := &mno.RoutingResult{ShouldUse: false} // Default to no route

	lookupPrefix := strings.TrimPrefix(msisdn, "+")
	if lookupPrefix == "" {
		result.Error = fmt.Errorf("invalid MSISDN format for routing: %s", msisdn)
		slog.WarnContext(logCtx, "Routing failed due to invalid MSISDN format", slog.String("msisdn", msisdn))
		return result, nil // Logical error, not internal
	}

	// Use enriched context if message details are available
	logCtx = logging.ContextWithMSISDN(logCtx, msisdn) // Add helper if needed

	slog.DebugContext(logCtx, "Attempting to find route", slog.String("prefix", lookupPrefix))

	// This query returns only the mno_id (int32)
	routedMnoID, err := r.dbQueries.GetApplicableRoutingRule(logCtx, lookupPrefix)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This is a valid outcome, not an internal error
			result.Error = fmt.Errorf("no route rule matched for prefix of %s", msisdn)
			slog.InfoContext(logCtx, "No applicable routing rule found") // Info level as it's expected outcome
			return result, nil                                           // Return nil error, result indicates no route
		}
		// This indicates a DB problem
		result.Error = fmt.Errorf("db error finding route: %w", err)
		slog.ErrorContext(logCtx, "Database error during routing lookup", slog.Any("error", err))
		return result, err // Return the internal DB error as well
	}

	// Route found
	result.ShouldUse = true
	result.MNOID = routedMnoID
	slog.InfoContext(logCtx, "Route found successfully", slog.Int("mno_id", int(routedMnoID)), slog.String("mno_id", msisdn))
	return result, nil
}
