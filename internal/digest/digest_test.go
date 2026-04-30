package digest

import (
	"testing"
	"time"
)

func TestCompute_Deterministic(t *testing.T) {
	fields := map[string]string{"host": "localhost", "port": "8080", "state": "closed"}
	a := Compute(fields)
	b := Compute(fields)
	if a != b {
		t.Fatalf("expected identical digests, got %q and %q", a, b)
	}
}

func TestCompute_OrderIndependent(t *testing.T) {
	a := Compute(map[string]string{"x": "1", "y": "2"})
	b := Compute(map[string]string{"y": "2", "x": "1"})
	if a != b {
		t.Fatalf("digest should be order-independent, got %q and %q", a, b)
	}
}

func TestCompute_DifferentInputsDifferentDigests(t *testing.T) {
	a := Compute(map[string]string{"host": "a"})
	b := Compute(map[string]string{"host": "b"})
	if a == b {
		t.Fatal("different inputs should produce different digests")
	}
}

func TestStore_And_Get(t *testing.T) {
	d := New(Config{TTL: time.Minute})
	d.Store("key1", "abc123")
	e, ok := d.Get("key1")
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if e.Digest != "abc123" {
		t.Fatalf("expected digest abc123, got %q", e.Digest)
	}
}

func TestGet_MissingKey(t *testing.T) {
	d := New(Config{})
	_, ok := d.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestGet_ExpiredEntry(t *testing.T) {
	d := New(Config{TTL: time.Millisecond})
	d.Store("k", "v")
	time.Sleep(5 * time.Millisecond)
	_, ok := d.Get("k")
	if ok {
		t.Fatal("expected expired entry to return false")
	}
}

func TestEvict_RemovesEntry(t *testing.T) {
	d := New(Config{})
	d.Store("k", "v")
	d.Evict("k")
	_, ok := d.Get("k")
	if ok {
		t.Fatal("expected evicted entry to be absent")
	}
}

func TestLen_ReflectsStored(t *testing.T) {
	d := New(Config{})
	if d.Len() != 0 {
		t.Fatalf("expected 0, got %d", d.Len())
	}
	d.Store("a", "1")
	d.Store("b", "2")
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
	d.Evict("a")
	if d.Len() != 1 {
		t.Fatalf("expected 1, got %d", d.Len())
	}
}
