package throttle

import (
	"sync"
	"time"
)

// Throttle limits how frequently an action can fire per key using a token
// bucket with a fixed refill interval.
type Throttle struct {
	mu       sync.Mutex
	tokens   map[string]int
	lastFill map[string]time.Time
	max      int
	interval time.Duration
}

// Config holds throttle configuration.
type Config struct {
	// MaxBurst is the maximum number of events allowed per interval.
	MaxBurst int `json:"max_burst"`
	// Interval is the refill period as a duration string.
	Interval string `json:"interval"`
}

// New creates a Throttle from Config.
func New(cfg Config) (*Throttle, error) {
	d, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		return nil, err
	}
	max := cfg.MaxBurst
	if max <= 0 {
		max = 1
	}
	return &Throttle{
		tokens:   make(map[string]int),
		lastFill: make(map[string]time.Time),
		max:      max,
		interval: d,
	}, nil
}

// Allow returns true if the key has remaining tokens, consuming one.
func (t *Throttle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	last, seen := t.lastFill[key]
	if !seen {
		t.tokens[key] = t.max
		t.lastFill[key] = now
	} else if now.Sub(last) >= t.interval {
		t.tokens[key] = t.max
		t.lastFill[key] = now
	}

	if t.tokens[key] <= 0 {
		return false
	}
	t.tokens[key]--
	return true
}

// Remaining returns the current token count for a key without consuming.
func (t *Throttle) Remaining(key string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if v, ok := t.tokens[key]; ok {
		return v
	}
	return t.max
}

// Reset clears throttle state for a key.
func (t *Throttle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tokens, key)
	delete(t.lastFill, key)
}
