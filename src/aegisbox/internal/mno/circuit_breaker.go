package mno

import (
	"log/slog"
	"sync"
	"time"
)

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures before opening circuit
	SuccessThreshold int           // Number of successes in half-open before closing
	Timeout          time.Duration // Time to wait before transitioning to half-open
	RequestTimeout   time.Duration // Timeout for individual requests
	VolumeThreshold  int           // Minimum requests before evaluating
	Logger           *slog.Logger
	MNOID            int32
}

type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	successCount    int
	requestCount    int
	lastFailure     time.Time
	lastSuccess     time.Time
	config          CircuitBreakerConfig
	lastStateChange time.Time
}

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 10 * time.Second
	}
	if config.VolumeThreshold == 0 {
		config.VolumeThreshold = 10
	}

	return &CircuitBreaker{
		state:  CircuitClosed,
		config: config,
	}
}

func (cb *CircuitBreaker) logStateTransition(from, to CircuitState) {
	if cb.config.Logger != nil {
		cb.config.Logger.Info("circuit_breaker_state_change",
			slog.Int("mno_id", int(cb.config.MNOID)),
			slog.String("from_state", from.String()),
			slog.String("to_state", to.String()),
			slog.Int("failure_count", cb.failureCount),
			slog.Int("success_count", cb.successCount),
		)
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if now.Sub(cb.lastStateChange) >= cb.config.Timeout {
			prevState := cb.state
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			cb.failureCount = 0
			cb.lastStateChange = now
			cb.logStateTransition(prevState, cb.state)
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.requestCount++
	cb.lastSuccess = time.Now()

	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			prevState := cb.state
			cb.state = CircuitClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.lastStateChange = time.Now()
			cb.logStateTransition(prevState, cb.state)
		}
	case CircuitClosed:
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.requestCount++
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitHalfOpen:
		prevState := cb.state
		cb.state = CircuitOpen
		cb.failureCount = 0
		cb.successCount = 0
		cb.lastStateChange = time.Now()
		cb.logStateTransition(prevState, cb.state)
	case CircuitClosed:
		cb.failureCount++
		if cb.failureCount >= cb.config.FailureThreshold && cb.requestCount >= cb.config.VolumeThreshold {
			prevState := cb.state
			cb.state = CircuitOpen
			cb.lastStateChange = time.Now()
			cb.logStateTransition(prevState, cb.state)
		}
	}
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	prevState := cb.state
	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.requestCount = 0
	cb.lastStateChange = time.Now()
	cb.logStateTransition(prevState, cb.state)
}

func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"request_count":     cb.requestCount,
		"last_failure":      cb.lastFailure,
		"last_success":      cb.lastSuccess,
		"last_state_change": cb.lastStateChange,
	}
}
