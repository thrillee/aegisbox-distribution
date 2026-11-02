package linkscan

import (
	"context"
	"fmt"
	"log/slog"

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
}

// Querier is the part of database.Querier the enqueuer uses.
type Querier interface {
	IsSimulatorMNO(ctx context.Context, mnoID int32) (bool, error)
	CountRecentVisitsForGroup(ctx context.Context, arg database.CountRecentVisitsForGroupParams) (int32, error)
	InsertLinkScanJob(ctx context.Context, arg database.InsertLinkScanJobParams) error
	RecordMessageLinkExtraction(ctx context.Context, arg database.RecordMessageLinkExtractionParams) error
}

// Enqueuer queues the links of messages sent through a simulator MNO.
type Enqueuer struct {
	q   Querier
	cfg Config
}

func NewEnqueuer(q Querier, cfg Config) *Enqueuer {
	if cfg.SamplePerHour < 0 {
		cfg.SamplePerHour = 0
	}
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
		return nil
	}

	links := Extract(text)
	outcome := OutcomeLinks
	switch {
	case e.cfg.SkipOTP && LooksLikeOTP(text):
		outcome = OutcomeSkippedOTP
	case len(links) == 0:
		outcome = OutcomeNoLinks
	}

	if outcome == OutcomeLinks {
		for _, l := range links {
			if err := e.insert(ctx, messageID, spID, l); err != nil {
				return err
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
		return fmt.Errorf("record link extraction for message %d: %w", messageID, err)
	}
	slog.InfoContext(ctx, "Link extraction for simulator message",
		slog.Int64("message_id", messageID), slog.String("outcome", outcome), slog.Int("links", len(links)))
	return nil
}

func (e *Enqueuer) insert(ctx context.Context, messageID int64, spID int32, l Link) error {
	mode, reason := l.Mode, l.Reason
	if mode == ModeVisit {
		visits, err := e.q.CountRecentVisitsForGroup(ctx, database.CountRecentVisitsForGroupParams{
			ServiceProviderID: spID,
			GroupKey:          l.GroupKey,
		})
		if err != nil {
			return fmt.Errorf("count visits for %s: %w", l.GroupKey, err)
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
	}); err != nil {
		return fmt.Errorf("queue link %s for message %d: %w", l.Normalized, messageID, err)
	}
	return nil
}
