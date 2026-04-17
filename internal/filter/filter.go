package filter

import "time"

// Rule defines criteria for suppressing or allowing alerts.
type Rule struct {
	// OnlyBetween restricts alerts to a time window (hour of day, 0-23).
	FromHour int `json:"from_hour"`
	ToHour   int `json:"to_hour"`
	// States limits alerts to specific port states ("open", "closed").
	States []string `json:"states"`
}

// Filter decides whether an event should be forwarded.
type Filter struct {
	rules []Rule
}

// New creates a Filter with the given rules.
func New(rules []Rule) *Filter {
	return &Filter{rules: rules}
}

// Allow returns true when the event passes all configured rules.
// If no rules are configured every event is allowed.
func (f *Filter) Allow(state string, at time.Time) bool {
	if len(f.rules) == 0 {
		return true
	}
	for _, r := range f.rules {
		if !r.matchesHour(at) {
			return false
		}
		if !r.matchesState(state) {
			return false
		}
	}
	return true
}

func (r Rule) matchesHour(at time.Time) bool {
	if r.FromHour == 0 && r.ToHour == 0 {
		return true
	}
	h := at.Hour()
	if r.FromHour <= r.ToHour {
		return h >= r.FromHour && h < r.ToHour
	}
	// wraps midnight
	return h >= r.FromHour || h < r.ToHour
}

func (r Rule) matchesState(state string) bool {
	if len(r.States) == 0 {
		return true
	}
	for _, s := range r.States {
		if s == state {
			return true
		}
	}
	return false
}
