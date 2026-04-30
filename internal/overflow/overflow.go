// Package overflow tracks how many events were dropped due to capacity limits.
package overflow

import (
	"sync"
	"time"
)

// Entry records a single overflow event.
type Entry struct {
	Key       string
	Dropped   int64
	LastDropAt time.Time
}

// Tracker counts dropped events per key.
type Tracker struct {
	mu      sync.Mutex
	counts  map[string]*Entry
	maxKeys int
}

// Config holds Tracker configuration.
type Config struct {
	// MaxKeys is the maximum number of distinct keys tracked (0 = unlimited).
	MaxKeys int
}

// New returns a new Tracker.
func New(cfg Config) *Tracker {
	max := cfg.MaxKeys
	if max <= 0 {
		max = 0
	}
	return &Tracker{
		counts:  make(map[string]*Entry),
		maxKeys: max,
	}
}

// Record increments the drop count for key by n.
// If MaxKeys is set and the key is new, the record is silently ignored.
func (t *Tracker) Record(key string, n int64) {
	if n <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.counts[key]
	if !ok {
		if t.maxKeys > 0 && len(t.counts) >= t.maxKeys {
			return
		}
		e = &Entry{Key: key}
		t.counts[key] = e
	}
	e.Dropped += n
	e.LastDropAt = time.Now()
}

// Get returns the Entry for key, or zero value if not found.
func (t *Tracker) Get(key string) Entry {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e, ok := t.counts[key]; ok {
		return *e
	}
	return Entry{Key: key}
}

// Reset clears the drop count for key.
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.counts, key)
}

// All returns a snapshot of all tracked entries.
func (t *Tracker) All() []Entry {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Entry, 0, len(t.counts))
	for _, e := range t.counts {
		out = append(out, *e)
	}
	return out
}
