package circuit

import (
	"testing"
	"time"
)

func TestAllow_ClosedByDefault(t *testing.T) {
	b := New(Config{})
	if !b.Allow() {
		t.Fatal("expected allow in closed state")
	}
}

func TestRecordFailure_OpensAfterThreshold(t *testing.T) {
	b := New(Config{FailureThreshold: 2})
	b.RecordFailure()
	if b.State() != StateClosed {
		t.Fatal("should still be closed after 1 failure")
	}
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatalf("expected open, got %s", b.State())
	}
	if b.Allow() {
		t.Fatal("should not allow when open")
	}
}

func TestOpenTimeout_TransitionsToHalfOpen(t *testing.T) {
	b := New(Config{FailureThreshold: 1, OpenTimeout: 20 * time.Millisecond})
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatal("expected open")
	}
	time.Sleep(30 * time.Millisecond)
	if !b.Allow() {
		t.Fatal("expected allow after timeout")
	}
	if b.State() != StateHalfOpen {
		t.Fatalf("expected half-open, got %s", b.State())
	}
}

func TestHalfOpen_SuccessCloses(t *testing.T) {
	b := New(Config{FailureThreshold: 1, SuccessThreshold: 1, OpenTimeout: 10 * time.Millisecond})
	b.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	b.Allow() // transition to half-open
	b.RecordSuccess()
	if b.State() != StateClosed {
		t.Fatalf("expected closed after success in half-open, got %s", b.State())
	}
}

func TestHalfOpen_FailureReopens(t *testing.T) {
	b := New(Config{FailureThreshold: 1, OpenTimeout: 10 * time.Millisecond})
	b.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	b.Allow()
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatalf("expected open after failure in half-open, got %s", b.State())
	}
}

func TestStateString(t *testing.T) {
	cases := map[State]string{
		StateClosed:   "closed",
		StateOpen:     "open",
		StateHalfOpen: "half-open",
		State(99):     "unknown",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", s, got, want)
		}
	}
}
