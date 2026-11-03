package linkscan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
)

func i32(v int32) *int32 { return &v }

// --- admin lists ---

func TestDomainListsMatchAndPrecedence(t *testing.T) {
	rows := []database.ListLinkDomainListsRow{
		{ID: 1, Domain: "Example.com", ListType: ListAllow},
		{ID: 2, Domain: "bad.example.com", ListType: ListBlock},
		{ID: 3, Domain: "promo.ng", ListType: ListBlock},
		{ID: 4, Domain: "promo.ng", ListType: ListAllow, ServiceProviderID: i32(7)},
	}
	d := NewDomainLists(func(context.Context) ([]database.ListLinkDomainListsRow, error) { return rows, nil }, time.Minute)
	ctx := context.Background()

	if m, ok := d.Match(ctx, 1, "www.example.com"); !ok || m.ListType != ListAllow || !m.Global {
		t.Fatalf("subdomain of an allowed domain: %+v %v", m, ok)
	}
	if m, ok := d.Match(ctx, 1, "x.bad.example.com"); !ok || m.ListType != ListBlock {
		t.Fatalf("block of a subdomain should win over allow of the parent: %+v", m)
	}
	if _, ok := d.Match(ctx, 1, "notexample.com"); ok {
		t.Fatal("a different domain ending in the same letters must not match")
	}
	if m, _ := d.Match(ctx, 1, "promo.ng"); m.ListType != ListBlock {
		t.Fatalf("global block for other customers: %+v", m)
	}
	if m, _ := d.Match(ctx, 7, "promo.ng"); m.ListType != ListAllow || m.Global {
		t.Fatalf("the customer's own allow wins over a global block: %+v", m)
	}
	if m, ok := d.Match(ctx, 1, "shop.com", "bad.example.com"); !ok || m.Host != "bad.example.com" {
		t.Fatalf("final host should be checked too: %+v", m)
	}
	var nilLists *DomainLists
	if _, ok := nilLists.Match(ctx, 1, "example.com"); ok {
		t.Fatal("nil lists match nothing")
	}
}

func TestDomainListsKeepLastGoodOnError(t *testing.T) {
	calls := 0
	d := NewDomainLists(func(context.Context) ([]database.ListLinkDomainListsRow, error) {
		calls++
		if calls == 1 {
			return []database.ListLinkDomainListsRow{{Domain: "x.com", ListType: ListBlock}}, nil
		}
		return nil, fmt.Errorf("db down")
	}, time.Millisecond)
	now := time.Now()
	d.now = func() time.Time { return now }
	if _, ok := d.Match(context.Background(), 1, "x.com"); !ok {
		t.Fatal("first load")
	}
	now = now.Add(time.Second)
	if _, ok := d.Match(context.Background(), 1, "x.com"); !ok {
		t.Fatal("lists must survive a failed refresh")
	}
}

