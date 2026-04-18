package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleHealth_ContainsUptimeAndStartedAt(t *testing.T) {
	s := New(":0")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	s.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var st Status
	if err := json.NewDecoder(rec.Body).Decode(&st); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if st.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
	if st.StartedAt.IsZero() {
		t.Error("expected non-zero started_at")
	}
}

func TestHandleHealth_ReflectsSetState(t *testing.T) {
	s := New(":0")
	s.SetState("localhost:8080", "open")
	s.SetState("localhost:9090", "closed")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	s.handleHealth(rec, req)

	var st Status
	_ = json.NewDecoder(rec.Body).Decode(&st)

	if st.Ports["localhost:8080"] != "open" {
		t.Errorf("expected open, got %q", st.Ports["localhost:8080"])
	}
	if st.Ports["localhost:9090"] != "closed" {
		t.Errorf("expected closed, got %q", st.Ports["localhost:9090"])
	}
}

func TestHandleHealth_ContentType(t *testing.T) {
	s := New(":0")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	s.handleHealth(rec, req)

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("unexpected content-type: %s", ct)
	}
}

func TestSetState_Concurrent(t *testing.T) {
	s := New(":0")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			s.SetState("host:80", "open")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		s.SetState("host:80", "closed")
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent SetState timed out")
	}
}
