package routing

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Reporter renders routing configuration as human-readable text.
type Reporter struct {
	router *Router
}

// NewReporter creates a Reporter backed by the given Router.
func NewReporter(r *Router) *Reporter {
	return &Reporter{router: r}
}

// WriteTable writes a tabular view of routes and their weights to w.
func (rp *Reporter) WriteTable(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "DESTINATION\tWEIGHT\tSHARE")
	fmt.Fprintln(tw, strings.Repeat("-", 30))
	total := rp.router.Total()
	for _, r := range rp.router.Routes() {
		share := float64(r.Weight) / float64(total) * 100
		fmt.Fprintf(tw, "%s\t%d\t%.1f%%\n", r.Name, r.Weight, share)
	}
	return tw.Flush()
}

// Summary writes a one-line summary of the routing configuration to w.
func (rp *Reporter) Summary(w io.Writer) {
	routes := rp.router.Routes()
	names := make([]string, len(routes))
	for i, r := range routes {
		names[i] = fmt.Sprintf("%s(%d)", r.Name, r.Weight)
	}
	fmt.Fprintf(w, "routing: %d destinations, total weight %d [%s]\n",
		len(routes), rp.router.Total(), strings.Join(names, ", "))
}
