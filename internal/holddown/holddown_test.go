package holddown_test

import (
	"testing"
	"time"

	"portwatch/internal/holddown"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestObserve_NotConfirmedBeforeWindow(t *testing.T) {
	now := time.Now()
	h := holddown.WithClock(holddown.New(holddown.Config{Window: 10 * time.Second}), fixedClock(now))

	if h.Observe("k", true) {
		t.Fatal("expected false before window elapses")
	}
}

func TestObserve_ConfirmedAfterWindow(t *testing.T) {
	now := time.Now()
	h := holddown.New(holddown.Config{Window: 5 * time.Second})
	h = holddown.WithClock(h, fixedClock(now))

	h.Observe("k", true) // start timer

	// Advance clock past window.
	h = holddown.WithClock(h, fixedClock(now.Add(6*time.Second)))
	if !h.Observe("k", true) {
		t.Fatal("expected true after window elapses")
	}
}

func TestObserve_StateFlipResetsTimer(t *testing.T) {
	now := time.Now()
	h := holddown.WithClock(holddown.New(holddown.Config{Window: 5 * time.Second}), fixedClock(now))

	h.Observe("k", true)

	// Flip state — timer should restart.
	h = holddown.WithClock(h, fixedClock(now.Add(6*time.Second)))
	if h.Observe("k", false) {
		t.Fatal("state flip should reset timer; expected false")
	}
}

func TestObserve_IndependentKeys(t *testing.T) {
	now := time.Now()
	h := holddown.WithClock(holddown.New(holddown.Config{Window: 5 * time.Second}), fixedClock(now))

	h.Observe("a", true)
	h.Observe("b", true)

	h = holddown.WithClock(h, fixedClock(now.Add(6*time.Second)))

	if !h.Observe("a", true) {
		t.Fatal("key a should be confirmed")
	}
	if !h.Observe("b", true) {
		t.Fatal("key b should be confirmed independently")
	}
}

func TestReset_ClearsPending(t *testing.T) {
	now := time.Now()
	h := holddown.WithClock(holddown.New(holddown.Config{Window: 5 * time.Second}), fixedClock(now))

	h.Observe("k", true)
	h.Reset("k")

	h = holddown.WithClock(h, fixedClock(now.Add(6*time.Second)))
	if h.Observe("k", true) {
		t.Fatal("after reset timer should restart; expected false")
	}
}

func TestDefault_WindowApplied(t *testing.T) {
	h := holddown.New(holddown.Config{})
	if h.Window() != 5*time.Second {
		t.Fatalf("expected default window 5s, got %v", h.Window())
	}
}
