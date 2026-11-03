package linkscan

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/idna"

	"github.com/thrillee/aegisbox/internal/database"
)

// List types of link_domain_lists.
const (
	ListAllow = "allow"
	ListBlock = "block"
)

// ListMatch is the admin list entry that decided a link.
type ListMatch struct {
	ListType string
	Domain   string
	Global   bool // applies to every customer
	Host     string
}

type listEntry struct {
	domain   string
	listType string
	spID     *int32
}

// DomainLists keeps the admin allow and block lists in memory, refreshed every refresh interval,
// so checking a link costs no query.
type DomainLists struct {
	load    func(context.Context) ([]database.ListLinkDomainListsRow, error)
	refresh time.Duration
	now     func() time.Time

	mu       sync.Mutex
	entries  []listEntry
	loadedAt time.Time
}

func NewDomainLists(load func(context.Context) ([]database.ListLinkDomainListsRow, error), refresh time.Duration) *DomainLists {
	if refresh <= 0 {
		refresh = 30 * time.Second
	}
	return &DomainLists{load: load, refresh: refresh, now: time.Now}
}

func (d *DomainLists) current(ctx context.Context) []listEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.loadedAt.IsZero() || d.now().Sub(d.loadedAt) >= d.refresh {
		rows, err := d.load(ctx)
		if err != nil {
			// Keep the last good lists; try again on the next check.
			slog.ErrorContext(ctx, "Link scan: could not load allow and block lists", slog.Any("error", err))
		} else {
			entries := make([]listEntry, 0, len(rows))
			for _, r := range rows {
				entries = append(entries, listEntry{domain: NormalizeListDomain(r.Domain), listType: r.ListType, spID: r.ServiceProviderID})
			}
			d.entries = entries
			d.loadedAt = d.now()
		}
	}
	return d.entries
}

// Match returns the entry that applies to any of hosts for the customer: the customer's own
// entries win over global ones, and within the same scope block wins over allow.
func (d *DomainLists) Match(ctx context.Context, spID int32, hosts ...string) (ListMatch, bool) {
	if d == nil {
		return ListMatch{}, false
	}
	var best ListMatch
	bestRank := -1
	for _, e := range d.current(ctx) {
		if e.spID != nil && *e.spID != spID {
			continue
		}
		for _, h := range hosts {
			if !domainMatches(h, e.domain) {
				continue
			}
			rank := 0
			if e.spID != nil {
				rank += 2
			}
			if e.listType == ListBlock {
				rank++
			}
			if rank > bestRank {
				bestRank = rank
				best = ListMatch{ListType: e.listType, Domain: e.domain, Global: e.spID == nil, Host: h}
			}
		}
	}
	return best, bestRank >= 0
}

// domainMatches reports whether host is domain or one of its subdomains.
func domainMatches(host, domain string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" || domain == "" {
		return false
	}
	return host == domain || strings.HasSuffix(host, "."+domain)
}

// NormalizeListDomain turns what an admin typed (a URL, a host, a domain with a leading dot or
// "*.") into the stored form: lowercase punycode host without port.
func NormalizeListDomain(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, "@"); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.LastIndex(s, ":"); i >= 0 && !strings.Contains(s, "]") {
		s = s[:i]
	}
	s = strings.TrimPrefix(s, "*.")
	s = strings.Trim(s, ".")
	if a, err := idna.Lookup.ToASCII(s); err == nil {
		s = a
	}
	return s
}
