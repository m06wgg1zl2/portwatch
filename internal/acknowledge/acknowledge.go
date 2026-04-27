package acknowledge

import (
	"sync"
	"time"
)

// State represents whether an alert has been acknowledged.
type State int

const (
	StatePending State = iota
	StateAcknowledged
)

// Entry holds acknowledgement metadata for a single key.
type Entry struct {
	AcknowledgedAt time.Time
	AcknowledgedBy string
	ExpiresAt      time.Time
}

// Store tracks acknowledgement state for alert keys.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	ttl     time.Duration
	clock   func() time.Time
}

// New returns a Store with the given TTL. Acknowledged keys expire after ttl.
func New(ttl time.Duration) *Store {
	return &Store{
		entries: make(map[string]Entry),
		ttl:     ttl,
		clock:   time.Now,
	}
}

// WithClock replaces the internal clock (for testing).
func WithClock(s *Store, fn func() time.Time) *Store {
	s.clock = fn
	return s
}

// Acknowledge marks key as acknowledged by the given actor until TTL elapses.
func (s *Store) Acknowledge(key, by string) {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = Entry{
		AcknowledgedAt: now,
		AcknowledgedBy: by,
		ExpiresAt:      now.Add(s.ttl),
	}
}

// IsAcknowledged returns true when key has an active, non-expired acknowledgement.
func (s *Store) IsAcknowledged(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok {
		return false
	}
	return s.clock().Before(e.ExpiresAt)
}

// Get returns the Entry for key and whether it exists and is active.
func (s *Store) Get(key string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok || !s.clock().Before(e.ExpiresAt) {
		return Entry{}, false
	}
	return e, true
}

// Unacknowledge removes any acknowledgement for key.
func (s *Store) Unacknowledge(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// Keys returns all currently active acknowledged keys.
func (s *Store) Keys() []string {
	now := s.clock()
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.entries))
	for k, e := range s.entries {
		if now.Before(e.ExpiresAt) {
			out = append(out, k)
		}
	}
	return out
}
