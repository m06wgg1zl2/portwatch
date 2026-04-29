package drain_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"portwatch/internal/drain"
)

func TestWait_NoWork(t *testing.T) {
	d := drain.New(drain.Config{Timeout: time.Second})
	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("expected nil error with no in-flight work, got %v", err)
	}
}

func TestWait_WorkCompletesInTime(t *testing.T) {
	d := drain.New(drain.Config{Timeout: time.Second})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		if !d.Acquire() {
			t.Fatal("Acquire returned false before Wait called")
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer d.Release()
			time.Sleep(20 * time.Millisecond)
		}()
	}

	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("expected clean drain, got %v", err)
	}
	wg.Wait()
}

func TestWait_TimeoutExceeded(t *testing.T) {
	d := drain.New(drain.Config{Timeout: 30 * time.Millisecond})

	if !d.Acquire() {
		t.Fatal("Acquire failed")
	}
	// Intentionally never call Release to force timeout.
	defer d.Release()

	ctx := context.Background()
	err := d.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	d := drain.New(drain.Config{Timeout: 5 * time.Second})

	if !d.Acquire() {
		t.Fatal("Acquire failed")
	}
	defer d.Release()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := d.Wait(ctx)
	if err != context.Canceled {
		t.Fatalf("expected Canceled, got %v", err)
	}
}

func TestAcquire_ReturnsFalseAfterWait(t *testing.T) {
	d := drain.New(drain.Config{Timeout: time.Second})

	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.Acquire() {
		t.Fatal("Acquire should return false after Wait closes the drain")
	}
}

func TestWait_DefaultTimeout(t *testing.T) {
	// Zero timeout should default to 10s — just verify it doesn't panic.
	d := drain.New(drain.Config{})
	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
