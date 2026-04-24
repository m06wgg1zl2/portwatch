// Package limiter provides a concurrent-request limiter that caps the number
// of in-flight operations for a given key using a semaphore approach.
package limiter

import (
	"errors"
	"sync"
	"time"
)

// ErrLimitExceeded is returned when the concurrency limit has been reached.
var ErrLimitExceeded = errors.New("limiter: concurrency limit exceeded")

// Config holds configuration for the Limiter.
type Config struct {
	// Max is the maximum number of concurrent operations allowed per key.
	Max int `json:"max"`
	// Timeout is how long Acquire will wait before returning ErrLimitExceeded.
	// Zero means non-blocking.
	Timeout time.Duration `json:"timeout_ms"`
}

type entry struct {
	sem chan struct{}
}

// Limiter tracks concurrent in-flight operations per key.
type Limiter struct {
	mu      sync.Mutex
	max     int
	timeout time.Duration
	keys    map[string]*entry
}

// New creates a Limiter from cfg. Max defaults to 1 if not positive.
func New(cfg Config) *Limiter {
	max := cfg.Max
	if max <= 0 {
		max = 1
	}
	return &Limiter{
		max:     max,
		timeout: cfg.Timeout,
		keys:    make(map[string]*entry),
	}
}

func (l *Limiter) entryFor(key string) *entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.keys[key]
	if !ok {
		e = &entry{sem: make(chan struct{}, l.max)}
		l.keys[key] = e
	}
	return e
}

// Acquire attempts to acquire a slot for key. Returns ErrLimitExceeded if the
// limit is reached and the optional timeout elapses.
func (l *Limiter) Acquire(key string) error {
	e := l.entryFor(key)
	if l.timeout == 0 {
		select {
		case e.sem <- struct{}{}:
			return nil
		default:
			return ErrLimitExceeded
		}
	}
	select {
	case e.sem <- struct{}{}:
		return nil
	case <-time.After(l.timeout):
		return ErrLimitExceeded
	}
}

// Release frees a previously acquired slot for key. It is a no-op if no slot
// is held.
func (l *Limiter) Release(key string) {
	e := l.entryFor(key)
	select {
	case <-e.sem:
	default:
	}
}

// InFlight returns the number of currently held slots for key.
func (l *Limiter) InFlight(key string) int {
	e := l.entryFor(key)
	return len(e.sem)
}
