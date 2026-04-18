package backoff

import (
	"math"
	"time"
)

// Strategy defines how delay grows between retries.
type Strategy int

const (
	Linear Strategy = iota
	Exponential
)

// Config holds backoff configuration.
type Config struct {
	Strategy  Strategy      `json:"strategy"`
	BaseDelay time.Duration `json:"base_delay"`
	MaxDelay  time.Duration `json:"max_delay"`
	Multiplier float64      `json:"multiplier"`
}

// Backoff computes delays for successive attempts.
type Backoff struct {
	cfg Config
}

// New returns a Backoff using the provided Config.
// Defaults are applied for zero values.
func New(cfg Config) *Backoff {
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 500 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 30 * time.Second
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}
	return &Backoff{cfg: cfg}
}

// Delay returns the wait duration for the given attempt (0-indexed).
func (b *Backoff) Delay(attempt int) time.Duration {
	var d time.Duration
	switch b.cfg.Strategy {
	case Exponential:
		factor := math.Pow(b.cfg.Multiplier, float64(attempt))
		d = time.Duration(float64(b.cfg.BaseDelay) * factor)
	default: // Linear
		d = b.cfg.BaseDelay * time.Duration(attempt+1)
	}
	if d > b.cfg.MaxDelay {
		d = b.cfg.MaxDelay
	}
	return d
}

// Reset is a no-op placeholder kept for interface symmetry.
func (b *Backoff) Reset() {}
