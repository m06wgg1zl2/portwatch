package budget

import (
	"testing"
	"time"
)

func TestRatio_ZeroWhenNoEvents(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	if r := b.Ratio("svc"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestRatio_AllSuccess(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	for i := 0; i < 10; i++ {
		b.Record("svc", false)
	}
	if r := b.Ratio("svc"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestRatio_AllFailures(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	for i := 0; i < 4; i++ {
		b.Record("svc", true)
	}
	if r := b.Ratio("svc"); r != 1.0 {
		t.Fatalf("expected 1.0, got %v", r)
	}
}

func TestRatio_Mixed(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.1})
	for i := 0; i < 9; i++ {
		b.Record("svc", false)
	}
	b.Record("svc", true) // 1 out of 10 = 0.1
	got := b.Ratio("svc")
	if got < 0.09 || got > 0.11 {
		t.Fatalf("expected ~0.1, got %v", got)
	}
}

func TestBreached_BelowThreshold(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.2})
	for i := 0; i < 9; i++ {
		b.Record("svc", false)
	}
	b.Record("svc", true) // 10 % < 20 %
	if b.Breached("svc") {
		t.Fatal("expected budget not breached")
	}
}

func TestBreached_AboveThreshold(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	for i := 0; i < 5; i++ {
		b.Record("svc", true)
	}
	for i := 0; i < 5; i++ {
		b.Record("svc", false)
	} // 50 % > 5 %
	if !b.Breached("svc") {
		t.Fatal("expected budget breached")
	}
}

func TestReset_ClearsEvents(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	b.Record("svc", true)
	b.Reset("svc")
	if r := b.Ratio("svc"); r != 0 {
		t.Fatalf("expected 0 after reset, got %v", r)
	}
}

func TestIndependentKeys(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	b.Record("a", true)
	b.Record("a", true)
	b.Record("b", false)
	if b.Breached("b") {
		t.Fatal("key b should not be breached")
	}
	if !b.Breached("a") {
		t.Fatal("key a should be breached")
	}
}

func TestDefaults_Applied(t *testing.T) {
	b := New(Config{})
	if b.cfg.Window != time.Minute {
		t.Fatalf("expected default window 1m, got %v", b.cfg.Window)
	}
	if b.cfg.Threshold != 0.05 {
		t.Fatalf("expected default threshold 0.05, got %v", b.cfg.Threshold)
	}
}
