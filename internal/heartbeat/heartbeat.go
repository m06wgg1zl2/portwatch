package heartbeat

import (
	"sync"
	"time"
)

// Config holds heartbeat configuration.
type Config struct {
	Interval time.Duration `json:"interval"`
	MissedThreshold int    `json:"missed_threshold"`
}

// Status represents the current heartbeat status.
type Status int

const (
	StatusHealthy Status = iota
	StatusDegraded
	StatusDead
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusDegraded:
		return "degraded"
	case StatusDead:
		return "dead"
	default:
		return "unknown"
	}
}

// Heartbeat tracks periodic beats and derives liveness status.
type Heartbeat struct {
	mu       sync.RWMutex
	cfg      Config
	lastBeat time.Time
	beats    int64
	missed   int
}

// New creates a Heartbeat with the given config. Sensible defaults are applied.
func New(cfg Config) *Heartbeat {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.MissedThreshold <= 0 {
		cfg.MissedThreshold = 3
	}
	return &Heartbeat{cfg: cfg}
}

// Beat records a heartbeat at the current time.
func (h *Heartbeat) Beat() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastBeat = time.Now()
	h.beats++
	h.missed = 0
}

// Check evaluates the current status based on elapsed time since the last beat.
func (h *Heartbeat) Check() Status {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.lastBeat.IsZero() {
		return StatusDead
	}
	elapsed := time.Since(h.lastBeat)
	missed := int(elapsed / h.cfg.Interval)
	h.missed = missed
	switch {
	case missed == 0:
		return StatusHealthy
	case missed < h.cfg.MissedThreshold:
		return StatusDegraded
	default:
		return StatusDead
	}
}

// LastBeat returns the time of the most recent beat.
func (h *Heartbeat) LastBeat() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastBeat
}

// Beats returns the total number of beats recorded.
func (h *Heartbeat) Beats() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.beats
}

// Missed returns the number of missed intervals detected at the last Check.
func (h *Heartbeat) Missed() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.missed
}
