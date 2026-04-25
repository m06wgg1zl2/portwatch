package budget

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// Reporter formats budget statistics for human consumption.
type Reporter struct {
	b *Budget
}

// NewReporter returns a Reporter backed by b.
func NewReporter(b *Budget) *Reporter {
	return &Reporter{b: b}
}

// WriteTable writes a tab-separated table of per-key budget stats to w.
func (r *Reporter) WriteTable(w io.Writer, keys []string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tRATIO\tBREACHED")
	for _, k := range keys {
		ratio := r.b.Ratio(k)
		breached := r.b.Breached(k)
		fmt.Fprintf(tw, "%s\t%.4f\t%v\n", k, ratio, breached)
	}
	return tw.Flush()
}

// Summary writes a single-line summary to w.
func (r *Reporter) Summary(w io.Writer, keys []string) {
	breached := 0
	for _, k := range keys {
		if r.b.Breached(k) {
			breached++
		}
	}
	fmt.Fprintf(w, "budget: %d/%d keys breached threshold\n", breached, len(keys))
}
