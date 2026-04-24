// Package batch provides alert batching: it accumulates alerts over a
// configurable window and flushes them as a slice when the window closes
// or the buffer reaches its maximum capacity.
package batch

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/alert"
)

// Config holds parameters for the Batcher.
type Config struct {
	// Window is the maximum duration to hold alerts before flushing.
	Window time.Duration
	// MaxSize is the maximum number of alerts before an early flush.
	MaxSize int
}

// Batcher accumulates alerts and flushes them periodically.
type Batcher struct {
	mu      sync.Mutex
	cfg     Config
	buf     []*alert.Alert
	flushFn func([]*alert.Alert)
	timer   *time.Timer
}

// New creates a Batcher that calls flushFn with each completed batch.
// flushFn is invoked from an internal goroutine; callers must ensure it
// is safe for concurrent use.
func New(cfg Config, flushFn func([]*alert.Alert)) *Batcher {
	if cfg.Window <= 0 {
		cfg.Window = 5 * time.Second
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 50
	}
	return &Batcher{
		cfg:     cfg,
		flushFn: flushFn,
	}
}

// Add appends an alert to the current batch. If the batch reaches MaxSize
// it is flushed immediately; otherwise a window timer is (re)started.
func (b *Batcher) Add(a *alert.Alert) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buf = append(b.buf, a)

	if len(b.buf) >= b.cfg.MaxSize {
		b.flushLocked()
		return
	}

	if b.timer == nil {
		b.timer = time.AfterFunc(b.cfg.Window, b.windowExpired)
	}
}

// Flush forces an immediate flush of any buffered alerts.
func (b *Batcher) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}

// flushLocked flushes the buffer; must be called with b.mu held.
func (b *Batcher) flushLocked() {
	if len(b.buf) == 0 {
		return
	}
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	batch := b.buf
	b.buf = nil
	go b.flushFn(batch)
}

// windowExpired is called by the timer goroutine.
func (b *Batcher) windowExpired() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.timer = nil
	b.flushLocked()
}
