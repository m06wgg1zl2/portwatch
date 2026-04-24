package portcheck

import (
	"fmt"
	"net"
	"time"
)

// State represents the availability state of a port.
type State int

const (
	StateUnknown State = iota
	StateOpen
	StateClosed
)

func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Result holds the outcome of a single port check.
type Result struct {
	Host    string
	Port    int
	State   State
	Latency time.Duration
	Err     error
}

// String returns a human-readable summary of the Result.
func (r Result) String() string {
	if r.Err != nil {
		return fmt.Sprintf("%s:%d %s (latency: %s, err: %v)", r.Host, r.Port, r.State, r.Latency.Round(time.Millisecond), r.Err)
	}
	return fmt.Sprintf("%s:%d %s (latency: %s)", r.Host, r.Port, r.State, r.Latency.Round(time.Millisecond))
}

// Checker probes TCP port availability.
type Checker struct {
	Timeout time.Duration
}

// New returns a Checker with the given timeout.
func New(timeout time.Duration) *Checker {
	return &Checker{Timeout: timeout}
}

// Check attempts a TCP dial to host:port and returns a Result.
func (c *Checker) Check(host string, port int) Result {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, c.Timeout)
	latency := time.Since(start)

	if err != nil {
		return Result{Host: host, Port: port, State: StateClosed, Latency: latency, Err: err}
	}
	conn.Close()
	return Result{Host: host, Port: port, State: StateOpen, Latency: latency}
}
