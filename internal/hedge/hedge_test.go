package hedge_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"portwatch/internal/hedge"
)

func TestDo_FastSuccess(t *testing.T) {
	h := hedge.New(hedge.Config{SoftTimeout: 50 * time.Millisecond, HardTimeout: 500 * time.Millisecond})
	res := h.Do(context.Background(), func(_ context.Context) (interface{}, error) {
		return "ok", nil
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Value.(string) != "ok" {
		t.Fatalf("unexpected value: %v", res.Value)
	}
}

func TestDo_HedgeLaunchedOnSlowPrimary(t *testing.T) {
	var calls atomic.Int32
	h := hedge.New(hedge.Config{SoftTimeout: 20 * time.Millisecond, HardTimeout: 300 * time.Millisecond})

	res := h.Do(context.Background(), func(_ context.Context) (interface{}, error) {
		n := calls.Add(1)
		if n == 1 {
			time.Sleep(60 * time.Millisecond) // slower than soft timeout
		}
		return "done", nil
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if calls.Load() < 2 {
		t.Fatal("expected hedge to launch a second call")
	}
}

func TestDo_BothFail_ReturnsError(t *testing.T) {
	h := hedge.New(hedge.Config{SoftTimeout: 10 * time.Millisecond, HardTimeout: 200 * time.Millisecond})
	sentinel := errors.New("boom")

	res := h.Do(context.Background(), func(_ context.Context) (interface{}, error) {
		return nil, sentinel
	})
	if res.Err == nil {
		t.Fatal("expected an error")
	}
}

func TestDo_HardTimeoutReturnsError(t *testing.T) {
	h := hedge.New(hedge.Config{SoftTimeout: 10 * time.Millisecond, HardTimeout: 30 * time.Millisecond})

	res := h.Do(context.Background(), func(ctx context.Context) (interface{}, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	if res.Err == nil {
		t.Fatal("expected hard-timeout error")
	}
}

func TestDo_Defaults(t *testing.T) {
	h := hedge.New(hedge.Config{})
	if h.SoftTimeout() != 50*time.Millisecond {
		t.Fatalf("wrong soft timeout: %v", h.SoftTimeout())
	}
	if h.HardTimeout() != 500*time.Millisecond {
		t.Fatalf("wrong hard timeout: %v", h.HardTimeout())
	}
}

func TestDo_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	h := hedge.New(hedge.Config{SoftTimeout: 10 * time.Millisecond, HardTimeout: 100 * time.Millisecond})
	res := h.Do(ctx, func(_ context.Context) (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return "late", nil
	})
	// May succeed or fail depending on scheduling; we just ensure no panic.
	_ = res
}
