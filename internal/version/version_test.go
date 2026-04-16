package version

import (
	"strings"
	"testing"
)

func TestString_ContainsVersion(t *testing.T) {
	Version = "1.2.3"
	Commit = "abc1234"
	Date = "2024-01-01"

	s := String()

	if !strings.Contains(s, "1.2.3") {
		t.Errorf("expected version in output, got: %s", s)
	}
	if !strings.Contains(s, "abc1234") {
		t.Errorf("expected commit in output, got: %s", s)
	}
	if !strings.Contains(s, "2024-01-01") {
		t.Errorf("expected date in output, got: %s", s)
	}
}

func TestString_Prefix(t *testing.T) {
	Version = "0.1.0"
	s := String()
	if !strings.HasPrefix(s, "portwatch") {
		t.Errorf("expected string to start with 'portwatch', got: %s", s)
	}
}
