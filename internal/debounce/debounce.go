// Package debounce delays state-change notifications until a port has been
// in its new state for a minimum confirmation window, reducing false alerts
// caused by transient blips.
package debounce

import (
	"sync"
	"time"
)

// StateChange represents a confirmed state transition.
type StateChange struct {
	Key      string
	NewState bool // true = open, false = closed
	At       time.Time
}

// Debouncer holds per-key pending transitions and confirms them after Window.
type Debouncer struct {
	Window  time.Duration
	mu      sync.Mutex
	pending map[string]*entry
}

type entry struct {
	state bool
	since time.Time
}

// New creates a Debouncer with the given confirmation window.
func New(window time.Duration) *Debouncer {
	return &Debouncer{
		Window:  window,
		pending: make(map[string]*entry),
	}
}

// Observe records an observed state for key. It returns a confirmed
// StateChange and true only when the state has been stable for Window.
func (d *Debouncer) Observe(key string, state bool, now time.Time) (StateChange, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	e, exists := d.pending[key]
	if !exists || e.state != state {
		d.pending[key] = &entry{state: state, since: now}
		return StateChange{}, false
	}

	if now.Sub(e.since) >= d.Window {
		delete(d.pending, key)
		return StateChange{Key: key, NewState: state, At: now}, true
	}

	return StateChange{}, false
}

// Reset clears any pending transition for key.
func (d *Debouncer) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.pending, key)
}
