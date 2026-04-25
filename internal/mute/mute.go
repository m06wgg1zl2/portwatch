// Package mute provides a time-bounded suppression registry that silences
// alerts for a given key until an explicit expiry or manual unmute.
package mute

import (
	"sync"
	"time"
)

// Clock allows injecting a time source for testing.
type Clock func() time.Time

// Entry holds the mute expiry for a single key.
type Entry struct {
	Until   time.Time
	Reason  string
}

// Mute is a concurrency-safe registry of muted keys.
type Mute struct {
	mu      sync.RWMutex
	entries map[string]Entry
	clock   Clock
}

// New returns a Mute using the real wall clock.
func New() *Mute {
	return WithClock(time.Now)
}

// WithClock returns a Mute using the provided clock function.
func WithClock(clock Clock) *Mute {
	return &Mute{
		entries: make(map[string]Entry),
		clock:   clock,
	}
}

// Silence mutes key until the given deadline with an optional reason.
func (m *Mute) Silence(key string, until time.Time, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[key] = Entry{Until: until, Reason: reason}
}

// Unmute removes a mute entry for key immediately.
func (m *Mute) Unmute(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, key)
}

// IsMuted reports whether key is currently muted.
func (m *Mute) IsMuted(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[key]
	if !ok {
		return false
	}
	if m.clock().After(e.Until) {
		return false
	}
	return true
}

// Get returns the Entry for key and whether it exists and is still active.
func (m *Mute) Get(key string) (Entry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[key]
	if !ok {
		return Entry{}, false
	}
	if m.clock().After(e.Until) {
		return Entry{}, false
	}
	return e, true
}

// Keys returns all currently active muted keys.
func (m *Mute) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := m.clock()
	out := make([]string, 0, len(m.entries))
	for k, e := range m.entries {
		if !now.After(e.Until) {
			out = append(out, k)
		}
	}
	return out
}
