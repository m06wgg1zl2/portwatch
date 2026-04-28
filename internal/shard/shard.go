// Package shard provides consistent key-based sharding across a fixed number of buckets.
// It allows distributing work (e.g. alert routing, probe scheduling) evenly across
// named workers or partitions using a stable hash.
package shard

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

// Config holds the configuration for a Sharder.
type Config struct {
	// Buckets is the total number of shards to distribute across.
	// Defaults to 1 if zero or negative.
	Buckets int `json:"buckets"`
}

// Sharder maps arbitrary string keys to a stable bucket index.
type Sharder struct {
	mu      sync.RWMutex
	buckets int
	counts  map[int]int64
}

// New creates a Sharder with the given configuration.
// If cfg.Buckets is less than 1, it defaults to 1.
func New(cfg Config) *Sharder {
	b := cfg.Buckets
	if b < 1 {
		b = 1
	}
	return &Sharder{
		buckets: b,
		counts:  make(map[int]int64, b),
	}
}

// Bucket returns the shard index for the given key.
// The result is in the range [0, buckets).
func (s *Sharder) Bucket(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % s.buckets
	if idx < 0 {
		idx += s.buckets
	}

	s.mu.Lock()
	s.counts[idx]++
	s.mu.Unlock()

	return idx
}

// Owns reports whether the given key maps to the specified bucket index.
func (s *Sharder) Owns(key string, bucket int) bool {
	return s.Bucket(key) == bucket
}

// Distribution returns a snapshot of how many times each bucket has been
// assigned, sorted by bucket index.
func (s *Sharder) Distribution() []BucketStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]BucketStat, 0, s.buckets)
	for i := 0; i < s.buckets; i++ {
		stats = append(stats, BucketStat{
			Index: i,
			Count: s.counts[i],
		})
	}
	sort.Slice(stats, func(a, b int) bool {
		return stats[a].Index < stats[b].Index
	})
	return stats
}

// Reset clears all distribution counters.
func (s *Sharder) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts = make(map[int]int64, s.buckets)
}

// Buckets returns the total number of configured buckets.
func (s *Sharder) Buckets() int {
	return s.buckets
}

// BucketStat holds assignment statistics for a single bucket.
type BucketStat struct {
	Index int
	Count int64
}

// String returns a human-readable representation of the stat.
func (b BucketStat) String() string {
	return fmt.Sprintf("bucket[%d]: %d", b.Index, b.Count)
}
