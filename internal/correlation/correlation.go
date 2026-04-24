// Package correlation provides a lightweight correlation-ID tracker that
// groups related alerts by a shared key and exposes the current group size
// and first-seen timestamp. It is useful for understanding whether multiple
// port-state changes share a common root cause (e.g. a network partition).
package correlation

import (
	"sync"
	"time"
)

// Group holds metadata about a set of correlated events.
type Group struct {
	Key       string
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
}

// Tracker correlates events that share the same key within a rolling TTL
// window. Events older than TTL are silently evicted on the next operation.
type Tracker struct {
	mu     sync.Mutex
	groups map[string]*Group
	ttl    time.Duration
	clock  func() time.Time
}

// Config holds Tracker configuration.
type Config struct {
	// TTL is the duration after which an inactive group is evicted.
	// Defaults to 5 minutes if zero.
	TTL time.Duration
}

// New returns a Tracker configured with cfg.
func New(cfg Config) *Tracker {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Tracker{
		groups: make(map[string]*Group),
		ttl:    ttl,
		clock:  time.Now,
	}
}

// withClock replaces the internal clock; intended for testing only.
func (t *Tracker) withClock(fn func() time.Time) *Tracker {
	t.clock = fn
	return t
}

// Record registers an event under key and returns the updated Group.
// Stale groups (last seen beyond TTL) are evicted before recording.
func (t *Tracker) Record(key string) Group {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	t.evict(now)

	g, ok := t.groups[key]
	if !ok {
		g = &Group{
			Key:       key,
			FirstSeen: now,
		}
		t.groups[key] = g
	}
	g.Count++
	g.LastSeen = now
	return *g
}

// Get returns the Group for key and true if it exists, otherwise false.
func (t *Tracker) Get(key string) (Group, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.evict(t.clock())
	g, ok := t.groups[key]
	if !ok {
		return Group{}, false
	}
	return *g, true
}

// Reset removes the group for key, if present.
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.groups, key)
}

// Keys returns all active (non-evicted) group keys.
func (t *Tracker) Keys() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	t.evict(now)

	keys := make([]string, 0, len(t.groups))
	for k := range t.groups {
		keys = append(keys, k)
	}
	return keys
}

// evict removes groups whose LastSeen is older than TTL.
// Caller must hold t.mu.
func (t *Tracker) evict(now time.Time) {
	for k, g := range t.groups {
		if now.Sub(g.LastSeen) > t.ttl {
			delete(t.groups, k)
		}
	}
}
