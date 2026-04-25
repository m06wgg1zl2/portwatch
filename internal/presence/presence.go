// Package presence tracks whether a monitored target has been seen
// within a configurable TTL window, providing a simple liveness signal.
package presence

import (
	"sync"
	"time"
)

// Config holds tunables for the Tracker.
type Config struct {
	// TTL is how long a target is considered present after its last ping.
	TTL time.Duration `json:"ttl"`
}

func (c *Config) defaults() {
	if c.TTL <= 0 {
		c.TTL = 30 * time.Second
	}
}

// Status describes whether a target is currently present.
type Status int

const (
	Absent  Status = iota // never seen or TTL elapsed
	Present               // seen within TTL
)

func (s Status) String() string {
	if s == Present {
		return "present"
	}
	return "absent"
}

// Tracker records the last-seen time for arbitrary string keys and
// reports whether each key is still within its TTL.
type Tracker struct {
	mu      sync.Mutex
	cfg     Config
	lastSeen map[string]time.Time
	now     func() time.Time
}

// New creates a Tracker with the supplied Config.
func New(cfg Config) *Tracker {
	cfg.defaults()
	return &Tracker{
		cfg:      cfg,
		lastSeen: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Ping records that key was observed right now.
func (t *Tracker) Ping(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastSeen[key] = t.now()
}

// Check returns the current Status for key.
func (t *Tracker) Check(key string) Status {
	t.mu.Lock()
	defer t.mu.Unlock()
	ts, ok := t.lastSeen[key]
	if !ok {
		return Absent
	}
	if t.now().Sub(ts) > t.cfg.TTL {
		return Absent
	}
	return Present
}

// Forget removes key from the tracker.
func (t *Tracker) Forget(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastSeen, key)
}

// Keys returns all keys currently tracked (regardless of status).
func (t *Tracker) Keys() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, 0, len(t.lastSeen))
	for k := range t.lastSeen {
		out = append(out, k)
	}
	return out
}
