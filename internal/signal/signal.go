// Package signal provides edge-detection for boolean state transitions,
// emitting a named signal (rising or falling) when a value changes.
package signal

import (
	"sync"
	"time"
)

// Edge represents the direction of a state transition.
type Edge int

const (
	EdgeNone    Edge = iota
	EdgeRising       // false → true
	EdgeFalling      // true  → false
)

func (e Edge) String() string {
	switch e {
	case EdgeRising:
		return "rising"
	case EdgeFalling:
		return "falling"
	default:
		return "none"
	}
}

// Event is emitted when a tracked key changes state.
type Event struct {
	Key       string
	Edge      Edge
	Prev      bool
	Current   bool
	ChangedAt time.Time
}

// Detector tracks per-key boolean state and reports edge transitions.
type Detector struct {
	mu     sync.Mutex
	states map[string]bool
	init   map[string]bool
}

// New returns an initialised Detector.
func New() *Detector {
	return &Detector{
		states: make(map[string]bool),
		init:   make(map[string]bool),
	}
}

// Observe records the current value for key and returns the resulting Event.
// On the very first observation for a key no edge is emitted (EdgeNone).
func (d *Detector) Observe(key string, current bool) Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.init[key] {
		d.states[key] = current
		d.init[key] = true
		return Event{Key: key, Edge: EdgeNone, Prev: current, Current: current, ChangedAt: time.Now()}
	}

	prev := d.states[key]
	d.states[key] = current

	var edge Edge
	switch {
	case !prev && current:
		edge = EdgeRising
	case prev && !current:
		edge = EdgeFalling
	default:
		edge = EdgeNone
	}

	return Event{Key: key, Edge: edge, Prev: prev, Current: current, ChangedAt: time.Now()}
}

// Reset removes all tracked state for key.
func (d *Detector) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.states, key)
	delete(d.init, key)
}

// Keys returns all keys currently being tracked.
func (d *Detector) Keys() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]string, 0, len(d.init))
	for k := range d.init {
		out = append(out, k)
	}
	return out
}
