package linkscan

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
)

type fakeQ struct {
	simulator bool
	simMsg    *database.GetSimulatorMessageForLinkScanRow
	visits    map[string]int32
	jobs      []database.InsertLinkScanJobParams
	records   []database.RecordMessageLinkExtractionParams
}

func (f *fakeQ) IsSimulatorMNO(context.Context, int32) (bool, error) { return f.simulator, nil }
func (f *fakeQ) GetSimulatorMessageForLinkScan(context.Context, int64) (database.GetSimulatorMessageForLinkScanRow, error) {
	if f.simMsg == nil {
		return database.GetSimulatorMessageForLinkScanRow{}, pgx.ErrNoRows
	}
	return *f.simMsg, nil
}
func (f *fakeQ) CountRecentVisitsForGroup(_ context.Context, a database.CountRecentVisitsForGroupParams) (int32, error) {
	return f.visits[a.GroupKey], nil
}
func (f *fakeQ) InsertLinkScanJob(_ context.Context, a database.InsertLinkScanJobParams) error {
	f.jobs = append(f.jobs, a)
	if a.ScanMode == ModeVisit {
		if f.visits == nil {
			f.visits = map[string]int32{}
		}
		f.visits[a.GroupKey]++
	}
	return nil
}
func (f *fakeQ) RecordMessageLinkExtraction(_ context.Context, a database.RecordMessageLinkExtractionParams) error {
	f.records = append(f.records, a)
	return nil
}

var defaultCfg = Config{Enabled: true, SkipOTP: true, SamplePerHour: 3, ClickPercent: 100}

func TestEnqueuer_IgnoresRealMNOs(t *testing.T) {
	q := &fakeQ{simulator: false}
	if err := NewEnqueuer(q, defaultCfg).MessageSent(context.Background(), 1, 2, 3, "https://a.com"); err != nil {
		t.Fatal(err)
	}
	if len(q.jobs) != 0 || len(q.records) != 0 {
		t.Fatalf("real MNO traffic must not be touched: %+v %+v", q.jobs, q.records)
	}
}

func TestEnqueuer_Disabled(t *testing.T) {
	q := &fakeQ{simulator: true}
	_ = NewEnqueuer(q, Config{}).MessageSent(context.Background(), 1, 2, 3, "https://a.com")
	if len(q.jobs) != 0 || len(q.records) != 0 {
		t.Fatal("disabled extraction must do nothing")
	}
}

func TestEnqueuer_Outcomes(t *testing.T) {
	ctx := context.Background()

	q := &fakeQ{simulator: true}
	_ = NewEnqueuer(q, defaultCfg).MessageSent(ctx, 1, 2, 3, "Hello, no links here")
	if len(q.records) != 1 || q.records[0].Outcome != OutcomeNoLinks {
		t.Fatalf("want no_links, got %+v", q.records)
	}

	q = &fakeQ{simulator: true}
	_ = NewEnqueuer(q, defaultCfg).MessageSent(ctx, 1, 2, 3, "Your OTP is 123456 https://a.com")
	if len(q.jobs) != 0 || q.records[0].Outcome != OutcomeSkippedOTP || q.records[0].LinkCount != 1 {
		t.Fatalf("want skipped_otp with 1 link counted, got %+v %+v", q.jobs, q.records)
	}

	q = &fakeQ{simulator: true}
	_ = NewEnqueuer(q, Config{Enabled: true, SamplePerHour: 3, ClickPercent: 100}).MessageSent(ctx, 1, 2, 3, "Your OTP is 123456 https://a.com")
	if len(q.jobs) != 1 {
		t.Fatalf("with SkipOTP off the link must be queued, got %+v", q.jobs)
	}

	q = &fakeQ{simulator: true}
	_ = NewEnqueuer(q, defaultCfg).MessageSent(ctx, 1, 2, 3, "Go https://a.com/x or wa.me/234 or https://a.com/unsubscribe")
	if len(q.jobs) != 3 || q.records[0].Outcome != OutcomeLinks || q.records[0].LinkCount != 3 {
		t.Fatalf("got %+v %+v", q.jobs, q.records)
	}
	if j := q.jobs[0]; j.ScanMode != ModeVisit || j.Status != "queued" || j.ModeReason != nil {
		t.Errorf("web link: %+v", j)
	}
	if j := q.jobs[1]; j.ScanMode != ModeNone || j.Status != "skipped" || *j.ModeReason != ReasonNonWeb {
		t.Errorf("whatsapp link: %+v", j)
	}
	if j := q.jobs[2]; j.ScanMode != ModeReputation || j.Status != "queued" || *j.ModeReason != ReasonActionLink {
		t.Errorf("action link: %+v", j)
	}
}

