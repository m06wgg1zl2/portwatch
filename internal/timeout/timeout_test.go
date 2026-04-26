package timeout_test

import (
	"errors"
	"testing"
	"time"

	"portwatch/internal/timeout"
)

func TestFor_DefaultWhenNoOverride(t *testing.T) {
	m := timeout.New(timeout.Config{Default: 2 * time.Second})
	if got := m.For("any"); got != 2*time.Second {
		t.Fatalf("expected 2s, got %v", got)
	}
}

func TestFor_UsesOverride(t *testing.T) {
	m := timeout.New(timeout.Config{
		Default:   2 * time.Second,
		Overrides: map[string]time.Duration{"db": 500 * time.Millisecond},
	})
	if got := m.For("db"); got != 500*time.Millisecond {
		t.Fatalf("expected 500ms, got %v", got)
	}
}

func TestNew_DefaultFallback(t *testing.T) {
	m := timeout.New(timeout.Config{})
	if got := m.For("x"); got != 5*time.Second {
		t.Fatalf("expected 5s default, got %v", got)
	}
}

func TestSet_And_Delete(t *testing.T) {
	m := timeout.New(timeout.Config{Default: 1 * time.Second})
	m.Set("svc", 300*time.Millisecond)
	if got := m.For("svc"); got != 300*time.Millisecond {
		t.Fatalf("expected 300ms after Set, got %v", got)
	}
	m.Delete("svc")
	if got := m.For("svc"); got != 1*time.Second {
		t.Fatalf("expected default after Delete, got %v", got)
	}
}

func TestDo_SuccessWithinTimeout(t *testing.T) {
	m := timeout.New(timeout.Config{Default: 1 * time.Second})
	err := m.Do("op", func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_PropagatesError(t *testing.T) {
	m := timeout.New(timeout.Config{Default: 1 * time.Second})
	sentinel := errors.New("boom")
	err := m.Do("op", func() error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestDo_TimesOut(t *testing.T) {
	m := timeout.New(timeout.Config{Default: 30 * time.Millisecond})
	err := m.Do("slow", func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	if !errors.Is(err, timeout.ErrTimedOut) {
		t.Fatalf("expected ErrTimedOut, got %v", err)
	}
}

func TestDo_PerKeyOverrideTimeout(t *testing.T) {
	m := timeout.New(timeout.Config{Default: 200 * time.Millisecond})
	m.Set("fast", 20*time.Millisecond)
	err := m.Do("fast", func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if !errors.Is(err, timeout.ErrTimedOut) {
		t.Fatalf("expected ErrTimedOut for overridden key, got %v", err)
	}
}
