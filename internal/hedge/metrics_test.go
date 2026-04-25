package hedge_test

import (
	"testing"

	"portwatch/internal/hedge"
)

func TestMetrics_InitialZero(t *testing.T) {
	var m hedge.Metrics
	s := m.Snapshot()
	if s.Total != 0 || s.Hedged != 0 || s.Success != 0 || s.Failures != 0 {
		t.Fatalf("expected all zeros, got %+v", s)
	}
}

func TestMetrics_Increments(t *testing.T) {
	var m hedge.Metrics
	m.RecordAttempt()
	m.RecordAttempt()
	m.RecordHedge()
	m.RecordSuccess()
	m.RecordFailure()

	s := m.Snapshot()
	if s.Total != 2 {
		t.Fatalf("want Total=2, got %d", s.Total)
	}
	if s.Hedged != 1 {
		t.Fatalf("want Hedged=1, got %d", s.Hedged)
	}
	if s.Success != 1 {
		t.Fatalf("want Success=1, got %d", s.Success)
	}
	if s.Failures != 1 {
		t.Fatalf("want Failures=1, got %d", s.Failures)
	}
}

func TestMetrics_HedgeRate(t *testing.T) {
	var m hedge.Metrics
	m.RecordAttempt()
	m.RecordAttempt()
	m.RecordHedge()

	s := m.Snapshot()
	got := s.HedgeRate()
	if got != 0.5 {
		t.Fatalf("want 0.5, got %f", got)
	}
}

func TestMetrics_HedgeRate_ZeroTotal(t *testing.T) {
	var m hedge.Metrics
	s := m.Snapshot()
	if s.HedgeRate() != 0 {
		t.Fatal("expected 0 rate when total is zero")
	}
}

func TestMetrics_Snapshot_IsIndependent(t *testing.T) {
	var m hedge.Metrics
	m.RecordAttempt()
	s1 := m.Snapshot()
	m.RecordAttempt()
	s2 := m.Snapshot()
	if s1.Total == s2.Total {
		t.Fatal("snapshots should be independent")
	}
}
