// Package dedup provides alert deduplication based on a fingerprint and TTL window.
// Duplicate alerts sharing the same key are suppressed until the TTL expires.
package dedup

import (
	"sync"
	"time"
)

// Clock allows time to be injected for testing.
type Clock func() time.Time

// entry holds the first-seen timestamp for a dedup key.
type entry struct {
	firstSeen time.Time
	count     int
}

// Deduplicator suppresses repeated events within a TTL window.
type Deduplicator struct {
	mu      sync.Mutex
	entries map[string]*entry
	ttl     time.Duration
	clock   Clock
}

// Config holds configuration for the Deduplicator.
type Config struct {
	// TTL is the duration during which duplicate keys are suppressed.
	TTL time.Duration `json:"ttl"`
}

// New creates a Deduplicator with the given TTL.
// A zero or negative TTL disables deduplication (all events pass through).
func New(cfg Config) *Deduplicator {
	return &Deduplicator{
		entries: make(map[string]*entry),
		ttl:     cfg.TTL,
		clock:   time.Now,
	}
}

// WithClock replaces the internal clock; intended for testing.
func WithClock(d *Deduplicator, clk Clock) *Deduplicator {
	d.clock = clk
	return d
}

// IsDuplicate returns true if key has been seen within the TTL window.
// The first call for a key always returns false and starts the window.
func (d *Deduplicator) IsDuplicate(key string) bool {
	if d.ttl <= 0 {
		return false
	}
	now := d.clock()
	d.mu.Lock()
	defer d.mu.Unlock()

	if e, ok := d.entries[key]; ok {
		if now.Sub(e.firstSeen) < d.ttl {
			e.count++
			return true
		}
		// TTL expired — reset the window.
		e.firstSeen = now
		e.count = 1
		return false
	}
	d.entries[key] = &entry{firstSeen: now, count: 1}
	return false
}

// Count returns how many times key has been seen in the current window.
// Returns 0 if the key is unknown or its window has expired.
func (d *Deduplicator) Count(key string) int {
	if d.ttl <= 0 {
		return 0
	}
	now := d.clock()
	d.mu.Lock()
	defer d.mu.Unlock()
	if e, ok := d.entries[key]; ok && now.Sub(e.firstSeen) < d.ttl {
		return e.count
	}
	return 0
}

// Reset removes the dedup entry for key, allowing the next event to pass through.
func (d *Deduplicator) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, key)
}
