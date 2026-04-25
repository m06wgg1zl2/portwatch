package quorum

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// Reporter renders quorum state to an io.Writer.
type Reporter struct {
	q *Quorum
}

// NewReporter creates a Reporter backed by q.
func NewReporter(q *Quorum) *Reporter {
	return &Reporter{q: q}
}

// WriteTable writes a formatted table of current observation counts to w.
func (r *Reporter) WriteTable(w io.Writer) error {
	r.q.mu.Lock()
	snap := make(map[string]int, len(r.q.obs))
	for k, v := range r.q.obs {
		snap[k] = len(v)
	}
	required := r.q.cfg.Required
	window := r.q.cfg.Window
	r.q.mu.Unlock()

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tOBSERVATIONS\tREQUIRED\tWINDOW")
	for k, cnt := range snap {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", k, cnt, required, window)
	}
	return tw.Flush()
}

// Summary writes a one-line summary to w.
func (r *Reporter) Summary(w io.Writer) {
	r.q.mu.Lock()
	keys := len(r.q.obs)
	req := r.q.cfg.Required
	r.q.mu.Unlock()
	fmt.Fprintf(w, "quorum: %d active key(s), consensus requires %d observation(s)\n", keys, req)
}