func TestNormalizeListDomain(t *testing.T) {
	for in, want := range map[string]string{
		"https://Login.Example.com/path?q=1": "login.example.com",
		"*.example.com":                      "example.com",
		".example.com.":                      "example.com",
		"example.com:8443":                   "example.com",
		"bücher.de":                          "xn--bcher-kva.de",
	} {
		if got := NormalizeListDomain(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestWorkerListDecidesWithoutVisit(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{res: &ScanResult{FinalURL: sp("https://x.com/")}}
	w := newTestWorker(t, q, s)
	w.SetChecks(NewDomainLists(func(context.Context) ([]database.ListLinkDomainListsRow, error) {
		return []database.ListLinkDomainListsRow{{Domain: "evil-promo.com", ListType: ListBlock}}, nil
	}, time.Minute), nil, nil)
	q.jobs = []database.LinkScanJob{job(1, "https://win.evil-promo.com/claim", ModeVisit, nil)}
	w.RunOnce(context.Background())
	if len(s.calls) != 0 {
		t.Fatalf("a blocked domain must not be opened: %v", s.calls)
	}
	if len(q.scans) != 1 || q.scans[0].ScanKind != "list" || q.scans[0].Verdict != VerdictMalicious {
		t.Fatalf("scans %+v", q.scans)
	}
	if q.completed[1] == 0 {
		t.Fatal("job not completed")
	}
}

func TestWorkerListOnFinalHost(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{res: &ScanResult{FinalURL: sp("https://phish.badsite.com/login"), Status: ip(200), ShortenerExpanded: true}}
	w := newTestWorker(t, q, s)
	w.SetChecks(NewDomainLists(func(context.Context) ([]database.ListLinkDomainListsRow, error) {
		return []database.ListLinkDomainListsRow{{Domain: "badsite.com", ListType: ListBlock}}, nil
	}, time.Minute), nil, nil)
	q.jobs = []database.LinkScanJob{job(1, "https://bit.ly/abc", ModeVisit, nil)}
	w.RunOnce(context.Background())
	if len(q.scans) != 2 || q.scans[1].ScanKind != "list" || q.scans[1].Verdict != VerdictMalicious {
		t.Fatalf("want the visit and then a list result: %+v", q.scans)
	}
	if !strings.Contains(string(q.scans[1].Reasons), "led there") {
		t.Fatalf("reason should say the link led there: %s", q.scans[1].Reasons)
	}
}

// --- domain age ---

type fakeAgeStore struct {
	rows map[string]database.LinkDomainAge
}

func (f *fakeAgeStore) GetLinkDomainAge(_ context.Context, d string) (database.LinkDomainAge, error) {
	if r, ok := f.rows[d]; ok {
		return r, nil
	}
	return database.LinkDomainAge{}, pgx.ErrNoRows
}
func (f *fakeAgeStore) UpsertLinkDomainAge(_ context.Context, a database.UpsertLinkDomainAgeParams) error {
	f.rows[a.Domain] = database.LinkDomainAge{Domain: a.Domain, RegisteredAt: a.RegisteredAt, Status: a.Status,
		CheckedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}
	return nil
}

func TestDomainAgerRDAP(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	hits := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/bootstrap":
			fmt.Fprintf(w, `{"services":[[["com","net"],["%s/rdap/"]]]}`, srv.URL)
		case r.URL.Path == "/rdap/domain/fresh-promo.com":
			hits++
			fmt.Fprint(w, `{"events":[{"eventAction":"last changed","eventDate":"2026-11-01T00:00:00Z"},{"eventAction":"registration","eventDate":"2026-10-24T12:00:00Z"}]}`)
		case r.URL.Path == "/rdap/domain/missing.com":
			http.NotFound(w, r)
		default:
			http.Error(w, "unexpected "+r.URL.Path, 500)
		}
	}))
	defer srv.Close()

	store := &fakeAgeStore{rows: map[string]database.LinkDomainAge{}}
	a := NewDomainAger(store, srv.Client())
	a.bootstrap = srv.URL + "/bootstrap"
	a.now = func() time.Time { return now }
	ctx := context.Background()

	dom, days := a.AgeDays(ctx, "login.fresh-promo.com")
	if dom != "fresh-promo.com" || days == nil || *days != 10 {
		t.Fatalf("got %q %v", dom, days)
	}
	if _, again := a.AgeDays(ctx, "www.fresh-promo.com"); again == nil || hits != 1 {
		t.Fatalf("second lookup should come from the cache (hits %d)", hits)
	}
	if _, d := a.AgeDays(ctx, "missing.com"); d != nil || store.rows["missing.com"].Status != AgeUnknown {
		t.Fatalf("not found should be unknown: %v %+v", d, store.rows["missing.com"])
	}
	if _, d := a.AgeDays(ctx, "bank.com.ng"); d != nil || store.rows["bank.com.ng"].Status != AgeUnknown {
		t.Fatal(".ng has no RDAP server: unknown, not scored")
	}
	if _, d := a.AgeDays(ctx, "10.0.0.1"); d != nil {
		t.Fatal("IP addresses have no domain age")
	}
	var nilAger *DomainAger
	if _, d := nilAger.AgeDays(ctx, "x.com"); d != nil {
		t.Fatal("nil ager")
	}
}

