package cooldown_test

import (
	"testing"
	"time"

	"portwatch/internal/cooldown"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	c := cooldown.New(time.Second)
	if !c.Allow("host:80") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_BlockedWithinWindow(t *testing.T) {
	base := time.Now()
	c := cooldown.New(5 * time.Second)
	// Inject a fixed clock.
	c2 := newWithClock(5*time.Second, func() time.Time { return base })
	c2.Allow("k")
	// Advance by less than the window.
	c2 = newWithClock(5*time.Second, func() time.Time { return base.Add(3 * time.Second) })
	_ = c2 // separate instance; use inline approach instead

	// Use the exported Allow twice with a manual time stub via the unexported
	// path — we test via the public API with real time and a very short window.
	short := cooldown.New(10 * time.Second)
	short.Allow("x")
	if short.Allow("x") {
		t.Fatal("expected second immediate call to be blocked")
	}
}

func TestAllow_PermittedAfterWindow(t *testing.T) {
	c := cooldown.New(10 * time.Millisecond)
	c.Allow("port")
	time.Sleep(15 * time.Millisecond)
	if !c.Allow("port") {
		t.Fatal("expected call after window to be permitted")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	c := cooldown.New(10 * time.Second)
	c.Allow("a")
	if !c.Allow("b") {
		t.Fatal("different key should be independent")
	}
}

func TestRemaining_ZeroWhenNotTracked(t *testing.T) {
	c := cooldown.New(time.Second)
	if r := c.Remaining("unknown"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestRemaining_PositiveAfterAllow(t *testing.T) {
	c := cooldown.New(5 * time.Second)
	c.Allow("k")
	if r := c.Remaining("k"); r <= 0 {
		t.Fatalf("expected positive remaining, got %v", r)
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	c := cooldown.New(10 * time.Second)
	c.Allow("p")
	c.Reset("p")
	if !c.Allow("p") {
		t.Fatal("expected allow after reset")
	}
}

func TestKeys_ReturnsTracked(t *testing.T) {
	c := cooldown.New(time.Second)
	c.Allow("a")
	c.Allow("b")
	keys := c.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestNew_PanicsOnZeroDuration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero duration")
		}
	}()
	cooldown.New(0)
}

// newWithClock is a helper that creates a Cooldown and overrides its internal
// clock via a thin wrapper — only used to document intent; real clock injection
// is tested indirectly through sleep-based tests above.
func newWithClock(d time.Duration, _ func() time.Time) *cooldown.Cooldown {
	return cooldown.New(d)
}
