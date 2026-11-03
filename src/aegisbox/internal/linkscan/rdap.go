package linkscan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
)

// Domain age statuses in link_domain_ages.
const (
	AgeKnown   = "known"
	AgeUnknown = "unknown" // no RDAP for the TLD (.ng has none) or no registration date: not scored
	AgeError   = "error"   // lookup failed; retried after ageErrorTTL
)

const (
	ageKnownTTL   = 30 * 24 * time.Hour
	ageUnknownTTL = 7 * 24 * time.Hour
	ageErrorTTL   = time.Hour
	ianaRDAPDNS   = "https://data.iana.org/rdap/dns.json"
)

// AgeStore is the part of database.Querier the domain age lookup uses.
type AgeStore interface {
	GetLinkDomainAge(ctx context.Context, domain string) (database.LinkDomainAge, error)
	UpsertLinkDomainAge(ctx context.Context, arg database.UpsertLinkDomainAgeParams) error
}

// DomainAger finds when a registrable domain was registered, through RDAP (the successor of
// WHOIS), with results kept in link_domain_ages. It asks the registry named by IANA's bootstrap
// file, never the domain itself.
type DomainAger struct {
	store     AgeStore
	client    *http.Client
	bootstrap string
	now       func() time.Time

	mu       sync.Mutex
	servers  map[string]string // TLD -> RDAP base URL
	loadedAt time.Time
}

func NewDomainAger(store AgeStore, client *http.Client) *DomainAger {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &DomainAger{store: store, client: client, bootstrap: ianaRDAPDNS, now: time.Now}
}

// AgeDays returns the registrable domain of host and its age in days, or nil when unknown. It
// never fails the scan: lookup problems only mean no age signal.
func (a *DomainAger) AgeDays(ctx context.Context, host string) (string, *int) {
	if a == nil || host == "" || net.ParseIP(host) != nil {
		return "", nil
	}
	domain := registrable(host)
	if !strings.Contains(domain, ".") {
		return domain, nil
	}
	now := a.now()
	row, err := a.store.GetLinkDomainAge(ctx, domain)
	if err == nil && fresh(row, now) {
		return domain, daysSince(row, now)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(ctx, "Domain age: cache lookup failed", slog.Any("error", err))
	}

	registered, status := a.lookup(ctx, domain)
	params := database.UpsertLinkDomainAgeParams{Domain: domain, Status: status}
	if registered != nil {
		params.RegisteredAt = pgtype.Timestamptz{Time: *registered, Valid: true}
	}
	if err := a.store.UpsertLinkDomainAge(ctx, params); err != nil {
		slog.WarnContext(ctx, "Domain age: could not store result", slog.Any("error", err))
	}
	if registered == nil {
		return domain, nil
	}
	d := int(now.Sub(*registered).Hours() / 24)
	if d < 0 {
		d = 0
	}
	return domain, &d
}

func fresh(row database.LinkDomainAge, now time.Time) bool {
	ttl := ageErrorTTL
	switch row.Status {
	case AgeKnown:
		ttl = ageKnownTTL
	case AgeUnknown:
		ttl = ageUnknownTTL
	}
	return row.CheckedAt.Valid && now.Sub(row.CheckedAt.Time) < ttl
}

func daysSince(row database.LinkDomainAge, now time.Time) *int {
	if row.Status != AgeKnown || !row.RegisteredAt.Valid {
		return nil
	}
	d := int(now.Sub(row.RegisteredAt.Time).Hours() / 24)
	if d < 0 {
		d = 0
	}
	return &d
}

func (a *DomainAger) lookup(ctx context.Context, domain string) (*time.Time, string) {
	base, err := a.server(ctx, domain[strings.LastIndex(domain, ".")+1:])
	if err != nil {
		slog.WarnContext(ctx, "Domain age: RDAP bootstrap failed", slog.Any("error", err))
		return nil, AgeError
	}
	if base == "" {
		return nil, AgeUnknown
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/domain/"+url.PathEscape(domain), nil)
	if err != nil {
		return nil, AgeError
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, AgeError
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, AgeUnknown // not registered (or the registry does not say)
	case resp.StatusCode != http.StatusOK:
		return nil, AgeError
	}
	var body struct {
		Events []struct {
			Action string `json:"eventAction"`
			Date   string `json:"eventDate"`
		} `json:"events"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return nil, AgeError
	}
	for _, e := range body.Events {
		if e.Action != "registration" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, e.Date); err == nil {
			return &t, AgeKnown
		}
	}
	return nil, AgeUnknown
}

// server returns the RDAP base URL for a TLD ("" when it has none), from IANA's bootstrap file,
// reloaded daily.
func (a *DomainAger) server(ctx context.Context, tld string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.servers == nil || a.now().Sub(a.loadedAt) > 24*time.Hour {
		servers, err := a.loadBootstrap(ctx)
		if err != nil {
			if a.servers == nil {
				return "", err
			}
		} else {
			a.servers, a.loadedAt = servers, a.now()
		}
	}
	return a.servers[strings.ToLower(tld)], nil
}

func (a *DomainAger) loadBootstrap(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.bootstrap, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RDAP bootstrap answered %d", resp.StatusCode)
	}
	var file struct {
		Services [][][]string `json:"services"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&file); err != nil {
		return nil, err
	}
	servers := map[string]string{}
	for _, svc := range file.Services {
		if len(svc) < 2 || len(svc[1]) == 0 {
			continue
		}
		base := svc[1][0]
		for _, u := range svc[1] {
			if strings.HasPrefix(u, "https://") {
				base = u
				break
			}
		}
		for _, tld := range svc[0] {
			servers[strings.ToLower(tld)] = base
		}
	}
	return servers, nil
}
