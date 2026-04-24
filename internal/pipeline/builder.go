package pipeline

import (
	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/cooldown"
	"github.com/user/portwatch/internal/filter"
	"github.com/user/portwatch/internal/sampler"
)

// filterStage wraps a *filter.Filter as a pipeline Stage.
type filterStage struct{ f *filter.Filter }

func (fs filterStage) Name() string             { return "time-filter" }
func (fs filterStage) Allow(a alert.Alert) bool { return fs.f.Allow(a.Host) }

// cooldownStage wraps a *cooldown.Cooldown as a pipeline Stage.
type cooldownStage struct{ c *cooldown.Cooldown }

func (cs cooldownStage) Name() string             { return "cooldown" }
func (cs cooldownStage) Allow(a alert.Alert) bool { return cs.c.Allow(a.Host) }

// samplerStage wraps a *sampler.Sampler as a pipeline Stage.
type samplerStage struct{ s *sampler.Sampler }

func (ss samplerStage) Name() string             { return "sampler" }
func (ss samplerStage) Allow(_ alert.Alert) bool { return ss.s.Allow() }

// Builder constructs a Pipeline from well-known components.
type Builder struct {
	stages []Stage
}

// NewBuilder creates an empty Builder.
func NewBuilder() *Builder { return &Builder{} }

// WithFilter appends a time-window filter stage.
func (b *Builder) WithFilter(f *filter.Filter) *Builder {
	b.stages = append(b.stages, filterStage{f})
	return b
}

// WithCooldown appends a per-key cooldown stage.
func (b *Builder) WithCooldown(c *cooldown.Cooldown) *Builder {
	b.stages = append(b.stages, cooldownStage{c})
	return b
}

// WithSampler appends a probabilistic sampling stage.
func (b *Builder) WithSampler(s *sampler.Sampler) *Builder {
	b.stages = append(b.stages, samplerStage{s})
	return b
}

// WithStage appends an arbitrary custom stage.
func (b *Builder) WithStage(s Stage) *Builder {
	b.stages = append(b.stages, s)
	return b
}

// Build returns the assembled Pipeline.
func (b *Builder) Build() *Pipeline { return New(b.stages...) }
