package batch_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/batch"
)

func TestWriteTable_ContainsHeaders(t *testing.T) {
	s := &batch.Stats{}
	r := batch.NewReporter(s)
	var buf bytes.Buffer
	r.WriteTable(&buf)
	out := buf.String()
	for _, h := range []string{"METRIC", "VALUE"} {
		if !strings.Contains(out, h) {
			t.Errorf("expected header %q in output: %s", h, out)
		}
	}
}

func TestWriteTable_ContainsEntry(t *testing.T) {
	s := &batch.Stats{
		TotalFlushed: 4,
		TotalAlerts:  20,
		LastFlushTime: time.Now(),
	}
	r := batch.NewReporter(s)
	var buf bytes.Buffer
	r.WriteTable(&buf)
	out := buf.String()
	if !strings.Contains(out, "4") {
		t.Errorf("expected flush count in output: %s", out)
	}
	if !strings.Contains(out, "20") {
		t.Errorf("expected alert count in output: %s", out)
	}
}

func TestSummary_NoFlushes(t *testing.T) {
	s := &batch.Stats{}
	r := batch.NewReporter(s)
	var buf bytes.Buffer
	r.Summary(&buf)
	if !strings.Contains(buf.String(), "no flushes") {
		t.Errorf("expected 'no flushes' message, got: %s", buf.String())
	}
}

func TestSummary_WithFlushes(t *testing.T) {
	s := &batch.Stats{
		TotalFlushed:  2,
		TotalAlerts:   10,
		LastFlushTime: time.Now(),
	}
	r := batch.NewReporter(s)
	var buf bytes.Buffer
	r.Summary(&buf)
	out := buf.String()
	if !strings.Contains(out, "2 flushes") {
		t.Errorf("expected flush count in summary: %s", out)
	}
	if !strings.Contains(out, "5.0 per flush") {
		t.Errorf("expected avg per flush in summary: %s", out)
	}
}
