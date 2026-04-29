package heartbeat_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/heartbeat"
)

func TestStatus_DeadBeforeAnyBeat(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{Interval: 100 * time.Millisecond, MissedThreshold: 2})
	if got := h.Check(); got != heartbeat.StatusDead {
		t.Fatalf("expected Dead before any beat, got %s", got)
	}
}

func TestStatus_HealthyAfterBeat(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{Interval: 5 * time.Second, MissedThreshold: 2})
	h.Beat()
	if got := h.Check(); got != heartbeat.StatusHealthy {
		t.Fatalf("expected Healthy immediately after beat, got %s", got)
	}
}

func TestStatus_DegradedAfterOneMissed(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{Interval: 20 * time.Millisecond, MissedThreshold: 3})
	h.Beat()
	time.Sleep(45 * time.Millisecond) // ~2 missed intervals
	got := h.Check()
	if got != heartbeat.StatusDegraded {
		t.Fatalf("expected Degraded, got %s", got)
	}
}

func TestStatus_DeadAfterThresholdMissed(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{Interval: 20 * time.Millisecond, MissedThreshold: 2})
	h.Beat()
	time.Sleep(60 * time.Millisecond) // >3 missed intervals
	got := h.Check()
	if got != heartbeat.StatusDead {
		t.Fatalf("expected Dead, got %s", got)
	}
}

func TestBeat_IncrementsCount(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{})
	for i := 0; i < 5; i++ {
		h.Beat()
	}
	if got := h.Beats(); got != 5 {
		t.Fatalf("expected 5 beats, got %d", got)
	}
}

func TestLastBeat_UpdatedOnBeat(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{})
	before := time.Now()
	h.Beat()
	after := time.Now()
	lb := h.LastBeat()
	if lb.Before(before) || lb.After(after) {
		t.Fatalf("LastBeat %v not in expected range [%v, %v]", lb, before, after)
	}
}

func TestStatusString(t *testing.T) {
	cases := []struct {
		s    heartbeat.Status
		want string
	}{
		{heartbeat.StatusHealthy, "healthy"},
		{heartbeat.StatusDegraded, "degraded"},
		{heartbeat.StatusDead, "dead"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("Status(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestDefaults_Applied(t *testing.T) {
	h := heartbeat.New(heartbeat.Config{})
	h.Beat()
	// With default 30s interval, immediately after a beat should be Healthy.
	if got := h.Check(); got != heartbeat.StatusHealthy {
		t.Fatalf("expected Healthy with defaults, got %s", got)
	}
}
