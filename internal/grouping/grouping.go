package grouping

import (
	"sync"
	"time"
)

// Group holds a collection of keys that have been associated together
// along with the time the group was last updated.
type Group struct {
	Keys      []string
	UpdatedAt time.Time
}

// Grouper clusters arbitrary string keys into named groups and tracks
// membership over time. It is safe for concurrent use.
type Grouper struct {
	mu     sync.RWMutex
	groups map[string]*Group
	ttl    time.Duration
	clock  func() time.Time
}

// New returns a Grouper that evicts group members that have not been
// refreshed within ttl. Pass zero to disable eviction.
func New(ttl time.Duration) *Grouper {
	return &Grouper{
		groups: make(map[string]*Group),
		ttl:    ttl,
		clock:  time.Now,
	}
}

// Add associates key with the named group, refreshing its timestamp.
func (g *Grouper) Add(group, key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.evict()
	gr, ok := g.groups[group]
	if !ok {
		gr = &Group{}
		g.groups[group] = gr
	}
	for _, k := range gr.Keys {
		if k == key {
			gr.UpdatedAt = g.clock()
			return
		}
	}
	gr.Keys = append(gr.Keys, key)
	gr.UpdatedAt = g.clock()
}

// Members returns the current keys belonging to group, or nil if the
// group does not exist.
func (g *Grouper) Members(group string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	gr, ok := g.groups[group]
	if !ok {
		return nil
	}
	out := make([]string, len(gr.Keys))
	copy(out, gr.Keys)
	return out
}

// Groups returns the names of all active groups.
func (g *Grouper) Groups() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]string, 0, len(g.groups))
	for name := range g.groups {
		out = append(out, name)
	}
	return out
}

// Remove deletes key from the named group. If the group becomes empty
// it is removed entirely.
func (g *Grouper) Remove(group, key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	gr, ok := g.groups[group]
	if !ok {
		return
	}
	filtered := gr.Keys[:0]
	for _, k := range gr.Keys {
		if k != key {
			filtered = append(filtered, k)
		}
	}
	if len(filtered) == 0 {
		delete(g.groups, group)
		return
	}
	gr.Keys = filtered
}

// evict removes groups whose UpdatedAt is older than ttl. Caller must
// hold the write lock.
func (g *Grouper) evict() {
	if g.ttl == 0 {
		return
	}
	now := g.clock()
	for name, gr := range g.groups {
		if now.Sub(gr.UpdatedAt) > g.ttl {
			delete(g.groups, name)
		}
	}
}
