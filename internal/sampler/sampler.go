// Package sampler provides probabilistic sampling for event streams,
// allowing a configurable fraction of events to pass through.
package sampler

import (
	"math/rand"
	"sync"
	"time"
)

// Sampler decides whether an event should be processed based on a
// configured sample rate in the range (0.0, 1.0].
type Sampler struct {
	mu   sync.Mutex
	rate float64
	rng  *rand.Rand
	hits uint64
	drops uint64
}

// Config holds the configuration for a Sampler.
type Config struct {
	// Rate is the fraction of events to allow through, e.g. 0.25 = 25%.
	// Values <= 0 are treated as 0 (drop all); values >= 1 are treated as 1 (allow all).
	Rate float64 `json:"rate"`
}

// New creates a Sampler from the provided Config.
func New(cfg Config) *Sampler {
	rate := cfg.Rate
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Allow returns true if the event should be processed according to the
// configured sample rate. It is safe for concurrent use.
func (s *Sampler) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rate <= 0 {
		s.drops++
		return false
	}
	if s.rate >= 1 {
		s.hits++
		return true
	}
	if s.rng.Float64() < s.rate {
		s.hits++
		return true
	}
	s.drops++
	return false
}

// Stats returns the cumulative hit and drop counts since creation or last Reset.
func (s *Sampler) Stats() (hits, drops uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits, s.drops
}

// Reset clears the cumulative hit and drop counters.
func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hits = 0
	s.drops = 0
}

// Rate returns the configured sample rate.
func (s *Sampler) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rate
}
