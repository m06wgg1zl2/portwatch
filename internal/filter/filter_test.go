package filter

import (
	"testing"
	"time"
)

func at(hour int) time.Time {
	return time.Date(2024, 1, 1, hour, 0, 0, 0, time.UTC)
}

func TestAllow_NoRules(t *testing.T) {
	f := New(nil)
	if !f.Allow("closed", at(3)) {
		t.Fatal("expected allow with no rules")
	}
}

func TestAllow_HourWindow_Inside(t *testing.T) {
	f := New([]Rule{{FromHour: 8, ToHour: 18}})
	if !f.Allow("open", at(10)) {
		t.Fatal("expected allow inside window")
	}
}

func TestAllow_HourWindow_Outside(t *testing.T) {
	f := New([]Rule{{FromHour: 8, ToHour: 18}})
	if f.Allow("open", at(3)) {
		t.Fatal("expected deny outside window")
	}
}

func TestAllow_HourWindow_WrapsMidnight(t *testing.T) {
	f := New([]Rule{{FromHour: 22, ToHour: 6}})
	if !f.Allow("open", at(23)) {
		t.Fatal("expected allow at 23 in wrap window")
	}
	if !f.Allow("open", at(2)) {
		t.Fatal("expected allow at 2 in wrap window")
	}
	if f.Allow("open", at(10)) {
		t.Fatal("expected deny at 10 in wrap window")
	}
}

func TestAllow_StateFilter_Match(t *testing.T) {
	f := New([]Rule{{States: []string{"closed"}}})
	if !f.Allow("closed", at(12)) {
		t.Fatal("expected allow for matching state")
	}
}

func TestAllow_StateFilter_NoMatch(t *testing.T) {
	f := New([]Rule{{States: []string{"closed"}}})
	if f.Allow("open", at(12)) {
		t.Fatal("expected deny for non-matching state")
	}
}

func TestAllow_CombinedRules(t *testing.T) {
	f := New([]Rule{{FromHour: 9, ToHour: 17, States: []string{"closed"}}})
	if !f.Allow("closed", at(10)) {
		t.Fatal("expected allow")
	}
	if f.Allow("open", at(10)) {
		t.Fatal("expected deny: wrong state")
	}
	if f.Allow("closed", at(20)) {
		t.Fatal("expected deny: outside hours")
	}
}
