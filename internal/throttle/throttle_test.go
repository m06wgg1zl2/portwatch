package throttle

import (
	"testing"
	"time"
)

func newTestThrottle(burst int, interval string) *Throttle {
	t, err := New(Config{MaxBurst: burst, Interval: interval})
	if err != nil {
		panic(err)
	}
	return t
}

func TestAllow_ConsumesTokens(t *testing.T) {
	th := newTestThrottle(3, "1m")
	for i := 0; i < 3; i++ {
		if !th.Allow("k") {
			t.Fatalf("expected allow on call %d", i+1)
		}
	}
	if th.Allow("k") {
		t.Fatal("expected deny after burst exhausted")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	th := newTestThrottle(1, "1m")
	if !th.Allow("a") {
		t.Fatal("expected allow for key a")
	}
	if !th.Allow("b") {
		t.Fatal("expected allow for key b")
	}
	if th.Allow("a") {
		t.Fatal("expected deny for key a after burst")
	}
}

func TestAllow_RefillAfterInterval(t *testing.T) {
	th := newTestThrottle(1, "50ms")
	if !th.Allow("k") {
		t.Fatal("first call should be allowed")
	}
	if th.Allow("k") {
		t.Fatal("second call should be denied")
	}
	time.Sleep(60 * time.Millisecond)
	if !th.Allow("k") {
		t.Fatal("should be allowed after interval")
	}
}

func TestRemaining_DefaultsToMax(t *testing.T) {
	th := newTestThrottle(5, "1m")
	if r := th.Remaining("new"); r != 5 {
		t.Fatalf("expected 5, got %d", r)
	}
}

func TestRemaining_DecreasesOnAllow(t *testing.T) {
	th := newTestThrottle(3, "1m")
	th.Allow("k")
	th.Allow("k")
	if r := th.Remaining("k"); r != 1 {
		t.Fatalf("expected 1 remaining, got %d", r)
	}
}

func TestReset_ClearsState(t *testing.T) {
	th := newTestThrottle(1, "1m")
	th.Allow("k")
	if th.Allow("k") {
		t.Fatal("should be denied before reset")
	}
	th.Reset("k")
	if !th.Allow("k") {
		t.Fatal("should be allowed after reset")
	}
}

func TestNew_InvalidInterval(t *testing.T) {
	_, err := New(Config{MaxBurst: 1, Interval: "bad"})
	if err == nil {
		t.Fatal("expected error for invalid interval")
	}
}
