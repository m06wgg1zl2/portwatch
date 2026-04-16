package monitor

import (
	"net"
	"testing"
	"time"

	"portwatch/internal/config"
)

func freePort(t *testing.T) (string, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not bind: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf("%d", port), func() { l.Close() }
}

func TestMonitor_InitialStateOpen(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer l.Close()
	port := fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)

	cfg := &config.Config{
		IntervalSeconds: 1,
		Targets: []config.Target{
			{Host: "127.0.0.1", Port: port, Callbacks: nil},
		},
	}

	m := New(cfg)
	m.checkAll()

	key := "127.0.0.1:" + port
	if state, ok := m.states[key]; !ok || state.String() != "open" {
		t.Errorf("expected open, got %v (ok=%v)", state, ok)
	}
}

func TestMonitor_StateChangeTrigger(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)
	key := "127.0.0.1:" + port

	cfg := &config.Config{
		IntervalSeconds: 1,
		Targets: []config.Target{
			{Host: "127.0.0.1", Port: port, Callbacks: nil},
		},
	}

	m := New(cfg)
	m.checkAll()
	if m.states[key].String() != "open" {
		t.Fatal("expected open initially")
	}

	l.Close()
	time.Sleep(50 * time.Millisecond)
	m.checkAll()
	if m.states[key].String() != "closed" {
		t.Errorf("expected closed after listener closed, got %s", m.states[key])
	}
}

func TestMonitor_DoneChannel(t *testing.T) {
	cfg := &config.Config{IntervalSeconds: 60, Targets: nil}
	m := New(cfg)
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		m.Run(done)
		close(finished)
	}()
	close(done)
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after done channel closed")
	}
}