func TestScoreDomainAgeAndListing(t *testing.T) {
	l := Extract("https://promo-deals.com/x")[0]
	d := 5
	l.DomainAgeDays, l.AgeDomain = &d, "promo-deals.com"
	_, score, reasons := Score(l, nil)
	if score != 20 || reasons[0].Code != "new_domain" {
		t.Fatalf("new domain: %d %+v", score, reasons)
	}
	d = 60
	if _, score, _ := Score(l, nil); score != 10 {
		t.Fatalf("young domain: %d", score)
	}
	d = 400
	l.ListedBy = "URLhaus: malware_download (online)"
	v, score, _ := Score(l, nil)
	if v != VerdictMalicious || score != 70 {
		t.Fatalf("listed URL: %s %d", v, score)
	}
}

func TestURLhaus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth-Key") != "k" {
			http.Error(w, "no key", 401)
			return
		}
		_ = r.ParseForm()
		if strings.Contains(r.Form.Get("url"), "bad") {
			fmt.Fprint(w, `{"query_status":"ok","url_status":"online","threat":"malware_download"}`)
			return
		}
		fmt.Fprint(w, `{"query_status":"no_results"}`)
	}))
	defer srv.Close()
	u := NewURLhaus("k", srv.Client())
	u.endpoint = srv.URL
	if hit, err := u.Lookup(context.Background(), "https://bad.example/x.apk"); err != nil || !strings.Contains(hit, "malware_download") {
		t.Fatalf("listed: %q %v", hit, err)
	}
	if hit, err := u.Lookup(context.Background(), "https://good.example/"); err != nil || hit != "" {
		t.Fatalf("not listed: %q %v", hit, err)
	}
	if NewURLhaus("", nil) != nil {
		t.Fatal("no key: off")
	}
}

// --- Phase 2 gate ---

type fakeGateQ struct {
	text     string
	holds    map[int64]database.MessageLinkHold
	statuses map[int64]string
	verdicts map[int64][]database.ListHoldJobVerdictsRow
	now      time.Time
}

func newFakeGateQ() *fakeGateQ {
	return &fakeGateQ{holds: map[int64]database.MessageLinkHold{}, statuses: map[int64]string{},
		verdicts: map[int64][]database.ListHoldJobVerdictsRow{}, now: time.Now()}
}
func (f *fakeGateQ) GetMessageTextForLinkScan(context.Context, int64) (string, error) {
	return f.text, nil
}
func (f *fakeGateQ) GetMessageLinkHold(_ context.Context, id int64) (database.MessageLinkHold, error) {
	if h, ok := f.holds[id]; ok {
		return h, nil
	}
	return database.MessageLinkHold{}, pgx.ErrNoRows
}
func (f *fakeGateQ) InsertMessageLinkHold(_ context.Context, a database.InsertMessageLinkHoldParams) (int64, error) {
	if _, ok := f.holds[a.MessageID]; ok {
		return 0, nil
	}
	f.holds[a.MessageID] = database.MessageLinkHold{MessageID: a.MessageID, ServiceProviderID: a.ServiceProviderID,
		Mode: a.Mode, Status: HoldWaiting, FailOpen: a.FailOpen, BlockSuspicious: a.BlockSuspicious, ChargeBlocked: a.ChargeBlocked,
		HoldUntil: pgtype.Timestamptz{Time: f.now.Add(time.Duration(a.HoldSeconds) * time.Second), Valid: true}}
	return 1, nil
}
func (f *fakeGateQ) SetMessageProcessingStatus(_ context.Context, a database.SetMessageProcessingStatusParams) (int64, error) {
	if f.statuses[a.ID] != a.FromStatus {
		return 0, nil
	}
	f.statuses[a.ID] = a.ToStatus
	return 1, nil
}
func (f *fakeGateQ) ListOpenLinkHolds(context.Context, int32) ([]database.MessageLinkHold, error) {
	var out []database.MessageLinkHold
	for _, h := range f.holds {
		if h.Status == HoldWaiting || (h.Status == HoldReview && h.CreatedAt.Valid && f.now.Sub(h.CreatedAt.Time) > reviewExpiry) {
			out = append(out, h)
		}
	}
	return out, nil
}
func (f *fakeGateQ) ListHoldJobVerdicts(_ context.Context, id int64) ([]database.ListHoldJobVerdictsRow, error) {
	return f.verdicts[id], nil
}
func (f *fakeGateQ) DecideMessageLinkHold(_ context.Context, a database.DecideMessageLinkHoldParams) (int64, error) {
	h, ok := f.holds[a.MessageID]
	if !ok || h.Status != a.FromStatus {
		return 0, nil
	}
	h.Status, h.Decision, h.WorstVerdict, h.DecidedBy = a.ToStatus, &a.Decision, a.WorstVerdict, a.DecidedBy
	f.holds[a.MessageID] = h
	return 1, nil
}

