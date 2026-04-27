// Package holddown implements a hold-down timer that requires a state to
// remain stable for a minimum duration before confirming the transition.
// This prevents flapping from triggering repeated callbacks.
package holddown

import (
	"sync"
	"time"
)

// Config holds configuration for the hold-down timer.
type Config struct {
	// Window is the minimum duration a state must be stable before confirmation.
	Window time.Duration `json:"window"`
}

type entry struct {
	state     bool
	pendingAt time.Time
}

// HoldDown tracks per-key state stability over a configurable window.
type HoldDown struct {
	mu      sync.Mutex
	window  time.Duration
	pending map[string]entry
	clock   func() time.Time
}

// New creates a HoldDown with the given config.
func New(cfg Config) *HoldDown {
	w := cfg.Window
	if w <= 0 {
		w = 5 * time.Second
	}
	return &HoldDown{
		window:  w,
		pending: make(map[string]entry),
		clock:   time.Now,
	}
}

// WithClock replaces the internal clock; intended for testing.
func WithClock(h *HoldDown, fn func() time.Time) *HoldDown {
	h.clock = fn
	return h
}

// Observe records the current state for key. It returns true only when the
// state has been continuously stable for at least the configured window.
func (h *HoldDown) Observe(key string, state bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.clock()
	e, ok := h.pending[key]

	if !ok || e.state != state {
		// New key or state flipped — restart the timer.
		h.pending[key] = entry{state: state, pendingAt: now}
		return false
	}

	if now.Sub(e.pendingAt) >= h.window {
		return true
	}
	return false
}

// Reset clears the hold-down timer for key.
func (h *HoldDown) Reset(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pending, key)
}

// Window returns the configured stability window.
func (h *HoldDown) Window() time.Duration {
	return h.window
}
