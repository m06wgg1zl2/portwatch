package digest

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Entry holds a computed digest and when it was last updated.
type Entry struct {
	Digest    string
	UpdatedAt time.Time
}

// Digester computes and caches stable fingerprints for alert payloads.
type Digester struct {
	mu      sync.RWMutex
	cache   map[string]Entry
	ttl     time.Duration
}

// Config holds options for the Digester.
type Config struct {
	// TTL controls how long a cached digest entry is retained.
	TTL time.Duration
}

// New returns a Digester with the provided configuration.
// A zero TTL disables expiry.
func New(cfg Config) *Digester {
	if cfg.TTL == 0 {
		cfg.TTL = 10 * time.Minute
	}
	return &Digester{
		cache: make(map[string]Entry),
		ttl:   cfg.TTL,
	}
}

// Compute derives a deterministic SHA-256 fingerprint from the supplied
// key/value pairs. Keys are sorted before hashing so insertion order does
// not affect the result.
func Compute(fields map[string]string) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fields[k])
		b.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", sum)
}

// Store caches a digest under the given key, replacing any existing entry.
func (d *Digester) Store(key, digest string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache[key] = Entry{Digest: digest, UpdatedAt: time.Now()}
}

// Get returns the cached Entry for key and whether it was found and
// still within TTL.
func (d *Digester) Get(key string) (Entry, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	e, ok := d.cache[key]
	if !ok {
		return Entry{}, false
	}
	if d.ttl > 0 && time.Since(e.UpdatedAt) > d.ttl {
		return Entry{}, false
	}
	return e, true
}

// Evict removes the entry for key if present.
func (d *Digester) Evict(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.cache, key)
}

// Len returns the number of entries currently in the cache (including
// potentially expired ones not yet evicted).
func (d *Digester) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.cache)
}
