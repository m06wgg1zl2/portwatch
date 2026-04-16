package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAdd_InMemory(t *testing.T) {
	h, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Add("localhost", 8080, "open"); err != nil {
		t.Fatal(err)
	}
	entries := h.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Host != "localhost" || entries[0].Port != 8080 || entries[0].State != "open" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestAdd_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = h.Add("example.com", 443, "closed")
	_ = h.Add("example.com", 443, "open")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 persisted entries, got %d", len(entries))
	}
}

func TestLoad_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	initial := []Entry{{Host: "h", Port: 22, State: "open"}}
	data, _ := json.Marshal(initial)
	_ = os.WriteFile(path, data, 0o644)

	h, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.All()) != 1 {
		t.Fatalf("expected pre-loaded entry")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	h, _ := New("")
	_ = h.Add("a", 1, "open")
	entries := h.All()
	entries[0].Host = "mutated"
	if h.All()[0].Host == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}
