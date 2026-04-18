package schedule

import (
	"time"
)

// Schedule defines a recurring check interval with optional jitter.
type Schedule struct {
	interval time.Duration
	jitter   time.Duration
}

// Config holds the configuration for a Schedule.
type Config struct {
	IntervalSeconds int `json:"interval_seconds"`
	JitterSeconds   int `json:"jitter_seconds"`
}

// New creates a Schedule from Config.
func New(cfg Config) *Schedule {
	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	jitter := time.Duration(cfg.JitterSeconds) * time.Second
	return &Schedule{interval: interval, jitter: jitter}
}

// Next returns the duration until the next tick, applying jitter if configured.
func (s *Schedule) Next() time.Duration {
	if s.jitter <= 0 {
		return s.interval
	}
	// Use nanosecond modulo for simple pseudo-jitter without importing math/rand.
	nano := time.Now().UnixNano()
	jitterNs := nano % int64(s.jitter)
	return s.interval + time.Duration(jitterNs)
}

// Ticker returns a channel that fires according to the schedule.
// The caller is responsible for stopping the returned *time.Ticker.
func (s *Schedule) Ticker() *time.Ticker {
	return time.NewTicker(s.Next())
}

// Interval returns the base interval.
func (s *Schedule) Interval() time.Duration {
	return s.interval
}

// Jitter returns the configured jitter duration.
func (s *Schedule) Jitter() time.Duration {
	return s.jitter
}
