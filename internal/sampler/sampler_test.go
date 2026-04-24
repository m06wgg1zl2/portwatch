package sampler_test

import (
	"testing"

	"github.com/user/portwatch/internal/sampler"
)

func TestAllow_RateZero_DropsAll(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 0})
	for i := 0; i < 100; i++ {
		if s.Allow() {
			t.Fatal("expected all events to be dropped at rate 0")
		}
	}
	_, drops := s.Stats()
	if drops != 100 {
		t.Fatalf("expected 100 drops, got %d", drops)
	}
}

func TestAllow_RateOne_AllowsAll(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 1.0})
	for i := 0; i < 100; i++ {
		if !s.Allow() {
			t.Fatal("expected all events to be allowed at rate 1.0")
		}
	}
	hits, _ := s.Stats()
	if hits != 100 {
		t.Fatalf("expected 100 hits, got %d", hits)
	}
}

func TestAllow_RateHalf_Approximate(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 0.5})
	const n = 10_000
	allowed := 0
	for i := 0; i < n; i++ {
		if s.Allow() {
			allowed++
		}
	}
	// Expect roughly 50% ± 5%
	if allowed < n*45/100 || allowed > n*55/100 {
		t.Fatalf("expected ~50%% allowed, got %d/%d", allowed, n)
	}
}

func TestAllow_NegativeRate_ClampedToZero(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: -0.5})
	if s.Rate() != 0 {
		t.Fatalf("expected rate clamped to 0, got %f", s.Rate())
	}
	if s.Allow() {
		t.Fatal("expected drop with negative rate")
	}
}

func TestAllow_ExceedsOne_ClampedToOne(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 1.5})
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate clamped to 1.0, got %f", s.Rate())
	}
	if !s.Allow() {
		t.Fatal("expected allow with rate > 1")
	}
}

func TestReset_ClearsStats(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 1.0})
	for i := 0; i < 10; i++ {
		s.Allow()
	}
	s.Reset()
	hits, drops := s.Stats()
	if hits != 0 || drops != 0 {
		t.Fatalf("expected zeroed stats after Reset, got hits=%d drops=%d", hits, drops)
	}
}

func TestStats_HitsPlusDropsEqualsTotal(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 0.3})
	const n = 1000
	for i := 0; i < n; i++ {
		s.Allow()
	}
	hits, drops := s.Stats()
	if hits+drops != n {
		t.Fatalf("expected hits+drops == %d, got %d+%d=%d", n, hits, drops, hits+drops)
	}
}
