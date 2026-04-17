package alert

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestDispatch_CallsAllHandlers(t *testing.T) {
	d := NewDispatcher()
	var count int64

	for i := 0; i < 3; i++ {
		d.Register(func(a Alert) {
			atomic.AddInt64(&count, 1)
		})
	}

	d.Dispatch(New("localhost", 9000, LevelInfo, "up"))

	if count != 3 {
		t.Errorf("expected 3 handler calls, got %d", count)
	}
}

func TestDispatch_ReceivesCorrectAlert(t *testing.T) {
	d := NewDispatcher()
	var received Alert
	var mu sync.Mutex

	d.Register(func(a Alert) {
		mu.Lock()
		received = a
		mu.Unlock()
	})

	sent := New("db.internal", 5432, LevelError, "connection refused")
	d.Dispatch(sent)

	mu.Lock()
	defer mu.Unlock()
	if received.Host != sent.Host || received.Port != sent.Port {
		t.Errorf("received alert mismatch: got %+v", received)
	}
}

func TestDispatch_PanicHandlerDoesNotCrash(t *testing.T) {
	d := NewDispatcher()
	d.Register(func(a Alert) {
		panic("handler panic")
	})
	// Should not panic the test
	d.Dispatch(New("localhost", 80, LevelWarn, "test"))
}

func TestDispatch_NoHandlers(t *testing.T) {
	d := NewDispatcher()
	// Should complete without error
	d.Dispatch(New("localhost", 80, LevelInfo, "ok"))
}

func TestDispatch_PanicHandlerAllowsOtherHandlers(t *testing.T) {
	d := NewDispatcher()
	var called int64

	d.Register(func(a Alert) {
		panic("intentional panic")
	})
	d.Register(func(a Alert) {
		atomic.AddInt64(&called, 1)
	})

	d.Dispatch(New("localhost", 443, LevelWarn, "test"))

	if called != 1 {
		t.Errorf("expected non-panicking handler to be called, got %d calls", called)
	}
}