type fakeRejecter struct {
	rejected map[int64]bool // id -> refund
	q        *fakeGateQ
}

func (r *fakeRejecter) RejectForLinks(_ context.Context, id int64, _ string, refund bool) error {
	r.rejected[id] = refund
	r.q.statuses[id] = StatusBlocked
	return nil
}

func gateFor(t *testing.T, s Settings) (*Gate, *fakeGateQ, *fakeRejecter, *fakeQ) {
	t.Helper()
	gq := newFakeGateQ()
	eq := &fakeQ{}
	settings := NewSettingsCache(func(context.Context) ([]database.ListActiveLinkScanSettingsRow, error) {
		return []database.ListActiveLinkScanSettingsRow{{ServiceProviderID: 5, Mode: s.Mode, HoldSeconds: s.HoldSeconds,
			FailOpen: s.FailOpen, BlockSuspicious: s.BlockSuspicious, ChargeBlocked: s.ChargeBlocked}}, nil
	}, time.Minute)
	enq := NewEnqueuer(eq, defaultCfg)
	enq.SetSettings(settings)
	rj := &fakeRejecter{rejected: map[int64]bool{}, q: gq}
	g := NewGate(gq, settings, enq, rj)
	g.now = func() time.Time { return gq.now }
	return g, gq, rj, eq
}

func vrow(status string, verdict string) database.ListHoldJobVerdictsRow {
	r := database.ListHoldJobVerdictsRow{Status: status}
	if verdict != "" {
		r.Verdict = &verdict
	}
	return r
}

func TestGateOtherCustomersAreNeverHeld(t *testing.T) {
	g, _, _, eq := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 30})
	held, err := g.BeforeSend(context.Background(), 1, 99)
	if err != nil || held || len(eq.jobs) != 0 {
		t.Fatalf("customer without settings: held %v err %v jobs %d", held, err, len(eq.jobs))
	}
	g2, _, _, eq2 := gateFor(t, Settings{Mode: ModeMonitor, HoldSeconds: 30})
	if held, _ := g2.BeforeSend(context.Background(), 1, 5); held || len(eq2.jobs) != 0 {
		t.Fatal("monitor mode never holds")
	}
}

func TestGateNoLinksOrOTPSendsAtOnce(t *testing.T) {
	g, gq, _, _ := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 30})
	gq.text = "Your order has shipped. Thanks!"
	if held, _ := g.BeforeSend(context.Background(), 1, 5); held {
		t.Fatal("no links: send")
	}
	gq.text = "Your OTP is 482913. Do not share. https://bank.example.com/help"
	if held, _ := g.BeforeSend(context.Background(), 2, 5); held {
		t.Fatal("OTP messages are not held")
	}
	gq.text = "Chat with us on https://wa.me/2348031234567"
	if held, _ := g.BeforeSend(context.Background(), 3, 5); held {
		t.Fatal("links without a verdict (WhatsApp) do not hold the message")
	}
}

