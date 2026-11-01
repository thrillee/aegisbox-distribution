package sms

import (
	"context"
	"errors"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno"
)

type guardQuerier struct {
	database.Querier
	blocked bool
	err     error
}

func (q *guardQuerier) IsSimulatorBlockedForSP(context.Context, database.IsSimulatorBlockedForSPParams) (bool, error) {
	return q.blocked, q.err
}

// countingProvider records whether a connector was asked for, i.e. whether sending went ahead.
type countingProvider struct{ calls int }

func (p *countingProvider) GetConnector(context.Context, int32) (mno.Connector, error) {
	p.calls++
	return nil, errors.New("not used")
}

// A non-test customer's message must never reach a simulator MNO, and a failed check must refuse
// too (fail closed).
func TestSendRefusesSimulatorForNonTestCustomer(t *testing.T) {
	mnoID := int32(5)
	for name, q := range map[string]*guardQuerier{
		"blocked":      {blocked: true},
		"check failed": {err: errors.New("db down")},
	} {
		provider := &countingProvider{}
		s := &DefaultSender{dbQueries: q, mnoConnectorProvider: provider}
		res, err := s.Send(context.Background(), database.GetMessageDetailsForSendingRow{ID: 1, ServiceProviderID: 2, RoutedMnoID: &mnoID})
		if err != nil {
			t.Fatalf("%s: unexpected error %v", name, err)
		}
		if res == nil || res.WasSubmitted || !errors.Is(res.Error, ErrSimulatorRouteBlocked) {
			t.Fatalf("%s: expected an ErrSimulatorRouteBlocked result, got %+v", name, res)
		}
		if provider.calls != 0 {
			t.Fatalf("%s: a connector was requested; the message would have been sent", name)
		}
	}
}
