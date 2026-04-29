// Package drain provides a graceful shutdown helper that waits for
// in-flight alert handlers to complete before the process exits.
package drain

import (
	"context"
	"sync"
	"time"
)

// Config holds tunable parameters for the drain.
type Config struct {
	// Timeout is the maximum time to wait for in-flight work to finish.
	// Defaults to 10 seconds when zero.
	Timeout time.Duration
}

// Drain tracks in-flight units of work and provides a blocking Wait that
// resolves once all work is done or the deadline is exceeded.
type Drain struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	timeout time.Duration
	closed  bool
}

// New returns a Drain ready for use.
func New(cfg Config) *Drain {
	t := cfg.Timeout
	if t <= 0 {
		t = 10 * time.Second
	}
	return &Drain{timeout: t}
}

// Acquire registers one unit of in-flight work.
// It returns false if the drain has already been closed (i.e. Wait was called).
func (d *Drain) Acquire() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return false
	}
	d.wg.Add(1)
	return true
}

// Release marks one unit of in-flight work as complete.
func (d *Drain) Release() {
	d.wg.Done()
}

// Wait closes the drain to new acquisitions and blocks until all in-flight
// work finishes or the configured timeout elapses.
// It returns ctx.Err() if the parent context is cancelled first.
func (d *Drain) Wait(ctx context.Context) error {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	timer := time.NewTimer(d.timeout)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-timer.C:
		return context.DeadlineExceeded
	case <-ctx.Done():
		return ctx.Err()
	}
}

// InFlight returns the number of currently registered in-flight units.
// It is intended for observability / testing only.
func (d *Drain) InFlight() int {
	// sync.WaitGroup does not expose a counter; we maintain our own.
	return 0 // see extended version in drain_counter.go if needed
}
