package deadletter_test

import (
	"testing"

	"portwatch/internal/alert"
	"portwatch/internal/deadletter"
)

func makeAlert(host string) alert.Alert {
	return alert.New(host, 80, alert.LevelCritical, "port closed")
}

func TestPush_AddsEntry(t *testing.T) {
	q := deadletter.New(deadletter.Config{MaxSize: 10})
	q.Push(makeAlert("host-a"), "timeout", 3)

	if q.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", q.Len())
	}
	entries := q.All()
	if entries[0].Reason != "timeout" {
		t.Errorf("expected reason 'timeout', got %q", entries[0].Reason)
	}
	if entries[0].Attempts != 3 {
		t.Errorf("expected attempts 3, got %d", entries[0].Attempts)
	}
}

func TestPush_EvictsOldestWhenFull(t *testing.T) {
	q := deadletter.New(deadletter.Config{MaxSize: 3})
	q.Push(makeAlert("first"), "r", 1)
	q.Push(makeAlert("second"), "r", 1)
	q.Push(makeAlert("third"), "r", 1)
	q.Push(makeAlert("fourth"), "r", 1)

	if q.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", q.Len())
	}
	entries := q.All()
	if entries[0].Alert.Host != "second" {
		t.Errorf("expected oldest surviving entry to be 'second', got %q", entries[0].Alert.Host)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	q := deadletter.New(deadletter.Config{})
	q.Push(makeAlert("x"), "err", 1)

	a := q.All()
	a[0].Reason = "mutated"

	b := q.All()
	if b[0].Reason == "mutated" {
		t.Error("All() should return an independent copy")
	}
}

func TestClear_RemovesAll(t *testing.T) {
	q := deadletter.New(deadletter.Config{})
	q.Push(makeAlert("a"), "e", 1)
	q.Push(makeAlert("b"), "e", 1)
	q.Clear()

	if q.Len() != 0 {
		t.Errorf("expected 0 entries after Clear, got %d", q.Len())
	}
}

func TestNew_DefaultMaxSize(t *testing.T) {
	q := deadletter.New(deadletter.Config{MaxSize: 0})
	for i := 0; i < 105; i++ {
		q.Push(makeAlert("h"), "e", 1)
	}
	if q.Len() != 100 {
		t.Errorf("expected default cap of 100, got %d", q.Len())
	}
}

func TestPush_FailedAtIsSet(t *testing.T) {
	q := deadletter.New(deadletter.Config{})
	q.Push(makeAlert("z"), "boom", 2)

	entry := q.All()[0]
	if entry.FailedAt.IsZero() {
		t.Error("FailedAt should not be zero")
	}
}
