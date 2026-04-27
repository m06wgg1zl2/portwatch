// Package cascade implements a failure cascade detector that tracks
// dependent port groups and suppresses downstream alerts when an upstream
// port is already known to be down.
package cascade

import (
	"fmt"
	"sync"
	"time"
)

// Dependency maps an upstream key to the set of downstream keys that depend on it.
type Dependency struct {
	Upstream   string
	Downstream []string
}

// Config holds the configuration for the cascade detector.
type Config struct {
	// Dependencies defines upstream→downstream relationships.
	Dependencies []Dependency
	// TTL is how long an upstream failure suppresses downstream alerts.
	// Defaults to 60 seconds if zero.
	TTL time.Duration
}

// entry records when an upstream failure was observed.
type entry struct {
	observedAt time.Time
}

// Cascade detects whether a downstream alert should be suppressed because
// a known upstream dependency is already failing.
type Cascade struct {
	mu       sync.RWMutex
	cfg      Config
	failing  map[string]entry   // upstream key → failure entry
	deps     map[string][]string // downstream key → upstream keys
	clock    func() time.Time
}

// New creates a Cascade detector from the provided config.
// A zero TTL defaults to 60 seconds.
func New(cfg Config) *Cascade {
	if cfg.TTL <= 0 {
		cfg.TTL = 60 * time.Second
	}

	// Build reverse index: downstream → list of upstreams
	deps := make(map[string][]string)
	for _, dep := range cfg.Dependencies {
		for _, ds := range dep.Downstream {
			deps[ds] = append(deps[ds], dep.Upstream)
		}
	}

	return &Cascade{
		cfg:     cfg,
		failing: make(map[string]entry),
		deps:    deps,
		clock:   time.Now,
	}
}

// RecordFailure marks the given upstream key as currently failing.
func (c *Cascade) RecordFailure(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failing[key] = entry{observedAt: c.clock()}
}

// RecordRecovery clears a previously recorded upstream failure.
func (c *Cascade) RecordRecovery(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.failing, key)
}

// IsCascade returns true if the given downstream key should be suppressed
// because at least one of its upstream dependencies is currently failing
// within the configured TTL window.
func (c *Cascade) IsCascade(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	upstreams, ok := c.deps[key]
	if !ok {
		return false
	}

	for _, up := range upstreams {
		if e, found := c.failing[up]; found {
			if now.Sub(e.observedAt) <= c.cfg.TTL {
				return true
			}
			// TTL expired — evict stale entry
			delete(c.failing, up)
		}
	}
	return false
}

// FailingUpstreams returns the list of upstream keys that are currently
// marked as failing (within TTL) for the given downstream key.
func (c *Cascade) FailingUpstreams(key string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	upstreams, ok := c.deps[key]
	if !ok {
		return nil
	}

	var result []string
	for _, up := range upstreams {
		if e, found := c.failing[up]; found {
			if now.Sub(e.observedAt) <= c.cfg.TTL {
				result = append(result, up)
			} else {
				delete(c.failing, up)
			}
		}
	}
	return result
}

// String returns a human-readable summary of current upstream failures.
func (c *Cascade) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("cascade{failing=%d, deps=%d}", len(c.failing), len(c.deps))
}
