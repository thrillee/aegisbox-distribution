package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/thrillee/aegisbox/internal/database"
)

// SenderHealthTracker manages sender ID health metrics and blocking
type SenderHealthTracker struct {
	dbQueries database.Querier
	logger    *slog.Logger
}

// SenderHealthConfig holds configuration for sender health tracking
type SenderHealthConfig struct {
	BlockThreshold   float64 // Delivery rate below this triggers auto-block
	UnblockThreshold float64 // Delivery rate above this for potential unblock
	DLRTimeoutSecs   int32   // Seconds to wait for DLR
	WindowSize       int32   // Number of messages in rolling window
}

func NewSenderHealthTracker(q database.Querier, logger *slog.Logger) *SenderHealthTracker {
	return &SenderHealthTracker{
		dbQueries: q,
		logger:    logger,
	}
}

// RecordDeliveryResult records a delivery outcome for a sender-MNO combination
func (t *SenderHealthTracker) RecordDeliveryResult(
	ctx context.Context,
	senderID int32,
	mnoID int32,
	messageSegmentID int64,
	status string, // "delivered", "failed", "no_dlr"
	errorCode *string,
	sentAt time.Time,
	dlrReceivedAt *time.Time,
) error {
	// Record attempt in ring buffer
	dlrTime := pgtype.Timestamptz{}
	if dlrReceivedAt != nil {
		dlrTime = pgtype.Timestamptz{Time: *dlrReceivedAt, Valid: true}
	}

	_, err := t.dbQueries.RecordSenderDeliveryAttempt(ctx, database.RecordSenderDeliveryAttemptParams{
		SenderID:         senderID,
		MnoID:            mnoID,
		MessageSegmentID: &messageSegmentID,
		Status:           status,
		ErrorCode:        errorCode,
		SentAt:           pgtype.Timestamptz{Time: sentAt, Valid: true},
		DlrReceivedAt:    dlrTime,
		RecordedDate:     pgtype.Date{Time: sentAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to record delivery attempt: %w", err)
	}

	// Recalculate health metrics
	err = t.updateHealthMetrics(ctx, senderID, mnoID, status)
	if err != nil {
		t.logger.Error("Failed to update health metrics",
			slog.Int("sender_id", int(senderID)),
			slog.Int("mno_id", int(mnoID)),
			slog.Any("error", err))
		// Don't fail the whole operation if health update fails
	}

	return nil
}

// IsSenderBlockedForMNO checks if a sender ID is blocked for a specific MNO
func (t *SenderHealthTracker) IsSenderBlockedForMNO(ctx context.Context, senderIDString string, mnoID int32) (bool, error) {
	// For now, we'll need to pass the actual sender ID, not the string
	// This method signature needs refinement based on how we track senders
	// TODO: Implement proper sender ID lookup or change signature
	return false, nil
}

// IsSenderBlockedForMNOByID checks if a sender ID is blocked using the ID directly
func (t *SenderHealthTracker) IsSenderBlockedForMNOByID(ctx context.Context, senderID int32, mnoID int32) (bool, error) {
	health, err := t.dbQueries.GetSenderIdHealthForMNO(ctx, database.GetSenderIdHealthForMNOParams{
		SenderID: senderID,
		MnoID:    mnoID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No health record yet - not blocked
			return false, nil
		}
		return false, fmt.Errorf("failed to get sender health: %w", err)
	}

	return health.IsBlocked != nil && *health.IsBlocked, nil
}

// updateHealthMetrics recalculates health from ring buffer and updates health table
func (t *SenderHealthTracker) updateHealthMetrics(
	ctx context.Context,
	senderID int32,
	mnoID int32,
	latestStatus string,
) error {
	// Get config for window size and thresholds
	config, err := t.dbQueries.GetGlobalRoutingConfig(ctx)
	if err != nil {
		// Use defaults if config not found
		t.logger.Warn("Failed to get routing config, using defaults", slog.Any("error", err))
		windowSize := int32(1000)
		blockThreshold := pgtype.Numeric{Int: big.NewInt(5000), Exp: -2, Valid: true}   // 50.00
		unblockThreshold := pgtype.Numeric{Int: big.NewInt(7000), Exp: -2, Valid: true} // 70.00
		config = database.RoutingConfig{
			SenderIDHealthWindowSize: &windowSize,
			SenderIDBlockThreshold:   blockThreshold,
			SenderIDUnblockThreshold: unblockThreshold,
		}
	}

	// Calculate metrics from ring buffer
	windowSize := int32(1000)
	if config.SenderIDHealthWindowSize != nil {
		windowSize = *config.SenderIDHealthWindowSize
	}
	metrics, err := t.dbQueries.CalculateSenderHealthFromRingBuffer(ctx, database.CalculateSenderHealthFromRingBufferParams{
		SenderID: senderID,
		MnoID:    mnoID,
		Limit:    windowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to calculate health metrics: %w", err)
	}

	// Get current health record to check previous blocking status
	currentHealth, err := t.dbQueries.GetSenderIdHealthForMNO(ctx, database.GetSenderIdHealthForMNOParams{
		SenderID: senderID,
		MnoID:    mnoID,
	})

	isCurrentlyBlocked := false
	autoBlocked := false
	blockedReason := ""
	if err == nil {
		isCurrentlyBlocked = currentHealth.IsBlocked != nil && *currentHealth.IsBlocked
		autoBlocked = currentHealth.AutoBlocked != nil && *currentHealth.AutoBlocked
		if currentHealth.BlockedReason != nil {
			blockedReason = *currentHealth.BlockedReason
		}
	}

	// Determine if should block based on thresholds
	blockThreshold := numericToFloat64(config.SenderIDBlockThreshold)

	shouldBlock := metrics.DeliveryRate < blockThreshold && metrics.TotalSent >= 10 // Min 10 messages before blocking

	if shouldBlock && !isCurrentlyBlocked {
		blockedReason = fmt.Sprintf("Auto-blocked: delivery rate %.2f%% below threshold %.2f%%",
			metrics.DeliveryRate, blockThreshold)
		isCurrentlyBlocked = true
		autoBlocked = true
	}

	// Determine timestamps
	now := time.Now()
	var lastSuccess, lastFailure, lastNoDLR, lastAttempt *time.Time
	lastAttempt = &now

	if latestStatus == "delivered" {
		lastSuccess = &now
	} else if latestStatus == "failed" {
		lastFailure = &now
	} else if latestStatus == "no_dlr" {
		lastNoDLR = &now
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

	lastNoDlrTS := pgtype.Timestamptz{}
	if lastNoDLR != nil {
		lastNoDlrTS = pgtype.Timestamptz{Time: *lastNoDLR, Valid: true}
	}

	lastAttemptTS := pgtype.Timestamptz{}
	if lastAttempt != nil {
		lastAttemptTS = pgtype.Timestamptz{Time: *lastAttempt, Valid: true}
	}

	// Upsert health record
	_, err = t.dbQueries.UpsertSenderIdHealth(ctx, database.UpsertSenderIdHealthParams{
		SenderID:        senderID,
		MnoID:           mnoID,
		TotalSent:       metrics.TotalSent,
		TotalDelivered:  metrics.TotalDelivered,
		TotalFailed:     metrics.TotalFailed,
		TotalNoDlr:      metrics.TotalNoDlr,
		DeliveryRate:    float64ToPgNumeric(metrics.DeliveryRate),
		DlrResponseRate: float64ToPgNumeric(metrics.DlrResponseRate),
		HealthScore:     float64ToPgNumeric(metrics.HealthScore),
		IsBlocked:       &isCurrentlyBlocked,
		BlockedReason:   &blockedReason,
		AutoBlocked:     &autoBlocked,
		LastSuccessAt:   lastSuccessTS,
		LastFailureAt:   lastFailureTS,
		LastNoDlrAt:     lastNoDlrTS,
		LastAttemptAt:   lastAttemptTS,
		WindowSize:      config.SenderIDHealthWindowSize,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert sender health: %w", err)
	}

	if shouldBlock && !isCurrentlyBlocked {
		t.logger.Warn("Sender ID auto-blocked due to poor health",
			slog.Int("sender_id", int(senderID)),
			slog.Int("mno_id", int(mnoID)),
			slog.Float64("delivery_rate", metrics.DeliveryRate),
			slog.String("reason", blockedReason))
	}

	return nil
}
