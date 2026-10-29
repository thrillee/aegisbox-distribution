package sms

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/thrillee/aegisbox/internal/database"
)

// routeHealthMockQuerier records the sender ID route health is written under.
type routeHealthMockQuerier struct {
	database.Querier

	attemptSenderIDs []*int32
	upsertSenderIDs  []*int32
}

func (m *routeHealthMockQuerier) RecordRouteDeliveryAttempt(ctx context.Context, arg database.RecordRouteDeliveryAttemptParams) (database.RouteDeliveryAttempt, error) {
	m.attemptSenderIDs = append(m.attemptSenderIDs, arg.SenderID)
	return database.RouteDeliveryAttempt{}, nil
}

func (m *routeHealthMockQuerier) GetGlobalRoutingConfig(ctx context.Context) (database.RoutingConfig, error) {
	return database.RoutingConfig{}, nil
}

func (m *routeHealthMockQuerier) CalculateRouteHealthFromRingBuffer(ctx context.Context, arg database.CalculateRouteHealthFromRingBufferParams) (database.CalculateRouteHealthFromRingBufferRow, error) {
	return database.CalculateRouteHealthFromRingBufferRow{}, nil
}

func (m *routeHealthMockQuerier) GetRouteHealth(ctx context.Context, arg database.GetRouteHealthParams) (database.RouteHealth, error) {
	return database.RouteHealth{}, pgx.ErrNoRows
}

func (m *routeHealthMockQuerier) UpsertRouteHealth(ctx context.Context, arg database.UpsertRouteHealthParams) (database.RouteHealth, error) {
	m.upsertSenderIDs = append(m.upsertSenderIDs, arg.SenderID)
	return database.RouteHealth{}, nil
}

func TestRecordRouteAttempt_RecordsRouteWideHealth(t *testing.T) {
	q := &routeHealthMockQuerier{}
	tracker := NewRouteHealthTracker(q, slog.Default())
	approvedSender := int32(1)
	msgID := int64(99)

	if err := tracker.RecordRouteAttempt(context.Background(), 8, &approvedSender, &msgID, false, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(q.attemptSenderIDs) != 1 || q.attemptSenderIDs[0] != nil {
		t.Fatalf("attempt must be recorded route-wide (sender NULL), got %v", q.attemptSenderIDs)
	}
	if len(q.upsertSenderIDs) != 1 || q.upsertSenderIDs[0] != nil {
		t.Fatalf("route health must be written route-wide (sender NULL), got %v", q.upsertSenderIDs)
	}
}

func TestRecordDeliveryOutcome_FailureDoesNotCountTowardOutage(t *testing.T) {
	q := &routeHealthCountingQuerier{current: database.RouteHealth{ConsecutiveFailures: int32Ptr(2), ConsecutiveSuccesses: int32Ptr(0), IsAvailable: boolPtr(true)}}
	tracker := NewRouteHealthTracker(q, slog.Default())
	msgID := int64(1)

	if err := tracker.RecordDeliveryOutcome(context.Background(), 8, &msgID, false, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := *q.upserted.ConsecutiveFailures; got != 2 {
		t.Fatalf("failed DLR must leave consecutive failures at 2, got %d", got)
	}

	if err := tracker.RecordRouteAttempt(context.Background(), 8, nil, &msgID, false, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := *q.upserted.ConsecutiveFailures; got != 3 {
		t.Fatalf("failed submission must raise consecutive failures to 3, got %d", got)
	}

	if err := tracker.RecordDeliveryOutcome(context.Background(), 8, &msgID, true, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := *q.upserted.ConsecutiveFailures; got != 0 {
		t.Fatalf("delivered DLR must clear consecutive failures, got %d", got)
	}
}

// routeHealthCountingQuerier keeps one route health row and applies each upsert to it.
type routeHealthCountingQuerier struct {
	routeHealthMockQuerier
	current  database.RouteHealth
	upserted database.UpsertRouteHealthParams
}

func (m *routeHealthCountingQuerier) GetRouteHealth(ctx context.Context, arg database.GetRouteHealthParams) (database.RouteHealth, error) {
	return m.current, nil
}

func (m *routeHealthCountingQuerier) UpsertRouteHealth(ctx context.Context, arg database.UpsertRouteHealthParams) (database.RouteHealth, error) {
	m.upserted = arg
	m.current.ConsecutiveFailures = arg.ConsecutiveFailures
	m.current.ConsecutiveSuccesses = arg.ConsecutiveSuccesses
	m.current.IsAvailable = arg.IsAvailable
	m.current.AutoUnavailable = arg.AutoUnavailable
	return m.current, nil
}

func int32Ptr(v int32) *int32 { return &v }
func boolPtr(v bool) *bool    { return &v }

func TestRouteHealthDecision(t *testing.T) {
	now := time.Now()
	at := func(ago time.Duration) pgtype.Timestamptz {
		return pgtype.Timestamptz{Time: now.Add(-ago), Valid: true}
	}
	row := func(failures int32, available, auto bool, lastAttempt pgtype.Timestamptz) database.GetAllApplicableRoutesForRoutingRow {
		return database.GetAllApplicableRoutesForRoutingRow{ConsecutiveFailures: failures, RouteIsAvailable: available, RouteAutoUnavailable: auto, RouteLastAttemptAt: lastAttempt}
	}
	cases := []struct {
		name string
		row  database.GetAllApplicableRoutesForRoutingRow
		want routeHealthState
	}{
		{"healthy route", row(0, true, false, at(time.Second)), routeUsable},
		{"a few failures", row(4, true, false, at(time.Second)), routeUsable},
		{"too many failures, just tried", row(5, true, false, at(10*time.Second)), routeExcluded},
		{"too many failures, retry due", row(7, true, false, at(2*time.Minute)), routeRetrying},
		{"auto unavailable, recovering after a success", row(0, false, true, at(time.Second)), routeRetrying},
		{"auto unavailable, still failing, just tried", row(6, false, true, at(5*time.Second)), routeExcluded},
		{"manually marked unavailable", row(0, false, false, at(time.Hour)), routeExcluded},
	}
	for _, tc := range cases {
		if got := routeHealthDecision(tc.row, 5, now); got != tc.want {
			t.Fatalf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
