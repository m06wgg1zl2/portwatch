// Package deadletter provides a dead-letter queue for alerts that fail
// all delivery attempts, retaining them for inspection or replay.
package deadletter

import (
	"sync"
	"time"

	"portwatch/internal/alert"
)

// Entry holds a failed alert alongside metadata about why it was rejected.
type Entry struct {
	Alert     alert.Alert
	Reason    string
	FailedAt  time.Time
	Attempts  int
}

// Queue is a bounded in-memory dead-letter store.
type Queue struct {
	mu      sync.Mutex
	entries []Entry
	maxSize int
}

// Config controls queue behaviour.
type Config struct {
	// MaxSize is the maximum number of entries retained.
	// Oldest entries are evicted when the limit is reached. Default: 100.
	MaxSize int
}

// New creates a Queue with the provided configuration.
func New(cfg Config) *Queue {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100
	}
	return &Queue{maxSize: cfg.MaxSize}
}

// Push adds a failed alert to the queue.
func (q *Queue) Push(a alert.Alert, reason string, attempts int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.entries) >= q.maxSize {
		// Evict the oldest entry.
		q.entries = q.entries[1:]
	}
	q.entries = append(q.entries, Entry{
		Alert:    a,
		Reason:   reason,
		FailedAt: time.Now(),
		Attempts: attempts,
	})
}

// All returns a copy of all queued entries, oldest first.
func (q *Queue) All() []Entry {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]Entry, len(q.entries))
	copy(out, q.entries)
	return out
}

// Len returns the current number of entries.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

// Clear removes all entries from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.entries = q.entries[:0]
}
