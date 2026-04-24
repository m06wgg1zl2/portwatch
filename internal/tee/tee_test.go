package tee_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"portwatch/internal/alert"
	"portwatch/internal/tee"
)

func makeAlert() alert.Alert {
	return alert.New("host", 8080, alert.LevelWarn, "test")
}

func TestNew_Len(t *testing.T) {
	tr := tee.New()
	if tr.Len() != 0 {
		t.Fatalf("expected 0 handlers, got %d", tr.Len())
	}

	tr.Add(func(alert.Alert) error { return nil })
	if tr.Len() != 1 {
		t.Fatalf("expected 1 handler, got %d", tr.Len())
	}
}

func TestSend_CallsAllHandlers(t *testing.T) {
	var count atomic.Int32
	h := func(alert.Alert) error { count.Add(1); return nil }

	tr := tee.New(h, h, h)
	if err := tr.Send(makeAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", count.Load())
	}
}

func TestSend_NoHandlers_NoError(t *testing.T) {
	tr := tee.New()
	if err := tr.Send(makeAlert()); err != nil {
		t.Fatalf("expected nil error for empty tee, got %v", err)
	}
}

func TestSend_ReceivesCorrectAlert(t *testing.T) {
	a := makeAlert()
	var got alert.Alert
	tr := tee.New(func(recv alert.Alert) error {
		got = recv
		return nil
	})
	_ = tr.Send(a)
	if got.Host != a.Host || got.Port != a.Port {
		t.Fatalf("handler received wrong alert: %+v", got)
	}
}

func TestSend_AggregatesErrors(t *testing.T) {
	err1 := errors.New("boom")
	err2 := errors.New("bang")

	tr := tee.New(
		func(alert.Alert) error { return err1 },
		func(alert.Alert) error { return nil },
		func(alert.Alert) error { return err2 },
	)

	err := tr.Send(makeAlert())
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	msg := err.Error()
	if !contains(msg, "boom") || !contains(msg, "bang") {
		t.Fatalf("error message missing details: %s", msg)
	}
}

func TestSend_PartialFailure_OtherHandlersStillRun(t *testing.T) {
	var ran atomic.Int32
	tr := tee.New(
		func(alert.Alert) error { return errors.New("fail") },
		func(alert.Alert) error { ran.Add(1); return nil },
	)
	_ = tr.Send(makeAlert())
	if ran.Load() != 1 {
		t.Fatal("second handler was not called despite first failing")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
