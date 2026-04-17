package ratelimit

import (
	"sync"
	"time"
)

// Limiter prevents callbacks from firing too frequently for a given key.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
}

// New creates a Limiter with the given cooldown duration.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
	}
}

// Allow returns true if the key has not fired within the cooldown window.
// If allowed, it records the current time for that key.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if t, ok := l.last[key]; ok {
		if now.Sub(t) < l.cooldown {
			return false
		}
	}
	l.last[key] = now
	return true
}

// Reset clears the recorded time for a key, allowing it to fire immediately.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, key)
}

// LastFired returns the last time a key was allowed, and whether it exists.
func (l *Limiter) LastFired(key string) (time.Time, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	t, ok := l.last[key]
	return t, ok
}
