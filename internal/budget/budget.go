// Package budget implements an error-budget tracker that measures the ratio
// of failures to total events over a rolling window and signals when the
// configured threshold is breached.
package budget

import (
	"sync"
	"time"
)

// Config holds the parameters for a Budget.
type Config struct {
	// Window is the rolling duration over which events are counted.
	Window time.Duration `json:"window"`
	// Threshold is the maximum allowed failure ratio (0.0–1.0).
	// A value of 0.05 means 5 % failures are tolerated.
	Threshold float64 `json:"threshold"`
}

type event struct {
	at      time.Time
	failed  bool
}

// Budget tracks the error budget for a named key.
type Budget struct {
	mu     sync.Mutex
	cfg    Config
	events map[string][]event
}

// New returns a Budget with the provided configuration.
func New(cfg Config) *Budget {
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Threshold <= 0 || cfg.Threshold > 1 {
		cfg.Threshold = 0.05
	}
	return &Budget{
		cfg:    cfg,
		events: make(map[string][]event),
	}
}

// Record appends an event for key. failed=true counts against the budget.
func (b *Budget) Record(key string, failed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.evict(key)
	b.events[key] = append(b.events[key], event{at: time.Now(), failed: failed})
}

// Ratio returns the current failure ratio for key (failures / total).
// Returns 0 when no events have been recorded.
func (b *Budget) Ratio(key string) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.evict(key)
	evs := b.events[key]
	if len(evs) == 0 {
		return 0
	}
	var failures int
	for _, e := range evs {
		if e.failed {
			failures++
		}
	}
	return float64(failures) / float64(len(evs))
}

// Breached reports whether the failure ratio for key exceeds the threshold.
func (b *Budget) Breached(key string) bool {
	return b.Ratio(key) > b.cfg.Threshold
}

// Reset clears all recorded events for key.
func (b *Budget) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.events, key)
}

// evict removes events outside the rolling window. Caller must hold mu.
func (b *Budget) evict(key string) {
	cutoff := time.Now().Add(-b.cfg.Window)
	evs := b.events[key]
	i := 0
	for i < len(evs) && evs[i].at.Before(cutoff) {
		i++
	}
	b.events[key] = evs[i:]
}
