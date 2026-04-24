// Package probe provides configurable multi-attempt port probing with
// backoff and circuit-breaker integration.
package probe

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/backoff"
	"github.com/user/portwatch/internal/circuit"
	"github.com/user/portwatch/internal/portcheck"
)

// Config holds tuning parameters for a Probe.
type Config struct {
	MaxAttempts int           `json:"max_attempts"`
	InitialWait time.Duration `json:"initial_wait"`
	MaxWait     time.Duration `json:"max_wait"`
	Exponential bool          `json:"exponential"`
}

// Result is the outcome of a single probe run.
type Result struct {
	Host     string
	Port     int
	Open     bool
	Attempts int
	Elapsed  time.Duration
	Err      error
}

// Probe combines a port checker, backoff strategy and circuit breaker.
type Probe struct {
	checker *portcheck.Checker
	bo      *backoff.Backoff
	cb      *circuit.Breaker
}

// New creates a Probe from cfg. cb may be nil to disable circuit breaking.
func New(cfg Config, cb *circuit.Breaker) *Probe {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	boCfg := backoff.Config{
		Initial:     cfg.InitialWait,
		Max:         cfg.MaxWait,
		Exponential: cfg.Exponential,
	}
	return &Probe{
		checker: portcheck.New(),
		bo:      backoff.New(boCfg),
		cb:      cb,
	}
}

// Run probes host:port up to MaxAttempts times, honouring the circuit breaker
// and backing off between failures.
func (p *Probe) Run(host string, port int) Result {
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", host, port)

	for attempt := 1; attempt <= p.bo.Config().MaxAttempts; attempt++ {
		if p.cb != nil && !p.cb.Allow() {
			return Result{
				Host:     host,
				Port:     port,
				Attempts: attempt,
				Elapsed:  time.Since(start),
				Err:      fmt.Errorf("circuit open: skipping probe of %s", addr),
			}
		}

		state := p.checker.Check(host, port)
		open := state == portcheck.StateOpen

		if p.cb != nil {
			if open {
				p.cb.RecordSuccess()
			} else {
				p.cb.RecordFailure()
			}
		}

		if open {
			return Result{Host: host, Port: port, Open: true, Attempts: attempt, Elapsed: time.Since(start)}
		}

		if attempt < p.bo.Config().MaxAttempts {
			time.Sleep(p.bo.Next(attempt))
		}
	}

	return Result{
		Host:     host,
		Port:     port,
		Open:     false,
		Attempts: p.bo.Config().MaxAttempts,
		Elapsed:  time.Since(start),
		Err:      fmt.Errorf("port %s unreachable after %d attempts", addr, p.bo.Config().MaxAttempts),
	}
}
