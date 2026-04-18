package backoff

import (
	"testing"
	"time"
)

func TestLinear_Delay(t *testing.T) {
	b := New(Config{
		Strategy:  Linear,
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  10 * time.Second,
	})
	if got := b.Delay(0); got != 100*time.Millisecond {
		t.Fatalf("attempt 0: want 100ms, got %v", got)
	}
	if got := b.Delay(2); got != 300*time.Millisecond {
		t.Fatalf("attempt 2: want 300ms, got %v", got)
	}
}

func TestExponential_Delay(t *testing.T) {
	b := New(Config{
		Strategy:   Exponential,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Multiplier: 2.0,
	})
	if got := b.Delay(0); got != 100*time.Millisecond {
		t.Fatalf("attempt 0: want 100ms, got %v", got)
	}
	if got := b.Delay(1); got != 200*time.Millisecond {
		t.Fatalf("attempt 1: want 200ms, got %v", got)
	}
	if got := b.Delay(3); got != 800*time.Millisecond {
		t.Fatalf("attempt 3: want 800ms, got %v", got)
	}
}

func TestDelay_CappedAtMax(t *testing.T) {
	b := New(Config{
		Strategy:   Exponential,
		BaseDelay:  1 * time.Second,
		MaxDelay:   3 * time.Second,
		Multiplier: 2.0,
	})
	for _, attempt := range []int{5, 10, 20} {
		if got := b.Delay(attempt); got > 3*time.Second {
			t.Fatalf("attempt %d: delay %v exceeds max", attempt, got)
		}
	}
}

func TestDefaults_Applied(t *testing.T) {
	b := New(Config{})
	if b.cfg.BaseDelay != 500*time.Millisecond {
		t.Fatalf("expected default BaseDelay 500ms, got %v", b.cfg.BaseDelay)
	}
	if b.cfg.MaxDelay != 30*time.Second {
		t.Fatalf("expected default MaxDelay 30s, got %v", b.cfg.MaxDelay)
	}
	if b.cfg.Multiplier != 2.0 {
		t.Fatalf("expected default Multiplier 2.0, got %v", b.cfg.Multiplier)
	}
}