func TestEnqueuer_SamplesPersonalLinks(t *testing.T) {
	q := &fakeQ{simulator: true}
	e := NewEnqueuer(q, defaultCfg)
	tokens := []string{"a8F3kQ29xZ", "b8F3kQ29xZ", "c8F3kQ29xZ", "d8F3kQ29xZ", "e8F3kQ29xZ"}
	for i, tok := range tokens {
		_ = e.MessageSent(context.Background(), int64(i+1), 2, 3, "Hi https://promo.example.com/r/"+tok)
	}
	var visits, sampled int
	for _, j := range q.jobs {
		switch {
		case j.ScanMode == ModeVisit:
			visits++
		case j.ModeReason != nil && *j.ModeReason == ReasonSampled:
			sampled++
		}
	}
	if visits != 3 || sampled != 2 {
		t.Fatalf("want 3 visits and 2 sampled out, got %d and %d", visits, sampled)
	}
}

func TestEnqueuer_AlwaysVisitSkipsSampling(t *testing.T) {
	q := &fakeQ{simulator: true}
	e := NewEnqueuer(q, Config{Enabled: true, SamplePerHour: 3, AlwaysVisit: true, ClickPercent: 100})
	tokens := []string{"a8F3kQ29xZ", "b8F3kQ29xZ", "c8F3kQ29xZ", "d8F3kQ29xZ", "e8F3kQ29xZ"}
	for i, tok := range tokens {
		_ = e.MessageSent(context.Background(), int64(i+1), 2, 3, "Hi https://promo.example.com/r/"+tok)
	}
	for _, j := range q.jobs {
		if j.ScanMode != ModeVisit {
			t.Fatalf("job for %s has mode %s, want every link visited", j.Url, j.ScanMode)
		}
	}
	if len(q.jobs) != len(tokens) {
		t.Fatalf("want %d jobs, got %d", len(tokens), len(q.jobs))
	}
}

func TestEnqueuer_VisitDelay(t *testing.T) {
	q := &fakeQ{simulator: true}
	e := NewEnqueuer(q, Config{Enabled: true, SamplePerHour: 100, ClickPercent: 100, VisitDelayMin: 2 * time.Second, VisitDelayMax: 5 * time.Second})
	for i := 0; i < 50; i++ {
		_ = e.MessageSent(context.Background(), int64(i+1), 2, 3, fmt.Sprintf("Hi https://site%d.example.com/x", i))
	}
	if len(q.jobs) != 50 {
		t.Fatalf("want 50 jobs, got %d", len(q.jobs))
	}
	seen := map[int32]bool{}
	for _, j := range q.jobs {
		if j.DelayMs < 2000 || j.DelayMs > 5000 {
			t.Fatalf("delay %dms outside 2s to 5s", j.DelayMs)
		}
		seen[j.DelayMs] = true
	}
	if len(seen) < 2 {
		t.Fatal("delays should vary")
	}
}

func TestEnqueuer_VisitDelaySkipsHeldAndNonVisit(t *testing.T) {
	e := NewEnqueuer(&fakeQ{}, Config{VisitDelayMin: time.Second, VisitDelayMax: 3 * time.Second})
	if d := e.visitDelay(ModeVisit, 1); d != 0 {
		t.Fatalf("held message links must not wait, got %v", d)
	}
	if d := e.visitDelay(ModeReputation, 0); d != 0 {
		t.Fatalf("links that are not opened must not wait, got %v", d)
	}
	if d := NewEnqueuer(&fakeQ{}, Config{}).visitDelay(ModeVisit, 0); d != 0 {
		t.Fatalf("no delay configured, got %v", d)
	}
	if d := NewEnqueuer(&fakeQ{}, Config{VisitDelayMin: 4 * time.Second, VisitDelayMax: 2 * time.Second}).visitDelay(ModeVisit, 0); d != 4*time.Second {
		t.Fatalf("max below min should use min, got %v", d)
	}
}

