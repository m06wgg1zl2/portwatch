package watchdog

import (
	"testing"
	"time"
)

func TestStatus_StaleWhenNoBeat(t *testing.T) {
	w := New(100 * time.Millisecond)
	if w.Status() != StatusStale {
		t.Fatal("expected stale before any beat")
	}
}

func TestStatus_HealthyAfterBeat(t *testing.T) {
	w := New(100 * time.Millisecond)
	w.Beat()
	if w.Status() != StatusHealthy {
		t.Fatal("expected healthy immediately after beat")
	}
}

func TestStatus_StaleAfterTimeout(t *testing.T) {
	now := time.Now()
	w := New(50 * time.Millisecond)
	w.now = func() time.Time { return now }
	w.Beat()
	w.now = func() time.Time { return now.Add(100 * time.Millisecond) }
	if w.Status() != StatusStale {
		t.Fatal("expected stale after timeout")
	}
}

func TestReset_ForcesStale(t *testing.T) {
	w := New(time.Second)
	w.Beat()
	if w.Status() != StatusHealthy {
		t.Fatal("expected healthy after beat")
	}
	w.Reset()
	if w.Status() != StatusStale {
		t.Fatal("expected stale after reset")
	}
}

func TestLastBeat_ZeroBeforeAnyBeat(t *testing.T) {
	w := New(time.Second)
	if !w.LastBeat().IsZero() {
		t.Fatal("expected zero time before any beat")
	}
}

func TestLastBeat_RecordsTime{
	fixed := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	w := New(time.Second)
	w.now = func() time.Time { return fixed }
	w.Beat()
	if !w.LastBeat().Equal(fixed) {
		t.Fatalf("expected %v, got %v", fixed, w.LastBeat())
	}
}

func TestStatusString(t *testing.T) {
	if StatusHealthy.String() != "healthy" {
		t.Errorf("unexpected: %s", StatusHealthy.String())
	}
	if StatusStale.String() != "stale" {
		t.Errorf("unexpected: %s", StatusStale.String())
	}
}
