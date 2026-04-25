package trend

import (
	"sync"
	"time"
)

// Direction represents the direction of a trend.
type Direction int

const (
	Flat     Direction = iota
	Rising             // more failures recently
	Falling            // fewer failures recently
)

func (d Direction) String() string {
	switch d {
	case Rising:
		return "rising"
	case Falling:
		return "falling"
	default:
		return "flat"
	}
}

// Config holds configuration for the Tracker.
type Config struct {
	// Window is how far back to look when computing a trend.
	Window time.Duration
	// MinSamples is the minimum number of samples required before a direction
	// is reported; below this threshold Flat is always returned.
	MinSamples int
}

type sample struct {
	at    time.Time
	value float64
}

// Tracker records time-series samples per key and reports whether the
// values are trending upward, downward, or are flat.
type Tracker struct {
	mu      sync.Mutex
	cfg     Config
	buckets map[string][]sample
}

// New creates a Tracker with the given Config.
func New(cfg Config) *Tracker {
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = 3
	}
	return &Tracker{
		cfg:     cfg,
		buckets: make(map[string][]sample),
	}
}

// Record adds a new sample for key at the current time.
func (t *Tracker) Record(key string, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.buckets[key] = append(t.buckets[key], sample{at: now, value: value})
	t.evict(key, now)
}

// Direction returns the current trend direction for the given key.
func (t *Tracker) Direction(key string) Direction {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.evict(key, now)
	samples := t.buckets[key]
	if len(samples) < t.cfg.MinSamples {
		return Flat
	}
	mid := len(samples) / 2
	first := avg(samples[:mid])
	second := avg(samples[mid:])
	switch {
	case second > first:
		return Rising
	case second < first:
		return Falling
	default:
		return Flat
	}
}

// Reset clears all samples for the given key.
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.buckets, key)
}

func (t *Tracker) evict(key string, now time.Time) {
	cutoff := now.Add(-t.cfg.Window)
	samples := t.buckets[key]
	i := 0
	for i < len(samples) && samples[i].at.Before(cutoff) {
		i++
	}
	t.buckets[key] = samples[i:]
}

func avg(samples []sample) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s.value
	}
	return sum / float64(len(samples))
}
