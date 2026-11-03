package sms

import (
	"context"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
)

type policyQuerier struct {
	database.Querier
	isTest      bool
	whitelisted map[string]bool
	gotMSISDN   string
}

func (q *policyQuerier) GetSimulatorRoutingPolicy(_ context.Context, a database.GetSimulatorRoutingPolicyParams) (database.GetSimulatorRoutingPolicyRow, error) {
	q.gotMSISDN = a.Msisdn
	return database.GetSimulatorRoutingPolicyRow{IsTest: q.isTest, IsWhitelisted: q.whitelisted[a.Msisdn]}, nil
}

func rowMnoIDs(rows []database.GetAllApplicableRoutesForRoutingRow) []int32 {
	out := []int32{}
	for _, r := range rows {
		out = append(out, r.MnoID)
	}
	return out
}

func TestApplySimulatorPolicy(t *testing.T) {
	sim := database.GetAllApplicableRoutesForRoutingRow{MnoID: 1, MnoIsSimulator: true}
	real1 := database.GetAllApplicableRoutesForRoutingRow{MnoID: 2}
	real2 := database.GetAllApplicableRoutesForRoutingRow{MnoID: 3}
	mixed := []database.GetAllApplicableRoutesForRoutingRow{sim, real1, real2}
	ctx := context.Background()

	cases := []struct {
		name       string
		isTest     bool
		msisdn     string
		candidates []database.GetAllApplicableRoutesForRoutingRow
		want       []int32
	}{
		{"test customer, not whitelisted: simulator only", true, "2348031111111", mixed, []int32{1}},
		{"test customer, whitelisted: real routes only", true, "+234 803 222 2222", mixed, []int32{2, 3}},
		{"test customer, whitelisted local format", true, "08032222222", mixed, []int32{2, 3}},
		{"real customer: simulator left out", false, "2348031111111", mixed, []int32{2, 3}},
		{"real customer on whitelisted number: still real", false, "2348032222222", mixed, []int32{2, 3}},
		{"test customer, group without simulator: unchanged", true, "2348031111111", []database.GetAllApplicableRoutesForRoutingRow{real1, real2}, []int32{2, 3}},
		{"test customer, whitelisted, simulator only group: no route", true, "2348032222222", []database.GetAllApplicableRoutesForRoutingRow{sim}, []int32{}},
		{"real customer, simulator only group: no route", false, "2348031111111", []database.GetAllApplicableRoutesForRoutingRow{sim}, []int32{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q := &policyQuerier{isTest: c.isTest, whitelisted: map[string]bool{"2348032222222": true}}
			r := &IntelligentRouter{BaseRouter: &BaseRouter{dbQueries: q}}
			got, err := r.applySimulatorPolicy(ctx, RoutingRequest{SPCredentialID: 7, MSISDN: c.msisdn}, c.candidates)
			if err != nil {
				t.Fatal(err)
			}
			g := rowMnoIDs(got)
			if len(g) != len(c.want) {
				t.Fatalf("got %v, want %v", g, c.want)
			}
			for i := range g {
				if g[i] != c.want[i] {
					t.Fatalf("got %v, want %v", g, c.want)
				}
			}
		})
	}
}

func TestNormalizeMSISDN(t *testing.T) {
	for in, want := range map[string]string{
		"+2348031234567":   "2348031234567",
		"08031234567":      "2348031234567",
		"234 803 123 4567": "2348031234567",
		"(0803) 123-4567":  "2348031234567",
		"447700900123":     "447700900123",
		"":                 "",
	} {
		if got := NormalizeMSISDN(in); got != want {
			t.Errorf("NormalizeMSISDN(%q) = %q, want %q", in, got, want)
		}
	}
}
