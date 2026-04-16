package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry records a single state-change event.
type Entry struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
}

// History manages an in-memory log of state-change entries with optional
// persistence to a JSON file.
type History struct {
	mu      sync.RWMutex
	entries []Entry
	path    string
}

// New creates a History. If path is non-empty the file is loaded on startup.
func New(path string) (*History, error) {
	h := &History{path: path}
	if path != "" {
		if err := h.load(); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return h, nil
}

// Add appends a new entry and persists if a path is configured.
func (h *History) Add(host string, port int, state string) error {
	e := Entry{Host: host, Port: port, State: state, Timestamp: time.Now().UTC()}
	h.mu.Lock()
	h.entries = append(h.entries, e)
	h.mu.Unlock()
	if h.path != "" {
		return h.save()
	}
	return nil
}

// All returns a copy of all recorded entries.
func (h *History) All() []Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Entry, len(h.entries))
	copy(out, h.entries)
	return out
}

func (h *History) load() error {
	data, err := os.ReadFile(h.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &h.entries)
}

func (h *History) save() error {
	h.mu.RLock()
	data, err := json.MarshalIndent(h.entries, "", "  ")
	h.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0o644)
}
