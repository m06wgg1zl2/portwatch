package audit

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	l := New(10)
	r := NewReporter(l)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	for _, hdr := range []string{"TIMESTAMP", "LEVEL", "SOURCE", "MESSAGE"} {
		if !strings.Contains(buf.String(), hdr) {
			t.Errorf("table missing header %q", hdr)
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	l := New(10)
	l.Add(LevelWarn, "probe", "connection refused")
	r := NewReporter(l)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatalf("WriteTable error: %v", err)
	}
	for _, want := range []string{"WARN", "probe", "connection refused"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("table missing %q\n%s", want, buf.String())
		}
	}
}

func TestSummary_NoEvents(t *testing.T) {
	l := New(10)
	r := NewReporter(l)
	var buf bytes.Buffer
	if err := r.Summary(&buf); err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	if !strings.Contains(buf.String(), "total=0") {
		t.Errorf("expected total=0, got: %s", buf.String())
	}
}

func TestSummary_WithEvents(t *testing.T) {
	l := New(20)
	l.Add(LevelInfo, "monitor", "started")
	l.Add(LevelInfo, "monitor", "tick")
	l.Add(LevelWarn, "circuit", "half-open")
	l.Add(LevelError, "probe", "dial failed")
	r := NewReporter(l)
	var buf bytes.Buffer
	if err := r.Summary(&buf); err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	s := buf.String()
	for _, want := range []string{"total=4", "info=2", "warn=1", "error=1"} {
		if !strings.Contains(s, want) {
			t.Errorf("summary missing %q: %s", want, s)
		}
	}
}
