package audit

import (
	"strings"
	"sync"
	"testing"
)

func TestAdd_StoresEvent(t *testing.T) {
	l := New(10)
	l.Add(LevelInfo, "monitor", "port opened")
	if l.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", l.Len())
	}
	events := l.All()
	if events[0].Message != "port opened" {
		t.Errorf("unexpected message: %s", events[0].Message)
	}
	if events[0].Source != "monitor" {
		t.Errorf("unexpected source: %s", events[0].Source)
	}
}

func TestAdd_EvictsOldestWhenFull(t *testing.T) {
	l := New(3)
	l.Add(LevelInfo, "src", "first")
	l.Add(LevelInfo, "src", "second")
	l.Add(LevelInfo, "src", "third")
	l.Add(LevelInfo, "src", "fourth")
	events := l.All()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Message != "second" {
		t.Errorf("expected oldest to be evicted, got: %s", events[0].Message)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := New(10)
	l.Add(LevelWarn, "probe", "timeout")
	a := l.All()
	a[0].Message = "mutated"
	b := l.All()
	if b[0].Message == "mutated" {
		t.Error("All() should return an independent copy")
	}
}

func TestClear_RemovesAll(t *testing.T) {
	l := New(10)
	l.Add(LevelError, "circuit", "open")
	l.Clear()
	if l.Len() != 0 {
		t.Errorf("expected 0 after Clear, got %d", l.Len())
	}
}

func TestLevel_String(t *testing.T) {
	cases := map[Level]string{
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		Level(99):  "UNKNOWN",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("Level(%d).String() = %q, want %q", level, got, want)
		}
	}
}

func TestEvent_String_ContainsFields(t *testing.T) {
	l := New(10)
	l.Add(LevelError, "watchdog", "heartbeat missed")
	s := l.All()[0].String()
	for _, substr := range []string{"ERROR", "watchdog", "heartbeat missed"} {
		if !strings.Contains(s, substr) {
			t.Errorf("event string missing %q: %s", substr, s)
		}
	}
}

func TestAdd_Concurrent(t *testing.T) {
	l := New(1000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Add(LevelInfo, "test", "concurrent")
		}()
	}
	wg.Wait()
	if l.Len() != 100 {
		t.Errorf("expected 100 events, got %d", l.Len())
	}
}
