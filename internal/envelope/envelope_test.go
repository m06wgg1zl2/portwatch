package envelope_test

import (
	"testing"
	"time"

	"portwatch/internal/alert"
	"portwatch/internal/envelope"
)

func makeAlert() *alert.Alert {
	return alert.New("localhost", 8080, alert.LevelCritical)
}

func TestNew_FieldsPopulated(t *testing.T) {
	a := makeAlert()
	e := envelope.New(a)

	if e.Alert != a {
		t.Error("expected Alert to match input")
	}
	if e.TraceID == "" {
		t.Error("expected non-empty TraceID")
	}
	if e.Priority != 0 {
		t.Errorf("expected default priority 0, got %d", e.Priority)
	}
	if e.Labels == nil {
		t.Error("expected Labels map to be initialised")
	}
	if time.Since(e.CreatedAt) > 2*time.Second {
		t.Error("expected CreatedAt to be recent")
	}
}

func TestNew_UniqueTraceIDs(t *testing.T) {
	e1 := envelope.New(makeAlert())
	e2 := envelope.New(makeAlert())
	if e1.TraceID == e2.TraceID {
		t.Error("expected distinct trace IDs for separate envelopes")
	}
}

func TestWithPriority_ReturnsCopy(t *testing.T) {
	e := envelope.New(makeAlert())
	high := e.WithPriority(10)

	if high.Priority != 10 {
		t.Errorf("expected priority 10, got %d", high.Priority)
	}
	if e.Priority != 0 {
		t.Error("original envelope priority should be unchanged")
	}
	if high.TraceID != e.TraceID {
		t.Error("WithPriority should preserve TraceID")
	}
}

func TestSetLabel_And_Label(t *testing.T) {
	e := envelope.New(makeAlert())
	e.SetLabel("region", "eu-west")

	v, ok := e.Label("region")
	if !ok {
		t.Fatal("expected label to be present")
	}
	if v != "eu-west" {
		t.Errorf("expected eu-west, got %s", v)
	}
}

func TestLabel_Missing(t *testing.T) {
	e := envelope.New(makeAlert())
	_, ok := e.Label("nonexistent")
	if ok {
		t.Error("expected missing label to return false")
	}
}

func TestString_ContainsTraceAndPriority(t *testing.T) {
	e := envelope.New(makeAlert())
	s := e.String()
	if s == "" {
		t.Error("expected non-empty String output")
	}
	for _, want := range []string{"trace=", "priority=", "alert="} {
		if !contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
