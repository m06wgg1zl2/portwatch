package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status holds the current health snapshot.
type Status struct {
	Uptime    string            `json:"uptime"`
	StartedAt time.Time         `json:"started_at"`
	Ports     map[string]string `json:"ports"`
}

// Server exposes a simple HTTP health endpoint.
type Server struct {
	mu        sync.RWMutex
	startedAt time.Time
	states    map[string]string
	addr      string
}

// New creates a new health check server on addr (e.g. ":9090").
func New(addr string) *Server {
	return &Server{
		addr:      addr,
		startedAt: time.Now(),
		states:    make(map[string]string),
	}
}

// SetState updates the recorded state for a port key.
func (s *Server) SetState(key, state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[key] = state
}

// ListenAndServe starts the HTTP server; blocks until error.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	srv := &http.Server{Addr: s.addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return srv.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	copy := make(map[string]string, len(s.states))
	for k, v := range s.states {
		copy[k] = v
	}
	s.mu.RUnlock()

	status := Status{
		Uptime:    time.Since(s.startedAt).Round(time.Second).String(),
		StartedAt: s.startedAt,
		Ports:     copy,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}
