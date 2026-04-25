package quorum

import (
	"testing"
)

func TestObserve_NotMetBeforeRequired(t *testing.T) {
	q := New(Config{Required: 3, Window: 10})
	if q.Observe("k", "open") {
		t.Fatal("expected false on first observation")
	}
	if q.Observe("k", "open") {
		t.Fatal("expected false on second observation")
	}
}

func TestObserve_MetAfterRequired(t *testing.T) {
	q := New(Config{Required: 3, Window: 10})
	q.Observe("k", "open")
	q.Observe("k", "open")
	if !q.Observe("k", "open") {
		t.Fatal("expected true after three matching observations")
	}
}

func TestObserve_MismatchResetsConsensus(t *testing.T) {
	q := New(Config{Required: 3, Window: 10})
	q.Observe("k", "open")
	q.Observe("k", "open")
	q.Observe("k", "closed") // breaks the run
	if q.Observe("k", "open") {
		t.Fatal("expected false after mismatch interrupts run")
	}
}

func TestObserve_IndependentKeys(t *testing.T) {
	q := New(Config{Required: 2, Window: 10})
	q.Observe("a", "open")
	q.Observe("a", "open") // a reaches quorum

	// key b has only one observation – must not be in quorum
	if q.Observe("b", "open") {
		t.Fatal("key b should not yet have quorum")
	}
}

func TestObserve_WindowEvictsOldEntries(t *testing.T) {
	q := New(Config{Required: 3, Window: 4})
	// Fill window with "closed"
	for i := 0; i < 4; i++ {
		q.Observe("k", "closed")
	}
	// Now push three "open" – old "closed" entries should be evicted
	q.Observe("k", "open")
	q.Observe("k", "open")
	if !q.Observe("k", "open") {
		t.Fatal("expected quorum after window eviction")
	}
}

func TestReset_ClearsObservations(t *testing.T) {
	q := New(Config{Required: 2, Window: 10})
	q.Observe("k", "open")
	q.Reset("k")
	if q.Count("k") != 0 {
		t.Fatalf("expected count 0 after reset, got %d", q.Count("k"))
	}
	if q.Observe("k", "open") {
		t.Fatal("expected false immediately after reset")
	}
}

func TestDefaults_Applied(t *testing.T) {
	q := New(Config{})
	if q.cfg.Required != 3 {
		t.Fatalf("expected default required=3, got %d", q.cfg.Required)
	}
	if q.cfg.Window != 10 {
		t.Fatalf("expected default window=10, got %d", q.cfg.Window)
	}
}

func TestString_ContainsFields(t *testing.T) {
	q := New(Config{Required: 2, Window: 5})
	s := q.String()
	for _, want := range []string{"required=2", "window=5", "keys=0"} {
		if !contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
