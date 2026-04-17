package retry

import (
	"errors"
	"testing"
	"time"
)

var errFail = errors.New("fail")

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	p := New(3, 0, 1.0)
	err := p.Do(func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	p := New(3, time.Millisecond, 1.0)
	calls := 0
	err := p.Do(func() error {
		calls++
		if calls < 3 {
			return errFail
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_AllAttemptsFail(t *testing.T) {
	p := New(3, time.Millisecond, 1.0)
	err := p.Do(func() error { return errFail })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDo_BackoffIncreasesDelay(t *testing.T) {
	p := New(3, 10*time.Millisecond, 2.0)
	start := time.Now()
	_ = p.Do(func() error { return errFail })
	elapsed := time.Since(start)
	// 10ms + 20ms = 30ms minimum between 3 attempts
	if elapsed < 25*time.Millisecond {
		t.Fatalf("expected backoff delay, elapsed %v", elapsed)
	}
}

func TestAttempts_ReturnsCount(t *testing.T) {
	p := New(5, time.Millisecond, 1.0)
	calls := 0
	n, err := p.Attempts(func() error {
		calls++
		if calls < 2 {
			return errFail
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 attempts, got %d", n)
	}
}

func TestNew_DefaultsMaxAttempts(t *testing.T) {
	p := New(0, 0, 0)
	if p.MaxAttempts != 1 {
		t.Fatalf("expected MaxAttempts=1, got %d", p.MaxAttempts)
	}
	if p.Backoff != 1.0 {
		t.Fatalf("expected Backoff=1.0, got %f", p.Backoff)
	}
}
