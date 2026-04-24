package escalation_test

import (
	"testing"
	"time"

	"portwatch/internal/escalation"
)

func defaultConfig() escalation.Config {
	return escalation.Config{
		Window: 10 * time.Second,
		Levels: []escalation.Level{
			{Name: "warn", Threshold: 2},
			{Name: "critical", Threshold: 5},
		},
	}
}

func TestLevel_NoneBeforeThreshold(t *testing.T) {
	e := escalation.New(defaultConfig())
	now := time.Now()
	e.Record("host:80", now)
	if got := e.Level("host:80", now); got != "" {
		t.Fatalf("expected empty level, got %q", got)
	}
}

func TestLevel_WarnAfterTwoFailures(t *testing.T) {
	e := escalation.New(defaultConfig())
	now := time.Now()
	e.Record("host:80", now)
	e.Record("host:80", now.Add(time.Second))
	if got := e.Level("host:80", now.Add(2*time.Second)); got != "warn" {
		t.Fatalf("expected warn, got %q", got)
	}
}

func TestLevel_CriticalAfterFiveFailures(t *testing.T) {
	e := escalation.New(defaultConfig())
	now := time.Now()
	for i := 0; i < 5; i++ {
		e.Record("host:80", now.Add(time.Duration(i)*time.Second))
	}
	if got := e.Level("host:80", now.Add(5*time.Second)); got != "critical" {
		t.Fatalf("expected critical, got %q", got)
	}
}

func TestLevel_EvictsOldFailures(t *testing.T) {
	e := escalation.New(defaultConfig())
	base := time.Now()
	// Record 5 failures in the distant past
	for i := 0; i < 5; i++ {
		e.Record("host:80", base)
	}
	// Advance beyond the window — all should be evicted
	now := base.Add(20 * time.Second)
	if got := e.Level("host:80", now); got != "" {
		t.Fatalf("expected empty after eviction, got %q", got)
	}
}

func TestLevel_IndependentKeys(t *testing.T) {
	e := escalation.New(defaultConfig())
	now := time.Now()
	for i := 0; i < 5; i++ {
		e.Record("a:80", now)
	}
	if got := e.Level("b:80", now); got != "" {
		t.Fatalf("key b should be unaffected, got %q", got)
	}
}

func TestReset_ClearsHistory(t *testing.T) {
	e := escalation.New(defaultConfig())
	now := time.Now()
	for i := 0; i < 5; i++ {
		e.Record("host:80", now)
	}
	e.Reset("host:80")
	if got := e.Level("host:80", now); got != "" {
		t.Fatalf("expected empty after reset, got %q", got)
	}
}

func TestLevel_UnknownKey(t *testing.T) {
	e := escalation.New(defaultConfig())
	if got := e.Level("unknown:9999", time.Now()); got != "" {
		t.Fatalf("expected empty for unknown key, got %q", got)
	}
}