func TestGateBlockModeRejectsMalicious(t *testing.T) {
	g, gq, rj, eq := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 30, FailOpen: true})
	gq.text = "Your account is locked, verify at https://gtbank-secure.xyz/login"
	gq.statuses[1] = StatusQueuedForSend
	held, err := g.BeforeSend(context.Background(), 1, 5)
	if err != nil || !held || gq.statuses[1] != StatusHeld {
		t.Fatalf("held %v err %v status %s", held, err, gq.statuses[1])
	}
	if len(eq.jobs) != 1 || eq.jobs[0].Priority != 1 {
		t.Fatalf("links must be queued first in line: %+v", eq.jobs)
	}
	// Asked again while waiting: still held, no new hold.
	gq.statuses[1] = StatusQueuedForSend
	if held, _ := g.BeforeSend(context.Background(), 1, 5); !held {
		t.Fatal("a waiting message must not be sent")
	}

	gq.verdicts[1] = []database.ListHoldJobVerdictsRow{vrow("queued", "")}
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldWaiting {
		t.Fatal("pending verdicts within the hold time: keep waiting")
	}
	gq.verdicts[1] = []database.ListHoldJobVerdictsRow{vrow("done", VerdictMalicious)}
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldRejected || *gq.holds[1].Decision != VerdictMalicious {
		t.Fatalf("hold %+v", gq.holds[1])
	}
	if refund, ok := rj.rejected[1]; !ok || !refund {
		t.Fatal("rejected and refunded (charge_blocked off)")
	}
}

func TestGateCleanReleases(t *testing.T) {
	g, gq, rj, _ := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 30})
	gq.text = "Sale today https://shop.example.com/sale and https://news.example.org/a"
	gq.statuses[1] = StatusQueuedForSend
	g.BeforeSend(context.Background(), 1, 5)
	gq.verdicts[1] = []database.ListHoldJobVerdictsRow{vrow("done", VerdictSafe), vrow("done", VerdictSuspicious)}
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldReleased || gq.statuses[1] != StatusQueuedForSend || len(rj.rejected) != 0 {
		t.Fatalf("suspicious is sent unless block_suspicious: %+v %s", gq.holds[1], gq.statuses[1])
	}
	if held, _ := g.BeforeSend(context.Background(), 1, 5); held {
		t.Fatal("a released message is sent")
	}
}

func TestGateBlockSuspiciousAndHoldReview(t *testing.T) {
	g, gq, rj, _ := gateFor(t, Settings{Mode: ModeHold, HoldSeconds: 30, BlockSuspicious: true})
	gq.text = "See https://deals.example.com/x"
	gq.statuses[1] = StatusQueuedForSend
	g.BeforeSend(context.Background(), 1, 5)
	gq.verdicts[1] = []database.ListHoldJobVerdictsRow{vrow("done", VerdictSuspicious)}
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldReview || len(rj.rejected) != 0 {
		t.Fatalf("hold mode sends a bad verdict to review: %+v", gq.holds[1])
	}
	// The admin releases it.
	if err := g.AdminDecide(context.Background(), 1, true, "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	if gq.holds[1].Status != HoldReleased || gq.statuses[1] != StatusQueuedForSend || *gq.holds[1].DecidedBy != "ada@example.com" {
		t.Fatalf("admin release: %+v %s", gq.holds[1], gq.statuses[1])
	}
	if err := g.AdminDecide(context.Background(), 1, false, "x"); err != ErrNotInReview {
		t.Fatalf("a decided message cannot be decided again: %v", err)
	}
}

