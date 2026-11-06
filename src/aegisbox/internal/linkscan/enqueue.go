package linkscan

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
)

// Extraction outcomes recorded in message_link_extractions.
const (
	OutcomeLinks      = "links"
	OutcomeNoLinks    = "no_links"
	OutcomeSkippedOTP = "skipped_otp"
)

// Config controls extraction. Scanning itself (step 3) has its own settings.
type Config struct {
	// Enabled turns link extraction for simulator traffic on or off.
	Enabled bool `envconfig:"LINKSCAN_EXTRACT_ENABLED" default:"true"`
	// SkipOTP skips messages that read like one time passwords.
	SkipOTP bool `envconfig:"LINKSCAN_SKIP_OTP" default:"true"`
	// SamplePerHour is how many links of one group (same customer, same host and path apart from a
	// per recipient token) are visited per hour; the rest get reputation checks only.
	SamplePerHour int `envconfig:"LINKSCAN_SAMPLE_PER_HOUR" default:"3"`
	// AlwaysVisit opens every visitable link: no sampling here, and the worker does not reuse an
	// earlier result for the same URL. Admin allow and block lists still decide first.
	AlwaysVisit bool `envconfig:"LINKSCAN_ALWAYS_VISIT" default:"false"`
	// VisitDelayMin and VisitDelayMax make a link wait a random time in that range before it is
	// opened, so visits don't follow every send at once. Links of held messages (Phase 2 hold or
	// block, waiting for a verdict before sending) are never delayed. Both 0 means no delay.
	VisitDelayMin time.Duration `envconfig:"LINKSCAN_VISIT_DELAY_MIN" default:"0s"`
	VisitDelayMax time.Duration `envconfig:"LINKSCAN_VISIT_DELAY_MAX" default:"0s"`
}

// maxVisitDelay caps the random delay; a longer wait would only fill the queue.
const maxVisitDelay = 24 * time.Hour

// visitDelay picks the wait before a link of this mode and priority is opened.
func (e *Enqueuer) visitDelay(mode string, priority int16) time.Duration {
	if mode != ModeVisit || priority > 0 || e.cfg.VisitDelayMax <= 0 {
		return 0
	}
	lo, hi := e.cfg.VisitDelayMin, e.cfg.VisitDelayMax
	if hi <= lo {
		return lo
	}
	return lo + time.Duration(rand.Int64N(int64(hi-lo)+1))
}

// Querier is the part of database.Querier the enqueuer uses.
type Querier interface {
	IsSimulatorMNO(ctx context.Context, mnoID int32) (bool, error)
	CountRecentVisitsForGroup(ctx context.Context, arg database.CountRecentVisitsForGroupParams) (int32, error)
	InsertLinkScanJob(ctx context.Context, arg database.InsertLinkScanJobParams) error
	RecordMessageLinkExtraction(ctx context.Context, arg database.RecordMessageLinkExtractionParams) error
}

// Enqueuer queues the links of messages sent through a simulator MNO, and (Phase 2) of real
// messages of customers whose scan mode is not off.
type Enqueuer struct {
	q        Querier
	cfg      Config
	settings *SettingsCache
}

// SetSettings enables Phase 2: real messages of customers in monitor mode are queued after
// sending (customers in hold or block mode are queued before sending, by the gate).
func (e *Enqueuer) SetSettings(s *SettingsCache) { e.settings = s }

func NewEnqueuer(q Querier, cfg Config) *Enqueuer {
	if cfg.SamplePerHour < 0 {
		cfg.SamplePerHour = 0
	}
	cfg.VisitDelayMin = min(max(cfg.VisitDelayMin, 0), maxVisitDelay)
	cfg.VisitDelayMax = min(max(cfg.VisitDelayMax, 0), maxVisitDelay)
	return &Enqueuer{q: q, cfg: cfg}
}

