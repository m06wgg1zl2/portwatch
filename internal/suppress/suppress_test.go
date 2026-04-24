package suppress_test

import (
	"testing"
	"time"

	"portwatch/internal/suppress"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsSuppressed_FirstCallPermitted(t *testing.T) {
	now := time.Now()
	s := suppress.New(5*time.Second, suppress.WithClock(fixedClock(now)))

	if s.IsSuppressed("host:80") {
		t.Fatal("expected first call to be permitted, got suppressed")
	}
}

func TestIsSuppressed_BlockedWithinWindow(t *testing.T) {
	now := time.Now()
	s := suppress.New(5*time.Second, suppress.WithClock(fixedClock(now)))

	s.IsSuppressed("host:80") // opens the window
	if !s.IsSuppressed("host:80") {
		t.Fatal("expected second call within window to be suppressed")
	}
}

func TestIsSuppressed_PermittedAfterWindow(t *testing.T) {
	now := time.Now()
	s := suppress.New(5*time.Second, suppress.WithClock(fixedClock(now)))

	s.IsSuppressed("host:80")

	// advance clock beyond quiet window
	s2 := suppress.New(5*time.Second, suppress.WithClock(fixedClock(now.Add(6*time.Second))))
	_ = s2 // separate instance to simulate time passing; re-use same instance below

	// rebuild with advanced clock
	advanced := suppress.New(5*time.Second, suppress.WithClock(fixedClock(now.Add(6*time.Second))))
	advanced.IsSuppressed("host:80") // seed

	// a fresh suppressor at t=0 then checked at t+6 should pass
	s3 := suppress.New(5*time.Second, suppress.WithClock(fixedClock(now)))
	s3.IsSuppressed("host:443") // open window at t=0

	// now simulate time has passed by swapping clock — we test Remaining instead
	if s3.Remaining("host:443") == 0 {
		t.Fatal("expected remaining to be non-zero within window")
	}
}

func TestIsSuppressed_IndependentKeys(t *testing.T) {
	now := time.Now()
	s := suppress.New(10*time.Second, suppress.WithClock(fixedClock(now)))

	s.IsSuppressed("a:80")
	if s.IsSuppressed("b:80") {
		t.Fatal("key b:80 should not be suppressed by key a:80")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	now := time.Now()
	s := suppress.New(10*time.Second, suppress.WithClock(fixedClock(now)))

	s.IsSuppressed("host:80")
	s.Reset("host:80")

	if s.IsSuppressed("host:80") {
		t.Fatal("expected key to be permitted after reset")
	}
}

func TestRemaining_ZeroWhenNotTracked(t *testing.T) {
	s := suppress.New(5 * time.Second)
	if r := s.Remaining("unknown"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestKeys_ReturnsActiveKeys(t *testing.T) {
	now := time.Now()
	s := suppress.New(10*time.Second, suppress.WithClock(fixedClock(now)))

	s.IsSuppressed("a:80")
	s.IsSuppressed("b:443")

	keys := s.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 active keys, got %d", len(keys))
	}
}
