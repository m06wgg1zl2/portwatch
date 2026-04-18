package metrics

import (
	"sync"
	"time"
)

// Counter tracks named event counts and last-seen timestamps.
type Counter struct {
	mu      sync.RWMutex
	counts  map[string]int64
	lastSeen map[string]time.Time
}

// New returns an initialised Counter.
func New() *Counter {
	return &Counter{
		counts:   make(map[string]int64),
		lastSeen: make(map[string]time.Time),
	}
}

// Inc increments the counter for key by 1.
func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[key]++
	c.lastSeen[key] = time.Now()
}

// Get returns the current count and last-seen time for key.
func (c *Counter) Get(key string) (int64, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.counts[key], c.lastSeen[key]
}

// Reset zeroes the counter for key.
func (c *Counter) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.counts, key)
	delete(c.lastSeen, key)
}

// Snapshot returns a copy of all counts.
func (c *Counter) Snapshot() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int64, len(c.counts))
	for k, v := range c.counts {
		out[k] = v
	}
	return out
}
