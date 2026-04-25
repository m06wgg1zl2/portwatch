package fanout_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/fanout"
)

func makeAlert() alert.Alert {
	return alert.New("host", 8080, "open", alert.LevelInfo)
}

func TestLen_EmptyFanout(t *testing.T) {
	f := fanout.New()
	if f.Len() != 0 {
		t.Fatalf("expected 0, got %d", f.Len())
	}
}

func TestRegister_IncrementsLen(t *testing.T) {
	f := fanout.New()
	f.Register("a", func(_ context.Context, _ alert.Alert) error { return nil })
	f.Register("b", func(_ context.Context, _ alert.Alert) error { return nil })
	if f.Len() != 2 {
		t.Fatalf("expected 2, got %d", f.Len())
	}
}

func TestUnregister_RemovesSink(t *testing.T) {
	f := fanout.New()
	f.Register("a", func(_ context.Context, _ alert.Alert) error { return nil })
	f.Unregister("a")
	if f.Len() != 0 {
		t.Fatalf("expected 0 after unregister, got %d", f.Len())
	}
}

func TestSend_CallsAllHandlers(t *testing.T) {
	f := fanout.New()
	var calls int64
	for _, name := range []string{"x", "y", "z"} {
		f.Register(name, func(_ context.Context, _ alert.Alert) error {
			atomic.AddInt64(&calls, 1)
			return nil
		})
	}
	results := f.Send(context.Background(), makeAlert())
	if int(atomic.LoadInt64(&calls)) != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestSend_CollectsErrors(t *testing.T) {
	f := fanout.New()
	f.Register("ok", func(_ context.Context, _ alert.Alert) error { return nil })
	f.Register("fail", func(_ context.Context, _ alert.Alert) error { return errors.New("boom") })

	results := f.Send(context.Background(), makeAlert())
	var errCount int
	for _, r := range results {
		if r.Error != nil {
			errCount++
		}
	}
	if errCount != 1 {
		t.Fatalf("expected 1 error, got %d", errCount)
	}
}

func TestSend_NoHandlers_ReturnsEmpty(t *testing.T) {
	f := fanout.New()
	results := f.Send(context.Background(), makeAlert())
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}

func TestRegister_OverwriteDoesNotDuplicateOrder(t *testing.T) {
	f := fanout.New()
	f.Register("a", func(_ context.Context, _ alert.Alert) error { return nil })
	f.Register("a", func(_ context.Context, _ alert.Alert) error { return nil })
	if f.Len() != 1 {
		t.Fatalf("expected 1, got %d", f.Len())
	}
	results := f.Send(context.Background(), makeAlert())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
