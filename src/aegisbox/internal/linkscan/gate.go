package linkscan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
)

// Phase 2: scanning of real customer traffic, per customer (link_scan_settings). Off unless an
// admin sets a mode for a customer.
const (
	ModeOff     = "off"
	ModeMonitor = "monitor" // scan after sending; never delays or blocks
	ModeHold    = "hold"    // wait for verdicts; a bad one goes to admin review
	ModeBlock   = "block"   // wait for verdicts; a bad one rejects the message
)

// Hold statuses in message_link_holds.
const (
	HoldWaiting  = "waiting"
	HoldReview   = "review"
	HoldReleased = "released"
	HoldRejected = "rejected"
)

// Message processing statuses used by the gate.
const (
	StatusQueuedForSend = "queued_for_send"
	StatusHeld          = "held_link_scan"
	StatusBlocked       = "blocked_link_scan"
)

// reviewExpiry is how long a message waits for an admin in review before it is rejected.
const reviewExpiry = 24 * time.Hour

// Settings is one customer's Phase 2 configuration.
type Settings struct {
	Mode            string
	HoldSeconds     int32
	FailOpen        bool
	BlockSuspicious bool
	ChargeBlocked   bool
}

// Holds reports whether messages wait for verdicts before sending.
func (s Settings) Holds() bool { return s.Mode == ModeHold || s.Mode == ModeBlock }

// SettingsCache keeps the customers with a mode other than off in memory, refreshed every 30 s,
// so the send path costs no query for the others.
type SettingsCache struct {
	load    func(context.Context) ([]database.ListActiveLinkScanSettingsRow, error)
	refresh time.Duration
	now     func() time.Time

	mu       sync.Mutex
	bySP     map[int32]Settings
	loadedAt time.Time
}

func NewSettingsCache(load func(context.Context) ([]database.ListActiveLinkScanSettingsRow, error), refresh time.Duration) *SettingsCache {
	if refresh <= 0 {
		refresh = 30 * time.Second
	}
	return &SettingsCache{load: load, refresh: refresh, now: time.Now}
}

