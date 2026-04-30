package reaper

import (
	"sync"
	"testing"
	"time"
)

func TestTrack_And_Remove(t *testing.T) {
	r := New(Config{Interval: time.Second}, nil)
	r.Track("key1", 5*time.Second)
	r.mu.Lock()
	_, ok := r.entries["key1"]
	r.mu.Unlock()
	if !ok {
		t.Fatal("expected key1 to be tracked")
	}
	r.Remove("key1")
	r.mu.Lock()
	_, ok = r.entries["key1"]
	r.mu.Unlock()
	if ok {
		t.Fatal("expected key1 to be removed")
	}
}

func TestReap_ExpiredKeyCallsCallback(t *testing.T) {
	var mu sync.Mutex
	var reaped []string
	cb := func(key string) {
		mu.Lock()
		reaped = append(reaped, key)
		mu.Unlock()
	}
	r := New(Config{Interval: 50 * time.Millisecond}, cb)
	r.Track("expired", 1*time.Millisecond)
	r.Track("alive", 10*time.Second)
	time.Sleep(20 * time.Millisecond)
	r.reap()
	mu.Lock()
	defer mu.Unlock()
	if len(reaped) != 1 || reaped[0] != "expired" {
		t.Fatalf("expected [expired], got %v", reaped)
	}
}

func TestReap_NonExpiredKeyNotCalled(t *testing.T) {
	var called bool
	cb := func(key string) { called = true }
	r := New(Config{Interval: time.Second}, cb)
	r.Track("alive", 10*time.Second)
	r.reap()
	if called {
		t.Fatal("callback should not fire for non-expired key")
	}
}

func TestStart_Stop_DoesNotPanic(t *testing.T) {
	r := New(Config{Interval: 50 * time.Millisecond}, nil)
	r.Start()
	time.Sleep(80 * time.Millisecond)
	r.Stop()
}

func TestDefaultInterval_Applied(t *testing.T) {
	r := New(Config{}, nil)
	if r.interval != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", r.interval)
	}
}

func TestReap_RemovesExpiredFromMap(t *testing.T) {
	r := New(Config{Interval: time.Second}, nil)
	r.Track("gone", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	r.reap()
	r.mu.Lock()
	_, ok := r.entries["gone"]
	r.mu.Unlock()
	if ok {
		t.Fatal("expected expired key to be removed from map")
	}
}
