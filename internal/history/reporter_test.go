package history

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	h, _ := New("")
	_ = h.Add("localhost", 9000, "open")

	r := NewReporter(h)
	var buf bytes.Buffer
	if err := r.WriteTable(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, col := range []string{"TIMESTAMP", "HOST", "PORT", "STATE"} {
		if !strings.Contains(out, col) {
			t.Errorf("expected column %q in output", col)
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	h, _ := New("")
	_ = h.Add("myhost", 3306, "closed")

	r := NewReporter(h)
	var buf bytes.Buffer
	_ = r.WriteTable(&buf)
	out := buf.String()

	if !strings.Contains(out, "myhost") {
		t.Error("expected host in output")
	}
	if !strings.Contains(out, "3306") {
		t.Error("expected port in output")
	}
	if !strings.Contains(out, "closed") {
		t.Error("expected state in output")
	}
}

func TestSummary(t *testing.T) {
	h, _ := New("")
	_ = h.Add("a", 1, "open")
	_ = h.Add("b", 2, "closed")

)
	s := r.Summary()
	if !strings.Contains(s, "2") {
		t.Errorf("expected count 2 in summary, got: %s", s)
	}
}
