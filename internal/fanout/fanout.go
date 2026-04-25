// Package fanout distributes a single alert to multiple named sinks
// with configurable concurrency and error collection.
package fanout

import (
	"context"
	"fmt"
	"sync"

	"github.com/user/portwatch/internal/alert"
)

// Handler is a function that receives an alert.
type Handler func(ctx context.Context, a alert.Alert) error

// Result holds the outcome of a single sink invocation.
type Result struct {
	Name  string
	Error error
}

// Fanout sends an alert to all registered sinks concurrently.
type Fanout struct {
	mu       sync.RWMutex
	sinks    map[string]Handler
	ordered  []string
}

// New returns an empty Fanout.
func New() *Fanout {
	return &Fanout{
		sinks: make(map[string]Handler),
	}
}

// Register adds a named sink. Duplicate names overwrite the previous entry.
func (f *Fanout) Register(name string, h Handler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.sinks[name]; !exists {
		f.ordered = append(f.ordered, name)
	}
	f.sinks[name] = h
}

// Unregister removes a named sink. It is a no-op if the name does not exist.
func (f *Fanout) Unregister(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sinks, name)
	updated := f.ordered[:0]
	for _, n := range f.ordered {
		if n != name {
			updated = append(updated, n)
		}
	}
	f.ordered = updated
}

// Send dispatches the alert to every registered sink concurrently and
// returns a slice of Result, one per sink, preserving registration order.
func (f *Fanout) Send(ctx context.Context, a alert.Alert) []Result {
	f.mu.RLock()
	names := make([]string, len(f.ordered))
	copy(names, f.ordered)
	handlers := make(map[string]Handler, len(f.sinks))
	for k, v := range f.sinks {
		handlers[k] = v
	}
	f.mu.RUnlock()

	results := make([]Result, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Add(1)
		go func(idx int, n string, h Handler) {
			defer wg.Done()
			var err error
			if h == nil {
				err = fmt.Errorf("nil handler for sink %q", n)
			} else {
				err = h(ctx, a)
			}
			results[idx] = Result{Name: n, Error: err}
		}(i, name, handlers[name])
	}
	wg.Wait()
	return results
}

// Len returns the number of registered sinks.
func (f *Fanout) Len() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.sinks)
}
