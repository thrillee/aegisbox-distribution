package linkscan

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
)

// ScanConfig controls the scan worker (steps 3 and 4).
type ScanConfig struct {
	// Enabled turns scanning on. Off by default; extraction (step 2) runs either way.
	Enabled bool `envconfig:"LINKSCAN_ENABLED" default:"false"`
	// URL of the linkscanner service, for example http://linkscanner:8080.
	URL   string `envconfig:"LINKSCAN_URL" default:"http://linkscanner:8080"`
	Token string `envconfig:"LINKSCAN_TOKEN"`
	// MaxConcurrency is how many links are scanned at once (keep it at or below the scanner's).
	MaxConcurrency int           `envconfig:"LINKSCAN_MAX_CONCURRENCY" default:"2"`
	PollInterval   time.Duration `envconfig:"LINKSCAN_POLL_INTERVAL" default:"5s"`
	RequestTimeout time.Duration `envconfig:"LINKSCAN_REQUEST_TIMEOUT" default:"90s"`
	MaxAttempts    int           `envconfig:"LINKSCAN_MAX_ATTEMPTS" default:"3"`
	// AlwaysVisit opens a link even when the same URL was visited recently (see Config.AlwaysVisit).
	AlwaysVisit bool `envconfig:"LINKSCAN_ALWAYS_VISIT" default:"false"`
	// ScreenshotDir is where screenshots are written (a private volume).
	ScreenshotDir       string        `envconfig:"LINKSCAN_SCREENSHOT_DIR" default:"/data/linkscans"`
	ScreenshotRetention time.Duration `envconfig:"LINKSCAN_SCREENSHOT_RETENTION" default:"168h"`
	ResultRetention     time.Duration `envconfig:"LINKSCAN_RESULT_RETENTION" default:"672h"`
	// BacklogWarn logs a warning when this many jobs wait, or the oldest waits 30 minutes.
	BacklogWarn int `envconfig:"LINKSCAN_BACKLOG_WARN" default:"500"`
	// RDAPEnabled looks up domain registration dates (new domains add to the score).
	RDAPEnabled bool `envconfig:"LINKSCAN_RDAP_ENABLED" default:"true"`
	// URLhausAuthKey turns on URLhaus lookups. Set it only after confirming URLhaus's terms for
	// commercial use.
	URLhausAuthKey string `envconfig:"LINKSCAN_URLHAUS_AUTH_KEY"`
	// DatabaseHost is blocked for scanned pages (set from DATABASE_URL, not the environment).
	DatabaseHost string `ignored:"true"`
}

// Cache lifetimes of a visit result for the same URL.
const (
	cacheSafe      = 24 * time.Hour
	cacheMalicious = 7 * 24 * time.Hour
	// A page that could not be loaded is checked again after an hour.
	cacheUnreachable = time.Hour
)

// WorkerQuerier is the part of database.Querier the scan worker uses.
type WorkerQuerier interface {
	ClaimLinkScanJobs(ctx context.Context, batchSize int32) ([]database.LinkScanJob, error)
	CompleteLinkScanJob(ctx context.Context, arg database.CompleteLinkScanJobParams) error
	FailLinkScanJob(ctx context.Context, arg database.FailLinkScanJobParams) error
	DeferLinkScanJob(ctx context.Context, arg database.DeferLinkScanJobParams) error
	InsertLinkScan(ctx context.Context, arg database.InsertLinkScanParams) (int64, error)
	FindCachedLinkScan(ctx context.Context, normalizedUrl string) (database.FindCachedLinkScanRow, error)
	FindRecentGroupVisit(ctx context.Context, groupKey string) (database.FindRecentGroupVisitRow, error)
	ListOldScreenshots(ctx context.Context, before pgtype.Timestamptz) ([]database.ListOldScreenshotsRow, error)
	ClearScreenshotPath(ctx context.Context, id int64) error
	DeleteOldLinkScanData(ctx context.Context, before pgtype.Timestamptz) (int32, error)
	CountLinkScanBacklog(ctx context.Context) (database.CountLinkScanBacklogRow, error)
}

