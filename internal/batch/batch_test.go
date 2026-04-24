package batch_test

import (
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/batch"
)

func makeAlert(msg string) *alert.Alert {
	return alert.New("host", 80, alert.LevelInfo, msg)
}

func TestAdd_FlushesOnMaxSize(t *testing.T) {
	var mu sync.Mutex
	var got [][]*alert.Alert

	b := batch.New(batch.Config{Window: time.Second, MaxSize: 3}, func(batch []*alert.Alert) {
		mu.Lock()
		got = append(got, batch)
		mu.Unlock()
	})

	b.Add(makeAlert("a"))
	b.Add(makeAlert("b"))
	b.Add(makeAlert("c")) // triggers flush

	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(got))
	}
	if len(got[0]) != 3 {
		t.Fatalf("expected 3 alerts in batch, got %d", len(got[0]))
	}
}

func TestAdd_FlushesAfterWindow(t *testing.T) {
	var mu sync.Mutex
	var got [][]*alert.Alert

	b := batch.New(batch.Config{Window: 40 * time.Millisecond, MaxSize: 100}, func(batch []*alert.Alert) {
		mu.Lock()
		got = append(got, batch)
		mu.Unlock()
	})

	b.Add(makeAlert("x"))
	b.Add(makeAlert("y"))

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(got))
	}
	if len(got[0]) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(got[0]))
	}
}

func TestFlush_EmptyBufferIsNoop(t *testing.T) {
	called := false
	b := batch.New(batch.Config{Window: time.Second, MaxSize: 10}, func(_ []*alert.Alert) {
		called = true
	})
	b.Flush()
	time.Sleep(20 * time.Millisecond)
	if called {
		t.Fatal("flush of empty buffer should not invoke flushFn")
	}
}

func TestFlush_ForcesEarlyFlush(t *testing.T) {
	var mu sync.Mutex
	var got [][]*alert.Alert

	b := batch.New(batch.Config{Window: 10 * time.Second, MaxSize: 100}, func(batch []*alert.Alert) {
		mu.Lock()
		got = append(got, batch)
		mu.Unlock()
	})

	b.Add(makeAlert("force"))
	b.Flush()

	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(got))
	}
}

func TestAdd_DefaultsApplied(t *testing.T) {
	// Zero config should not panic.
	b := batch.New(batch.Config{}, func(_ []*alert.Alert) {})
	b.Add(makeAlert("default"))
	b.Flush()
}
