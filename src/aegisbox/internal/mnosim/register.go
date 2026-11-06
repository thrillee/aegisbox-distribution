package mnosim

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MNOName is the name the embedded simulator is registered under in the mnos table.
const MNOName = "mnosim"

// Registration says how to record the embedded simulator in the database, so nobody has to add
// the MNO, its SMPP connection or its routes by hand.
type Registration struct {
	// RoutingGroupRefs are routing_groups.reference values that should send to the simulator.
	// Each gets an active simulator route for every prefix group it already routes (or for every
	// prefix group, if it has no routes yet).
	RoutingGroupRefs []string
	// Price is charged per SMS to test customers on those groups when they have no price for the
	// simulator yet. The charging code refuses a zero price.
	Price string
}

// ParseRoutingGroupRefs splits a comma separated MNOSIM_ROUTING_GROUPS value.
func ParseRoutingGroupRefs(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// connectHostPort turns the listen address into the address the gateway dials. An empty or
// wildcard host becomes 127.0.0.1, because the simulator runs in the gateway's own process.
func connectHostPort(addr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("simulator address %q: %w", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("simulator port %q: %w", portStr, err)
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return host, port, nil
}

// Register creates or updates the simulator MNO, its SMPP connection and the requested routes and
// prices in one transaction. Running it again changes nothing unless the configuration changed.
func Register(ctx context.Context, pool *pgxpool.Pool, cfg Config, reg Registration) error {
	host, port, err := connectHostPort(cfg.Addr)
	if err != nil {
		return err
	}
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var mnoID int32
		if err := tx.QueryRow(ctx, `
			INSERT INTO mnos (name, country_code, status, is_simulator)
			VALUES ($1, 'NG', 'active', TRUE)
			ON CONFLICT (name) DO UPDATE SET status = 'active', is_simulator = TRUE, updated_at = NOW()
			RETURNING id`, MNOName).Scan(&mnoID); err != nil {
			return fmt.Errorf("register simulator MNO: %w", err)
		}

		tag, err := tx.Exec(ctx, `
			UPDATE mno_connections
			SET status = 'active', host = $2, port = $3, system_id = $4, password = $5,
			    bind_type = 'trx', use_tls = FALSE, updated_at = NOW()
			WHERE id = (SELECT id FROM mno_connections WHERE mno_id = $1 AND protocol = 'smpp' ORDER BY id LIMIT 1)`,
			mnoID, host, port, cfg.SystemID, cfg.Password)
		if err != nil {
			return fmt.Errorf("update simulator connection: %w", err)
		}
		if tag.RowsAffected() == 0 {
			if _, err := tx.Exec(ctx, `
				INSERT INTO mno_connections (mno_id, protocol, status, priority, system_id, password, host, port, use_tls, bind_type)
				VALUES ($1, 'smpp', 'active', 1, $2, $3, $4, $5, FALSE, 'trx')`,
				mnoID, cfg.SystemID, cfg.Password, host, port); err != nil {
				return fmt.Errorf("create simulator connection: %w", err)
			}
		}

		for _, ref := range reg.RoutingGroupRefs {
			var groupID int32
			err := tx.QueryRow(ctx, `SELECT id FROM routing_groups WHERE reference = $1`, ref).Scan(&groupID)
			if err == pgx.ErrNoRows {
				slog.WarnContext(ctx, "MNOSIM_ROUTING_GROUPS names a routing group that does not exist; skipped", slog.String("reference", ref))
				continue
			}
			if err != nil {
				return fmt.Errorf("find routing group %q: %w", ref, err)
			}

			routes, err := tx.Exec(ctx, `
				WITH prefix_groups AS (
				    SELECT DISTINCT msisdn_prefix_group_id AS id FROM routing_assignments WHERE routing_group_id = $1
				    UNION
				    SELECT id FROM msisdn_prefix_groups
				    WHERE NOT EXISTS (SELECT 1 FROM routing_assignments WHERE routing_group_id = $1)
				)
				INSERT INTO routing_assignments (routing_group_id, msisdn_prefix_group_id, mno_id, status, priority, is_active, weight, comment)
				SELECT $1, pg.id, $2, 'active', 1, TRUE, 100, 'Simulator route added by MNOSIM_ROUTING_GROUPS'
				FROM prefix_groups pg
				WHERE NOT EXISTS (
				    SELECT 1 FROM routing_assignments ra
				    WHERE ra.routing_group_id = $1 AND ra.msisdn_prefix_group_id = pg.id AND ra.mno_id = $2
				)`, groupID, mnoID)
			if err != nil {
				return fmt.Errorf("add simulator routes to %q: %w", ref, err)
			}

			// Without a price for the simulator, charging fails before the message is sent.
			prices, err := tx.Exec(ctx, `
				INSERT INTO pricing_rules (service_provider_id, mno_id, currency_code, price_per_sms)
				SELECT DISTINCT sp.id, $2::INT, sp.default_currency_code, $3::NUMERIC
				FROM sp_credentials c
				JOIN service_providers sp ON sp.id = c.service_provider_id
				WHERE c.routing_group_id = $1 AND sp.is_test
				ON CONFLICT (service_provider_id, mno_id, currency_code) WHERE mno_id IS NOT NULL DO NOTHING`,
				groupID, mnoID, reg.Price)
			if err != nil {
				return fmt.Errorf("add simulator prices for %q: %w", ref, err)
			}
			slog.InfoContext(ctx, "Simulator routes checked",
				slog.String("routing_group", ref), slog.Int64("routes_added", routes.RowsAffected()),
				slog.Int64("prices_added", prices.RowsAffected()))
		}

		slog.InfoContext(ctx, "Simulated MNO registered",
			slog.Int("mno_id", int(mnoID)), slog.String("host", host), slog.Int("port", port),
			slog.String("system_id", cfg.SystemID))
		return nil
	})
}
