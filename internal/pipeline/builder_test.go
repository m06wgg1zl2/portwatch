package pipeline_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/cooldown"
	"github.com/user/portwatch/internal/filter"
	"github.com/user/portwatch/internal/pipeline"
	"github.com/user/portwatch/internal/sampler"
)

func TestBuilder_EmptyBuild(t *testing.T) {
	p := pipeline.NewBuilder().Build()
	if p.Len() != 0 {
		t.Fatalf("expected 0 stages, got %d", p.Len())
	}
}

func TestBuilder_WithFilter(t *testing.T) {
	f := filter.New(filter.Config{})
	p := pipeline.NewBuilder().WithFilter(f).Build()
	if p.Len() != 1 {
		t.Fatalf("expected 1 stage, got %d", p.Len())
	}
	if p.Stages()[0] != "time-filter" {
		t.Fatalf("unexpected stage name: %s", p.Stages()[0])
	}
}

func TestBuilder_WithCooldown(t *testing.T) {
	c := cooldown.New(cooldown.Config{Window: time.Second})
	p := pipeline.NewBuilder().WithCooldown(c).Build()
	if p.Len() != 1 {
		t.Fatalf("expected 1 stage, got %d", p.Len())
	}
	if p.Stages()[0] != "cooldown" {
		t.Fatalf("unexpected stage name: %s", p.Stages()[0])
	}
}

func TestBuilder_WithSampler(t *testing.T) {
	s := sampler.New(sampler.Config{Rate: 1.0})
	p := pipeline.NewBuilder().WithSampler(s).Build()
	if p.Len() != 1 {
		t.Fatalf("expected 1 stage, got %d", p.Len())
	}
	if p.Stages()[0] != "sampler" {
		t.Fatalf("unexpected stage name: %s", p.Stages()[0])
	}
}

func TestBuilder_ChainedStages(t *testing.T) {
	f := filter.New(filter.Config{})
	c := cooldown.New(cooldown.Config{Window: time.Second})
	s := sampler.New(sampler.Config{Rate: 1.0})
	p := pipeline.NewBuilder().WithFilter(f).WithCooldown(c).WithSampler(s).Build()
	if p.Len() != 3 {
		t.Fatalf("expected 3 stages, got %d", p.Len())
	}
	names := p.Stages()
	if names[0] != "time-filter" || names[1] != "cooldown" || names[2] != "sampler" {
		t.Fatalf("unexpected stage order: %v", names)
	}
}

func TestBuilder_WithCustomStage(t *testing.T) {
	p := pipeline.NewBuilder().WithStage(allowStage{"custom"}).Build()
	if p.Len() != 1 || p.Stages()[0] != "custom" {
		t.Fatalf("unexpected pipeline: %v", p.Stages())
	}
}
