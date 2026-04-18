package schedule_test

import (
	"testing"
	"time"

	"portwatch/internal/schedule"
)

func TestSchedule_MultipleTicksWithinWindow(t *testing.T) {
	s := schedule.New(schedule.Config{IntervalSeconds: 0})
	// Force a very short interval via the exported Ticker path.
	// We re-create using a tiny interval via Config workaround: use 1s and
	// accept the test takes ~2 ticks.
	_ = s

	// Direct integration: build schedule with 1-second interval and collect 2 ticks.
	s2 := schedule.New(schedule.Config{IntervalSeconds: 1})
	tk := s2.Ticker()
	defer tk.Stop()

	count := 0
	timeout := time.After(3 * time.Second)
	for count < 2 {
		select {
		case <-tk.C:
			count++
		case <-timeout:
			t.Fatalf("only received %d ticks before timeout", count)
		}
	}
}
