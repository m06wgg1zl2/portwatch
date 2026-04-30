package overflow_test

import (
	"testing"

	"portwatch/internal/overflow"
)

func TestRecord_IncrementsCount(t *testing.T) {
	tr := overflow.New(overflow.Config{})
	tr.Record("q1", 3)
	tr.Record("q1", 2)
	e := tr.Get("q1")
	if e.Dropped != 5 {
		t.Fatalf("expected 5 dropped, got %d", e.Dropped)
	}
}

func TestGet_ZeroForUnknownKey(t *testing.T) {
	tr := overflow.New(overflow.Config{})
	e := tr.Get("missing")
	if e.Dropped != 0 {
		t.Fatalf("expected 0, got %d", e.Dropped)
	}
	if e.Key != "missing" {
		t.Fatalf("expected key 'missing', got %q", e.Key)
	}
}

func TestRecord_SetsLastDropAt(t *testing.T) {
	tr := overflow.New(overflow.Config{})
	tr.Record("q1", 1)
	e := tr.Get("q1")
	if e.LastDropAt.IsZero() {
		t.Fatal("expected LastDropAt to be set")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	tr := overflow.New(overflow.Config{})
	tr.Record("q1", 10)
	tr.Reset("q1")
	e := tr.Get("q1")
	if e.Dropped != 0 {
		t.Fatalf("expected 0 after reset, got %d", e.Dropped)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	tr := overflow.New(overflow.Config{})
	tr.Record("a", 1)
	tr.Record("b", 2)
	all := tr.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestMaxKeys_IgnoresNewKeysAtLimit(t *testing.T) {
	tr := overflow.New(overflow.Config{MaxKeys: 2})
	tr.Record("a", 1)
	tr.Record("b", 1)
	tr.Record("c", 5) // should be silently ignored
	if e := tr.Get("c"); e.Dropped != 0 {
		t.Fatalf("expected key 'c' to be ignored, got %d drops", e.Dropped)
	}
	if len(tr.All()) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(tr.All()))
	}
}

func TestRecord_NonPositiveIsNoop(t *testing.T) {
	tr := overflow.New(overflow.Config{})
	tr.Record("x", 0)
	tr.Record("x", -3)
	if e := tr.Get("x"); e.Dropped != 0 {
		t.Fatalf("expected 0 for non-positive record, got %d", e.Dropped)
	}
}
