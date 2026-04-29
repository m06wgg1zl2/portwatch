package heartbeat

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Reporter formats heartbeat state for human-readable output.
type Reporter struct {
	h *Heartbeat
}

// NewReporter creates a Reporter backed by the given Heartbeat.
func NewReporter(h *Heartbeat) *Reporter {
	return &Reporter{h: h}
}

// WriteTable writes a formatted table of heartbeat metrics to w.
func (r *Reporter) WriteTable(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "METRIC\tVALUE")
	fmt.Fprintf(tw, "status\t%s\n", r.h.Check())
	fmt.Fprintf(tw, "beats\t%d\n", r.h.Beats())
	fmt.Fprintf(tw, "missed\t%d\n", r.h.Missed())
	lb := r.h.LastBeat()
	if lb.IsZero() {
		fmt.Fprintf(tw, "last_beat\t%s\n", "never")
	} else {
		fmt.Fprintf(tw, "last_beat\t%s\n", lb.Format(time.RFC3339))
	}
	return tw.Flush()
}

// Summary writes a single-line summary to w.
func (r *Reporter) Summary(w io.Writer) {
	status := r.h.Check()
	lb := r.h.LastBeat()
	lastStr := "never"
	if !lb.IsZero() {
		lastStr = lb.Format(time.RFC3339)
	}
	fmt.Fprintf(w, "heartbeat: status=%s beats=%d missed=%d last_beat=%s\n",
		status, r.h.Beats(), r.h.Missed(), lastStr)
}
