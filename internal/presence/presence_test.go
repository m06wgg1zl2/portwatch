package presence

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestCheck_AbsentBeforePing(t *testing.T) {
	tr := New(Config{TTL: time.Second})
	if got := tr.Check("host:80"); got != Absent {
		t.Fatalf("expected Absent before ping, got %s", got)
	}
}

func TestCheck_PresentAfterPing(t *testing.T) {
	tr := New(Config{TTL: time.Second})
	tr.Ping("host:80")
	if got := tr.Check("host:80"); got != Present {
		t.Fatalf("expected Present after ping, got %s", got)
	}
}

func TestCheck_AbsentAfterTTLExpires(t *testing.T) {
	now := time.Now()
	tr := New(Config{TTL: 5 * time.Second})
	tr.now = fixedClock(now)
	tr.Ping("host:80")

	// advance clock beyond TTL
	tr.now = fixedClock(now.Add(6 * time.Second))
	if got := tr.Check("host:80"); got != Absent {
		t.Fatalf("expected Absent after TTL, got %s", got)
	}
}

func TestCheck_PresentWithinTTL(t *testing.T) {
	now := time.Now()
	tr := New(Config{TTL: 10 * time.Second})
	tr.now = fixedClock(now)
	tr.Ping("host:80")

	tr.now = fixedClock(now.Add(9 * time.Second))
	if got := tr.Check("host:80"); got != Present {
		t.Fatalf("expected Present within TTL, got %s", got)
	}
}

func TestForget_RemovesKey(t *testing.T) {
	tr := New(Config{TTL: time.Minute})
	tr.Ping("host:80")
	tr.Forget("host:80")
	if got := tr.Check("host:80"); got != Absent {
		t.Fatalf("expected Absent after Forget, got %s", got)
	}
}

func TestKeys_ReturnsTrackedKeys(t *testing.T) {
	tr := New(Config{TTL: time.Minute})
	tr.Ping("a:1")
	tr.Ping("b:2")
	keys := tr.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestDefaultTTL_Applied(t *testing.T) {
	tr := New(Config{})
	if tr.cfg.TTL != 30*time.Second {
		t.Fatalf("expected default TTL 30s, got %s", tr.cfg.TTL)
	}
}

func TestStatusString(t *testing.T) {
	if Present.String() != "present" {
		t.Fatalf("unexpected string for Present")
	}
	if Absent.String() != "absent" {
		t.Fatalf("unexpected string for Absent")
	}
}