// MessageSent is called after a submit was accepted by the MNO. It does nothing unless the MNO is
// a simulator. It is safe to call twice for one message (failover): jobs and the extraction record
// are unique per message.
func (e *Enqueuer) MessageSent(ctx context.Context, messageID int64, spID int32, mnoID int32, text string) error {
	if e == nil || !e.cfg.Enabled {
		return nil
	}
	isSim, err := e.q.IsSimulatorMNO(ctx, mnoID)
	if err != nil {
		return fmt.Errorf("check simulator MNO %d: %w", mnoID, err)
	}
	if !isSim {
		// Real traffic is scanned after sending only for customers in monitor mode.
		if s, ok := e.settings.Get(ctx, spID); !ok || s.Mode != ModeMonitor {
			return nil
		}
	}
	_, err = e.queue(ctx, messageID, spID, text, 0)
	return err
}

// HoldLinks queues a real message's links before it is sent (hold and block modes), first in the
// scan queue. It returns how many of them get a verdict (0: send at once).
func (e *Enqueuer) HoldLinks(ctx context.Context, messageID int64, spID int32, text string) (int, error) {
	if e == nil {
		return 0, nil
	}
	return e.queue(ctx, messageID, spID, text, 1)
}

func (e *Enqueuer) queue(ctx context.Context, messageID int64, spID int32, text string, priority int16) (int, error) {
	links := Extract(text)
	outcome := OutcomeLinks
	switch {
	case e.cfg.SkipOTP && LooksLikeOTP(text):
		outcome = OutcomeSkippedOTP
	case len(links) == 0:
		outcome = OutcomeNoLinks
	}

	verdicts := 0
	if outcome == OutcomeLinks {
		for _, l := range links {
			mode, err := e.insert(ctx, messageID, spID, l, priority)
			if err != nil {
				return 0, err
			}
			if mode != ModeNone {
				verdicts++
			}
		}
	}

	// Written after the jobs, so a crash in between leaves the message to be recorded by a retry
	// rather than marked done without its links.
	if err := e.q.RecordMessageLinkExtraction(ctx, database.RecordMessageLinkExtractionParams{
		MessageID: messageID,
		Outcome:   outcome,
		LinkCount: int32(len(links)),
	}); err != nil {
		return 0, fmt.Errorf("record link extraction for message %d: %w", messageID, err)
	}
	slog.InfoContext(ctx, "Link extraction for message",
		slog.Int64("message_id", messageID), slog.String("outcome", outcome), slog.Int("links", len(links)),
		slog.Bool("before_send", priority > 0))
	return verdicts, nil
}

func (e *Enqueuer) insert(ctx context.Context, messageID int64, spID int32, l Link, priority int16) (string, error) {
	mode, reason := l.Mode, l.Reason
	if mode == ModeVisit && !e.cfg.AlwaysVisit {
		visits, err := e.q.CountRecentVisitsForGroup(ctx, database.CountRecentVisitsForGroupParams{
			ServiceProviderID: spID,
			GroupKey:          l.GroupKey,
		})
		if err != nil {
			return "", fmt.Errorf("count visits for %s: %w", l.GroupKey, err)
		}
		if int(visits) >= e.cfg.SamplePerHour {
			mode, reason = ModeReputation, ReasonSampled
		}
	}

	status := "queued"
	if mode == ModeNone {
		status = "skipped"
	}
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	if err := e.q.InsertLinkScanJob(ctx, database.InsertLinkScanJobParams{
		MessageID:         messageID,
		ServiceProviderID: spID,
		Url:               l.Raw,
		NormalizedUrl:     l.Normalized,
		Host:              l.Host,
		GroupKey:          l.GroupKey,
		LinkKind:          l.Kind,
		ScanMode:          mode,
		ModeReason:        reasonPtr,
		Status:            status,
		Priority:          priority,
		DelayMs:           int32(e.visitDelay(mode, priority) / time.Millisecond),
	}); err != nil {
		return "", fmt.Errorf("queue link %s for message %d: %w", l.Normalized, messageID, err)
	}
	return mode, nil
}
