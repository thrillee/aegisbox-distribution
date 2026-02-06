package sms

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	WindowSize        time.Duration
}

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tb.tokens = min(tb.capacity, tb.tokens+(elapsed.Seconds()*tb.refillRate))
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type RateLimitedSender struct {
	sender         Sender
	buckets        map[int32]*TokenBucket
	mu             sync.RWMutex
	fallbackStatus string
}

func NewRateLimitedSender(sender Sender, defaultConfig RateLimitConfig) *RateLimitedSender {
	return &RateLimitedSender{
		sender:         sender,
		buckets:        make(map[int32]*TokenBucket),
		fallbackStatus: "retry_pending",
	}
}

func (rl *RateLimitedSender) GetBucket(mnoID int32, config RateLimitConfig) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, ok := rl.buckets[mnoID]; ok {
		return bucket
	}

	bucket := NewTokenBucket(float64(config.BurstSize), config.RequestsPerSecond)
	rl.buckets[mnoID] = bucket
	return bucket
}

func (rl *RateLimitedSender) Send(ctx context.Context, msg database.GetMessageDetailsForSendingRow) (*mno.SubmitResult, error) {
	if msg.RoutedMnoID == nil {
		return &mno.SubmitResult{WasSubmitted: false, Error: nil}, nil
	}

	mnoID := *msg.RoutedMnoID
	config := RateLimitConfig{
		RequestsPerSecond: 50,
		BurstSize:         100,
	}

	bucket := rl.GetBucket(mnoID, config)

	if !bucket.Take() {
		slog.WarnContext(ctx, "Rate limit exceeded for MNO, message will retry",
			slog.Int("mno_id", int(mnoID)),
			slog.String("status", rl.fallbackStatus),
		)
		return &mno.SubmitResult{WasSubmitted: false, Error: nil}, nil
	}

	return rl.sender.Send(ctx, msg)
}

type RateLimitedProcessor struct {
	*Processor
	rateLimitedSender *RateLimitedSender
}

func NewRateLimitedProcessor(deps ProcessorDependencies, rateLimitConfig RateLimitConfig) *RateLimitedProcessor {
	processor := NewProcessor(deps)
	rateLimitedSender := NewRateLimitedSender(processor.sender, rateLimitConfig)
	processor.sender = rateLimitedSender

	return &RateLimitedProcessor{
		Processor:         processor,
		rateLimitedSender: rateLimitedSender,
	}
}
