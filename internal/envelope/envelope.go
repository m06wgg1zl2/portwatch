// Package envelope wraps an alert with routing metadata such as target labels,
// priority, and a trace ID so that downstream pipeline stages can make
// routing decisions without inspecting alert internals.
package envelope

import (
	"fmt"
	"math/rand"
	"time"

	"portwatch/internal/alert"
)

// Envelope carries an Alert together with routing metadata.
type Envelope struct {
	Alert     *alert.Alert
	TraceID   string
	Priority  int
	Labels    map[string]string
	CreatedAt time.Time
}

// New wraps a into a new Envelope with a generated trace ID and the current
// timestamp. Priority defaults to 0 (normal).
func New(a *alert.Alert) *Envelope {
	return &Envelope{
		Alert:     a,
		TraceID:   newTraceID(),
		Priority:  0,
		Labels:    make(map[string]string),
		CreatedAt: time.Now(),
	}
}

// WithPriority returns a shallow copy of e with the given priority.
func (e *Envelope) WithPriority(p int) *Envelope {
	copy := *e
	copy.Priority = p
	return &copy
}

// SetLabel attaches a key/value label to the envelope.
func (e *Envelope) SetLabel(key, value string) {
	e.Labels[key] = value
}

// Label returns the value for key and whether it was present.
func (e *Envelope) Label(key string) (string, bool) {
	v, ok := e.Labels[key]
	return v, ok
}

// String returns a compact human-readable representation.
func (e *Envelope) String() string {
	return fmt.Sprintf("envelope trace=%s priority=%d alert=%s",
		e.TraceID, e.Priority, e.Alert)
}

func newTraceID() string {
	return fmt.Sprintf("%016x", rand.Int63()) //nolint:gosec
}
