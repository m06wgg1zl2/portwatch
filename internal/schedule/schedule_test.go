package schedule

import (
	"testing"
	"time"
)

func TestNew_DefaultInterval(t *testing.T) {
	s := New(Config{})
	if s.Interval() != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", s.Interval())
	}
}

func TestNew_CustomInterval(t *testing.T) {
	s := New(Config{IntervalSeconds: 10})
	if s.Interval() != 10*time.Second {
		t.Fatalf("expected 10s, got %v", s.Interval())
	}
}

func TestNew_Jitter(t *testing.T) {
	s := New(Config{IntervalSeconds: 10, JitterSeconds: 2})
	if s.Jitter() != 2*time.Second {
		t.Fatalf("expected 2s jitter, got %v", s.Jitter())
	}
}

func TestNext_NoJitter(t *testing.T) {
	s := New(Config{IntervalSeconds: 5})
	if s.Next() != 5*time.Second {
		t.Fatalf("expected exact interval without jitter")
	}
}

func TestNext_WithJitter_GteBase(t *testing.T) {
	s := New(Config{IntervalSeconds: 5, JitterSeconds: 3})
	next := s.Next()
	if next < 5*time.Second {
		t.Fatalf("next %v should be >= base interval", next)
	}
	if next > 8*time.Second {
		t.Fatalf("next %v should be <= base + jitter", next)
	}
}

func TestTicker_Fires(t *testing.T) {
	s := New(Config{IntervalSeconds: 0}) // defaults to 30s but we override below
	s.interval = 50 * time.Millisecond
	s.jitter = 0
	tk := s.Ticker()
	defer tk.Stop()
	select {
	case <-tk.C:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("ticker did not fire")
	}
}
