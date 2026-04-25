package mute_test

import (
	"testing"
	"time"

	"portwatch/internal/mute"
)

func fixedClock(t time.Time) mute.Clock {
	return func() time.Time { return t }
}

func TestIsMuted_FalseBeforeSilence(t *testing.T) {
	m := mute.New()
	if m.IsMuted("host:9000") {
		t.Fatal("expected not muted before any silence call")
	}
}

func TestIsMuted_TrueWithinWindow(t *testing.T) {
	now := time.Now()
	m := mute.WithClock(fixedClock(now))
	m.Silence("host:9000", now.Add(10*time.Minute), "maintenance")
	if !m.IsMuted("host:9000") {
		t.Fatal("expected key to be muted within window")
	}
}

func TestIsMuted_FalseAfterExpiry(t *testing.T) {
	now := time.Now()
	m := mute.WithClock(fixedClock(now.Add(time.Hour)))
	m.Silence("host:9000", now.Add(5*time.Minute), "maintenance")
	if m.IsMuted("host:9000") {
		t.Fatal("expected key to be unmuted after expiry")
	}
}

func TestUnmute_RemovesKey(t *testing.T) {
	now := time.Now()
	m := mute.WithClock(fixedClock(now))
	m.Silence("host:9000", now.Add(time.Hour), "")
	m.Unmute("host:9000")
	if m.IsMuted("host:9000") {
		t.Fatal("expected key to be removed after Unmute")
	}
}

func TestGet_ReturnsMuteEntry(t *testing.T) {
	now := time.Now()
	m := mute.WithClock(fixedClock(now))
	until := now.Add(30 * time.Minute)
	m.Silence("host:9000", until, "planned downtime")
	e, ok := m.Get("host:9000")
	if !ok {
		t.Fatal("expected Get to return entry")
	}
	if e.Reason != "planned downtime" {
		t.Fatalf("expected reason 'planned downtime', got %q", e.Reason)
	}
	if !e.Until.Equal(until) {
		t.Fatalf("expected Until %v, got %v", until, e.Until)
	}
}

func TestGet_MissingKey(t *testing.T) {
	m := mute.New()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected Get to return false for missing key")
	}
}

func TestKeys_ReturnsActiveMuted(t *testing.T) {
	now := time.Now()
	m := mute.WithClock(fixedClock(now))
	m.Silence("a", now.Add(time.Hour), "")
	m.Silence("b", now.Add(time.Hour), "")
	m.Silence("c", now.Add(-time.Second), "") // already expired
	keys := m.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 active keys, got %d", len(keys))
	}
}

func TestKeys_IndependentKeys(t *testing.T) {
	now := time.Now()
	m := mute.WithClock(fixedClock(now))
	m.Silence("x", now.Add(time.Minute), "")
	m.Silence("y", now.Add(time.Minute), "")
	m.Unmute("x")
	keys := m.Keys()
	for _, k := range keys {
		if k == "x" {
			t.Fatal("expected x to be removed from Keys after Unmute")
		}
	}
}
