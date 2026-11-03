package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
)

// NormalizeMSISDN returns a number as digits in international format, as stored in the
// simulator whitelist: "+234 803 123 4567" and "08031234567" both become "2348031234567".
// Local Nigerian numbers (11 digits starting with 0) get the 234 country code.
func NormalizeMSISDN(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) == 11 && d[0] == '0' {
		d = "234" + d[1:]
	}
	return d
}

// applySimulatorPolicy decides which of a routing group's routes may carry a message, so test
// traffic stays on the simulator and real traffic stays off it:
//
//   - A customer that is not a test customer never uses a simulator route (the sender would refuse
//     it anyway; dropping it here lets the message take a real route instead of failing).
//   - A test customer's message to a whitelisted number uses only real routes, so the phone gets
//     the message and the real DLR comes back.
//   - Any other test customer message uses only simulator routes when the group has one, so it can
//     never fail over to a real network and reach a real phone.
//   - A test customer whose group has no simulator route is routed normally.
func (r *IntelligentRouter) applySimulatorPolicy(
	ctx context.Context,
	req RoutingRequest,
	candidates []database.GetAllApplicableRoutesForRoutingRow,
) ([]database.GetAllApplicableRoutesForRoutingRow, error) {
	hasSim := false
	for _, c := range candidates {
		if c.MnoIsSimulator {
			hasSim = true
			break
		}
	}
	if !hasSim {
		return candidates, nil
	}

	msisdn := NormalizeMSISDN(req.MSISDN)
	policy, err := r.dbQueries.GetSimulatorRoutingPolicy(ctx, database.GetSimulatorRoutingPolicyParams{
		Msisdn:         msisdn,
		SpCredentialID: req.SPCredentialID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("service provider credential %d not found", req.SPCredentialID)
		}
		return nil, fmt.Errorf("check simulator routing policy: %w", err)
	}

	wantSim := policy.IsTest && !policy.IsWhitelisted
	kept := make([]database.GetAllApplicableRoutesForRoutingRow, 0, len(candidates))
	for _, c := range candidates {
		if c.MnoIsSimulator == wantSim {
			kept = append(kept, c)
		}
	}

	switch {
	case policy.IsTest && policy.IsWhitelisted:
		slog.InfoContext(ctx, "Whitelisted number: sending through a real network, not the simulator",
			slog.String("msisdn", msisdn), slog.Int("real_routes", len(kept)))
		if len(kept) == 0 {
			return nil, nil // reported by the caller as no route
		}
	case policy.IsTest:
		slog.DebugContext(ctx, "Test customer: simulator routes only", slog.Int("simulator_routes", len(kept)))
	default:
		slog.DebugContext(ctx, "Customer is not a test customer: simulator routes left out", slog.Int("routes", len(kept)))
	}
	return kept, nil
}

// noRouteAfterSimulatorPolicy explains an empty candidate list after applySimulatorPolicy.
func noRouteAfterSimulatorPolicy(routingGroupID, prefixGroupID int32) error {
	return fmt.Errorf("no usable route for prefix group %d in routing group %d: simulator routes only "+
		"carry test customers' messages to numbers that are not whitelisted; add a real route to the "+
		"group for whitelisted numbers and other customers", prefixGroupID, routingGroupID)
}