// Scanner calls the linkscanner service.
type Scanner interface {
	Scan(ctx context.Context, rawURL string, blockHosts []string) (*ScanResult, error)
}

// HTTPScanner is the linkscanner HTTP client.
type HTTPScanner struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// ErrScannerBusy means the scanner refused the job for now (queue full); it is retried.
var ErrScannerBusy = errors.New("link scanner busy")

// ErrScannerUnavailable wraps failures that say nothing about the page: the scanner cannot be
// reached, is starting, is not configured or rejects the token. Such jobs wait without using up
// their attempts, so links are not lost while the scanner is down.
var ErrScannerUnavailable = errors.New("link scanner unavailable")

func (s *HTTPScanner) Scan(ctx context.Context, rawURL string, blockHosts []string) (*ScanResult, error) {
	body, _ := json.Marshal(map[string]any{"url": rawURL, "blockHosts": blockHosts})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.BaseURL, "/")+"/scan", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	resp, err := s.Client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("call link scanner: %w", err)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			// The page took the scanner too long: counts as an attempt.
			return nil, fmt.Errorf("call link scanner: %w", err)
		}
		return nil, fmt.Errorf("%w: %v", ErrScannerUnavailable, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read link scanner response: %w", err)
	}
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, ErrScannerBusy
	case resp.StatusCode == http.StatusServiceUnavailable, resp.StatusCode == http.StatusUnauthorized,
		resp.StatusCode == http.StatusBadGateway, resp.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("%w: answered %d: %s", ErrScannerUnavailable, resp.StatusCode, strings.TrimSpace(string(data)))
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("link scanner answered %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var res ScanResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("decode link scanner response: %w", err)
	}
	return &res, nil
}

// Worker scans queued links, scores them and stores the results.
type Worker struct {
	q       WorkerQuerier
	scanner Scanner
	cfg     ScanConfig
	now     func() time.Time

	lists   *DomainLists // admin allow and block lists (optional)
	ager    *DomainAger  // RDAP domain age (optional)
	urlhaus *URLhaus     // URLhaus lookups (optional)
}

// SetChecks adds the admin lists, domain age and reputation lookups. Any may be nil.
func (w *Worker) SetChecks(lists *DomainLists, ager *DomainAger, urlhaus *URLhaus) {
	w.lists, w.ager, w.urlhaus = lists, ager, urlhaus
}

func NewWorker(q WorkerQuerier, cfg ScanConfig) *Worker {
	if cfg.MaxConcurrency < 1 {
		cfg.MaxConcurrency = 1
	}
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 3
	}
	return &Worker{
		q:       q,
		scanner: &HTTPScanner{BaseURL: cfg.URL, Token: cfg.Token, Client: &http.Client{Timeout: cfg.RequestTimeout}},
		cfg:     cfg,
		now:     time.Now,
	}
}

// DatabaseHostFromURL returns the host of a Postgres connection string, so scanned pages can be
// kept away from the database.
func DatabaseHostFromURL(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return ""
	}
	if h := u.Hostname(); h != "" {
		return h
	}
	if h := u.Query().Get("host"); h != "" && !strings.HasPrefix(h, "/") {
		return h
	}
	return ""
}

// Validate reports settings that would make scanning fail or unsafe.
func (c ScanConfig) Validate() error {
	if len(c.Token) < 16 {
		return errors.New("LINKSCAN_TOKEN must be set (at least 16 characters, the same as the scanner's)")
	}
	if c.URL == "" {
		return errors.New("LINKSCAN_URL must be set")
	}
	if c.ScreenshotDir == "" {
		return errors.New("LINKSCAN_SCREENSHOT_DIR must be set")
	}
	return nil
}

