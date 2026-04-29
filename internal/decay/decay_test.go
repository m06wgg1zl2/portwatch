package decay

import (
	"testing"
	"time"
)

func TestAdd_ReturnsInitialScore(t *testing.T) {
	s := New(Config{HalfLife: time.Second})
	score := s.Add("k", 10)
	if score != 10 {
		t.Fatalf("expected 10, got %f", score)
	}
}

func TestScore_ZeroForUnknownKey(t *testing.T) {
	s := New(Config{})
	if got := s.Score("missing"); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestScore_DecaysOverTime(t *testing.T) {
	// Use a very short half-life so we can observe decay quickly.
	s := New(Config{HalfLife: 50 * time.Millisecond})
	s.Add("k", 100)
	time.Sleep(50 * time.Millisecond)
	got := s.Score("k")
	// After one half-life the score should be near 50 (allow ±15 for timing).
	if got < 35 || got > 65 {
		t.Fatalf("expected ~50 after one half-life, got %f", got)
	}
}

func TestAdd_AccumulatesBeforeDecay(t *testing.T) {
	s := New(Config{HalfLife: time.Minute})
	s.Add("k", 5)
	score := s.Add("k", 5)
	// Both adds happen nearly simultaneously; score should be close to 10.
	if score < 9.9 || score > 10.1 {
		t.Fatalf("expected ~10, got %f", score)
	}
}

func TestReset_ClearsScore(t *testing.T) {
	s := New(Config{})
	s.Add("k", 42)
	s.Reset("k")
	if got := s.Score("k"); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestScore_IndependentKeys(t *testing.T) {
	s := New(Config{HalfLife: time.Minute})
	s.Add("a", 10)
	s.Add("b", 20)
	if got := s.Score("a"); got < 9.9 || got > 10.1 {
		t.Fatalf("key a: expected ~10, got %f", got)
	}
	if got := s.Score("b"); got < 19.9 || got > 20.1 {
		t.Fatalf("key b: expected ~20, got %f", got)
	}
}

func TestDefaults_Applied(t *testing.T) {
	s := New(Config{})
	if s.cfg.HalfLife != 30*time.Second {
		t.Fatalf("expected default half-life 30s, got %v", s.cfg.HalfLife)
	}
}
