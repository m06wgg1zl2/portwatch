package limiter_test

import (
	"sync"
	"testing"
	"time"

	"portwatch/internal/limiter"
)

func TestAcquire_FirstSlotPermitted(t *testing.T) {
	l := limiter.New(limiter.Config{Max: 2})
	if err := l.Acquire("host:80"); err != nil {
		t.Fatalf("expected first acquire to succeed, got %v", err)
	}
}

func TestAcquire_BlockedAtLimit(t *testing.T) {
	l := limiter.New(limiter.Config{Max: 1})
	if err := l.Acquire("host:80"); err != nil {
		t.Fatalf("unexpected error on first acquire: %v", err)
	}
	if err := l.Acquire("host:80"); err != limiter.ErrLimitExceeded {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}
}

func TestAcquire_PermittedAfterRelease(t *testing.T) {
	l := limiter.New(limiter.Config{Max: 1})
	_ = l.Acquire("host:80")
	l.Release("host:80")
	if err := l.Acquire("host:80"); err != nil {
		t.Fatalf("expected acquire after release to succeed, got %v", err)
	}
}

func TestAcquire_IndependentKeys(t *testing.T) {
	l := limiter.New(limiter.Config{Max: 1})
	_ = l.Acquire("host:80")
	if err := l.Acquire("host:443"); err != nil {
		t.Fatalf("expected independent key to succeed, got %v", err)
	}
}

func TestInFlight_TracksCount(t *testing.T) {
	l := limiter.New(limiter.Config{Max: 3})
	_ = l.Acquire("host:80")
	_ = l.Acquire("host:80")
	if got := l.InFlight("host:80"); got != 2 {
		t.Fatalf("expected 2 in-flight, got %d", got)
	}
	l.Release("host:80")
	if got := l.InFlight("host:80"); got != 1 {
		t.Fatalf("expected 1 in-flight after release, got %d", got)
	}
}

func TestAcquire_WithTimeout_Blocks(t *testing.T) {
	l := limiter.New(limiter.Config{Max: 1, Timeout: 50 * time.Millisecond})
	_ = l.Acquire("host:80")
	start := time.Now()
	err := l.Acquire("host:80")
	elapsed := time.Since(start)
	if err != limiter.ErrLimitExceeded {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}
	if elapsed < 40*time.Millisecond {
		t.Fatalf("expected to block for ~50ms, only waited %v", elapsed)
	}
}

func TestAcquire_Concurrent_DoesNotExceedMax(t *testing.T) {
	const max = 3
	l := limiter.New(limiter.Config{Max: max})
	var (
		mu      sync.Mutex
		peak    int
		current int
		wg      sync.WaitGroup
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Acquire("key"); err != nil {
				return
			}
			mu.Lock()
			current++
			if current > peak {
				peak = current
			}
			mu.Unlock()
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			current--
			mu.Unlock()
			l.Release("key")
		}()
	}
	wg.Wait()
	if peak > max {
		t.Fatalf("peak concurrency %d exceeded max %d", peak, max)
	}
}
