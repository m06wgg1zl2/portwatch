package routing_test

import (
	"strings"
	"testing"

	"portwatch/internal/routing"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	r, _ := routing.New([]routing.Route{
		{Name: "primary", Weight: 8},
		{Name: "fallback", Weight: 2},
	})
	rp := routing.NewReporter(r)
	var sb strings.Builder
	if err := rp.WriteTable(&sb); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	out := sb.String()
	for _, want := range []string{"DESTINATION", "WEIGHT", "SHARE"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected header %q in output", want)
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	r, _ := routing.New([]routing.Route{
		{Name: "alpha", Weight: 5},
	})
	rp := routing.NewReporter(r)
	var sb strings.Builder
	_ = rp.WriteTable(&sb)
	out := sb.String()
	if !strings.Contains(out, "alpha") {
		t.Error("expected route name 'alpha' in table output")
	}
	if !strings.Contains(out, "100.0%") {
		t.Error("expected 100.0% share for single route")
	}
}

func TestSummary_ContainsNames(t *testing.T) {
	r, _ := routing.New([]routing.Route{
		{Name: "web", Weight: 3},
		{Name: "ops", Weight: 7},
	})
	rp := routing.NewReporter(r)
	var sb strings.Builder
	rp.Summary(&sb)
	out := sb.String()
	for _, want := range []string{"web", "ops", "10", "2 destinations"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in summary output", want)
		}
	}
}
