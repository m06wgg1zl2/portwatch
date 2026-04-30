package reaper

import (
	"sync"
	"time"
)

// Entry holds a key and its expiry time.
type Entry struct {
	Key       string
	ExpiresAt time.Time
}

// OnExpire is called when an entry is reaped.
type OnExpire func(key string)

// Reaper periodically removes expired keys and fires callbacks.
type Reaper struct {
	mu       sync.Mutex
	entries  map[string]time.Time
	callback OnExpire
	interval time.Duration
	stop     chan struct{}
}

// Config holds reaper configuration.
type Config struct {
	// Interval between reap passes.
	Interval time.Duration
}

// New creates a Reaper with the given config and expiry callback.
func New(cfg Config, cb OnExpire) *Reaper {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	return &Reaper{
		entries:  make(map[string]time.Time),
		callback: cb,
		interval: cfg.Interval,
		stop:     make(chan struct{}),
	}
}

// Track registers a key with a TTL.
func (r *Reaper) Track(key string, ttl time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[key] = time.Now().Add(ttl)
}

// Remove unregisters a key.
func (r *Reaper) Remove(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, key)
}

// Start begins the background reap loop.
func (r *Reaper) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.reap()
			case <-r.stop:
				return
			}
		}
	}()
}

// Stop halts the background loop.
func (r *Reaper) Stop() {
	close(r.stop)
}

func (r *Reaper) reap() {
	now := time.Now()
	r.mu.Lock()
	var expired []string
	for k, exp := range r.entries {
		if now.After(exp) {
			expired = append(expired, k)
			delete(r.entries, k)
		}
	}
	r.mu.Unlock()
	for _, k := range expired {
		if r.callback != nil {
			r.callback(k)
		}
	}
}
