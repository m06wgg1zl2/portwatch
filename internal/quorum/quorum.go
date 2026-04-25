// Package quorum implements a multi-probe consensus check: a port state
// change is only accepted when a configurable number of independent
// observations agree on the new state.
package quorum

import (
	"fmt"
	"sync"
)

// Config holds quorum parameters.
type Config struct {
	// Required is the number of agreeing observations needed to confirm a
	// state change. Must be >= 1.
	Required int `json:"required"`
	// Window is the maximum number of recent observations retained per key.
	// Older entries beyond this limit are evicted.
	Window int `json:"window"`
}

func (c *Config) defaults() {
	if c.Required <= 0 {
		c.Required = 3
	}
	if c.Window <= 0 {
		c.Window = 10
	}
}

// Quorum tracks observations per key and reports whether consensus has been
// reached for a given value.
type Quorum struct {
	mu  sync.Mutex
	cfg Config
	obs map[string][]string
}

// New creates a Quorum with the supplied config.
func New(cfg Config) *Quorum {
	cfg.defaults()
	return &Quorum{
		cfg: cfg,
		obs: make(map[string][]string),
	}
}

// Observe records a new observation for key with the given value and returns
// true when the required number of consecutive matching observations is met.
func (q *Quorum) Observe(key, value string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	buf := append(q.obs[key], value)
	if len(buf) > q.cfg.Window {
		buf = buf[len(buf)-q.cfg.Window:]
	}
	q.obs[key] = buf

	if len(buf) < q.cfg.Required {
		return false
	}
	tail := buf[len(buf)-q.cfg.Required:]
	for _, v := range tail {
		if v != value {
			return false
		}
	}
	return true
}

// Reset clears all observations for key.
func (q *Quorum) Reset(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.obs, key)
}

// Count returns the number of stored observations for key.
func (q *Quorum) Count(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.obs[key])
}

// String returns a human-readable summary.
func (q *Quorum) String() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	return fmt.Sprintf("quorum(required=%d window=%d keys=%d)",
		q.cfg.Required, q.cfg.Window, len(q.obs))
}
