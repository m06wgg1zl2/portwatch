package correlation_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/correlation"
)

func makeAlert(host string, level alert.Level) alert.Alert {
	return alert.New(host, 9000, level, "test")
}

func TestAdd_NewGroup(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 2,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	groups := c.Groups()

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
}

func TestAdd_SameGroupAccumulates(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 2,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	c.Add("group-a", makeAlert("host2", alert.LevelCritical))

	members := c.Members("group-a")
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
}

func TestCorrelated_BelowMinEvents(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 3,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	c.Add("group-a", makeAlert("host2", alert.LevelWarn))

	if c.Correlated("group-a") {
		t.Error("expected not correlated below min events")
	}
}

func TestCorrelated_MetMinEvents(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 2,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	c.Add("group-a", makeAlert("host2", alert.LevelCritical))

	if !c.Correlated("group-a") {
		t.Error("expected correlated after meeting min events")
	}
}

func TestMembers_MissingGroup(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 1,
	})

	members := c.Members("nonexistent")
	if members != nil && len(members) != 0 {
		t.Errorf("expected empty members for missing group, got %v", members)
	}
}

func TestGroups_ReturnsAllKeys(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 1,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	c.Add("group-b", makeAlert("host2", alert.LevelInfo))
	c.Add("group-c", makeAlert("host3", alert.LevelCritical))

	groups := c.Groups()
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
}

func TestEviction_AfterWindow(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 50 * time.Millisecond,
		MinEvents: 2,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	time.Sleep(80 * time.Millisecond)
	c.Add("group-a", makeAlert("host2", alert.LevelWarn))

	// Only the second event should remain after eviction; min not met
	if c.Correlated("group-a") {
		t.Error("expected not correlated after window eviction")
	}
}

func TestClear_RemovesGroup(t *testing.T) {
	c := correlation.New(correlation.Config{
		Window: 5 * time.Second,
		MinEvents: 1,
	})

	c.Add("group-a", makeAlert("host1", alert.LevelWarn))
	c.Clear("group-a")

	groups := c.Groups()
	if len(groups) != 0 {
		t.Errorf("expected 0 groups after clear, got %d", len(groups))
	}
}
