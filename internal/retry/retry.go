package retry

import (
	"time"
)

// Policy defines retry behaviour for failed checks or notifications.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     float64 // multiplier applied to delay after each attempt
}

// Doer is a function that returns an error; retried on non-nil result.
type Doer func() error

// New returns a Policy with sensible defaults.
func New(maxAttempts int, delay time.Duration, backoff float64) *Policy {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if backoff <= 0 {
		backoff = 1.0
	}
	return &Policy{
		MaxAttempts: maxAttempts,
		Delay:       delay,
		Backoff:     backoff,
	}
}

// Do executes fn up to MaxAttempts times, waiting between attempts.
// Returns the last error if all attempts fail, nil on first success.
func (p *Policy) Do(fn Doer) error {
	delay := p.Delay
	var err error
	for i := 0; i < p.MaxAttempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < p.MaxAttempts-1 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * p.Backoff)
		}
	}
	return err
}

// Attempts returns how many times fn was called before success or exhaustion.
func (p *Policy) Attempts(fn Doer) (int, error) {
	delay := p.Delay
	var err error
	for i := 0; i < p.MaxAttempts; i++ {
		if err = fn(); err == nil {
			return i + 1, nil
		}
		if i < p.MaxAttempts-1 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * p.Backoff)
		}
	}
	return p.MaxAttempts, err
}
