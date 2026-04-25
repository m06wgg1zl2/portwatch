package fanout

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Reporter renders fanout result summaries as human-readable tables.
type Reporter struct {
	w io.Writer
}

// NewReporter returns a Reporter that writes to w.
func NewReporter(w io.Writer) *Reporter {
	return &Reporter{w: w}
}

// WriteTable writes a tabular summary of the provided results.
func (r *Reporter) WriteTable(results []Result) {
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SINK\tSTATUS")
	fmt.Fprintln(tw, "----\t------")
	for _, res := range results {
		status := "ok"
		if res.Error != nil {
			status = "error: " + res.Error.Error()
		}
		fmt.Fprintf(tw, "%s\t%s\n", res.Name, status)
	}
	_ = tw.Flush()
}

// Summary writes a one-line summary: total sinks, success count, error count.
func (r *Reporter) Summary(results []Result) {
	var errs []string
	for _, res := range results {
		if res.Error != nil {
			errs = append(errs, res.Name)
		}
	}
	ok := len(results) - len(errs)
	if len(errs) == 0 {
		fmt.Fprintf(r.w, "fanout: %d/%d sinks succeeded\n", ok, len(results))
	} else {
		fmt.Fprintf(r.w, "fanout: %d/%d sinks succeeded, failed: %s\n",
			ok, len(results), strings.Join(errs, ", "))
	}
}
