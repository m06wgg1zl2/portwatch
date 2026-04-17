package ratelimit_test

import (
	"testing"
	"time"

	"portwatch/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(1 * time.Second)
	if !l.Allow("host:80") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_BlockedWithinCooldown(t *testing.T) {
	l := ratelimit.New(1 * time.Second)
	l.Allow("host:80")
	if l.Allow("host:80") {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestAllow_PermittedAfterCooldown(t *testing.T) {
	l := ratelimit.New(20 * time.Millisecond)
	l.Allow("host:80")
	time.Sleep(30 * time.Millisecond)
	if !l.Allow("host:80") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	l := ratelimit.New(1 * time.Second)
	l.Allow("host:80")
	if !l.Allow("host:443") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	l := ratelimit.New(1 * time.Second)
	l.Allow("host:80")
	l.Reset("host:80")
	if !l.Allow("host:80") {
		t.Fatal("expected allow after reset")
	}
}

func TestLastFired_ReturnsTime(t *testing.T) {
	l := ratelimit.New(1 * time.Second)
	before := time.Now()
	l.Allow("host:80")
	after := time.Now()

	t2, ok := l.LastFired("host:80")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if t2.Before(before) || t2.After(after) {
		t.Errorf("last fired time %v out of expected range", t2)
	}
}

func TestLastFired_MissingKey(t *testing.T) {
	l := ratelimit.New(1 * time.Second)
	_, ok := l.LastFired("missing:9999")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}
