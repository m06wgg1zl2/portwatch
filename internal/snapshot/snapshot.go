package snapshot

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// State holds the last known state for a monitored target.
type State struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Open      bool      `json:"open"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store persists and retrieves port states.
type Store struct {
	mu     sync.RWMutex
	states map[string]State
	path   string
}

// New creates a Store, loading existing state from path if present.
func New(path string) (*Store, error) {
	s := &Store{path: path, states: make(map[string]State)}
	if _, err := os.Stat(path); err == nil {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Set updates the state for a given key.
func (s *Store) Set(key string, st State) error {
	s.mu.Lock()
	st.UpdatedAt = time.Now()
	s.states[key] = st
	s.mu.Unlock()
	return s.persist()
}

// Get returns the state for a key and whether it exists.
func (s *Store) Get(key string) (State, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.states[key]
	return st, ok
}

// All returns a copy of all states.
func (s *Store) All() map[string]State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]State, len(s.states))
	for k, v := range s.states {
		copy[k] = v
	}
	return copy
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.states)
}

func (s *Store) persist() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.states, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
