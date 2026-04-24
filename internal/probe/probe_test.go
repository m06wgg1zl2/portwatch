package probe_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/user/portwatch/internal/circuit"
	"github.com/user/portwatch/internal/probe"
)

func startTCP(t *testing.T) (host string, port int, stop func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return addr.IP.String(), addr.Port, func() { l.Close() }
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("free port: %v", err)
	}
	port, _ := strconv.Atoi(l.Addr().(*net.TCPAddr).AddrPort().Port().String())
	l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestRun_OpenPort(t *testing.T) {
	host, port, stop := startTCP(t)
	defer stop()

	p := probe.New(probe.Config{MaxAttempts: 1}, nil)
	r := p.Run(host, port)

	if !r.Open {
		t.Fatalf("expected open, got closed; err=%v", r.Err)
	}
	if r.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", r.Attempts)
	}
}

func TestRun_ClosedPort(t *testing.T) {
	p := probe.New(probe.Config{MaxAttempts: 2, InitialWait: time.Millisecond}, nil)
	r := p.Run("127.0.0.1", 1)

	if r.Open {
		t.Fatal("expected closed port")
	}
	if r.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", r.Attempts)
	}
	if r.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestRun_CircuitOpen_Skips(t *testing.T) {
	cb := circuit.New(circuit.Config{Threshold: 1, Timeout: 10 * time.Second})
	cb.RecordFailure() // trip the breaker

	p := probe.New(probe.Config{MaxAttempts: 3}, cb)
	r := p.Run("127.0.0.1", 1)

	if r.Open {
		t.Fatal("expected circuit to block probe")
	}
	if r.Err == nil {
		t.Error("expected circuit-open error")
	}
}

func TestRun_ElapsedNonZero(t *testing.T) {
	host, port, stop := startTCP(t)
	defer stop()

	p := probe.New(probe.Config{MaxAttempts: 1}, nil)
	r := p.Run(host, port)

	if r.Elapsed <= 0 {
		t.Errorf("expected positive elapsed, got %v", r.Elapsed)
	}
}
