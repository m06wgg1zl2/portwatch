// Package pipeline chains multiple middleware-style handlers that each decide
// whether to allow a notification to proceed. If any stage returns false the
// pipeline is halted and the alert is suppressed.
package pipeline

import "github.com/user/portwatch/internal/alert"

// Stage is a single processing step in the pipeline.
type Stage interface {
	// Name returns a human-readable label used in logs and reports.
	Name() string
	// Allow returns true when the alert may continue through the pipeline.
	Allow(a alert.Alert) bool
}

// Pipeline executes a sequence of Stages in order.
type Pipeline struct {
	stages []Stage
}

// New creates a Pipeline with the supplied stages.
func New(stages ...Stage) *Pipeline {
	s := make([]Stage, len(stages))
	copy(s, stages)
	return &Pipeline{stages: s}
}

// Run passes the alert through every stage. It returns the name of the first
// stage that blocked the alert, or an empty string when all stages passed.
func (p *Pipeline) Run(a alert.Alert) (blockedBy string) {
	for _, s := range p.stages {
		if !s.Allow(a) {
			return s.Name()
		}
	}
	return ""
}

// Stages returns a snapshot of the current stage list.
func (p *Pipeline) Stages() []string {
	names := make([]string, len(p.stages))
	for i, s := range p.stages {
		names[i] = s.Name()
	}
	return names
}

// Len returns the number of stages in the pipeline.
func (p *Pipeline) Len() int { return len(p.stages) }
