// Package stagger spreads a set of N tasks evenly across a time window
// to avoid thundering-herd effects when many ports are checked simultaneously.
package stagger

import (
	"sync"
	"time"
)

// Config holds stagger parameters.
type Config struct {
	// Window is the total duration across which tasks are spread.
	Window time.Duration
	// Count is the expected number of tasks to distribute.
	Count int
}

// Stagger distributes task slots evenly across a window.
type Stagger struct {
	mu     sync.Mutex
	cfg    Config
	step   time.Duration
	next   map[string]time.Time
	clock  func() time.Time
}

// New returns a Stagger using the given config.
// Count must be >= 1; Window must be > 0.
func New(cfg Config) *Stagger {
	if cfg.Count < 1 {
		cfg.Count = 1
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Second
	}
	return &Stagger{
		cfg:   cfg,
		step:  cfg.Window / time.Duration(cfg.Count),
		next:  make(map[string]time.Time),
		clock: time.Now,
	}
}

// Delay returns how long the caller should wait before executing the task
// identified by key. The first call for a key assigns a deterministic slot;
// subsequent calls advance by one full window.
func (s *Stagger) Delay(key string, index int) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	if t, ok := s.next[key]; ok {
		delay := t.Sub(now)
		s.next[key] = t.Add(s.cfg.Window)
		if delay < 0 {
			return 0
		}
		return delay
	}

	// Assign initial slot based on index.
	offset := time.Duration(index) * s.step
	if offset >= s.cfg.Window {
		offset = offset % s.cfg.Window
	}
	slot := now.Add(offset)
	s.next[key] = slot.Add(s.cfg.Window)
	delay := slot.Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}

// Reset removes the scheduling state for key, allowing it to be re-slotted.
func (s *Stagger) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.next, key)
}

// Step returns the computed interval between consecutive task slots.
func (s *Stagger) Step() time.Duration {
	return s.step
}
