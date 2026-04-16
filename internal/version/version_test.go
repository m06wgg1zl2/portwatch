package version

import (
	"strings"
	"testing"
)

func TestString_ContainsVersion(t *testing.T) {
	Version = "1.2.3"
	Commit = "abc1234"
	BuildDate = "2024-01-01"

	s := String()
	if !strings.Contains(s, "1.2.3") {
		t.Errorf("expected version in output, got: %s", s)
	}
	if !strings.Contains(s, "abc1234") {
		t.Errorf("expected commit in output, got: %s", s)
	}
	if !strings.Contains(s, "2024-01-01") {
		t.Errorf("expected build date in output, got: %s", s)
	}
}

func TestString_Prefix(t *testing.T) {
	Version = "dev"
	s := String()
	if !strings.HasPrefix(s, "portwatch ") {
		t.Errorf("expected string to start with 'portwatch ', got: %s", s)
	}
}

func TestGet_ReturnsInfo(t *testing.T) {
	Version = "0.9.0"
	Commit = "deadbeef"
	BuildDate = "2024-06-15"

	info := Get()
	if info.Version != "0.9.0" {
		t.Errorf("expected Version=0.9.0, got %s", info.Version)
	}
	if info.Commit != "deadbeef" {
		t.Errorf("expected Commit=deadbeef, got %s", info.Commit)
	}
	if info.BuildDate != "2024-06-15" {
		t.Errorf("expected BuildDate=2024-06-15, got %s", info.BuildDate)
	}
}
