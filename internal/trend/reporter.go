package trend

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// Reporter formats trend state for human consumption.
type Reporter struct {
	tracker *Tracker
}

// NewReporter returns a Reporter backed by the given Tracker.
func NewReporter(tr *Tracker) *Reporter {
	return &Reporter{tracker: tr}
}

// WriteTable writes a tab-separated table of keys and their current trend
// direction to w.
func (r *Reporter) WriteTable(w io.Writer) error {
	r.tracker.mu.Lock()
	keys := make([]string, 0, len(r.tracker.buckets))
	for k := range r.tracker.buckets {
		keys = append(keys, k)
	}
	r.tracker.mu.Unlock()

	sort.Strings(keys)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tSAMPLES\tDIRECTION")
	for _, k := range keys {
		r.tracker.mu.Lock()
		n := len(r.tracker.buckets[k])
		r.tracker.mu.Unlock()
		d := r.tracker.Direction(k)
		fmt.Fprintf(tw, "%s\t%d\t%s\n", k, n, d)
	}
	return tw.Flush()
}

// Summary writes a one-line summary of all tracked keys to w.
func (r *Reporter) Summary(w io.Writer) error {
	r.tracker.mu.Lock()
	total := len(r.tracker.buckets)
	r.tracker.mu.Unlock()
	_, err := fmt.Fprintf(w, "trend: tracking %d key(s)\n", total)
	return err
}
