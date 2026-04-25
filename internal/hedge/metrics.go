package hedge

import "sync/atomic"

// Metrics tracks runtime counters for a Hedger.
type Metrics struct {
	total    atomic.Int64
	hedged   atomic.Int64
	success  atomic.Int64
	failures atomic.Int64
}

// RecordAttempt increments the total-attempts counter.
func (m *Metrics) RecordAttempt() { m.total.Add(1) }

// RecordHedge increments the hedged-attempts counter.
func (m *Metrics) RecordHedge() { m.hedged.Add(1) }

// RecordSuccess increments the success counter.
func (m *Metrics) RecordSuccess() { m.success.Add(1) }

// RecordFailure increments the failure counter.
func (m *Metrics) RecordFailure() { m.failures.Add(1) }

// Snapshot returns a point-in-time copy of all counters.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		Total:    m.total.Load(),
		Hedged:   m.hedged.Load(),
		Success:  m.success.Load(),
		Failures: m.failures.Load(),
	}
}

// MetricsSnapshot is an immutable copy of Metrics counters.
type MetricsSnapshot struct {
	Total    int64
	Hedged   int64
	Success  int64
	Failures int64
}

// HedgeRate returns the fraction of attempts that triggered a hedge.
// Returns 0 when Total is zero.
func (s MetricsSnapshot) HedgeRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Hedged) / float64(s.Total)
}
