// Package routing provides weighted round-robin and priority-based
// routing of alerts to named handler groups.
package routing

import (
	"fmt"
	"math/rand"
	"sync"
)

// Route maps a named destination to a relative weight.
type Route struct {
	Name   string
	Weight int
}

// Router selects a destination for each alert based on configured weights.
type Router struct {
	mu     sync.Mutex
	routes []Route
	total  int
}

// New creates a Router from the provided routes.
// Returns an error if no routes are provided or any weight is non-positive.
func New(routes []Route) (*Router, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("routing: at least one route is required")
	}
	total := 0
	for _, r := range routes {
		if r.Weight <= 0 {
			return nil, fmt.Errorf("routing: weight for %q must be positive, got %d", r.Name, r.Weight)
		}
		total += r.Weight
	}
	copy := make([]Route, len(routes))
	for i, r := range routes {
		copy[i] = r
	}
	return &Router{routes: copy, total: total}, nil
}

// Select returns a destination name chosen proportionally by weight.
func (r *Router) Select() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := rand.Intn(r.total) //nolint:gosec
	cumulative := 0
	for _, route := range r.routes {
		cumulative += route.Weight
		if n < cumulative {
			return route.Name
		}
	}
	return r.routes[len(r.routes)-1].Name
}

// Routes returns a copy of the configured routes.
func (r *Router) Routes() []Route {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Route, len(r.routes))
	copy(out, r.routes)
	return out
}

// Total returns the sum of all route weights.
func (r *Router) Total() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.total
}
