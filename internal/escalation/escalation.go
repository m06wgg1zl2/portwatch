// Package escalation provides a configurable alert escalation policy
// that upgrades an alert's severity after repeated failures within a window.
package escalation

import (
	"sync"
	"time"
)

// Level represents a named severity tier.
type Level struct {
	Name      string
	Threshold int // number of failures before this level activates
}

// Config holds escalation policy settings.
type Config struct {
	Levels  []Level
	Window  time.Duration // rolling window over which failures are counted
}

// entry tracks failure timestamps for a single key.
type entry struct {
	mu        sync.Mutex
	timestamps []time.Time
}

// Escalator evaluates the current severity level for a given key.
type Escalator struct {
	cfg    Config
	mu     sync.Mutex
	entries map[string]*entry
}

// New returns a new Escalator with the given config.
// Levels should be ordered from lowest to highest threshold.
func New(cfg Config) *Escalator {
	return &Escalator{
		cfg:     cfg,
		entries: make(map[string]*entry),
	}
}

// Record registers a failure event for key at the given time.
func (e *Escalator) Record(key string, at time.Time) {
	e.mu.Lock()
	en, ok := e.entries[key]
	if !ok {
		en = &entry{}
		e.entries[key] = en
	}
	e.mu.Unlock()

	en.mu.Lock()
	defer en.mu.Unlock()
	en.timestamps = append(en.timestamps, at)
	e.evict(en, at)
}

// Level returns the highest escalation level name reached for key, or "" if
// no threshold has been crossed.
func (e *Escalator) Level(key string, now time.Time) string {
	e.mu.Lock()
	en, ok := e.entries[key]
	e.mu.Unlock()
	if !ok {
		return ""
	}

	en.mu.Lock()
	defer en.mu.Unlock()
	e.evict(en, now)

	count := len(en.timestamps)
	current := ""
	for _, lvl := range e.cfg.Levels {
		if count >= lvl.Threshold {
			current = lvl.Name
		}
	}
	return current
}

// Reset clears the failure history for key.
func (e *Escalator) Reset(key string) {
	e.mu.Lock()
	delete(e.entries, key)
	e.mu.Unlock()
}

// evict removes timestamps outside the rolling window. Must be called with en.mu held.
func (e *Escalator) evict(en *entry, now time.Time) {
	cutoff := now.Add(-e.cfg.Window)
	i := 0
	for i < len(en.timestamps) && en.timestamps[i].Before(cutoff) {
		i++
	}
	en.timestamps = en.timestamps[i:]
}
