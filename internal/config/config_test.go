package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "portwatch-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_, _ = f.WriteString(content)
	_ = f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `{
		"interval_seconds": 10,
		"ports": [
			{"host": "localhost", "port": 8080, "webhooks": ["http://example.com/hook"]},
			{"host": "127.0.0.1", "port": 5432, "shell": "echo changed"}
		]
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("expected 10s interval, got %v", cfg.Interval)
	}
	if len(cfg.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(cfg.Ports))
	}
	if cfg.Ports[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Ports[0].Port)
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	path := writeTempConfig(t, `{"ports": [{"host": "localhost", "port": 80}]}`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected default 30s, got %v", cfg.Interval)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	path := writeTempConfig(t, `{"ports": [{"host": "localhost", "port": 99999}]}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for invalid port")
	}
}

func TestLoad_MissingHost(t *testing.T) {
	path := writeTempConfig(t, `{"ports": [{"port": 80}]}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing host")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
