package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/thrillee/aegisbox/internal/database"
)

// RouteHealthTracker tracks route-level performance and availability
type RouteHealthTracker struct {
	dbQueries database.Querier
	logger    *slog.Logger
}

// NewRouteHealthTracker creates a new route health tracker
func NewRouteHealthTracker(q database.Querier, logger *slog.Logger) *RouteHealthTracker {
	return &RouteHealthTracker{
		dbQueries: q,
		logger:    logger,
	}
}

// RecordRouteAttempt records the outcome of submitting a message to a route. Failed submissions
// (connection or operator errors) count toward taking the route out of service.
func (t *RouteHealthTracker) RecordRouteAttempt(
	ctx context.Context,
	routingAssignmentID int32,
	senderID *int32, // the message's approved sender; not used for route health (see below)
	messageID *int64,
	success bool,
	errorCode *string,
	errorDescription *string,
) error {
	return t.record(ctx, routingAssignmentID, messageID, success, true, errorCode, errorDescription)
}

// RecordDeliveryOutcome records a DLR result for a route. It feeds the route's health score, but a
// failed delivery (UNDELIV, EXPIRED and the like) is about the handset or number, not the route, so
// it does not count toward taking the route out of service. A delivered message still shows the
// route works and clears its failure count.
func (t *RouteHealthTracker) RecordDeliveryOutcome(
	ctx context.Context,
	routingAssignmentID int32,
	messageID *int64,
	delivered bool,
	errorCode *string,
) error {
	return t.record(ctx, routingAssignmentID, messageID, delivered, false, errorCode, nil)
}

