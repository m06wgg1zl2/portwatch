package routing_test

import (
	"testing"

	"portwatch/internal/routing"
)

func TestNew_NoRoutes(t *testing.T) {
	_, err := routing.New(nil)
	if err == nil {
		t.Fatal("expected error for empty routes")
	}
}

func TestNew_ZeroWeight(t *testing.T) {
	_, err := routing.New([]routing.Route{{Name: "a", Weight: 0}})
	if err == nil {
		t.Fatal("expected error for zero weight")
	}
}

func TestNew_NegativeWeight(t *testing.T) {
	_, err := routing.New([]routing.Route{{Name: "a", Weight: -1}})
	if err == nil {
		t.Fatal("expected error for negative weight")
	}
}

func TestSelect_SingleRoute(t *testing.T) {
	r, err := routing.New([]routing.Route{{Name: "only", Weight: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 20; i++ {
		if got := r.Select(); got != "only" {
			t.Fatalf("expected 'only', got %q", got)
		}
	}
}

func TestSelect_DistributionApproximate(t *testing.T) {
	routes := []routing.Route{
		{Name: "a", Weight: 1},
		{Name: "b", Weight: 9},
	}
	r, err := routing.New(routes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	counts := map[string]int{}
	const iterations = 10000
	for i := 0; i < iterations; i++ {
		counts[r.Select()]++
	}
	// "b" should be selected roughly 90% of the time; allow ±5%
	bRatio := float64(counts["b"]) / float64(iterations)
	if bRatio < 0.85 || bRatio > 0.95 {
		t.Errorf("expected b ratio ~0.90, got %.3f", bRatio)
	}
}

func TestRoutes_ReturnsCopy(t *testing.T) {
	input := []routing.Route{{Name: "x", Weight: 5}}
	r, _ := routing.New(input)
	copy1 := r.Routes()
	copy1[0].Name = "mutated"
	copy2 := r.Routes()
	if copy2[0].Name == "mutated" {
		t.Error("Routes() should return an independent copy")
	}
}

func TestTotal_SumsWeights(t *testing.T) {
	r, _ := routing.New([]routing.Route{
		{Name: "a", Weight: 3},
		{Name: "b", Weight: 7},
	})
	if got := r.Total(); got != 10 {
		t.Errorf("expected total 10, got %d", got)
	}
}
