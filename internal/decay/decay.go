// Package decay implements an exponential decay scorer that reduces a
// per-key score over time, useful for tracking fading signal strength.
package decay

import (
	"math"
	"sync"
	"time"
)

// Config holds tuning parameters for the decay scorer.
type Config struct {
	// HalfLife is the duration after which a score decays to half its value.
	HalfLife time.Duration
}

func (c *Config) applyDefaults() {
	if c.HalfLife <= 0 {
		c.HalfLife = 30 * time.Second
	}
}

type entry struct {
	score     float64
	updatedAt time.Time
}

// Scorer tracks per-key scores that decay exponentially over time.
type Scorer struct {
	cfg     Config
	mu      sync.Mutex
	entries map[string]entry
}

// New returns a Scorer configured with cfg.
func New(cfg Config) *Scorer {
	cfg.applyDefaults()
	return &Scorer{
		cfg:     cfg,
		entries: make(map[string]entry),
	}
}

// Add increments the current (decayed) score for key by delta and records now
// as the last update time.
func (s *Scorer) Add(key string, delta float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	e := s.entries[key]
	current := s.decayed(e, now)
	current += delta
	s.entries[key] = entry{score: current, updatedAt: now}
	return current
}

// Score returns the current decayed score for key.
func (s *Scorer) Score(key string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.decayed(s.entries[key], time.Now())
}

// Reset removes the tracked score for key.
func (s *Scorer) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// decayed computes the score after applying exponential decay since e.updatedAt.
func (s *Scorer) decayed(e entry, now time.Time) float64 {
	if e.updatedAt.IsZero() || e.score == 0 {
		return 0
	}
	elapsed := now.Sub(e.updatedAt).Seconds()
	hl := s.cfg.HalfLife.Seconds()
	return e.score * math.Pow(0.5, elapsed/hl)
}
