package alert

import (
	"strings"
	"testing"
	"time"
)

func TestNew_Fields(t *testing.T) {
	before := time.Now().UTC()
	a := New("localhost", 8080, LevelWarn, "port closed")
	after := time.Now().UTC()

	if a.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", a.Host)
	}
	if a.Port != 8080 {
		t.Errorf("expected port 8080, got %d", a.Port)
	}
	if a.Level != LevelWarn {
		t.Errorf("expected level warn, got %s", a.Level)
	}
	if a.Message != "port closed" {
		t.Errorf("unexpected message: %s", a.Message)
	}
	if a.Timestamp.Before(before) || a.Timestamp.After(after) {
		t.Error("timestamp out of expected range")
	}
}

func TestString_ContainsFields(t *testing.T) {
	a := New("example.com", 443, LevelError, "unreachable")
	s := a.String()

	for _, substr := range []string{"example.com", "443", "error", "unreachable"} {
		if !strings.Contains(s, substr) {
			t.Errorf("expected %q in alert string: %s", substr, s)
		}
	}
}

func TestLevel_Constants(t *testing.T) {
	levels := []Level{LevelInfo, LevelWarn, LevelError}
	expected := []string{"info", "warn", "error"}
	for i, l := range levels {
		if string(l) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], l)
		}
	}
}
