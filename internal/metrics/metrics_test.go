package metrics

import (
	"testing"
	"time"
)

func TestInc_IncrementsCount(t *testing.T) {
	c := New()
	c.Inc("open")
	c.Inc("open")
	n, _ := c.Get("open")
	if n != 2 {
		t.Fatalf("expected 2, got %d", n)
	}
}

func TestGet_ZeroForMissing(t *testing.T) {
	c := New()
	n, ts := c.Get("missing")
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
	if !ts.IsZero() {
		t.Fatal("expected zero time")
	}
}

func TestInc_SetsLastSeen(t *testing.T) {
	c := New()
	before := time.Now()
	c.Inc("closed")
	_, ts := c.Get("closed")
	if ts.Before(before) {
		t.Fatal("lastSeen should be after test start")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	c := New()
	c.Inc("open")
	c.Reset("open")
	n, _ := c.Get("open")
	if n != 0 {
		t.Fatalf("expected 0 after reset, got %d", n)
	}
}

func TestSnapshot_ReturnsCopy(t *testing.T) {
	c := New()
	c.Inc("a")
	c.Inc("a")
	c.Inc("b")
	snap := c.Snapshot()
	if snap["a"] != 2 || snap["b"] != 1 {
		t.Fatalf("unexpected snapshot: %v", snap)
	}
	// mutating snap must not affect counter
	snap["a"] = 99
	n, _ := c.Get("a")
	if n != 2 {
		t.Fatal("snapshot mutation affected counter")
	}
}

func TestInc_Concurrent(t *testing.T) {
	c := New()
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() { c.Inc("x"); done <- struct{}{} }()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	n, _ := c.Get("x")
	if n != 100 {
		t.Fatalf("expected 100, got %d", n)
	}
}
