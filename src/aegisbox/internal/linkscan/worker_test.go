package linkscan

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thrillee/aegisbox/internal/database"
)

type fakeWQ struct {
	jobs      []database.LinkScanJob
	scans     []database.InsertLinkScanParams
	completed map[int64]int64
	failed    map[int64]string
	deferred  map[int64]string
	cachedURL map[string]int64
	groupHit  map[string]int64
	cleared   []int64
	old       []database.ListOldScreenshotsRow
}

func newFakeWQ() *fakeWQ {
	return &fakeWQ{completed: map[int64]int64{}, failed: map[int64]string{}, deferred: map[int64]string{}, cachedURL: map[string]int64{}, groupHit: map[string]int64{}}
}
func (f *fakeWQ) ClaimLinkScanJobs(context.Context, int32) ([]database.LinkScanJob, error) {
	j := f.jobs
	f.jobs = nil
	return j, nil
}
func (f *fakeWQ) CompleteLinkScanJob(_ context.Context, a database.CompleteLinkScanJobParams) error {
	f.completed[a.ID] = *a.LinkScanID
	return nil
}
func (f *fakeWQ) FailLinkScanJob(_ context.Context, a database.FailLinkScanJobParams) error {
	f.failed[a.ID] = *a.LastError
	return nil
}
func (f *fakeWQ) DeferLinkScanJob(_ context.Context, a database.DeferLinkScanJobParams) error {
	f.deferred[a.ID] = *a.LastError
	return nil
}
func (f *fakeWQ) InsertLinkScan(_ context.Context, a database.InsertLinkScanParams) (int64, error) {
	f.scans = append(f.scans, a)
	return int64(100 + len(f.scans)), nil
}
func (f *fakeWQ) FindCachedLinkScan(_ context.Context, u string) (database.FindCachedLinkScanRow, error) {
	if id, ok := f.cachedURL[u]; ok {
		return database.FindCachedLinkScanRow{ID: id, Verdict: "safe"}, nil
	}
	return database.FindCachedLinkScanRow{}, pgx.ErrNoRows
}
func (f *fakeWQ) FindRecentGroupVisit(_ context.Context, g string) (database.FindRecentGroupVisitRow, error) {
	if id, ok := f.groupHit[g]; ok {
		return database.FindRecentGroupVisitRow{ID: id, Verdict: "malicious"}, nil
	}
	return database.FindRecentGroupVisitRow{}, pgx.ErrNoRows
}
func (f *fakeWQ) ListOldScreenshots(context.Context, pgtype.Timestamptz) ([]database.ListOldScreenshotsRow, error) {
	o := f.old
	f.old = nil
	return o, nil
}
func (f *fakeWQ) ClearScreenshotPath(_ context.Context, id int64) error {
	f.cleared = append(f.cleared, id)
	return nil
}
func (f *fakeWQ) DeleteOldLinkScanData(context.Context, pgtype.Timestamptz) (int32, error) {
	return 0, nil
}
func (f *fakeWQ) CountLinkScanBacklog(context.Context) (database.CountLinkScanBacklogRow, error) {
	return database.CountLinkScanBacklogRow{}, nil
}

type fakeScanner struct {
	res      *ScanResult
	err      error
	calls    []string
	blockArg []string
}

func (s *fakeScanner) Scan(_ context.Context, u string, block []string) (*ScanResult, error) {
	s.calls = append(s.calls, u)
	s.blockArg = block
	return s.res, s.err
}

// A minimal valid JPEG header is enough for the file check.
var jpeg = base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0xFF, 0xE0, 1, 2, 3})

func newTestWorker(t *testing.T, q *fakeWQ, s *fakeScanner) *Worker {
	w := NewWorker(q, ScanConfig{MaxConcurrency: 4, MaxAttempts: 3, ScreenshotDir: t.TempDir(),
		ScreenshotRetention: time.Hour, ResultRetention: time.Hour, DatabaseHost: "db.example.com"})
	w.scanner = s
	return w
}

