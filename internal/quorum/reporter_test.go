package quorum

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	q := New(Config{Required: 2, Window: 5})
	r := NewReporter(q)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	for _, hdr := range []string{"KEY", "OBSERVATIONS", "REQUIRED", "WINDOW"} {
		if !strings.Contains(buf.String(), hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	q := New(Config{Required: 2, Window: 5})
	q.Observe("host:9200", "open")
	r := NewReporter(q)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	if !strings.Contains(buf.String(), "host:9200") {
		t.Errorf("expected key in table output, got:\n%s", buf.String())
	}
}

func TestSummary_NoKeys(t *testing.T) {
	q := New(Config{Required: 3, Window: 10})
	r := NewReporter(q)
	var buf bytes.Buffer
	r.Summary(&buf)
	out := buf.String()
	if !strings.Contains(out, "0 active key") {
		t.Errorf("expected zero keys in summary, got: %s", out)
	}
	if !strings.Contains(out, "3 observation") {
		t.Errorf("expected required count in summary, got: %s", out)
	}
}

func TestSummary_WithKeys(t *testing.T) {
	q := New(Config{Required: 2, Window: 5})
	q.Observe("a", "open")
	q.Observe("b", "closed")
	r := NewReporter(q)
	var buf bytes.Buffer
	r.Summary(&buf)
	if !strings.Contains(buf.String(), "2 active key") {
		t.Errorf("expected 2 active keys in summary, got: %s", buf.String())
	}
}
