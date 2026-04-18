package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	c := New()
	r := NewReporter(c)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "KEY") || !strings.Contains(buf.String(), "COUNT") {
		t.Fatalf("missing headers: %s", buf.String())
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	c := New()
	c.Inc("port:8080:open")
	c.Inc("port:8080:open")
	r := NewReporter(c)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "port:8080:open") {
		t.Fatalf("expected key in output: %s", out)
	}
	if !strings.Contains(out, "2") {
		t.Fatalf("expected count 2 in output: %s", out)
	}
}

func TestSummary_NoEvents(t *testing.T) {
	c := New()
	r := NewReporter(c)
	s := r.Summary()
	if !strings.Contains(s, "0") {
		t.Fatalf("expected zero total: %s", s)
	}
}

func TestSummary_WithEvents(t *testing.T) {
	c := New()
	c.Inc("a")
	c.Inc("a")
	c.Inc("b")
	r := NewReporter(c)
	s := r.Summary()
	if !strings.Contains(s, "3") {
		t.Fatalf("expected total 3: %s", s)
	}
	if !strings.Contains(s, "2 key") {
		t.Fatalf("expected 2 keys: %s", s)
	}
}