// Start runs the worker and the retention job until ctx is cancelled.
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	if err := os.MkdirAll(w.cfg.ScreenshotDir, 0o750); err != nil {
		slog.Error("Link scanning: screenshot folder is not writable; screenshots will not be kept",
			slog.String("dir", w.cfg.ScreenshotDir), slog.Any("error", err))
	}
	slog.Info("Link scan worker started", slog.String("scanner", w.cfg.URL), slog.Int("concurrency", w.cfg.MaxConcurrency))

	wg.Add(2)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(w.cfg.PollInterval)
		defer ticker.Stop()
		for {
			n := w.RunOnce(ctx)
			if n > 0 {
				continue // more may be waiting
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			w.Cleanup(ctx)
			w.checkBacklog(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// StartCleanupOnly runs only the retention job, for when scanning is off: links of simulator
// messages are still queued (step 2) and must not pile up forever.
func (w *Worker) StartCleanupOnly(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			w.Cleanup(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// RunOnce claims a batch and processes it. It returns the number of jobs claimed.
func (w *Worker) RunOnce(ctx context.Context) int {
	if ctx.Err() != nil {
		return 0
	}
	jobs, err := w.q.ClaimLinkScanJobs(ctx, int32(w.cfg.MaxConcurrency))
	if err != nil {
		slog.ErrorContext(ctx, "Link scan worker: claim failed", slog.Any("error", err))
		return 0
	}
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(j database.LinkScanJob) {
			defer wg.Done()
			w.process(ctx, j)
		}(j)
	}
	wg.Wait()
	return len(jobs)
}

func linkFromJob(j database.LinkScanJob) Link {
	return Link{
		Raw:        j.Url,
		Normalized: j.NormalizedUrl,
		Host:       j.Host,
		Kind:       j.LinkKind,
		Mode:       j.ScanMode,
		GroupKey:   j.GroupKey,
		IsIP:       net.ParseIP(j.Host) != nil,
		IsPunycode: strings.HasPrefix(j.Host, "xn--") || strings.Contains(j.Host, ".xn--"),
	}
}

func (w *Worker) process(ctx context.Context, j database.LinkScanJob) {
	logCtx := ctx
	log := slog.With(slog.Int64("link_scan_job_id", j.ID), slog.Int64("message_id", j.MessageID), slog.String("url", j.NormalizedUrl))
	l := linkFromJob(j)

	// An admin allow or block entry for the link's own domain decides without opening it.
	if m, ok := w.lists.Match(logCtx, j.ServiceProviderID, j.Host); ok {
		w.finishWithList(logCtx, j, l, m, 0)
		return
	}

	// Same URL visited recently: reuse that result, unless every link must be opened.
	if w.cfg.AlwaysVisit && j.ScanMode == ModeVisit {
		// Fall through to a fresh visit.
	} else if cached, err := w.q.FindCachedLinkScan(logCtx, j.NormalizedUrl); err == nil {
		if w.finalHostListed(logCtx, j, l, cached.ID, cached.FinalHost) {
			return
		}
		w.complete(logCtx, j, cached.ID)
		log.Info("Link scan: reused recent result", slog.String("verdict", cached.Verdict))
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		w.fail(logCtx, j, fmt.Errorf("cache lookup: %w", err))
		return
	}

	if j.ScanMode != ModeVisit {
		// Not opened (action link, or sampled out of a per recipient campaign). A sampled link
		// shares the verdict of a visited link of its group; otherwise only static checks apply.
		if j.ModeReason != nil && *j.ModeReason == ReasonSampled {
			if g, err := w.q.FindRecentGroupVisit(logCtx, j.GroupKey); err == nil {
				if w.finalHostListed(logCtx, j, l, g.ID, g.FinalHost) {
					return
				}
				w.complete(logCtx, j, g.ID)
				log.Info("Link scan: sampled link shares its group's result", slog.String("verdict", g.Verdict))
				return
			}
			// The visited links of its group may still be in the queue: wait for them, and fall
			// back to static checks on the last attempt.
			if int(j.Attempts) < w.cfg.MaxAttempts {
				w.fail(logCtx, j, errors.New("waiting for a visited link of the same group"))
				return
			}
		}
		id, err := w.store(logCtx, l, "static", nil)
		if err != nil {
			w.fail(logCtx, j, err)
			return
		}
		w.complete(logCtx, j, id)
		return
	}

	var blockHosts []string
	if w.cfg.DatabaseHost != "" {
		blockHosts = []string{w.cfg.DatabaseHost}
	}
	res, err := w.scanner.Scan(logCtx, j.NormalizedUrl, blockHosts)
	if err != nil {
		if errors.Is(err, ErrScannerBusy) || errors.Is(err, ErrScannerUnavailable) {
			// Not a problem with the page: wait without using up an attempt.
			w.deferJob(logCtx, j, err)
			return
		}
		w.fail(logCtx, j, err)
		return
	}
	id, err := w.store(logCtx, l, "visit", res)
	if err != nil {
		w.fail(logCtx, j, err)
		return
	}
	if w.finalHostListed(logCtx, j, l, id, finalHostOf(res)) {
		return
	}
	w.complete(logCtx, j, id)
}

func finalHostOf(res *ScanResult) *string {
	if res == nil || res.FinalURL == nil {
		return nil
	}
	u, err := url.Parse(*res.FinalURL)
	if err != nil || u.Hostname() == "" {
		return nil
	}
	h := strings.ToLower(u.Hostname())
	return &h
}

// finalHostListed applies the admin lists to the page a link led to (a shortener to a blocked
// domain, for example). It completes the job and returns true when a list decided.
func (w *Worker) finalHostListed(ctx context.Context, j database.LinkScanJob, l Link, baseScanID int64, finalHost *string) bool {
	if finalHost == nil || *finalHost == "" || *finalHost == j.Host {
		return false
	}
	m, ok := w.lists.Match(ctx, j.ServiceProviderID, *finalHost)
	if !ok {
		return false
	}
	w.finishWithList(ctx, j, l, m, baseScanID)
	return true
}

// finishWithList stores a result decided by an admin list (not reused for other jobs, because
// lists can be per customer) and completes the job with it.
func (w *Worker) finishWithList(ctx context.Context, j database.LinkScanJob, l Link, m ListMatch, baseScanID int64) {
	scope := "this customer's"
	if m.Global {
		scope = "the global"
	}
	verdict, score, code := VerdictSafe, 0, "admin_allow"
	if m.ListType == ListBlock {
		verdict, score, code = VerdictMalicious, 100, "admin_block"
	}
	detail := fmt.Sprintf("%s is on %s %s list (%s)", m.Host, scope, m.ListType, m.Domain)
	if baseScanID > 0 {
		detail += fmt.Sprintf("; the link led there (scan %d)", baseScanID)
	}
	reasons := []Reason{{Code: code, Points: score, Detail: detail}}
	now := w.now()
	p := database.InsertLinkScanParams{
		NormalizedUrl: l.Normalized,
		GroupKey:      l.GroupKey,
		ScanKind:      "list",
		Verdict:       verdict,
		Score:         int32(score),
		Reasons:       mustJSON(reasons),
		Redirects:     []byte("[]"),
		Forms:         []byte("[]"),
		Downloads:     []byte("[]"),
		Blocked:       []byte("[]"),
		ExpiresAt:     pgtype.Timestamptz{Time: now, Valid: true},
	}
	if m.Host != j.Host {
		h := m.Host
		p.FinalHost = &h
	}
	id, err := w.q.InsertLinkScan(ctx, p)
	if err != nil {
		w.fail(ctx, j, fmt.Errorf("save list result: %w", err))
		return
	}
	slog.InfoContext(ctx, "Link decided by admin list", slog.String("url", l.Normalized),
		slog.String("list", m.ListType), slog.String("domain", m.Domain), slog.Int64("link_scan_id", id))
	w.complete(ctx, j, id)
}

func ptrInt32(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" {
		return []byte("[]")
	}
	return b
}

// enrich adds the domain age of the page's domain (the final one after redirects) and reputation
// list hits for the link and the final URL.
func (w *Worker) enrich(ctx context.Context, l *Link, res *ScanResult) {
	host := l.Host
	if fh := finalHostOf(res); fh != nil {
		host = *fh
	}
	if l.Kind == KindWeb || l.Kind == "" {
		l.AgeDomain, l.DomainAgeDays = w.ager.AgeDays(ctx, host)
	}
	if w.urlhaus != nil {
		urls := []string{l.Normalized}
		if res != nil && res.FinalURL != nil && *res.FinalURL != l.Normalized {
			urls = append(urls, *res.FinalURL)
		}
		for _, u := range urls {
			hit, err := w.urlhaus.Lookup(ctx, u)
			if err != nil {
				slog.WarnContext(ctx, "URLhaus lookup failed", slog.Any("error", err))
				break
			}
			if hit != "" {
				l.ListedBy = hit
				break
			}
		}
	}
}

// store scores a result (res nil for static checks), writes its screenshot and saves it.
func (w *Worker) store(ctx context.Context, l Link, kind string, res *ScanResult) (int64, error) {
	w.enrich(ctx, &l, res)
	verdict, score, reasons := Score(l, res)
	now := w.now()
	expires := now.Add(cacheSafe)
	switch {
	case verdict == VerdictMalicious:
		expires = now.Add(cacheMalicious)
	case res != nil && res.ErrorKind != nil && *res.ErrorKind != "blocked":
		// The page could not be checked (DNS, timeout...): try again sooner.
		expires = now.Add(cacheUnreachable)
	}
	p := database.InsertLinkScanParams{
		NormalizedUrl: l.Normalized,
		GroupKey:      l.GroupKey,
		ScanKind:      kind,
		Verdict:       verdict,
		Score:         int32(score),
		Reasons:       mustJSON(reasons),
		Redirects:     []byte("[]"),
		Forms:         []byte("[]"),
		Downloads:     []byte("[]"),
		Blocked:       []byte("[]"),
		ExpiresAt:     pgtype.Timestamptz{Time: expires, Valid: true},
	}
	if kind == "static" {
		// A static check is not a visit: it is not reused for other jobs (FindCachedLinkScan
		// only takes visits), and expires with the results.
		p.ExpiresAt = pgtype.Timestamptz{Time: now, Valid: true}
	}
	if res != nil {
		p.FinalUrl = res.FinalURL
		if res.FinalURL != nil {
			if u, err := url.Parse(*res.FinalURL); err == nil {
				h := strings.ToLower(u.Hostname())
				p.FinalHost = &h
			}
		}
		p.HttpStatus = ptrInt32(res.Status)
		p.Title = res.Title
		if res.Text != nil {
			t := *res.Text
			if len(t) > 20*1024 {
				t = t[:20*1024]
			}
			t = strings.ToValidUTF8(t, "")
			p.TextExcerpt = &t
		}
		p.Redirects = mustJSON(res.Redirects)
		p.Forms = mustJSON(res.Forms)
		p.Downloads = mustJSON(res.Downloads)
		p.Blocked = mustJSON(res.Blocked)
		p.TlsIssuer = res.TLSIssuer
		p.ScanError = res.Error
		d := int32(res.DurationMs)
		p.DurationMs = &d
		if res.Screenshot != nil && *res.Screenshot != "" {
			if path, err := w.saveScreenshot(now, *res.Screenshot); err != nil {
				slog.ErrorContext(ctx, "Link scan: could not save screenshot", slog.Any("error", err))
			} else {
				p.ScreenshotPath = &path
			}
		}
	}
	id, err := w.q.InsertLinkScan(ctx, p)
	if err != nil {
		if p.ScreenshotPath != nil {
			_ = os.Remove(filepath.Join(w.cfg.ScreenshotDir, *p.ScreenshotPath))
		}
		return 0, fmt.Errorf("save link scan: %w", err)
	}
	slog.InfoContext(ctx, "Link scanned", slog.String("url", l.Normalized), slog.String("kind", kind),
		slog.String("verdict", verdict), slog.Int("score", score), slog.Int64("link_scan_id", id))
	return id, nil
}

// saveScreenshot writes a JPEG and returns its path relative to the screenshot folder
// (YYYY-MM-DD/<random>.jpg).
func (w *Worker) saveScreenshot(now time.Time, b64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("decode screenshot: %w", err)
	}
	if len(data) < 3 || data[0] != 0xFF || data[1] != 0xD8 {
		return "", errors.New("screenshot is not a JPEG")
	}
	day := now.UTC().Format("2006-01-02")
	dir := filepath.Join(w.cfg.ScreenshotDir, day)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "scan-*.jpg")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return filepath.Join(day, filepath.Base(f.Name())), nil
}

func (w *Worker) complete(ctx context.Context, j database.LinkScanJob, scanID int64) {
	if err := w.q.CompleteLinkScanJob(ctx, database.CompleteLinkScanJobParams{ID: j.ID, LinkScanID: &scanID}); err != nil {
		slog.ErrorContext(ctx, "Link scan: could not complete job", slog.Int64("link_scan_job_id", j.ID), slog.Any("error", err))
	}
}

func (w *Worker) deferJob(ctx context.Context, j database.LinkScanJob, cause error) {
	msg := cause.Error()
	if len(msg) > 1000 {
		msg = msg[:1000]
	}
	slog.WarnContext(ctx, "Link scan postponed: scanner not available", slog.Int64("link_scan_job_id", j.ID), slog.String("error", msg))
	if err := w.q.DeferLinkScanJob(ctx, database.DeferLinkScanJobParams{ID: j.ID, LastError: &msg}); err != nil {
		slog.ErrorContext(ctx, "Link scan: could not postpone job", slog.Any("error", err))
	}
}

func (w *Worker) fail(ctx context.Context, j database.LinkScanJob, cause error) {
	msg := cause.Error()
	if len(msg) > 1000 {
		msg = msg[:1000]
	}
	slog.WarnContext(ctx, "Link scan attempt failed", slog.Int64("link_scan_job_id", j.ID),
		slog.Int("attempt", int(j.Attempts)), slog.String("error", msg))
	base := int32(30)
	if j.Priority > 0 {
		base = 3 // a real message is waiting on this verdict
	}
	if err := w.q.FailLinkScanJob(ctx, database.FailLinkScanJobParams{ID: j.ID, MaxAttempts: int32(w.cfg.MaxAttempts), LastError: &msg, BaseSeconds: base}); err != nil {
		slog.ErrorContext(ctx, "Link scan: could not record failed attempt", slog.Any("error", err))
	}
}

// Cleanup removes screenshots older than the screenshot retention and scan data older than the
// result retention.
func (w *Worker) Cleanup(ctx context.Context) {
	now := w.now()
	for {
		rows, err := w.q.ListOldScreenshots(ctx, pgtype.Timestamptz{Time: now.Add(-w.cfg.ScreenshotRetention), Valid: true})
		if err != nil {
			slog.ErrorContext(ctx, "Link scan retention: list screenshots", slog.Any("error", err))
			return
		}
		for _, r := range rows {
			if r.ScreenshotPath != nil {
				p := filepath.Join(w.cfg.ScreenshotDir, filepath.Clean("/"+*r.ScreenshotPath))
				if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
					slog.ErrorContext(ctx, "Link scan retention: delete screenshot", slog.String("path", p), slog.Any("error", err))
					continue
				}
			}
			if err := w.q.ClearScreenshotPath(ctx, r.ID); err != nil {
				slog.ErrorContext(ctx, "Link scan retention: clear screenshot path", slog.Any("error", err))
			}
		}
		if len(rows) < 500 || ctx.Err() != nil {
			break
		}
	}
	removeEmptyDayDirs(w.cfg.ScreenshotDir)
	n, err := w.q.DeleteOldLinkScanData(ctx, pgtype.Timestamptz{Time: now.Add(-w.cfg.ResultRetention), Valid: true})
	if err != nil {
		slog.ErrorContext(ctx, "Link scan retention: delete old results", slog.Any("error", err))
		return
	}
	if n > 0 {
		slog.InfoContext(ctx, "Link scan retention: old results deleted", slog.Int("scans", int(n)))
	}
}

func removeEmptyDayDirs(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			_ = os.Remove(filepath.Join(root, e.Name())) // only succeeds when empty
		}
	}
}

func (w *Worker) checkBacklog(ctx context.Context) {
	b, err := w.q.CountLinkScanBacklog(ctx)
	if err != nil {
		return
	}
	if int(b.Queued) >= w.cfg.BacklogWarn || b.OldestSeconds >= 1800 {
		slog.WarnContext(ctx, "Link scan queue is growing: is the linkscanner service up?",
			slog.Int("waiting", int(b.Queued)), slog.Int("oldest_wait_seconds", int(b.OldestSeconds)))
	}
}
