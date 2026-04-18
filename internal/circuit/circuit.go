package circuit

import (
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota // normal operation
	StateOpen                // blocking calls
	StateHalfOpen            // testing recovery
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	}
	return "unknown"
}

// Config holds circuit breaker settings.
type Config struct {
	FailureThreshold int
	SuccessThreshold int
	OpenTimeout      time.Duration
}

// Breaker is a simple circuit breaker keyed by a string.
type Breaker struct {
	mu       sync.Mutex
	cfg      Config
	failures int
	successes int
	state    State
	openedAt time.Time
}

// New returns a Breaker with the given config, applying defaults.
func New(cfg Config) *Breaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 1
	}
	if cfg.OpenTimeout <= 0 {
		cfg.OpenTimeout = 10 * time.Second
	}
	return &Breaker{cfg: cfg}
}

// Allow reports whether a call should proceed.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.openedAt) >= b.cfg.OpenTimeout {
			b.state = StateHalfOpen
			b.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful call.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	if b.state == StateHalfOpen {
		b.successes++
		if b.successes >= b.cfg.SuccessThreshold {
			b.state = StateClosed
		}
	}
}

// RecordFailure records a failed call.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.state == StateHalfOpen || b.failures >= b.cfg.FailureThreshold {
		b.state = StateOpen
		b.openedAt = time.Now()
		b.failures = 0
	}
}

// State returns the current circuit state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
