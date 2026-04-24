package replay_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/replay"
)

func makeAlert(host string, port int) alert.Alert {
	return alert.New(host, port, alert.LevelInfo, "open")
}

func TestPush_IncreasesLen(t *testing.T) {
	b := replay.New(replay.Config{Capacity: 10})
	b.Push(makeAlert("localhost", 8080))
	b.Push(makeAlert("localhost", 8081))
	if got := b.Len(); got != 2 {
		t.Fatalf("expected len 2, got %d", got)
	}
}

func TestPush_EvictsOldestWhenFull(t *testing.T) {
	b := replay.New(replay.Config{Capacity: 3})
	for i := 8080; i <= 8083; i++ {
		b.Push(makeAlert("localhost", i))
	}
	if got := b.Len(); got != 3 {
		t.Fatalf("expected len 3 after eviction, got %d", got)
	}
	var ports []int
	b.Replay(func(a alert.Alert) { ports = append(ports, a.Port) })
	if ports[0] != 8081 {
		t.Errorf("expected oldest surviving port 8081, got %d", ports[0])
	}
}

func TestReplay_OrderPreserved(t *testing.T) {
	b := replay.New(replay.Config{Capacity: 10})
	expected := []int{9001, 9002, 9003}
	for _, p := range expected {
		b.Push(makeAlert("host", p))
	}
	var got []int
	b.Replay(func(a alert.Alert) { got = append(got, a.Port) })
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("index %d: expected %d, got %d", i, p, got[i])
		}
	}
}

func TestReplay_TTL_SkipsOldEntries(t *testing.T) {
	b := replay.New(replay.Config{Capacity: 10, TTL: 50 * time.Millisecond})
	b.Push(makeAlert("host", 1111))
	time.Sleep(80 * time.Millisecond)
	b.Push(makeAlert("host", 2222))

	var ports []int
	b.Replay(func(a alert.Alert) { ports = append(ports, a.Port) })
	if len(ports) != 1 || ports[0] != 2222 {
		t.Errorf("expected only port 2222 after TTL, got %v", ports)
	}
}

func TestClear_RemovesAll(t *testing.T) {
	b := replay.New(replay.Config{Capacity: 10})
	b.Push(makeAlert("host", 3000))
	b.Push(makeAlert("host", 3001))
	b.Clear()
	if got := b.Len(); got != 0 {
		t.Fatalf("expected len 0 after clear, got %d", got)
	}
}

func TestNew_DefaultCapacity(t *testing.T) {
	b := replay.New(replay.Config{})
	for i := 0; i < 70; i++ {
		b.Push(makeAlert("host", 5000+i))
	}
	if got := b.Len(); got != 64 {
		t.Errorf("expected default capacity 64, got %d", got)
	}
}
