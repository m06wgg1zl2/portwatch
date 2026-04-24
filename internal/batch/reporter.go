package batch

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Stats holds aggregate statistics collected by a tracked Batcher.
type Stats struct {
	TotalFlushed  int
	TotalAlerts   int
	LastFlushTime time.Time
}

// Reporter formats batch Stats for human consumption.
type Reporter struct {
	stats *Stats
}

// NewReporter returns a Reporter for the given Stats pointer.
func NewReporter(s *Stats) *Reporter {
	return &Reporter{stats: s}
}

// WriteTable writes a tab-separated summary table to w.
func (r *Reporter) WriteTable(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "METRIC\tVALUE")
	fmt.Fprintf(tw, "total_flushes\t%d\n", r.stats.TotalFlushed)
	fmt.Fprintf(tw, "total_alerts\t%d\n", r.stats.TotalAlerts)
	if r.stats.LastFlushTime.IsZero() {
		fmt.Fprintf(tw, "last_flush\t-\n")
	} else {
		fmt.Fprintf(tw, "last_flush\t%s\n", r.stats.LastFlushTime.Format(time.RFC3339))
	}
	tw.Flush()
}

// Summary writes a one-line summary to w.
func (r *Reporter) Summary(w io.Writer) {
	if r.stats.TotalFlushed == 0 {
		fmt.Fprintln(w, "batch: no flushes recorded")
		return
	}
	avg := float64(r.stats.TotalAlerts) / float64(r.stats.TotalFlushed)
	fmt.Fprintf(w, "batch: %d flushes, %d alerts total (avg %.1f per flush), last %s\n",
		r.stats.TotalFlushed,
		r.stats.TotalAlerts,
		avg,
		r.stats.LastFlushTime.Format(time.RFC3339),
	)
}
