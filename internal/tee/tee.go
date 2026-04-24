// Package tee provides a fan-out stage that delivers a single alert to
// multiple independent handlers concurrently, collecting any errors.
package tee

import (
	"fmt"
	"strings"
	"sync"

	"portwatch/internal/alert"
)

// Handler is any function that accepts an alert.
type Handler func(a alert.Alert) error

// Tee fans an alert out to all registered handlers in parallel.
type Tee struct {
	handlers []Handler
}

// New returns a Tee that will deliver alerts to each of the supplied handlers.
func New(handlers ...Handler) *Tee {
	h := make([]Handler, len(handlers))
	copy(h, handlers)
	return &Tee{handlers: h}
}

// Add appends a handler to the fan-out list.
func (t *Tee) Add(h Handler) {
	t.handlers = append(t.handlers, h)
}

// Len returns the number of registered handlers.
func (t *Tee) Len() int { return len(t.handlers) }

// Send delivers the alert to every handler concurrently.
// It waits for all handlers to finish and returns a combined error if any
// handler returned a non-nil error.
func (t *Tee) Send(a alert.Alert) error {
	if len(t.handlers) == 0 {
		return nil
	}

	var (
		mu   sync.Mutex
		errs []string
		wg   sync.WaitGroup
	)

	for i, h := range t.handlers {
		wg.Add(1)
		go func(idx int, fn Handler) {
			defer wg.Done()
			if err := fn(a); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("handler[%d]: %s", idx, err))
				mu.Unlock()
			}
		}(i, h)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("tee: %s", strings.Join(errs, "; "))
	}
	return nil
}
