package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhook_Success(t *testing.T) {
	var received Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := New()
	if err := n.Webhook(ts.URL, "localhost", 8080, "open"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Host != "localhost" || received.Port != 8080 || received.State != "open" {
		t.Errorf("unexpected payload: %+v", received)
	}
}

func TestWebhook_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := New()
	if err := n.Webhook(ts.URL, "localhost", 9090, "closed"); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhook_BadURL(t *testing.T) {
	n := New()
	if err := n.Webhook("http://127.0.0.1:0/nope", "localhost", 1234, "open"); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestShell_Success(t *testing.T) {
	n := New()
	if err := n.Shell("echo {host}:{port} is {state}", "localhost", 8080, "open"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShell_Failure(t *testing.T) {
	n := New()
	if err := n.Shell("exit 1", "localhost", 8080, "closed"); err == nil {
		t.Fatal("expected error for failing shell command")
	}
}