func job(id int64, url, mode string, reason *string) database.LinkScanJob {
	l := Extract(url)[0]
	return database.LinkScanJob{ID: id, MessageID: id, NormalizedUrl: l.Normalized, Url: url, Host: l.Host,
		GroupKey: l.GroupKey, LinkKind: l.Kind, ScanMode: mode, ModeReason: reason, Attempts: 1}
}

func TestWorkerVisitStoresResultAndScreenshot(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{res: &ScanResult{FinalURL: sp("https://shop.example.com/"), Status: ip(200), Screenshot: &jpeg}}
	w := newTestWorker(t, q, s)
	q.jobs = []database.LinkScanJob{job(1, "https://shop.example.com/", ModeVisit, nil)}
	if n := w.RunOnce(context.Background()); n != 1 {
		t.Fatalf("claimed %d", n)
	}
	if len(s.calls) != 1 || s.blockArg[0] != "db.example.com" {
		t.Fatalf("scanner calls %v block %v", s.calls, s.blockArg)
	}
	if q.completed[1] != 101 || len(q.scans) != 1 || q.scans[0].Verdict != VerdictSafe || q.scans[0].ScanKind != "visit" {
		t.Fatalf("completed %v scans %+v", q.completed, q.scans)
	}
	p := q.scans[0].ScreenshotPath
	if p == nil {
		t.Fatal("no screenshot path")
	}
	if _, err := os.Stat(filepath.Join(w.cfg.ScreenshotDir, *p)); err != nil {
		t.Fatalf("screenshot file: %v", err)
	}
	// Retention removes it.
	q.old = []database.ListOldScreenshotsRow{{ID: 101, ScreenshotPath: p}}
	w.Cleanup(context.Background())
	if _, err := os.Stat(filepath.Join(w.cfg.ScreenshotDir, *p)); !errors.Is(err, os.ErrNotExist) || len(q.cleared) != 1 {
		t.Fatalf("screenshot not removed: %v %v", err, q.cleared)
	}
}

func TestWorkerRetentionCannotEscapeFolder(t *testing.T) {
	q := newFakeWQ()
	w := newTestWorker(t, q, &fakeScanner{})
	outside := filepath.Join(t.TempDir(), "keep.txt")
	os.WriteFile(outside, []byte("x"), 0o600)
	rel, _ := filepath.Rel(w.cfg.ScreenshotDir, outside)
	q.old = []database.ListOldScreenshotsRow{{ID: 1, ScreenshotPath: &rel}}
	w.Cleanup(context.Background())
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("file outside the screenshot folder was deleted: %v", err)
	}
}

func TestWorkerReusesCacheAndGroupResults(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{}
	w := newTestWorker(t, q, s)
	cached := job(1, "https://shop.example.com/", ModeVisit, nil)
	q.cachedURL[cached.NormalizedUrl] = 55
	sampled := job(2, "https://promo.example.com/r/a8F3kQ29xZ", ModeReputation, sp(ReasonSampled))
	q.groupHit[sampled.GroupKey] = 66
	q.jobs = []database.LinkScanJob{cached, sampled}
	w.RunOnce(context.Background())
	if len(s.calls) != 0 || q.completed[1] != 55 || q.completed[2] != 66 || len(q.scans) != 0 {
		t.Fatalf("calls %v completed %v scans %d", s.calls, q.completed, len(q.scans))
	}
}

func TestWorkerActionLinkGetsStaticCheckOnly(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{}
	w := newTestWorker(t, q, s)
	q.jobs = []database.LinkScanJob{job(1, "https://gtbank-verify.com/unsubscribe", ModeReputation, sp(ReasonActionLink))}
	w.RunOnce(context.Background())
	if len(s.calls) != 0 {
		t.Fatal("an action link must never be opened")
	}
	if len(q.scans) != 1 || q.scans[0].ScanKind != "static" || q.scans[0].Verdict != VerdictSuspicious {
		t.Fatalf("scans %+v", q.scans)
	}
}

