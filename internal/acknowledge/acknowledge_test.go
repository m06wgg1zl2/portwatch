package acknowledge

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsAcknowledged_FalseBeforeAck(t *testing.T) {
	s := New(time.Minute)
	if s.IsAcknowledged("host:8080") {
		t.Fatal("expected not acknowledged")
	}
}

func TestIsAcknowledged_TrueAfterAck(t *testing.T) {
	s := New(time.Minute)
	s.Acknowledge("host:8080", "ops")
	if !s.IsAcknowledged("host:8080") {
		t.Fatal("expected acknowledged")
	}
}

func TestIsAcknowledged_FalseAfterExpiry(t *testing.T) {
	now := time.Now()
	s := WithClock(New(time.Second), fixedClock(now))
	s.Acknowledge("host:8080", "ops")
	// Advance clock past TTL
	WithClock(s, fixedClock(now.Add(2*time.Second)))
	if s.IsAcknowledged("host:8080") {
		t.Fatal("expected acknowledgement to have expired")
	}
}

func TestGet_ReturnsEntryFields(t *testing.T) {
	now := time.Now()
	s := WithClock(New(time.Minute), fixedClock(now))
	s.Acknowledge("host:9090", "alice")
	e, ok := s.Get("host:9090")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.AcknowledgedBy != "alice" {
		t.Errorf("got AcknowledgedBy=%q, want %q", e.AcknowledgedBy, "alice")
	}
	if !e.AcknowledgedAt.Equal(now) {
		t.Errorf("unexpected AcknowledgedAt")
	}
}

func TestGet_MissingKey(t *testing.T) {
	s := New(time.Minute)
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestUnacknowledge_RemovesKey(t *testing.T) {
	s := New(time.Minute)
	s.Acknowledge("host:8080", "ops")
	s.Unacknowledge("host:8080")
	if s.IsAcknowledged("host:8080") {
		t.Fatal("expected key to be removed")
	}
}

func TestKeys_ReturnsActiveKeys(t *testing.T) {
	s := New(time.Minute)
	s.Acknowledge("a", "ops")
	s.Acknowledge("b", "ops")
	keys := s.Keys()
	if len(keys) != 2 {
		t.Errorf("got %d keys, want 2", len(keys))
	}
}

func TestKeys_ExcludesExpired(t *testing.T) {
	now := time.Now()
	s := WithClock(New(time.Second), fixedClock(now))
	s.Acknowledge("expired", "ops")
	WithClock(s, fixedClock(now.Add(2*time.Second)))
	s.Acknowledge("active", "ops")
	keys := s.Keys()
	if len(keys) != 1 || keys[0] != "active" {
		t.Errorf("expected only 'active', got %v", keys)
	}
}
