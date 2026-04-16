package history

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Reporter formats history entries for human-readable output.
type Reporter struct {
	h *History
}

// NewReporter wraps a History for reporting.
func NewReporter(h *History) *Reporter {
	return &Reporter{h: h}
}

// WriteTable writes a tab-aligned table of entries to w.
func (r *Reporter) WriteTable(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TIMESTAMP\tHOST\tPORT\tSTATE")
	for _, e := range r.h.All() {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Host,
			e.Port,
			e.State,
		)
	}
	return tw.Flush()
}

// Summary returns a brief string with total event count.
func (r *Reporter) Summary() string {
	return fmt.Sprintf("total events: %d", len(r.h.All()))
}