func (t *RouteHealthTracker) record(
	ctx context.Context,
	routingAssignmentID int32,
	messageID *int64,
	success bool,
	failureCountsForAvailability bool,
	errorCode *string,
	errorDescription *string,
) error {
	// The router reads only route-wide health (sender_id NULL) when it decides whether a route is
	// usable. Recording attempts per approved sender left that row empty for every message with an
	// approved sender, so failing routes were never taken out. Sender performance is tracked
	// separately in sender_id_health.
	var senderID *int32

	// Record attempt in ring buffer
	_, err := t.dbQueries.RecordRouteDeliveryAttempt(ctx, database.RecordRouteDeliveryAttemptParams{
		RoutingAssignmentID: routingAssignmentID,
		SenderID:            senderID,
		MessageID:           messageID,
		Success:             success,
		ErrorCode:           errorCode,
		ErrorDescription:    errorDescription,
		RecordedDate:        pgtype.Date{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to record route attempt: %w", err)
	}

	// Recalculate and update route health
	err = t.updateRouteHealth(ctx, routingAssignmentID, senderID, success, failureCountsForAvailability)
	if err != nil {
		t.logger.Error("Failed to update route health",
			slog.Int("routing_assignment_id", int(routingAssignmentID)),
			slog.Any("sender_id", senderID),
			slog.Any("error", err))
		// Don't fail the whole operation if health update fails
	}

	return nil
}

// RecordRouteSuccess is a convenience method for recording successful route attempts
func (t *RouteHealthTracker) RecordRouteSuccess(
	ctx context.Context,
	routingAssignmentID int32,
	senderID *int32,
	messageID *int64,
) error {
	return t.RecordRouteAttempt(ctx, routingAssignmentID, senderID, messageID, true, nil, nil)
}

// RecordRouteFailure is a convenience method for recording failed route attempts
func (t *RouteHealthTracker) RecordRouteFailure(
	ctx context.Context,
	routingAssignmentID int32,
	senderID *int32,
	messageID *int64,
	errorCode *string,
	errorDescription *string,
) error {
	return t.RecordRouteAttempt(ctx, routingAssignmentID, senderID, messageID, false, errorCode, errorDescription)
}

// IsRouteAvailable checks if a route is currently available
func (t *RouteHealthTracker) IsRouteAvailable(ctx context.Context, routingAssignmentID int32, senderID *int32) (bool, error) {
	health, err := t.dbQueries.GetRouteHealth(ctx, database.GetRouteHealthParams{
		RoutingAssignmentID: routingAssignmentID,
		SenderID:            senderID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No health record yet - assume available
			return true, nil
		}
		return false, fmt.Errorf("failed to get route health: %w", err)
	}

	return health.IsAvailable != nil && *health.IsAvailable, nil
}

// updateRouteHealth recalculates route health from ring buffer and updates health table
func (t *RouteHealthTracker) updateRouteHealth(
	ctx context.Context,
	routingAssignmentID int32,
	senderID *int32,
	latestSuccess bool,
	failureCountsForAvailability bool,
) error {
	// Get config for window size and thresholds
	config, err := t.dbQueries.GetGlobalRoutingConfig(ctx)
	if err != nil {
		// Use defaults if config not found
		t.logger.Warn("Failed to get routing config, using defaults", slog.Any("error", err))
		windowSize := int32(1000)
		config = database.RoutingConfig{
			RouteHealthWindowSize: &windowSize,
		}
	}

	// Calculate metrics from ring buffer
	windowSize := int32(1000)
	if config.RouteHealthWindowSize != nil {
		windowSize = *config.RouteHealthWindowSize
	}
	metrics, err := t.dbQueries.CalculateRouteHealthFromRingBuffer(ctx, database.CalculateRouteHealthFromRingBufferParams{
		RoutingAssignmentID: routingAssignmentID,
		SenderID:            senderID,
		Limit:               windowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to calculate route health metrics: %w", err)
	}

	// Get current health record to check consecutive counters
	currentHealth, err := t.dbQueries.GetRouteHealth(ctx, database.GetRouteHealthParams{
		RoutingAssignmentID: routingAssignmentID,
		SenderID:            senderID,
	})

	consecutiveFailures := int32(0)
	consecutiveSuccesses := int32(0)
	isAvailable := true
	autoUnavailable := false
	unavailableReason := ""

	if err == nil {
		// Update consecutive counters
		if latestSuccess {
			consecutiveSuccesses = int32PtrToInt32(currentHealth.ConsecutiveSuccesses, 0) + 1
			consecutiveFailures = 0
		} else if failureCountsForAvailability {
			consecutiveFailures = int32PtrToInt32(currentHealth.ConsecutiveFailures, 0) + 1
			consecutiveSuccesses = 0
		} else {
			// A failed delivery leaves the route's availability counters as they were.
			consecutiveFailures = int32PtrToInt32(currentHealth.ConsecutiveFailures, 0)
			consecutiveSuccesses = int32PtrToInt32(currentHealth.ConsecutiveSuccesses, 0)
		}

		// Check if should mark unavailable (5 consecutive failures)
		failureThreshold := int32(5)
		if config.RouteFailureThreshold != nil {
			failureThreshold = *config.RouteFailureThreshold
		}

		if consecutiveFailures >= failureThreshold {
			isAvailable = false
			autoUnavailable = true
			unavailableReason = fmt.Sprintf("Auto-unavailable: %d consecutive failures", consecutiveFailures)
		}

		// Check if should mark available again (3 consecutive successes after being unavailable)
		recoveryThreshold := int32(3)
		if config.RouteRecoveryThreshold != nil {
			recoveryThreshold = *config.RouteRecoveryThreshold
		}

		isCurrentlyAvailable := currentHealth.IsAvailable != nil && *currentHealth.IsAvailable
		if !isCurrentlyAvailable && consecutiveSuccesses >= recoveryThreshold {
			isAvailable = true
			autoUnavailable = false
			unavailableReason = fmt.Sprintf("Auto-recovered: %d consecutive successes", consecutiveSuccesses)

			t.logger.Info("Route auto-recovered and marked available",
				slog.Int("routing_assignment_id", int(routingAssignmentID)),
				slog.Any("sender_id", senderID),
				slog.Int("consecutive_successes", int(consecutiveSuccesses)))
		}

		// Preserve existing availability status if manually set
		if currentHealth.AutoUnavailable != nil && !*currentHealth.AutoUnavailable {
			// Manually marked unavailable - don't auto-recover
			isAvailable = isCurrentlyAvailable
			unavailableReason = stringPtrToStr(currentHealth.UnavailableReason)
		}
	} else {
		// First record - initialize consecutive counters
		if latestSuccess {
			consecutiveSuccesses = 1
		} else if failureCountsForAvailability {
			consecutiveFailures = 1
		}
	}

	// Determine timestamps
	now := time.Now()
	var lastSuccess, lastFailure, lastAttempt *time.Time
	lastAttempt = &now

	if latestSuccess {
		lastSuccess = &now
	} else {
		lastFailure = &now
	}

	// Convert pointers to pgtype
	lastSuccessTS := pgtype.Timestamptz{}
	if lastSuccess != nil {
		lastSuccessTS = pgtype.Timestamptz{Time: *lastSuccess, Valid: true}
	}

	lastFailureTS := pgtype.Timestamptz{}
	if lastFailure != nil {
		lastFailureTS = pgtype.Timestamptz{Time: *lastFailure, Valid: true}
	}

	lastAttemptTS := pgtype.Timestamptz{}
	if lastAttempt != nil {
		lastAttemptTS = pgtype.Timestamptz{Time: *lastAttempt, Valid: true}
	}

	// Upsert route health record
	_, err = t.dbQueries.UpsertRouteHealth(ctx, database.UpsertRouteHealthParams{
		RoutingAssignmentID:  routingAssignmentID,
		SenderID:             senderID,
		TotalAttempts:        metrics.TotalAttempts,
		SuccessfulAttempts:   metrics.SuccessfulAttempts,
		FailedAttempts:       metrics.FailedAttempts,
		SuccessRate:          float64ToPgNumeric(metrics.SuccessRate),
		HealthScore:          float64ToPgNumeric(metrics.HealthScore),
		IsAvailable:          &isAvailable,
		UnavailableReason:    &unavailableReason,
		AutoUnavailable:      &autoUnavailable,
		ConsecutiveFailures:  &consecutiveFailures,
		ConsecutiveSuccesses: &consecutiveSuccesses,
		LastAttemptAt:        lastAttemptTS,
		LastSuccessAt:        lastSuccessTS,
		LastFailureAt:        lastFailureTS,
		WindowSize:           config.RouteHealthWindowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert route health: %w", err)
	}

	// Log if route was auto-marked unavailable
	if !isAvailable && autoUnavailable {
		t.logger.Warn("Route auto-marked unavailable due to consecutive failures",
			slog.Int("routing_assignment_id", int(routingAssignmentID)),
			slog.Any("sender_id", senderID),
			slog.Int("consecutive_failures", int(consecutiveFailures)),
			slog.String("reason", unavailableReason))
	}

	return nil
}

// MarkRouteUnavailable manually marks a route as unavailable
func (t *RouteHealthTracker) MarkRouteUnavailable(
	ctx context.Context,
	routingAssignmentID int32,
	senderID *int32,
	reason string,
) error {
	// Get current health to preserve metrics
	health, err := t.dbQueries.GetRouteHealth(ctx, database.GetRouteHealthParams{
		RoutingAssignmentID: routingAssignmentID,
		SenderID:            senderID,
	})

	autoUnavailable := false
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get route health: %w", err)
		}
		// No existing record - create with defaults
		falseVal := false
		zero := int32(0)
		health = database.RouteHealth{
			RoutingAssignmentID:  routingAssignmentID,
			SenderID:             senderID,
			TotalAttempts:        0,
			SuccessfulAttempts:   0,
			FailedAttempts:       0,
			SuccessRate:          pgtype.Numeric{Valid: false},
			HealthScore:          pgtype.Numeric{Valid: false},
			IsAvailable:          &falseVal,
			UnavailableReason:    &reason,
			AutoUnavailable:      &autoUnavailable,
			ConsecutiveFailures:  &zero,
			ConsecutiveSuccesses: &zero,
		}
	}

	// Update to unavailable with manual flag
	falseVal := false
	_, err = t.dbQueries.UpsertRouteHealth(ctx, database.UpsertRouteHealthParams{
		RoutingAssignmentID:  routingAssignmentID,
		SenderID:             senderID,
		TotalAttempts:        health.TotalAttempts,
		SuccessfulAttempts:   health.SuccessfulAttempts,
		FailedAttempts:       health.FailedAttempts,
		SuccessRate:          health.SuccessRate,
		HealthScore:          health.HealthScore,
		IsAvailable:          &falseVal,
		UnavailableReason:    &reason,
		AutoUnavailable:      &autoUnavailable, // false = manual
		ConsecutiveFailures:  health.ConsecutiveFailures,
		ConsecutiveSuccesses: health.ConsecutiveSuccesses,
		LastAttemptAt:        health.LastAttemptAt,
		LastSuccessAt:        health.LastSuccessAt,
		LastFailureAt:        health.LastFailureAt,
		WindowSize:           health.WindowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to mark route unavailable: %w", err)
	}

	t.logger.Info("Route manually marked unavailable",
		slog.Int("routing_assignment_id", int(routingAssignmentID)),
		slog.Any("sender_id", senderID),
		slog.String("reason", reason))

	return nil
}

// MarkRouteAvailable manually marks a route as available
func (t *RouteHealthTracker) MarkRouteAvailable(
	ctx context.Context,
	routingAssignmentID int32,
	senderID *int32,
) error {
	// Get current health to preserve metrics
	health, err := t.dbQueries.GetRouteHealth(ctx, database.GetRouteHealthParams{
		RoutingAssignmentID: routingAssignmentID,
		SenderID:            senderID,
	})

	reason := ""
	autoUnavailable := false
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get route health: %w", err)
		}
		// No existing record - create with defaults
		trueVal := true
		zero := int32(0)
		health = database.RouteHealth{
			RoutingAssignmentID:  routingAssignmentID,
			SenderID:             senderID,
			TotalAttempts:        0,
			SuccessfulAttempts:   0,
			FailedAttempts:       0,
			SuccessRate:          pgtype.Numeric{Valid: false},
			HealthScore:          pgtype.Numeric{Valid: false},
			IsAvailable:          &trueVal,
			UnavailableReason:    &reason,
			AutoUnavailable:      &autoUnavailable,
			ConsecutiveFailures:  &zero,
			ConsecutiveSuccesses: &zero,
		}
	}

	// Reset consecutive counters when manually marking available
	trueVal := true
	zero := int32(0)
	_, err = t.dbQueries.UpsertRouteHealth(ctx, database.UpsertRouteHealthParams{
		RoutingAssignmentID:  routingAssignmentID,
		SenderID:             senderID,
		TotalAttempts:        health.TotalAttempts,
		SuccessfulAttempts:   health.SuccessfulAttempts,
		FailedAttempts:       health.FailedAttempts,
		SuccessRate:          health.SuccessRate,
		HealthScore:          health.HealthScore,
		IsAvailable:          &trueVal,
		UnavailableReason:    &reason,
		AutoUnavailable:      &autoUnavailable,
		ConsecutiveFailures:  &zero, // Reset on manual mark available
		ConsecutiveSuccesses: &zero,
		LastAttemptAt:        health.LastAttemptAt,
		LastSuccessAt:        health.LastSuccessAt,
		LastFailureAt:        health.LastFailureAt,
		WindowSize:           health.WindowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to mark route available: %w", err)
	}

	t.logger.Info("Route manually marked available",
		slog.Int("routing_assignment_id", int(routingAssignmentID)),
		slog.Any("sender_id", senderID))

	return nil
}
