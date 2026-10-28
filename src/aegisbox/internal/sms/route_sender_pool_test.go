package sms

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/thrillee/aegisbox/internal/database"
)

// routeSenderMockQuerier stubs only the queries the route sender pool uses; any other call
// panics through the nil embedded interface.
type routeSenderMockQuerier struct {
	database.Querier

	usesPool      bool
	picked        *database.PickRouteSenderRow
	pickedConnID  int32
	sentSource    *string
	validateCalls int
}

func (m *routeSenderMockQuerier) UsesRouteSenderPool(ctx context.Context, id int32) (bool, error) {
	return m.usesPool, nil
}

func (m *routeSenderMockQuerier) PickRouteSender(ctx context.Context, connID int32) (database.PickRouteSenderRow, error) {
	m.pickedConnID = connID
	if m.picked == nil {
		return database.PickRouteSenderRow{}, pgx.ErrNoRows
	}
	return *m.picked, nil
}

func (m *routeSenderMockQuerier) SetMessageSentSourceAddr(ctx context.Context, arg database.SetMessageSentSourceAddrParams) error {
	m.sentSource = arg.SentSourceAddr
	return nil
}

func (m *routeSenderMockQuerier) ValidateSenderID(ctx context.Context, arg database.ValidateSenderIDParams) (int32, error) {
	m.validateCalls++
	return 0, pgx.ErrNoRows
}

func TestValidate_RouteSenderPoolSkipsSenderApproval(t *testing.T) {
	q := &routeSenderMockQuerier{usesPool: true}
	res, err := NewDefaultValidator(q).Validate(context.Background(), 6, "ANYTHING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsValid || res.ApprovedSenderID != 0 {
		t.Fatalf("expected valid with no approved sender record, got %+v", res)
	}
	if q.validateCalls != 0 {
		t.Fatalf("approved sender lookup should be skipped, called %d times", q.validateCalls)
	}
}

func TestValidate_WithoutPoolStillRequiresApprovedSender(t *testing.T) {
	q := &routeSenderMockQuerier{usesPool: false}
	res, err := NewDefaultValidator(q).Validate(context.Background(), 6, "UNKNOWN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsValid || q.validateCalls != 1 {
		t.Fatalf("expected rejection after one approved sender lookup, got valid=%v calls=%d", res.IsValid, q.validateCalls)
	}
}

func TestResolveSenderID_WithoutPoolKeepsOriginalSender(t *testing.T) {
	q := &routeSenderMockQuerier{usesPool: false}
	s := &DefaultSender{dbQueries: q}
	got, err := s.resolveSenderID(context.Background(), database.GetMessageDetailsForSendingRow{ID: 1, OriginalSourceAddr: "CUSTOMER"}, 9)
	if err != nil || got != "CUSTOMER" {
		t.Fatalf("expected original sender, got %q err=%v", got, err)
	}
	if q.sentSource != nil {
		t.Fatalf("sent sender should not be recorded when it is unchanged")
	}
}

func TestResolveSenderID_PoolUsesConnectionSender(t *testing.T) {
	q := &routeSenderMockQuerier{usesPool: true, picked: &database.PickRouteSenderRow{ID: 3, SenderIDString: "SONDER"}}
	s := &DefaultSender{dbQueries: q}
	got, err := s.resolveSenderID(context.Background(), database.GetMessageDetailsForSendingRow{ID: 1, OriginalSourceAddr: "CUSTOMER"}, 9)
	if err != nil || got != "SONDER" {
		t.Fatalf("expected pool sender SONDER, got %q err=%v", got, err)
	}
	if q.pickedConnID != 9 {
		t.Fatalf("expected pick from connection 9, got %d", q.pickedConnID)
	}
	if q.sentSource == nil || *q.sentSource != "SONDER" {
		t.Fatalf("expected sent sender SONDER to be recorded, got %v", q.sentSource)
	}
}

func TestResolveSenderID_EmptyPoolFailsTheAttempt(t *testing.T) {
	q := &routeSenderMockQuerier{usesPool: true}
	s := &DefaultSender{dbQueries: q}
	if _, err := s.resolveSenderID(context.Background(), database.GetMessageDetailsForSendingRow{ID: 1, OriginalSourceAddr: "CUSTOMER"}, 9); err == nil {
		t.Fatalf("expected an error when the connection has no active pool sender")
	}
}

func mnoIDs(cs []RouteCandidate) []int32 {
	ids := make([]int32, len(cs))
	for i, c := range cs {
		ids[i] = c.MNOID
	}
	return ids
}

func TestOrderAttempts(t *testing.T) {
	cands := []RouteCandidate{{MNOID: 12}, {MNOID: 7}, {MNOID: 3}}
	cases := []struct {
		name           string
		routed, picked int32
		want           []int32
	}{
		{"routed route still usable goes first", 7, 12, []int32{7, 12, 3}},
		{"routed route excluded, fresh pick goes first", 99, 3, []int32{3, 12, 7}},
		{"already first keeps priority order", 12, 7, []int32{12, 7, 3}},
		{"neither found keeps priority order", 99, 98, []int32{12, 7, 3}},
	}
	for _, tc := range cases {
		got := mnoIDs(orderAttempts(cands, tc.routed, tc.picked))
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Fatalf("%s: got %v, want %v", tc.name, got, tc.want)
			}
		}
	}
	if got := mnoIDs(cands); got[0] != 12 || got[1] != 7 || got[2] != 3 {
		t.Fatalf("input slice was modified: %v", got)
	}
}
