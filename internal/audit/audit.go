// Package audit provides a structured event log for recording significant
// portwatch lifecycle and state-change events.
package audit

import (
	"fmt"
	"sync"
	"time"
)

// Level represents the severity of an audit event.
type Level int

const (
	LevelInfo  Level = iota
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time
	Level     Level
	Source    string
	Message   string
}

func (e Event) String() string {
	return fmt.Sprintf("[%s] %s (%s): %s",
		e.Timestamp.Format(time.RFC3339),
		e.Level.String(),
		e.Source,
		e.Message,
	)
}

// Log is an in-memory, thread-safe audit event store.
type Log struct {
	mu     sync.RWMutex
	events []Event
	cap    int
}

// New creates a new Log with the given maximum capacity.
// When capacity is reached, the oldest entry is evicted.
func New(capacity int) *Log {
	if capacity <= 0 {
		capacity = 256
	}
	return &Log{cap: capacity}
}

// Add appends a new event to the log.
func (l *Log) Add(level Level, source, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
	}
	if len(l.events) >= l.cap {
		l.events = l.events[1:]
	}
	l.events = append(l.events, e)
}

// All returns a copy of all stored events.
func (l *Log) All() []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// Len returns the number of events currently stored.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.events)
}

// Clear removes all events from the log.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = nil
}
