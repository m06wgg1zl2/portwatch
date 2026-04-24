// Package window provides a sliding time-window counter for tracking
// event frequency over a rolling duration (e.g. errors per minute).
package window

import (
	"sync"
	"time"
)

// entry holds a single timestamped event.
type entry struct {
	at time.Time
}

// Window is a thread-safe sliding-window counter keyed by an arbitrary string.
type Window struct {
	mu       sync.Mutex
	size     time.Duration
	buckets  map[string][]entry
}

// New creates a Window with the given rolling duration.
func New(size time.Duration) *Window {
	if size <= 0 {
		size = time.Minute
	}
	return &Window{
		size:    size,
		buckets: make(map[string][]entry),
	}
}

// Add records one event for the given key at the current time.
func (w *Window) Add(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buckets[key] = append(w.buckets[key], entry{at: time.Now()})
	w.evict(key)
}

// Count returns the number of events recorded for key within the window.
func (w *Window) Count(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(key)
	return len(w.buckets[key])
}

// Reset clears all events for the given key.
func (w *Window) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.buckets, key)
}

// Keys returns all keys that currently have at least one event in the window.
func (w *Window) Keys() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]string, 0, len(w.buckets))
	for k := range w.buckets {
		w.evict(k)
		if len(w.buckets[k]) > 0 {
			out = append(out, k)
		}
	}
	return out
}

// evict removes entries older than the window size. Caller must hold w.mu.
func (w *Window) evict(key string) {
	cutoff := time.Now().Add(-w.size)
	entries := w.buckets[key]
	i := 0
	for i < len(entries) && entries[i].at.Before(cutoff) {
		i++
	}
	if i > 0 {
		w.buckets[key] = entries[i:]
	}
	if len(w.buckets[key]) == 0 {
		delete(w.buckets, key)
	}
}
