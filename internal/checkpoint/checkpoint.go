// Package checkpoint provides periodic state snapshotting with dirty-flag
// tracking. It records the last-saved value for a key and reports whether
// the current value differs, enabling callers to decide when a flush is
// warranted.
package checkpoint

import (
	"sync"
	"time"
)

// Entry holds the last-checkpointed value and the time it was saved.
type Entry struct {
	Value     string
	SavedAt   time.Time
	Flushes   int
}

// Checkpoint tracks dirty state per key.
type Checkpoint struct {
	mu      sync.Mutex
	records map[string]Entry
	clock   func() time.Time
}

// New returns a ready-to-use Checkpoint.
func New() *Checkpoint {
	return &Checkpoint{
		records: make(map[string]Entry),
		clock:   time.Now,
	}
}

// WithClock replaces the internal clock (useful in tests).
func WithClock(c *Checkpoint, fn func() time.Time) *Checkpoint {
	c.clock = fn
	return c
}

// IsDirty returns true when current differs from the last saved value for key,
// or when no checkpoint has been recorded yet.
func (c *Checkpoint) IsDirty(key, current string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.records[key]
	if !ok {
		return true
	}
	return e.Value != current
}

// Save records current as the clean value for key and increments the flush
// counter.
func (c *Checkpoint) Save(key, current string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.records[key]
	e.Value = current
	e.SavedAt = c.clock()
	e.Flushes++
	c.records[key] = e
}

// Get returns the Entry for key and a boolean indicating whether it exists.
func (c *Checkpoint) Get(key string) (Entry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.records[key]
	return e, ok
}

// Reset removes the checkpoint for key, causing the next IsDirty call to
// return true regardless of the current value.
func (c *Checkpoint) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, key)
}

// Keys returns all keys that have at least one saved checkpoint.
func (c *Checkpoint) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, len(c.records))
	for k := range c.records {
		out = append(out, k)
	}
	return out
}
