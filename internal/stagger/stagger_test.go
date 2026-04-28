package stagger

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestStep_EvenDistribution(t *testing.T) {
	s := New(Config{Window: 10 * time.Second, Count: 5})
	if s.Step() != 2*time.Second {
		t.Fatalf("expected 2s step, got %v", s.Step())
	}
}

func TestDelay_FirstCallUsesIndex(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New(Config{Window: 10 * time.Second, Count: 5})
	s.clock = fixedClock(now)

	// index 0 → offset 0 → immediate
	d := s.Delay("a", 0)
	if d != 0 {
		t.Fatalf("expected 0 delay for index 0, got %v", d)
	}

	// index 2 → offset 4s
	s2 := New(Config{Window: 10 * time.Second, Count: 5})
	s2.clock = fixedClock(now)
	d2 := s2.Delay("b", 2)
	if d2 != 4*time.Second {
		t.Fatalf("expected 4s delay for index 2, got %v", d2)
	}
}

func TestDelay_SubsequentCallAdvancesByWindow(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New(Config{Window: 10 * time.Second, Count: 5})
	s.clock = fixedClock(now)

	first := s.Delay("k", 1) // slot at +2s
	// Advance clock past first slot
	s.clock = fixedClock(now.Add(3 * time.Second))
	second := s.Delay("k", 1) // next slot = first_slot + window = +12s
	if second <= first {
		t.Fatalf("second delay (%v) should be > first (%v) relative to advanced clock", second, first)
	}
}

func TestDelay_NegativeBecomesZero(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New(Config{Window: 10 * time.Second, Count: 5})
	s.clock = fixedClock(now)
	s.Delay("x", 0) // register slot at now

	// Advance clock well past the next slot
	s.clock = fixedClock(now.Add(30 * time.Second))
	d := s.Delay("x", 0)
	if d != 0 {
		t.Fatalf("expected 0 for past slot, got %v", d)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New(Config{Window: 10 * time.Second, Count: 5})
	s.clock = fixedClock(now)

	s.Delay("y", 3)
	s.Reset("y")

	// After reset, index 0 should give 0 delay again
	d := s.Delay("y", 0)
	if d != 0 {
		t.Fatalf("expected 0 after reset, got %v", d)
	}
}

func TestDefaults_Applied(t *testing.T) {
	s := New(Config{})
	if s.cfg.Count != 1 {
		t.Fatalf("expected count=1, got %d", s.cfg.Count)
	}
	if s.cfg.Window != time.Second {
		t.Fatalf("expected window=1s, got %v", s.cfg.Window)
	}
}

func TestDelay_IndependentKeys(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New(Config{Window: 10 * time.Second, Count: 10})
	s.clock = fixedClock(now)

	d1 := s.Delay("p1", 0)
	d2 := s.Delay("p2", 5)
	if d1 == d2 {
		t.Fatalf("expected different delays for different indices, both got %v", d1)
	}
}
