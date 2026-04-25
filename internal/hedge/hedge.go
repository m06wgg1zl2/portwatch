// Package hedge implements a hedged-request pattern: when a primary
// operation exceeds a soft deadline, a duplicate request is launched in
// parallel and the first successful result wins.
package hedge

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Result holds the outcome of a hedged operation.
type Result struct {
	Value interface{}
	Err   error
}

// Config controls hedge behaviour.
type Config struct {
	// SoftTimeout is how long to wait before launching the hedge request.
	SoftTimeout time.Duration
	// HardTimeout caps the total wait across all attempts.
	HardTimeout time.Duration
}

// Hedger runs operations with the hedged-request pattern.
type Hedger struct {
	cfg Config
}

// New returns a Hedger with the given config. Defaults are applied for
// zero values: SoftTimeout=50ms, HardTimeout=500ms.
func New(cfg Config) *Hedger {
	if cfg.SoftTimeout <= 0 {
		cfg.SoftTimeout = 50 * time.Millisecond
	}
	if cfg.HardTimeout <= 0 {
		cfg.HardTimeout = 500 * time.Millisecond
	}
	return &Hedger{cfg: cfg}
}

// Do executes fn up to twice using the hedged pattern. The first call
// starts immediately; if it does not return within SoftTimeout a second
// call is launched. The first non-error result is returned. If both
// calls fail, the error from the second attempt is returned.
func (h *Hedger) Do(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) Result {
	ctx, cancel := context.WithTimeout(ctx, h.HardTimeout)
	defer cancel()

	ch := make(chan Result, 2)

	launch := func() {
		v, err := fn(ctx)
		ch <- Result{Value: v, Err: err}
	}

	go launch()

	select {
	case res := <-ch:
		if res.Err == nil {
			return res
		}
		// Primary failed fast — launch hedge immediately.
		go launch()
	case <-time.After(h.cfg.SoftTimeout):
		// Primary is slow — launch hedge in parallel.
		go launch()
	}

	// Collect up to two results; return first success.
	var last Result
	var mu sync.Mutex
	for i := 0; i < 2; i++ {
		select {
		case res := <-ch:
			if res.Err == nil {
				return res
			}
			mu.Lock()
			last = res
			mu.Unlock()
		case <-ctx.Done():
			return Result{Err: errors.New("hedge: hard timeout exceeded")}
		}
	}
	return last
}

// SoftTimeout returns the configured soft deadline.
func (h *Hedger) SoftTimeout() time.Duration { return h.cfg.SoftTimeout }

// HardTimeout returns the configured hard deadline.
func (h *Hedger) HardTimeout() time.Duration { return h.cfg.HardTimeout }
