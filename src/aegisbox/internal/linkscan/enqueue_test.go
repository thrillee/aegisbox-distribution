package linkscan

import (
	"context"
	"testing"

	"github.com/thrillee/aegisbox/internal/database"
)

type fakeQ struct {
	simulator bool
	visits    map[string]int32
	jobs      []database.InsertLinkScanJobParams
	records   []database.RecordMessageLinkExtractionParams
}

func (f *fakeQ) IsSimulatorMNO(context.Context, int32) (bool, error) { return f.simulator, nil }
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

var defaultCfg = Config{Enabled: true, SkipOTP: true, SamplePerHour: 3}

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
	_ = NewEnqueuer(q, Config{Enabled: true, SamplePerHour: 3}).MessageSent(ctx, 1, 2, 3, "Your OTP is 123456 https://a.com")
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
	e := NewEnqueuer(q, Config{Enabled: true, SamplePerHour: 3, AlwaysVisit: true})
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
