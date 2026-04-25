package trend

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 2})
	rep := NewReporter(tr)
	var buf bytes.Buffer
	if err := rep.WriteTable(&buf); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	for _, hdr := range []string{"KEY", "SAMPLES", "DIRECTION"} {
		if !strings.Contains(buf.String(), hdr) {
			t.Errorf("missing header %q in output: %s", hdr, buf.String())
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 2})
	tr.Record("api", 1)
	tr.Record("api", 2)
	rep := NewReporter(tr)
	var buf bytes.Buffer
	if err := rep.WriteTable(&buf); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	if !strings.Contains(buf.String(), "api") {
		t.Errorf("expected key 'api' in table, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "rising") {
		t.Errorf("expected direction 'rising' in table, got: %s", buf.String())
	}
}

func TestSummary_NoKeys(t *testing.T) {
	tr := New(Config{})
	rep := NewReporter(tr)
	var buf bytes.Buffer
	if err := rep.Summary(&buf); err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	if !strings.Contains(buf.String(), "0") {
		t.Errorf("expected 0 keys in summary, got: %s", buf.String())
	}
}

func TestSummary_WithKeys(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 2})
	tr.Record("x", 1)
	tr.Record("y", 2)
	rep := NewReporter(tr)
	var buf bytes.Buffer
	if err := rep.Summary(&buf); err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	if !strings.Contains(buf.String(), "2") {
		t.Errorf("expected 2 keys in summary, got: %s", buf.String())
	}
}
