package metrics

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"
)

// Reporter writes metric summaries to an io.Writer.
type Reporter struct {
	counter *Counter
}

// NewReporter wraps a Counter for reporting.
func NewReporter(c *Counter) *Reporter {
	return &Reporter{counter: c}
}

// WriteTable writes a formatted table of all counters to w.
func (r *Reporter) WriteTable(w io.Writer) error {
	snap := r.counter.Snapshot()
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tCOUNT\tLAST SEEN")
	for _, k := range keys {
		n, ts := r.counter.Get(k)
		last := "never"
		if !ts.IsZero() {
			last = ts.Format(time.RFC3339)
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\n", k, n, last)
	}
	return tw.Flush()
}

// Summary returns a one-line string with total event count across all keys.
func (r *Reporter) Summary() string {
	snap := r.counter.Snapshot()
	var total int64
	for _, v := range snap {
		total += v
	}
	return fmt.Sprintf("total events: %d across %d key(s)", total, len(snap))
}
