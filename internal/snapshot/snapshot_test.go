package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func tempFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "snapshot.json")
}

func TestNew_EmptyWhenNoFile(t *testing.T) {
	s, err := New(tempFile(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.All()) != 0 {
		t.Error("expected empty store")
	}
}

func TestSet_And_Get(t *testing.T) {
	s, _ := New(tempFile(t))
	st := State{Host: "localhost", Port: 8080, Open: true}
	if err := s.Set("localhost:8080", st); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, ok := s.Get("localhost:8080")
	if !ok {
		t.Fatal("expected state to exist")
	}
	if got.Host != "localhost" || got.Port != 8080 || !got.Open {
		t.Errorf("unexpected state: %+v", got)
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestGet_MissingKey(t *testing.T) {
	s, _ := New(tempFile(t))
	_, ok := s.Get("missing:9999")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestPersistence_ReloadFromDisk(t *testing.T) {
	path := tempFile(t)
	s1, _ := New(path)
	_ = s1.Set("host:1234", State{Host: "host", Port: 1234, Open: false})

	s2, err := New(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	st, ok := s2.Get("host:1234")
	if !ok {
		t.Fatal("expected persisted state")
	}
	if st.Port != 1234 || st.Open {
		t.Errorf("unexpected reloaded state: %+v", st)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s, _ := New(tempFile(t))
	_ = s.Set("a:1", State{Host: "a", Port: 1, Open: true})
	all := s.All()
	delete(all, "a:1")
	if _, ok := s.Get("a:1"); !ok {
		t.Error("original store mutated by All()")
	}
}

func TestNew_InvalidJSON(t *testing.T) {
	path := tempFile(t)
	_ = os.WriteFile(path, []byte("not json"), 0644)
	_, err := New(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
