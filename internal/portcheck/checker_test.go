package portcheck_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/portcheck"
)

func startTCPServer(t *testing.T) (port int, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(portStr)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return p, func() { ln.Close() }
}

func TestCheck_Open(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	c := portcheck.New(2 * time.Second)
	res := c.Check("127.0.0.1", port)

	if res.State != portcheck.StateOpen {
		t.Errorf("expected StateOpen, got %s (err: %v)", res.State, res.Err)
	}
	if res.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestCheck_Closed(t *testing.T) {
	c := portcheck.New(500 * time.Millisecond)
	// Port 1 is almost certainly closed/refused in test environments.
	res := c.Check("127.0.0.1", 1)

	if res.State != portcheck.StateClosed {
		t.Errorf("expected StateClosed, got %s", res.State)
	}
	if res.Err == nil {
		t.Error("expected non-nil error for closed port")
	}
}

func TestStateString(t *testing.T) {
	cases := []struct {
		s    portcheck.State
		want string
	}{
		{portcheck.StateOpen, "open"},
		{portcheck.StateClosed, "closed"},
		{portcheck.StateUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
