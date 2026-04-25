package budget

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.1})
	rep := NewReporter(b)
	var buf bytes.Buffer
	if err := rep.WriteTable(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, hdr := range []string{"KEY", "RATIO", "BREACHED"} {
		if !strings.Contains(buf.String(), hdr) {
			t.Errorf("missing header %q in output:\n%s", hdr, buf.String())
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.1})
	b.Record("db", true)
	b.Record("db", false)
	rep := NewReporter(b)
	var buf bytes.Buffer
	_ = rep.WriteTable(&buf, []string{"db"})
	out := buf.String()
	if !strings.Contains(out, "db") {
		t.Errorf("expected key 'db' in output:\n%s", out)
	}
	if !strings.Contains(out, "0.5000") {
		t.Errorf("expected ratio 0.5000 in output:\n%s", out)
	}
}

func TestSummary_NoBreaches(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.5})
	b.Record("svc", false)
	rep := NewReporter(b)
	var buf bytes.Buffer
	rep.Summary(&buf, []string{"svc"})
	if !strings.Contains(buf.String(), "0/1") {
		t.Errorf("unexpected summary: %s", buf.String())
	}
}

func TestSummary_WithBreaches(t *testing.T) {
	b := New(Config{Window: time.Minute, Threshold: 0.05})
	for i := 0; i < 5; i++ {
		b.Record("svc", true)
	}
	rep := NewReporter(b)
	var buf bytes.Buffer
	rep.Summary(&buf, []string{"svc"})
	if !strings.Contains(buf.String(), "1/1") {
		t.Errorf("unexpected summary: %s", buf.String())
	}
}
