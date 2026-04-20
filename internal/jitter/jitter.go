// Package jitter provides utilities for adding randomised jitter to
// durations, useful for spreading out concurrent retry or poll attempts.
package jitter

import (
	"math/rand"
	"sync"
	"time"
)

// Jitter adds a bounded random offset to base durations.
type Jitter struct {
	mu      sync.Mutex
	rng     *rand.Rand
	factor  float64 // fraction of base to use as max jitter, e.g. 0.25
	maxJitter time.Duration // hard cap on the added jitter (0 = uncapped)
}

// New creates a Jitter instance.
//
//   factor    – fraction of the base duration to use as the jitter ceiling
//               (e.g. 0.25 means up to ±25 % of base). Must be > 0.
//   maxJitter – optional hard cap; pass 0 for no cap.
func New(factor float64, maxJitter time.Duration) *Jitter {
	if factor <= 0 {
		factor = 0.1
	}
	return &Jitter{
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		factor:    factor,
		maxJitter: maxJitter,
	}
}

// Apply returns base plus a random duration in [0, factor*base].
// If maxJitter > 0 the added offset is capped at maxJitter.
func (j *Jitter) Apply(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}

	ceil := time.Duration(float64(base) * j.factor)
	if j.maxJitter > 0 && ceil > j.maxJitter {
		ceil = j.maxJitter
	}

	j.mu.Lock()
	offset := time.Duration(j.rng.Int63n(int64(ceil) + 1))
	j.mu.Unlock()

	return base + offset
}

// ApplyFull returns base plus a random duration in [-factor*base, +factor*base]
// (full symmetric jitter). Useful when callers want to spread around a midpoint
// rather than always adding delay.
func (j *Jitter) ApplyFull(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}

	half := time.Duration(float64(base) * j.factor / 2)
	if j.maxJitter > 0 && half > j.maxJitter/2 {
		half = j.maxJitter / 2
	}
	if half == 0 {
		return base
	}

	j.mu.Lock()
	offset := time.Duration(j.rng.Int63n(int64(half)*2+1)) - half
	j.mu.Unlock()

	return base + offset
}
