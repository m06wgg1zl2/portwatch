package grouping

import (
	"testing"
	"time"
)

func TestAdd_NewGroup(t *testing.T) {
	g := New(0)
	g.Add("web", "host-1")
	members := g.Members("web")
	if len(members) != 1 || members[0] != "host-1" {
		t.Fatalf("expected [host-1], got %v", members)
	}
}

func TestAdd_DuplicateKeyNotAdded(t *testing.T) {
	g := New(0)
	g.Add("web", "host-1")
	g.Add("web", "host-1")
	if len(g.Members("web")) != 1 {
		t.Fatal("duplicate key should not be added twice")
	}
}

func TestAdd_MultipleKeys(t *testing.T) {
	g := New(0)
	g.Add("web", "host-1")
	g.Add("web", "host-2")
	if len(g.Members("web")) != 2 {
		t.Fatal("expected two members")
	}
}

func TestMembers_MissingGroup(t *testing.T) {
	g := New(0)
	if g.Members("missing") != nil {
		t.Fatal("expected nil for missing group")
	}
}

func TestGroups_ReturnsNames(t *testing.T) {
	g := New(0)
	g.Add("alpha", "k1")
	g.Add("beta", "k2")
	names := g.Groups()
	if len(names) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(names))
	}
}

func TestRemove_KeyFromGroup(t *testing.T) {
	g := New(0)
	g.Add("web", "host-1")
	g.Add("web", "host-2")
	g.Remove("web", "host-1")
	members := g.Members("web")
	if len(members) != 1 || members[0] != "host-2" {
		t.Fatalf("unexpected members after remove: %v", members)
	}
}

func TestRemove_EmptyGroupDeleted(t *testing.T) {
	g := New(0)
	g.Add("web", "host-1")
	g.Remove("web", "host-1")
	if g.Members("web") != nil {
		t.Fatal("empty group should be removed")
	}
}

func TestRemove_MissingGroupNoOp(t *testing.T) {
	g := New(0)
	g.Remove("none", "k") // must not panic
}

func TestEviction_AfterTTL(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	g := New(5 * time.Second)
	g.clock = func() time.Time { return now }
	g.Add("web", "host-1")

	// advance clock past TTL
	g.clock = func() time.Time { return now.Add(10 * time.Second) }
	g.Add("db", "host-2") // triggers evict

	if g.Members("web") != nil {
		t.Fatal("stale group should have been evicted")
	}
	if g.Members("db") == nil {
		t.Fatal("fresh group should still exist")
	}
}

func TestMembers_ReturnsCopy(t *testing.T) {
	g := New(0)
	g.Add("web", "host-1")
	m := g.Members("web")
	m[0] = "tampered"
	if g.Members("web")[0] != "host-1" {
		t.Fatal("Members should return a copy")
	}
}
