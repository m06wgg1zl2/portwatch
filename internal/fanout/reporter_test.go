package fanout_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/fanout"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	var buf strings.Builder
	r := fanout.NewReporter(&buf)
	r.WriteTable(nil)
	out := buf.String()
	if !strings.Contains(out, "SINK") || !strings.Contains(out, "STATUS") {
		t.Fatalf("missing headers in output: %q", out)
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	var buf strings.Builder
	r := fanout.NewReporter(&buf)
	results := []fanout.Result{
		{Name: "webhook", Error: nil},
		{Name: "slack", Error: errors.New("timeout")},
	}
	r.WriteTable(results)
	out := buf.String()
	if !strings.Contains(out, "webhook") {
		t.Errorf("expected 'webhook' in output: %q", out)
	}
	if !strings.Contains(out, "slack") {
		t.Errorf("expected 'slack' in output: %q", out)
	}
	if !strings.Contains(out, "timeout") {
		t.Errorf("expected error message in output: %q", out)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("expected 'ok' for successful sink: %q", out)
	}
}

func TestSummary_AllSuccess(t *testing.T) {
	var buf strings.Builder
	r := fanout.NewReporter(&buf)
	results := []fanout.Result{
		{Name: "a", Error: nil},
		{Name: "b", Error: nil},
	}
	r.Summary(results)
	out := buf.String()
	if !strings.Contains(out, "2/2") {
		t.Errorf("expected '2/2' in summary: %q", out)
	}
}

func TestSummary_WithErrors(t *testing.T) {
	var buf strings.Builder
	r := fanout.NewReporter(&buf)
	results := []fanout.Result{
		{Name: "ok-sink", Error: nil},
		{Name: "bad-sink", Error: errors.New("refused")},
	}
	r.Summary(results)
	out := buf.String()
	if !strings.Contains(out, "1/2") {
		t.Errorf("expected '1/2' in summary: %q", out)
	}
	if !strings.Contains(out, "bad-sink") {
		t.Errorf("expected failed sink name in summary: %q", out)
	}
}
