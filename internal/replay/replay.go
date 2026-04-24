// Package replay provides a bounded in-memory buffer of recent alerts that
// can be replayed to a new handler — useful for late-joining consumers or
// post-mortem inspection.
package replay

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/alert"
)

// Config holds configuration for the replay buffer.
type Config struct {
	// Capacity is the maximum number of alerts retained. Oldest are evicted.
	Capacity int `json:"capacity"`
	// TTL, when non-zero, causes alerts older than TTL to be skipped during replay.
	TTL time.Duration `json:"ttl"`
}

// Buffer is a thread-safe ring buffer of recent alerts.
type Buffer struct {
	mu       sync.Mutex
	entries  []entry
	capacity int
	ttl      time.Duration
}

type entry struct {
	at    time.Time
	alert alert.Alert
}

// New creates a Buffer with the given configuration.
// Capacity defaults to 64 if not set.
func New(cfg Config) *Buffer {
	cap := cfg.Capacity
	if cap <= 0 {
		cap = 64
	}
	return &Buffer{
		capacity: cap,
		ttl:      cfg.TTL,
	}
}

// Push records an alert into the buffer, evicting the oldest entry when full.
func (b *Buffer) Push(a alert.Alert) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.entries) >= b.capacity {
		b.entries = b.entries[1:]
	}
	b.entries = append(b.entries, entry{at: time.Now(), alert: a})
}

// Replay calls fn for each buffered alert in insertion order, skipping entries
// older than TTL (if TTL > 0). It holds no lock while calling fn.
func (b *Buffer) Replay(fn func(alert.Alert)) {
	b.mu.Lock()
	copy := make([]entry, len(b.entries))
	for i, e := range b.entries {
		copy[i] = e
	}
	b.mu.Unlock()

	cutoff := time.Time{}
	if b.ttl > 0 {
		cutoff = time.Now().Add(-b.ttl)
	}
	for _, e := range copy {
		if !cutoff.IsZero() && e.at.Before(cutoff) {
			continue
		}
		fn(e.alert)
	}
}

// Len returns the current number of buffered alerts.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

// Clear removes all buffered alerts.
func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = b.entries[:0]
}
