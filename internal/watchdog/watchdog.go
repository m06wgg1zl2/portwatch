package watchdog

import (
	"sync"
	"time"
)

// Status represents the current watchdog health state.
type Status int

const (
	StatusHealthy Status = iota
	StatusStale
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusStale:
		return "stale"
	default:
		return "unknown"
	}
}

// Watchdog tracks periodic heartbeats and reports staleness.
type Watchdog struct {
	mu       sync.Mutex
	timeout  time.Duration
	lastBeat time.Time
	now      func() time.Time
}

// New creates a Watchdog with the given staleness timeout.
func New(timeout time.Duration) *Watchdog {
	return &Watchdog{
		timeout: timeout,
		now:     time.Now,
	}
}

// Beat records a heartbeat at the current time.
func (w *Watchdog) Beat() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastBeat = w.now()
}

// Status returns the current health status.
func (w *Watchdog) Status() Status {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.lastBeat.IsZero() || w.now().Sub(w.lastBeat) > w.timeout {
		return StatusStale
	}
	return StatusHealthy
}

// LastBeat returns the time of the most recent heartbeat.
func (w *Watchdog) LastBeat() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastBeat
}

// Reset clears the last beat, forcing stale status.
func (w *Watchdog) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastBeat = time.Time{}
}
