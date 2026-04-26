package timeout

import (
	"errors"
	"sync"
	"time"
)

// ErrTimedOut is returned when an operation exceeds its deadline.
var ErrTimedOut = errors.New("operation timed out")

// Config holds timeout configuration.
type Config struct {
	// Default is the fallback timeout duration when no per-key override exists.
	Default time.Duration `json:"default"`
	// Overrides maps keys to specific timeout durations.
	Overrides map[string]time.Duration `json:"overrides"`
}

// Manager tracks per-key timeout durations and provides deadline enforcement.
type Manager struct {
	mu        sync.RWMutex
	defaultTT time.Duration
	overrides map[string]time.Duration
}

// New creates a Manager from the provided Config.
// If cfg.Default is zero, it defaults to 5 seconds.
func New(cfg Config) *Manager {
	d := cfg.Default
	if d <= 0 {
		d = 5 * time.Second
	}
	ov := make(map[string]time.Duration, len(cfg.Overrides))
	for k, v := range cfg.Overrides {
		ov[k] = v
	}
	return &Manager{
		defaultTT: d,
		overrides: ov,
	}
}

// For returns the timeout duration for the given key.
func (m *Manager) For(key string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if d, ok := m.overrides[key]; ok {
		return d
	}
	return m.defaultTT
}

// Set registers a per-key timeout override.
func (m *Manager) Set(key string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.overrides[key] = d
}

// Delete removes a per-key override, reverting to the default.
func (m *Manager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.overrides, key)
}

// Do executes fn within the timeout for key, returning ErrTimedOut if exceeded.
func (m *Manager) Do(key string, fn func() error) error {
	d := m.For(key)
	type result struct{ err error }
	ch := make(chan result, 1)
	go func() {
		ch <- result{err: fn()}
	}()
	select {
	case r := <-ch:
		return r.err
	case <-time.After(d):
		return ErrTimedOut
	}
}
