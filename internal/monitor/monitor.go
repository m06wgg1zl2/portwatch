package monitor

import (
	"log"
	"time"

	"portwatch/internal/config"
	"portwatch/internal/notify"
	"portwatch/internal/portcheck"
)

// Monitor watches configured ports and triggers notifications on state changes.
type Monitor struct {
	cfg      *config.Config
	checker  *portcheck.Checker
	notifier *notify.Notifier
	states   map[string]portcheck.State
}

// New creates a new Monitor from the provided config.
func New(cfg *config.Config) *Monitor {
	return &Monitor{
		cfg:      cfg,
		checker:  portcheck.New(),
		notifier: notify.New(),
		states:   make(map[string]portcheck.State),
	}
}

// Run starts the monitoring loop. It blocks until the done channel is closed.
func (m *Monitor) Run(done <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(m.cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Run an immediate check before waiting for the first tick.
	m.checkAll()

	for {
		select {
		case <-ticker.C:
			m.checkAll()
		case <-done:
			log.Println("monitor: shutting down")
			return
		}
	}
}

func (m *Monitor) checkAll() {
	for _, target := range m.cfg.Targets {
		key := target.Host + ":" + target.Port
		newState := m.checker.Check(target.Host, target.Port)
		prev, seen := m.states[key]

		if !seen || prev != newState {
			if seen {
				log.Printf("monitor: state change %s %s -> %s", key, prev, newState)
			} else {
				log.Printf("monitor: initial state %s %s", key, newState)
			}
			m.states[key] = newState
			m.dispatch(target, newState)
		}
	}
}

func (m *Monitor) dispatch(target config.Target, state portcheck.State) {
	for _, cb := range target.Callbacks {
		if err := m.notifier.Send(cb, target.Host, target.Port, state.String()); err != nil {
			log.Printf("monitor: callback error for %s:%s — %v", target.Host, target.Port, err)
		}
	}
}
