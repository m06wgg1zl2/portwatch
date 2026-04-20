package jitter_test

import (
	"testing"
	"time"

	"github.com/example/portwatch/internal/jitter"
)

func TestApply_ResultInRange(t *testing.T) {
	j := jitter.New(0.25, 0)
	base := 100 * time.Millisecond

	for i := 0; i < 200; i++ {
		got := j.Apply(base)
		if got < base {
			t.Fatalf("Apply returned %v, want >= %v", got, base)
		}
		max := base + time.Duration(float64(base)*0.25)
		if got > max {
			t.Fatalf("Apply returned %v, want <= %v", got, max)
		}
	}
}

func TestApply_HardCap(t *testing.T) {
	cap := 5 * time.Millisecond
	j := jitter.New(0.5, cap)
	base := 200 * time.Millisecond

	for i := 0; i < 200; i++ {
		got := j.Apply(base)
		if got < base {
			t.Fatalf("Apply returned %v, want >= %v", got, base)
		}
		if got > base+cap {
			t.Fatalf("Apply returned %v, exceeds cap %v", got, base+cap)
		}
	}
}

func TestApply_ZeroBase(t *testing.T) {
	j := jitter.New(0.25, 0)
	if got := j.Apply(0); got != 0 {
		t.Fatalf("expected 0 for zero base, got %v", got)
	}
}

func TestApplyFull_Symmetric(t *testing.T) {
	j := jitter.New(0.2, 0)
	base := 100 * time.Millisecond
	half := time.Duration(float64(base) * 0.2 / 2)

	for i := 0; i < 300; i++ {
		got := j.ApplyFull(base)
		if got < base-half || got > base+half {
			t.Fatalf("ApplyFull returned %v, want in [%v, %v]", got, base-half, base+half)
		}
	}
}

func TestApplyFull_ZeroBase(t *testing.T) {
	j := jitter.New(0.1, 0)
	if got := j.ApplyFull(0); got != 0 {
		t.Fatalf("expected 0 for zero base, got %v", got)
	}
}

func TestNew_DefaultFactor(t *testing.T) {
	// factor <= 0 should fall back to 0.1 without panicking
	j := jitter.New(-1, 0)
	base := 100 * time.Millisecond
	got := j.Apply(base)
	if got < base {
		t.Fatalf("Apply returned %v, want >= %v", got, base)
	}
}