func TestEnqueuer_SimulatorWaitsForDelivery(t *testing.T) {
	ctx := context.Background()
	cfg := defaultCfg
	cfg.SimulatorAfterDelivery = true
	q := &fakeQ{simulator: true}
	e := NewEnqueuer(q, cfg)
	if err := e.MessageSent(ctx, 1, 2, 3, "https://a.com/x"); err != nil {
		t.Fatal(err)
	}
	if len(q.jobs) != 0 || len(q.records) != 0 {
		t.Fatalf("simulator links must wait for delivery: %+v %+v", q.jobs, q.records)
	}

	q.simMsg = &database.GetSimulatorMessageForLinkScanRow{ServiceProviderID: 2, ShortMessage: "https://a.com/x"}
	if err := e.MessageDelivered(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if len(q.jobs) != 1 || q.jobs[0].ServiceProviderID != 2 || q.jobs[0].MessageID != 1 {
		t.Fatalf("delivered simulator message should queue its link: %+v", q.jobs)
	}
}

func TestEnqueuer_DeliveredIgnoresNonSimulator(t *testing.T) {
	cfg := defaultCfg
	cfg.SimulatorAfterDelivery = true
	q := &fakeQ{} // no simulator row
	if err := NewEnqueuer(q, cfg).MessageDelivered(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if len(q.jobs) != 0 || len(q.records) != 0 {
		t.Fatalf("non-simulator delivery must do nothing: %+v", q.jobs)
	}
}

func TestEnqueuer_DeliveredOffQueuesAtSend(t *testing.T) {
	q := &fakeQ{simulator: true, simMsg: &database.GetSimulatorMessageForLinkScanRow{ServiceProviderID: 2, ShortMessage: "https://a.com/x"}}
	e := NewEnqueuer(q, defaultCfg) // SimulatorAfterDelivery off
	_ = e.MessageDelivered(context.Background(), 1)
	if len(q.jobs) != 0 {
		t.Fatal("with the setting off, delivery must not queue links")
	}
	_ = e.MessageSent(context.Background(), 1, 2, 3, "https://a.com/x")
	if len(q.jobs) != 1 {
		t.Fatalf("with the setting off, links are queued at send: %+v", q.jobs)
	}
}

func TestEnqueuer_ClickPercent(t *testing.T) {
	ctx := context.Background()
	run := func(pct int) int {
		cfg := defaultCfg
		cfg.SimulatorAfterDelivery = true
		cfg.AlwaysVisit = true
		cfg.ClickPercent = pct
		q := &fakeQ{simMsg: &database.GetSimulatorMessageForLinkScanRow{ServiceProviderID: 2, ShortMessage: "https://a.com/x"}}
		e := NewEnqueuer(q, cfg)
		for i := 0; i < 1000; i++ {
			if err := e.MessageDelivered(ctx, int64(i)); err != nil {
				t.Fatal(err)
			}
		}
		if len(q.records) != 1000 {
			t.Fatalf("every delivered message must be recorded, got %d", len(q.records))
		}
		visits := 0
		for _, j := range q.jobs {
			if j.ScanMode == ModeVisit {
				visits++
			} else if j.ModeReason == nil || *j.ModeReason != ReasonNotClicked {
				t.Fatalf("unclicked link should be a reputation check marked not_clicked: %+v", j)
			}
		}
		return visits
	}
	if n := run(0); n != 0 {
		t.Fatalf("0%% must open nothing, got %d", n)
	}
	if n := run(100); n != 1000 {
		t.Fatalf("100%% must open every message, got %d", n)
	}
	if n := run(30); n < 220 || n > 380 {
		t.Fatalf("30%% of 1000 should be near 300, got %d", n)
	}
}
