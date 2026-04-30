package shard_test

import (
	"fmt"
	"testing"

	"github.com/user/portwatch/internal/shard"
)

func TestNew_DefaultBuckets(t *testing.T) {
	s := shard.New(shard.Config{})
	if s == nil {
		t.Fatal("expected non-nil Shard")
	}
}

func TestAssign_ConsistentForSameKey(t *testing.T) {
	s := shard.New(shard.Config{Buckets: 8})
	key := "host:9200"
	a := s.Assign(key)
	b := s.Assign(key)
	if a != b {
		t.Errorf("expected consistent bucket, got %d and %d", a, b)
	}
}

func TestAssign_BucketInRange(t *testing.T) {
	buckets := uint32(16)
	s := shard.New(shard.Config{Buckets: buckets})
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		b := s.Assign(key)
		if b >= buckets {
			t.Errorf("bucket %d out of range [0, %d)", b, buckets)
		}
	}
}

func TestAssign_DistributionReasonable(t *testing.T) {
	buckets := uint32(4)
	s := shard.New(shard.Config{Buckets: buckets})
	counts := make(map[uint32]int, buckets)
	for i := 0; i < 400; i++ {
		key := fmt.Sprintf("target-%d", i)
		counts[s.Assign(key)]++
	}
	for b, c := range counts {
		if c < 50 {
			t.Errorf("bucket %d under-represented: %d assignments", b, c)
		}
	}
}

func TestOwns_TrueForAssignedBucket(t *testing.T) {
	s := shard.New(shard.Config{Buckets: 8, Self: 3})
	// find a key that maps to bucket 3
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("probe-%d", i)
		if s.Assign(key) == 3 {
			if !s.Owns(key) {
				t.Errorf("expected Owns=true for key %q in bucket 3", key)
			}
			return
		}
	}
	t.Skip("no key hashed to bucket 3 in sample; increase iterations")
}

func TestOwns_FalseForOtherBucket(t *testing.T) {
	s := shard.New(shard.Config{Buckets: 8, Self: 0})
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("probe-%d", i)
		if s.Assign(key) != 0 {
			if s.Owns(key) {
				t.Errorf("expected Owns=false for key %q not in bucket 0", key)
			}
			return
		}
	}
	t.Skip("all keys hashed to bucket 0; increase iterations")
}

func TestBuckets_ReturnsConfiguredCount(t *testing.T) {
	s := shard.New(shard.Config{Buckets: 12})
	if got := s.Buckets(); got != 12 {
		t.Errorf("expected 12 buckets, got %d", got)
	}
}