func TestWorkerScanErrorUsesAnAttempt(t *testing.T) {
	q := newFakeWQ()
	w := newTestWorker(t, q, &fakeScanner{err: errors.New("link scanner answered 500: scan failed")})
	q.jobs = []database.LinkScanJob{job(1, "https://shop.example.com/", ModeVisit, nil)}
	w.RunOnce(context.Background())
	if q.failed[1] == "" || len(q.deferred) != 0 || len(q.completed) != 0 || len(q.scans) != 0 {
		t.Fatalf("failed %v deferred %v completed %v", q.failed, q.deferred, q.completed)
	}
}

func TestWorkerScannerDownPostponesWithoutUsingAttempts(t *testing.T) {
	for _, err := range []error{ErrScannerBusy, fmt.Errorf("%w: connection refused", ErrScannerUnavailable)} {
		q := newFakeWQ()
		w := newTestWorker(t, q, &fakeScanner{err: err})
		q.jobs = []database.LinkScanJob{job(1, "https://shop.example.com/", ModeVisit, nil)}
		w.RunOnce(context.Background())
		if q.deferred[1] == "" || len(q.failed) != 0 || len(q.completed) != 0 {
			t.Fatalf("%v: deferred %v failed %v", err, q.deferred, q.failed)
		}
	}
}

func TestHTTPScannerClassifiesUnavailable(t *testing.T) {
	for code, unavailable := range map[int]bool{503: true, 401: true, 502: true, 500: false, 400: false} {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(code)
		}))
		s := &HTTPScanner{BaseURL: srv.URL, Token: "t", Client: srv.Client()}
		_, err := s.Scan(context.Background(), "https://example.com/", nil)
		srv.Close()
		if err == nil || errors.Is(err, ErrScannerUnavailable) != unavailable {
			t.Errorf("%d: err %v, unavailable want %v", code, err, unavailable)
		}
	}
	// Nothing listening at all.
	s := &HTTPScanner{BaseURL: "http://127.0.0.1:1", Token: "t", Client: &http.Client{Timeout: 2 * time.Second}}
	if _, err := s.Scan(context.Background(), "https://example.com/", nil); !errors.Is(err, ErrScannerUnavailable) {
		t.Errorf("connection refused: %v", err)
	}
}

func TestDatabaseHostFromURL(t *testing.T) {
	for in, want := range map[string]string{
		"postgres://u:p@db.example.com:5432/aegis?sslmode=require": "db.example.com",
		"postgres://postgres@/aegis?host=/var/run/pg&port=5432":    "",
		"postgres://u@/aegis?host=10.1.2.3":                        "10.1.2.3",
	} {
		if got := DatabaseHostFromURL(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestWorkerSampledLinkWaitsForGroupVisit(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{}
	w := newTestWorker(t, q, s)
	j := job(1, "https://promo.example.com/r/a8F3kQ29xZ", ModeReputation, sp(ReasonSampled))
	q.jobs = []database.LinkScanJob{j}
	w.RunOnce(context.Background())
	if q.failed[1] == "" || len(q.scans) != 0 {
		t.Fatalf("first attempt should wait: failed %v scans %d", q.failed, len(q.scans))
	}
	j.Attempts = 3
	q.jobs = []database.LinkScanJob{j}
	w.RunOnce(context.Background())
	if len(q.scans) != 1 || q.scans[0].ScanKind != "static" || len(s.calls) != 0 {
		t.Fatalf("last attempt should use static checks: %+v", q.scans)
	}
}

func TestWorkerAlwaysVisitIgnoresCachedResult(t *testing.T) {
	q := newFakeWQ()
	s := &fakeScanner{res: &ScanResult{FinalURL: sp("https://shop.example.com/"), Status: ip(200)}}
	w := newTestWorker(t, q, s)
	w.cfg.AlwaysVisit = true
	j := job(1, "https://shop.example.com/", ModeVisit, nil)
	q.cachedURL[j.NormalizedUrl] = 55
	q.jobs = []database.LinkScanJob{j}
	w.RunOnce(context.Background())
	if len(s.calls) != 1 || q.completed[1] == 55 || len(q.scans) != 1 {
		t.Fatalf("calls %v completed %v scans %d", s.calls, q.completed, len(q.scans))
	}
}
