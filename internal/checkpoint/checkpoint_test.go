package checkpoint

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsDirty_TrueWhenNoCheckpoint(t *testing.T) {
	c := New()
	if !c.IsDirty("k", "v") {
		t.Fatal("expected dirty for unknown key")
	}
}

func TestIsDirty_FalseAfterSave(t *testing.T) {
	c := New()
	c.Save("k", "v")
	if c.IsDirty("k", "v") {
		t.Fatal("expected clean after save")
	}
}

func TestIsDirty_TrueWhenValueChanges(t *testing.T) {
	c := New()
	c.Save("k", "v1")
	if !c.IsDirty("k", "v2") {
		t.Fatal("expected dirty after value change")
	}
}

func TestSave_IncrementsFlushCount(t *testing.T) {
	c := New()
	c.Save("k", "a")
	c.Save("k", "b")
	c.Save("k", "b")
	e, ok := c.Get("k")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.Flushes != 3 {
		t.Fatalf("expected 3 flushes, got %d", e.Flushes)
	}
}

func TestSave_RecordsSavedAt(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	c := WithClock(New(), fixedClock(now))
	c.Save("k", "v")
	e, _ := c.Get("k")
	if !e.SavedAt.Equal(now) {
		t.Fatalf("expected SavedAt %v, got %v", now, e.SavedAt)
	}
}

func TestGet_MissingKey(t *testing.T) {
	c := New()
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestReset_MakesDirtyAgain(t *testing.T) {
	c := New()
	c.Save("k", "v")
	c.Reset("k")
	if !c.IsDirty("k", "v") {
		t.Fatal("expected dirty after reset")
	}
}

func TestKeys_ReturnsAllSaved(t *testing.T) {
	c := New()
	c.Save("a", "1")
	c.Save("b", "2")
	keys := c.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestKeys_EmptyWhenNoneStored(t *testing.T) {
	c := New()
	if len(c.Keys()) != 0 {
		t.Fatal("expected no keys")
	}
}

func TestIsDirty_IndependentKeys(t *testing.T) {
	c := New()
	c.Save("x", "hello")
	if !c.IsDirty("y", "hello") {
		t.Fatal("key y should be independently dirty")
	}
	if c.IsDirty("x", "hello") {
		t.Fatal("key x should be clean")
	}
}
