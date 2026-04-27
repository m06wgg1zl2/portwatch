package cascade_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/cascade"
)

func makeAlert(host string, port int, state alert.Level) alert.Alert {
	return alert.New(host, port, state, "test")
}

func defaultConfig() cascade.Config {
	return cascade.Config{
		Window:   200 * time.Millisecond,
		MinCount: 2,
		MaxDepth: 3,
	}
}

func TestCascade_BelowMinCount(t *testing.T) {
	c := cascade.New(defaultConfig())
	a := makeAlert("host-a", 8080, alert.LevelCritical)

	// Only one event — should not trigger cascade
	if c.Triggered(a.Host()) {
		t.Fatal("expected no cascade before min count reached")
	}
	c.Record(a)
	if c.Triggered(a.Host()) {
		t.Fatal("expected no cascade with only one event")
	}
}

func TestCascade_TriggeredAfterMinCount(t *testing.T) {
	c := cascade.New(defaultConfig())
	a := makeAlert("host-b", 9090, alert.LevelCritical)

	c.Record(a)
	c.Record(a)

	if !c.Triggered(a.Host()) {
		t.Fatal("expected cascade to trigger after min count")
	}
}

func TestCascade_WindowEvictsOldEvents(t *testing.T) {
	cfg := cascade.Config{
		Window:   50 * time.Millisecond,
		MinCount: 2,
		MaxDepth: 3,
	}
	c := cascade.New(cfg)
	a := makeAlert("host-c", 7070, alert.LevelCritical)

	c.Record(a)
	time.Sleep(80 * time.Millisecond)
	c.Record(a)

	// First event expired; only one active in window
	if c.Triggered(a.Host()) {
		t.Fatal("expected no cascade after window eviction")
	}
}

func TestCascade_IndependentHosts(t *testing.T) {
	c := cascade.New(defaultConfig())
	a1 := makeAlert("host-x", 8080, alert.LevelCritical)
	a2 := makeAlert("host-y", 8080, alert.LevelCritical)

	c.Record(a1)
	c.Record(a1)

	if !c.Triggered(a1.Host()) {
		t.Fatal("expected cascade for host-x")
	}
	if c.Triggered(a2.Host()) {
		t.Fatal("expected no cascade for host-y")
	}
}

func TestCascade_Depth_DoesNotExceedMax(t *testing.T) {
	cfg := cascade.Config{
		Window:   200 * time.Millisecond,
		MinCount: 2,
		MaxDepth: 2,
	}
	c := cascade.New(cfg)
	a := makeAlert("host-d", 5050, alert.LevelCritical)

	for i := 0; i < 10; i++ {
		c.Record(a)
	}

	depth := c.Depth(a.Host())
	if depth > cfg.MaxDepth {
		t.Fatalf("expected depth <= %d, got %d", cfg.MaxDepth, depth)
	}
}

func TestCascade_Reset_ClearsHost(t *testing.T) {
	c := cascade.New(defaultConfig())
	a := makeAlert("host-e", 3030, alert.LevelCritical)

	c.Record(a)
	c.Record(a)

	if !c.Triggered(a.Host()) {
		t.Fatal("expected cascade before reset")
	}

	c.Reset(a.Host())

	if c.Triggered(a.Host()) {
		t.Fatal("expected no cascade after reset")
	}
}

func TestCascade_Depth_ZeroForUnknownHost(t *testing.T) {
	c := cascade.New(defaultConfig())
	if d := c.Depth("unknown"); d != 0 {
		t.Fatalf("expected depth 0 for unknown host, got %d", d)
	}
}
