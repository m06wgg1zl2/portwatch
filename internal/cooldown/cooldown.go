// Package cooldown provides a per-key cooldown tracker that enforces a minimum
// quiet period between successive events. Unlike ratelimit which counts tokens,
// cooldown simply requires that a minimum duration elapses before the same key
// is allowed to fire again.
package cooldown

import (
	"sync"
	"time"
)

// Cooldown tracks the last-allowed time for each key.
type Cooldown struct {
	mu       sync.Mutex
	duration time.Duration
	last     map[string]time.Time
	now      func() time.Time // injectable for testing
}

// New creates a Cooldown with the given minimum quiet period.
// Panics if duration is zero or negative.
func New(d time.Duration) *Cooldown {
	if d <= 0 {
		panic("cooldown: duration must be positive")
	}
	return &Cooldown{
		duration: d,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true and records the current time if the key has not fired
// within the cooldown window. Returns false otherwise.
func (c *Cooldown) Allow(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if t, ok := c.last[key]; ok && now.Sub(t) < c.duration {
		return false
	}
	c.last[key] = now
	return true
}

// Remaining returns the time left in the cooldown window for key.
// Returns 0 if the key is not in cooldown.
func (c *Cooldown) Remaining(key string) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	t, ok := c.last[key]
	if !ok {
		return 0
	}
	elapsed := c.now().Sub(t)
	if elapsed >= c.duration {
		return 0
	}
	return c.duration - elapsed
}

// Reset clears the cooldown state for key, allowing it to fire immediately.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, key)
}

// Keys returns the list of keys currently tracked (in cooldown or recently fired).
func (c *Cooldown) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.last))
	for k := range c.last {
		keys = append(keys, k)
	}
	return keys
}
