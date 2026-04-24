package dedup

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestIsDuplicate_FirstCallPermitted(t *testing.T) {
	d := New(Config{TTL: time.Minute})
	if d.IsDuplicate("key1") {
		t.Fatal("first call should not be a duplicate")
	}
}

func TestIsDuplicate_BlockedWithinWindow(t *testing.T) {
	now := time.Now()
	d := WithClock(New(Config{TTL: time.Minute}), fixedClock(now))
	d.IsDuplicate("key1")
	if !d.IsDuplicate("key1") {
		t.Fatal("second call within TTL should be a duplicate")
	}
}

func TestIsDuplicate_PermittedAfterTTL(t *testing.T) {
	now := time.Now()
	clk := fixedClock(now)
	d := WithClock(New(Config{TTL: time.Second}), clk)
	d.IsDuplicate("key1")

	// Advance past TTL.
	d = WithClock(d, fixedClock(now.Add(2*time.Second)))
	if d.IsDuplicate("key1") {
		t.Fatal("call after TTL expiry should not be a duplicate")
	}
}

func TestIsDuplicate_IndependentKeys(t *testing.T) {
	d := New(Config{TTL: time.Minute})
	d.IsDuplicate("alpha")
	if d.IsDuplicate("beta") {
		t.Fatal("different key should not be a duplicate")
	}
}

func TestIsDuplicate_ZeroTTL_NeverDuplicate(t *testing.T) {
	d := New(Config{TTL: 0})
	d.IsDuplicate("key1")
	if d.IsDuplicate("key1") {
		t.Fatal("zero TTL should never deduplicate")
	}
}

func TestCount_IncrementsOnDuplicate(t *testing.T) {
	now := time.Now()
	d := WithClock(New(Config{TTL: time.Minute}), fixedClock(now))
	d.IsDuplicate("k")
	d.IsDuplicate("k")
	d.IsDuplicate("k")
	if got := d.Count("k"); got != 3 {
		t.Fatalf("expected count 3, got %d", got)
	}
}

func TestCount_ZeroForUnknownKey(t *testing.T) {
	d := New(Config{TTL: time.Minute})
	if c := d.Count("missing"); c != 0 {
		t.Fatalf("expected 0 for unknown key, got %d", c)
	}
}

func TestReset_AllowsNextCall(t *testing.T) {
	now := time.Now()
	d := WithClock(New(Config{TTL: time.Minute}), fixedClock(now))
	d.IsDuplicate("key1")
	d.Reset("key1")
	if d.IsDuplicate("key1") {
		t.Fatal("after reset, next call should not be a duplicate")
	}
}
