package signal

import (
	"testing"
)

func TestObserve_FirstCall_EdgeNone(t *testing.T) {
	d := New()
	ev := d.Observe("port:8080", true)
	if ev.Edge != EdgeNone {
		t.Fatalf("expected EdgeNone on first observation, got %s", ev.Edge)
	}
	if !ev.Current {
		t.Fatalf("expected Current=true")
	}
}

func TestObserve_RisingEdge(t *testing.T) {
	d := New()
	d.Observe("k", false)
	ev := d.Observe("k", true)
	if ev.Edge != EdgeRising {
		t.Fatalf("expected EdgeRising, got %s", ev.Edge)
	}
	if ev.Prev != false || ev.Current != true {
		t.Fatalf("unexpected prev/current: %v/%v", ev.Prev, ev.Current)
	}
}

func TestObserve_FallingEdge(t *testing.T) {
	d := New()
	d.Observe("k", true)
	ev := d.Observe("k", false)
	if ev.Edge != EdgeFalling {
		t.Fatalf("expected EdgeFalling, got %s", ev.Edge)
	}
}

func TestObserve_NoChange_EdgeNone(t *testing.T) {
	d := New()
	d.Observe("k", true)
	ev := d.Observe("k", true)
	if ev.Edge != EdgeNone {
		t.Fatalf("expected EdgeNone on stable state, got %s", ev.Edge)
	}
}

func TestObserve_IndependentKeys(t *testing.T) {
	d := New()
	d.Observe("a", false)
	d.Observe("b", true)

	evA := d.Observe("a", true)  // rising for a
	evB := d.Observe("b", true)  // no change for b

	if evA.Edge != EdgeRising {
		t.Fatalf("expected EdgeRising for key a, got %s", evA.Edge)
	}
	if evB.Edge != EdgeNone {
		t.Fatalf("expected EdgeNone for key b, got %s", evB.Edge)
	}
}

func TestReset_ClearsState(t *testing.T) {
	d := New()
	d.Observe("k", true)
	d.Reset("k")

	// After reset, next observation should behave like first call.
	ev := d.Observe("k", false)
	if ev.Edge != EdgeNone {
		t.Fatalf("expected EdgeNone after reset, got %s", ev.Edge)
	}
}

func TestKeys_ReturnsTrackedKeys(t *testing.T) {
	d := New()
	d.Observe("x", true)
	d.Observe("y", false)

	keys := d.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestEdge_String(t *testing.T) {
	cases := []struct {
		edge Edge
		want string
	}{
		{EdgeNone, "none"},
		{EdgeRising, "rising"},
		{EdgeFalling, "falling"},
	}
	for _, tc := range cases {
		if got := tc.edge.String(); got != tc.want {
			t.Errorf("Edge(%d).String() = %q, want %q", tc.edge, got, tc.want)
		}
	}
}
