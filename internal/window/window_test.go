package window

import (
	"testing"
	"time"
)

func TestAdd_And_Count(t *testing.T) {
	w := New(time.Second)
	w.Add("host:8080")
	w.Add("host:8080")
	if got := w.Count("host:8080"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestCount_ZeroForUnknownKey(t *testing.T) {
	w := New(time.Second)
	if got := w.Count("missing"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestEviction_AfterWindow(t *testing.T) {
	w := New(50 * time.Millisecond)
	w.Add("k")
	w.Add("k")
	time.Sleep(80 * time.Millisecond)
	if got := w.Count("k"); got != 0 {
		t.Fatalf("expected 0 after window expired, got %d", got)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	w := New(time.Second)
	w.Add("k")
	w.Reset("k")
	if got := w.Count("k"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestKeys_ReturnsActiveKeys(t *testing.T) {
	w := New(time.Second)
	w.Add("a")
	w.Add("b")
	keys := w.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestKeys_ExcludesExpiredKeys(t *testing.T) {
	w := New(50 * time.Millisecond)
	w.Add("stale")
	time.Sleep(80 * time.Millisecond)
	w.Add("fresh")
	keys := w.Keys()
	if len(keys) != 1 || keys[0] != "fresh" {
		t.Fatalf("expected only 'fresh', got %v", keys)
	}
}

func TestNew_DefaultSize(t *testing.T) {
	w := New(0)
	if w.size != time.Minute {
		t.Fatalf("expected default size of 1m, got %v", w.size)
	}
}

func TestIndependentKeys(t *testing.T) {
	w := New(time.Second)
	w.Add("a")
	w.Add("a")
	w.Add("b")
	if w.Count("a") != 2 {
		t.Fatalf("expected 2 for 'a'")
	}
	if w.Count("b") != 1 {
		t.Fatalf("expected 1 for 'b'")
	}
}
