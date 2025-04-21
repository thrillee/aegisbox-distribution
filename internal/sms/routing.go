package sms

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
)

// Compile-time check
var _ Router = (*DefaultRouter)(nil)

type DefaultRouter struct {
	dbQueries database.Querier
}

func NewDefaultRouter(q database.Querier) *DefaultRouter {
	return &DefaultRouter{dbQueries: q}
}

func (r *DefaultRouter) Route(ctx context.Context, msisdn string) (*RoutingResult, error) {
	result := &RoutingResult{ShouldUse: false} // Default to no route

	lookupPrefix := strings.TrimPrefix(msisdn, "+")
	if lookupPrefix == "" {
		result.Error = fmt.Errorf("invalid MSISDN format for routing: %s", msisdn)
		return result, nil // Logical error, not internal error
	}

	routedMnoID, err := r.dbQueries.GetApplicableRoutingRule(ctx, lookupPrefix)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This is a valid outcome, not an internal error
			log.Printf("No route found for MSISDN prefix %s", lookupPrefix)
			result.Error = fmt.Errorf("no route rule matched")
			return result, nil // Return nil error, result indicates no route
		}
		// This indicates a DB problem
		result.Error = fmt.Errorf("db error finding route: %w", err)
		return result, err // Return the internal DB error as well
	}

	result.ShouldUse = true
	result.MNOID = routedMnoID
	return result, nil
}