func TestGateReviewExpires(t *testing.T) {
	g, gq, rj, _ := gateFor(t, Settings{Mode: ModeHold, HoldSeconds: 30})
	gq.holds[1] = database.MessageLinkHold{MessageID: 1, Mode: ModeHold, Status: HoldReview,
		CreatedAt: pgtype.Timestamptz{Time: gq.now.Add(-25 * time.Hour), Valid: true}}
	gq.statuses[1] = StatusHeld
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldRejected || *gq.holds[1].Decision != "review_expired" || !rj.rejected[1] {
		t.Fatalf("review older than 24 h is rejected: %+v", gq.holds[1])
	}
}

func TestGateTimeoutFailOpenAndClosed(t *testing.T) {
	for _, failOpen := range []bool{true, false} {
		g, gq, rj, _ := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 10, FailOpen: failOpen, ChargeBlocked: true})
		gq.text = "Visit https://slow.example.com/"
		gq.statuses[1] = StatusQueuedForSend
		g.BeforeSend(context.Background(), 1, 5)
		gq.verdicts[1] = []database.ListHoldJobVerdictsRow{vrow("processing", "")}
		gq.now = gq.now.Add(11 * time.Second)
		g.RunOnce(context.Background())
		if failOpen {
			if gq.holds[1].Status != HoldReleased || *gq.holds[1].Decision != "timeout_released" {
				t.Fatalf("fail open: %+v", gq.holds[1])
			}
		} else {
			if gq.holds[1].Status != HoldRejected || *gq.holds[1].Decision != "timeout_rejected" {
				t.Fatalf("fail closed: %+v", gq.holds[1])
			}
			if refund := rj.rejected[1]; refund {
				t.Fatal("charge_blocked keeps the charge")
			}
		}
	}
}

func TestGateFailedScanCountsAsNoVerdict(t *testing.T) {
	g, gq, _, _ := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 30, FailOpen: false})
	gq.text = "Visit https://broken.example.com/"
	gq.statuses[1] = StatusQueuedForSend
	g.BeforeSend(context.Background(), 1, 5)
	gq.verdicts[1] = []database.ListHoldJobVerdictsRow{vrow("failed", "")}
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldRejected {
		t.Fatalf("fail closed with no verdict: %+v", gq.holds[1])
	}
}

func TestGateOverrideVerdictWins(t *testing.T) {
	g, gq, _, _ := gateFor(t, Settings{Mode: ModeBlock, HoldSeconds: 30})
	gq.text = "Visit https://fp.example.com/"
	gq.statuses[1] = StatusQueuedForSend
	g.BeforeSend(context.Background(), 1, 5)
	r := vrow("done", VerdictMalicious)
	safe := VerdictSafe
	r.OverrideVerdict = &safe
	gq.verdicts[1] = []database.ListHoldJobVerdictsRow{r}
	g.RunOnce(context.Background())
	if gq.holds[1].Status != HoldReleased {
		t.Fatalf("an admin's false positive override should release: %+v", gq.holds[1])
	}
}

func TestEnqueuerMonitorModeQueuesRealTraffic(t *testing.T) {
	q := &fakeQ{simulator: false}
	e := NewEnqueuer(q, defaultCfg)
	e.SetSettings(NewSettingsCache(func(context.Context) ([]database.ListActiveLinkScanSettingsRow, error) {
		return []database.ListActiveLinkScanSettingsRow{{ServiceProviderID: 2, Mode: ModeMonitor}, {ServiceProviderID: 3, Mode: ModeBlock}}, nil
	}, time.Minute))
	_ = e.MessageSent(context.Background(), 1, 2, 9, "https://a.com/x")
	if len(q.jobs) != 1 || q.jobs[0].Priority != 0 {
		t.Fatalf("monitor customer: %+v", q.jobs)
	}
	_ = e.MessageSent(context.Background(), 2, 3, 9, "https://b.com/x")
	_ = e.MessageSent(context.Background(), 3, 4, 9, "https://c.com/x")
	if len(q.jobs) != 1 {
		t.Fatalf("block mode is queued before sending, other customers never: %+v", q.jobs)
	}
}