// Get returns a customer's settings; ok is false when scanning of its real traffic is off.
func (c *SettingsCache) Get(ctx context.Context, spID int32) (Settings, bool) {
	if c == nil {
		return Settings{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bySP == nil || c.now().Sub(c.loadedAt) >= c.refresh {
		rows, err := c.load(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Link scan: could not load customer scan settings", slog.Any("error", err))
			if c.bySP == nil {
				return Settings{}, false
			}
		} else {
			m := make(map[int32]Settings, len(rows))
			for _, r := range rows {
				m[r.ServiceProviderID] = Settings{
					Mode: r.Mode, HoldSeconds: r.HoldSeconds, FailOpen: r.FailOpen,
					BlockSuspicious: r.BlockSuspicious, ChargeBlocked: r.ChargeBlocked,
				}
			}
			c.bySP, c.loadedAt = m, c.now()
		}
	}
	s, ok := c.bySP[spID]
	return s, ok && s.Mode != ModeOff
}

// Rejecter rejects a held message: no send, final status rejected, REJECTD DLR to the customer,
// and the debit reversed when refund is set. Implemented in the sms package.
type Rejecter interface {
	RejectForLinks(ctx context.Context, messageID int64, reason string, refund bool) error
}

// GateQuerier is the part of database.Querier the gate uses.
type GateQuerier interface {
	GetMessageTextForLinkScan(ctx context.Context, id int64) (string, error)
	GetMessageLinkHold(ctx context.Context, messageID int64) (database.MessageLinkHold, error)
	InsertMessageLinkHold(ctx context.Context, arg database.InsertMessageLinkHoldParams) (int64, error)
	SetMessageProcessingStatus(ctx context.Context, arg database.SetMessageProcessingStatusParams) (int64, error)
	ListOpenLinkHolds(ctx context.Context, batchSize int32) ([]database.MessageLinkHold, error)
	ListHoldJobVerdicts(ctx context.Context, messageID int64) ([]database.ListHoldJobVerdictsRow, error)
	DecideMessageLinkHold(ctx context.Context, arg database.DecideMessageLinkHoldParams) (int64, error)
}

// Gate holds real messages of customers in hold or block mode until their links have verdicts.
type Gate struct {
	q        GateQuerier
	settings *SettingsCache
	enq      *Enqueuer
	rejecter Rejecter
	now      func() time.Time
}

func NewGate(q GateQuerier, settings *SettingsCache, enq *Enqueuer, rejecter Rejecter) *Gate {
	return &Gate{q: q, settings: settings, enq: enq, rejecter: rejecter, now: time.Now}
}

// BeforeSend is called for a message about to be sent. It returns true when the message must not
// be sent now: its links were queued and it waits for their verdicts (the gate worker sends it on
// or rejects it). Customers whose mode is not hold or block are never held.
func (g *Gate) BeforeSend(ctx context.Context, messageID int64, spID int32) (bool, error) {
	if g == nil {
		return false, nil
	}
	s, ok := g.settings.Get(ctx, spID)
	if !ok || !s.Holds() {
		return false, nil
	}
	hold, err := g.q.GetMessageLinkHold(ctx, messageID)
	switch {
	case err == nil:
		if hold.Status == HoldReleased {
			return false, nil
		}
		// Waiting, in review or rejected: never send from here. Put it back in the held state in
		// case it was queued again by hand.
		g.setStatus(ctx, messageID, StatusQueuedForSend, StatusHeld)
		return true, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return s.held(), fmt.Errorf("read link hold: %w", err)
	}

	text, err := g.q.GetMessageTextForLinkScan(ctx, messageID)
	if err != nil {
		return s.held(), fmt.Errorf("read message text: %w", err)
	}
	n, err := g.enq.HoldLinks(ctx, messageID, spID, text)
	if err != nil {
		return s.held(), err
	}
	if n == 0 {
		return false, nil // no links that get a verdict, or an OTP: send at once
	}
	if _, err := g.q.InsertMessageLinkHold(ctx, database.InsertMessageLinkHoldParams{
		MessageID:         messageID,
		ServiceProviderID: spID,
		Mode:              s.Mode,
		HoldSeconds:       s.HoldSeconds,
		FailOpen:          s.FailOpen,
		BlockSuspicious:   s.BlockSuspicious,
		ChargeBlocked:     s.ChargeBlocked,
	}); err != nil {
		return s.held(), fmt.Errorf("create link hold: %w", err)
	}
	g.setStatus(ctx, messageID, StatusQueuedForSend, StatusHeld)
	slog.InfoContext(ctx, "Message held for link verdicts", slog.Int64("message_id", messageID),
		slog.String("mode", s.Mode), slog.Int("links", n), slog.Int("hold_seconds", int(s.HoldSeconds)))
	return true, nil
}

// held is what to do with a message when the gate itself fails: send it when the customer is set
// to fail open, otherwise keep it queued and try again on the next pass.
func (s Settings) held() bool { return !s.FailOpen }

func (g *Gate) setStatus(ctx context.Context, id int64, from, to string) bool {
	n, err := g.q.SetMessageProcessingStatus(ctx, database.SetMessageProcessingStatusParams{ID: id, FromStatus: from, ToStatus: to})
	if err != nil {
		slog.ErrorContext(ctx, "Link hold: could not change message status", slog.Int64("message_id", id),
			slog.String("to", to), slog.Any("error", err))
		return false
	}
	return n == 1
}

// Start runs the gate worker: it decides held messages every interval.
func (g *Gate) Start(ctx context.Context, wg *sync.WaitGroup, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	slog.Info("Link hold gate started (real traffic of customers in hold or block mode)")
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			g.RunOnce(ctx)
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

// RunOnce decides the open holds whose verdicts are in or whose time is up.
func (g *Gate) RunOnce(ctx context.Context) {
	holds, err := g.q.ListOpenLinkHolds(ctx, 200)
	if err != nil {
		slog.ErrorContext(ctx, "Link hold gate: list holds failed", slog.Any("error", err))
		return
	}
	for _, h := range holds {
		g.decide(ctx, h)
	}
}

var verdictRank = map[string]int{VerdictSafe: 1, VerdictSuspicious: 2, VerdictMalicious: 3}

func (g *Gate) decide(ctx context.Context, h database.MessageLinkHold) {
	if h.Status == HoldReview {
		// ListOpenLinkHolds returns reviews only once they are past reviewExpiry.
		g.reject(ctx, h, HoldReview, "review_expired", nil, "",
			fmt.Sprintf("held for review and not decided within %s", reviewExpiry))
		return
	}

	jobs, err := g.q.ListHoldJobVerdicts(ctx, h.MessageID)
	if err != nil {
		slog.ErrorContext(ctx, "Link hold gate: read verdicts failed", slog.Int64("message_id", h.MessageID), slog.Any("error", err))
		return
	}
	worst, pending, noVerdict := "", 0, 0
	for _, j := range jobs {
		if j.OverrideVerdict != nil {
			j.Verdict = j.OverrideVerdict
		}
		switch {
		case j.Status == "queued" || j.Status == "processing":
			pending++
		case j.Verdict == nil:
			noVerdict++ // failed after its attempts
		case verdictRank[*j.Verdict] > verdictRank[worst]:
			worst = *j.Verdict
		}
	}
	var worstPtr *string
	if worst != "" {
		worstPtr = &worst
	}

	bad := worst == VerdictMalicious || (worst == VerdictSuspicious && h.BlockSuspicious)
	timedOut := g.now().After(h.HoldUntil.Time)
	switch {
	case bad:
		// One bad link is enough; no need to wait for the others.
		g.stop(ctx, h, worst, worstPtr, fmt.Sprintf("a link in the message was judged %s", worst))
	case pending == 0 && noVerdict == 0:
		g.release(ctx, h, HoldWaiting, "clean", worstPtr, "")
	case pending == 0 || timedOut:
		why := "no verdict in time"
		if pending == 0 {
			why = "a link could not be checked"
		}
		if h.FailOpen {
			g.release(ctx, h, HoldWaiting, "timeout_released", worstPtr, "")
		} else {
			g.stop(ctx, h, "timeout", worstPtr, why)
		}
	}
}

// stop rejects the message (block mode) or sends it to admin review (hold mode).
func (g *Gate) stop(ctx context.Context, h database.MessageLinkHold, decision string, worst *string, why string) {
	if h.Mode == ModeHold {
		n, err := g.q.DecideMessageLinkHold(ctx, database.DecideMessageLinkHoldParams{
			MessageID: h.MessageID, FromStatus: HoldWaiting, ToStatus: HoldReview, Decision: decision, WorstVerdict: worst,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Link hold: could not move to review", slog.Any("error", err))
			return
		}
		if n == 1 {
			slog.WarnContext(ctx, "Message held for admin review", slog.Int64("message_id", h.MessageID), slog.String("reason", why))
		}
		return
	}
	if decision == "timeout" {
		decision = "timeout_rejected"
	}
	g.reject(ctx, h, HoldWaiting, decision, worst, "", why)
}

func (g *Gate) release(ctx context.Context, h database.MessageLinkHold, from, decision string, worst *string, by string) bool {
	var byPtr *string
	if by != "" {
		byPtr = &by
	}
	n, err := g.q.DecideMessageLinkHold(ctx, database.DecideMessageLinkHoldParams{
		MessageID: h.MessageID, FromStatus: from, ToStatus: HoldReleased, Decision: decision, WorstVerdict: worst, DecidedBy: byPtr,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Link hold: could not release", slog.Any("error", err))
		return false
	}
	if n != 1 {
		return false // decided by someone else meanwhile
	}
	g.setStatus(ctx, h.MessageID, StatusHeld, StatusQueuedForSend)
	slog.InfoContext(ctx, "Message released after link check", slog.Int64("message_id", h.MessageID), slog.String("decision", decision))
	return true
}

func (g *Gate) reject(ctx context.Context, h database.MessageLinkHold, from, decision string, worst *string, by, why string) bool {
	var byPtr *string
	if by != "" {
		byPtr = &by
	}
	n, err := g.q.DecideMessageLinkHold(ctx, database.DecideMessageLinkHoldParams{
		MessageID: h.MessageID, FromStatus: from, ToStatus: HoldRejected, Decision: decision, WorstVerdict: worst, DecidedBy: byPtr,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Link hold: could not reject", slog.Any("error", err))
		return false
	}
	if n != 1 {
		return false
	}
	if err := g.rejecter.RejectForLinks(ctx, h.MessageID, "Blocked by link check: "+why, !h.ChargeBlocked); err != nil {
		slog.ErrorContext(ctx, "Link hold: reject failed", slog.Int64("message_id", h.MessageID), slog.Any("error", err))
	}
	slog.WarnContext(ctx, "Message rejected after link check", slog.Int64("message_id", h.MessageID), slog.String("decision", decision))
	return true
}

// ErrNotInReview is returned by AdminDecide when the message is not waiting for an admin.
var ErrNotInReview = errors.New("message is not held for review")

// AdminDecide releases or rejects a held message on an admin's request (portal review queue).
func (g *Gate) AdminDecide(ctx context.Context, messageID int64, release bool, by string) error {
	h, err := g.q.GetMessageLinkHold(ctx, messageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotInReview
		}
		return err
	}
	if h.Status != HoldReview && h.Status != HoldWaiting {
		return ErrNotInReview
	}
	if release {
		if !g.release(ctx, h, h.Status, "admin_released", h.WorstVerdict, by) {
			return ErrNotInReview
		}
		return nil
	}
	if !g.reject(ctx, h, h.Status, "admin_rejected", h.WorstVerdict, by, "rejected by an admin") {
		return ErrNotInReview
	}
	return nil
}
