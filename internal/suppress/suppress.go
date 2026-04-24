// Package suppress provides a mechanism to suppress repeated notifications
// for the same key until a configurable quiet period has elapsed.
package suppress

import (
	"sync"
	"time"
)

// Suppressor tracks suppression state per key.
type Suppressor struct {
	mu      sync.Mutex
	quiet   time.Duration
	entries map[string]time.Time
	now     func() time.Time
}

// Option configures a Suppressor.
type Option func(*Suppressor)

// WithClock overrides the clock used for time comparisons (useful in tests).
func WithClock(fn func() time.Time) Option {
	return func(s *Suppressor) { s.now = fn }
}

// New creates a Suppressor with the given quiet window duration.
// While a key is within its quiet window, IsSuppressed returns true.
func New(quiet time.Duration, opts ...Option) *Suppressor {
	s := &Suppressor{
		quiet:   quiet,
		entries: make(map[string]time.Time),
		now:     time.Now,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// IsSuppressed reports whether the key is currently suppressed.
// If the key is not suppressed, it is marked as active and the quiet
// window begins. Subsequent calls within the window return true.
func (s *Suppressor) IsSuppressed(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if t, ok := s.entries[key]; ok && now.Before(t) {
		return true
	}
	s.entries[key] = now.Add(s.quiet)
	return false
}

// Reset clears the suppression record for key, allowing the next call
// to IsSuppressed to pass through immediately.
func (s *Suppressor) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// Remaining returns how long the key remains suppressed.
// Returns zero if the key is not currently suppressed.
func (s *Suppressor) Remaining(key string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if t, ok := s.entries[key]; ok && now.Before(t) {
		return t.Sub(now)
	}
	return 0
}

// Keys returns all keys that are currently within their quiet window.
func (s *Suppressor) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	out := make([]string, 0, len(s.entries))
	for k, t := range s.entries {
		if now.Before(t) {
			out = append(out, k)
		}
	}
	return out
}
