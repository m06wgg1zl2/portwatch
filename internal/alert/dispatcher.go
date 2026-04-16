package alert

import (
	"log"
	"sync"
)

// Handler is a function that processes an Alert.
type Handler func(Alert)

// Dispatcher fans out alerts to registered handlers.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers []Handler
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Register adds a handler to the dispatcher.
func (d *Dispatcher) Register(h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, h)
}

// Dispatch sends the alert to all registered handlers concurrently.
func (d *Dispatcher) Dispatch(a Alert) {
	d.mu.RLock()
	handlers := make([]Handler, len(d.handlers))
	copy(handlers, d.handlers)
	d.mu.RUnlock()

	var wg sync.WaitGroup
	for _, h := range handlers {
		wg.Add(1)
		go func(fn Handler) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("alert handler panic: %v", r)
				}
			}()
			fn(a)
		}(h)
	}
	wg.Wait()
}
