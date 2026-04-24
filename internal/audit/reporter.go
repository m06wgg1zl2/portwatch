package audit

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// Reporter formats audit log entries for human-readable output.
type Reporter struct {
	log *Log
}

// NewReporter creates a Reporter backed by the given Log.
func NewReporter(l *Log) *Reporter {
	return &Reporter{log: l}
}

// WriteTable writes all audit events as an aligned table to w.
func (r *Reporter) WriteTable(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TIMESTAMP\tLEVEL\tSOURCE\tMESSAGE")
	fmt.Fprintln(tw, "---------\t-----\t------\t-------")
	for _, e := range r.log.All() {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			e.Level.String(),
			e.Source,
			e.Message,
		)
	}
	return tw.Flush()
}

// Summary writes a compact summary (total counts per level) to w.
func (r *Reporter) Summary(w io.Writer) error {
	counts := map[Level]int{}
	for _, e := range r.log.All() {
		counts[e.Level]++
	}
	total := counts[LevelInfo] + counts[LevelWarn] + counts[LevelError]
	_, err := fmt.Fprintf(w,
		"audit: total=%d info=%d warn=%d error=%d\n",
		total,
		counts[LevelInfo],
		counts[LevelWarn],
		counts[LevelError],
	)
	return err
}
