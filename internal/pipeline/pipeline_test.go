package pipeline_test

import (
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/pipeline"
)

// allowStage always permits the alert to pass.
type allowStage struct{ name string }

func (a allowStage) Name() string              { return a.name }
func (a allowStage) Allow(_ alert.Alert) bool  { return true }

// blockStage always halts the alert.
type blockStage struct{ name string }

func (b blockStage) Name() string             { return b.name }
func (b blockStage) Allow(_ alert.Alert) bool { return false }

func makeAlert() alert.Alert {
	return alert.New("localhost", 8080, alert.LevelCritical, "open")
}

func TestRun_AllowsWhenNoStages(t *testing.T) {
	p := pipeline.New()
	if blocked := p.Run(makeAlert()); blocked != "" {
		t.Fatalf("expected empty, got %q", blocked)
	}
}

func TestRun_AllowsWhenAllPass(t *testing.T) {
	p := pipeline.New(allowStage{"a"}, allowStage{"b"})
	if blocked := p.Run(makeAlert()); blocked != "" {
		t.Fatalf("expected empty, got %q", blocked)
	}
}

func TestRun_BlockedByFirstFailingStage(t *testing.T) {
	p := pipeline.New(allowStage{"pass"}, blockStage{"gate"}, allowStage{"never"})
	blocked := p.Run(makeAlert())
	if blocked != "gate" {
		t.Fatalf("expected gate, got %q", blocked)
	}
}

func TestRun_BlockedAtFirstStage(t *testing.T) {
	p := pipeline.New(blockStage{"early"}, allowStage{"late"})
	if blocked := p.Run(makeAlert()); blocked != "early" {
		t.Fatalf("expected early, got %q", blocked)
	}
}

func TestStages_ReturnsNames(t *testing.T) {
	p := pipeline.New(allowStage{"x"}, blockStage{"y"})
	names := p.Stages()
	if len(names) != 2 || names[0] != "x" || names[1] != "y" {
		t.Fatalf("unexpected names: %v", names)
	}
}

func TestLen_ReturnsCount(t *testing.T) {
	p := pipeline.New(allowStage{"a"}, allowStage{"b"}, allowStage{"c"})
	if p.Len() != 3 {
		t.Fatalf("expected 3, got %d", p.Len())
	}
}
