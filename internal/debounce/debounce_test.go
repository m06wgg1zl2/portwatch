package debounce

import (
	"testing"
	"time"
)

var base = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func TestObserve_NotConfirmedBeforeWindow(t *testing.T) {
	d := New(5 * time.Second)
	_, ok := d.Observe("host:80", false, base)
	if ok {
		t.Fatal("expected no confirmation on first observe")
	}
	_, ok = d.Observe("host:80", false, base.Add(4*time.Second))
	if ok {
		t.Fatal("expected no confirmation before window elapses")
	}
}

func TestObserve_ConfirmedAfterWindow(t *testing.T) {
	d := New(5 * time.Second)
	d.Observe("host:80", false, base)
	sc, ok := d.Observe("host:80", false, base.Add(5*time.Second))
	if !ok {
		t.Fatal("expected confirmation after window")
	}
	if sc.Key != "host:80" || sc.NewState != false {
		t.Fatalf("unexpected state change: %+v", sc)
	}
}

func TestObserve_StateFlipResetsTimer(t *testing.T) {
	d := New(5 * time.Second)
	d.Observe("host:80", false, base)
	d.Observe("host:80", true, base.Add(3*time.Second)) // flip resets
	_, ok := d.Observe("host:80", true, base.Add(6*time.Second)) // only 3s since flip
	if ok {
		t.Fatal("expected no confirmation; timer should have reset on flip")
	}
	_, ok = d.Observe("host:80", true, base.Add(8*time.Second)) // 5s since flip
	if !ok {
		t.Fatal("expected confirmation after full window from flip")
	}
}

func TestObserve_IndependentKeys(t *testing.T) {
	d := New(2 * time.Second)
	d.Observe("a:80", false, base)
	d.Observe("b:80", false, base)
	_, okA := d.Observe("a:80", false, base.Add(2*time.Second))
	_, okB := d.Observe("b:80", false, base.Add(1*time.Second))
	if !okA {
		t.Error("expected a:80 confirmed")
	}
	if okB {
		t.Error("expected b:80 not yet confirmed")
	}
}

func TestReset_ClearsPending(t *testing.T) {
	d := New(2 * time.Second)
	d.Observe("host:80", false, base)
	d.Reset("host:80")
	_, ok := d.Observe("host:80", false, base.Add(2*time.Second))
	if ok {
		t.Fatal("expected no confirmation after reset")
	}
}
