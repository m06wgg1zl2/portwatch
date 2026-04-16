package version

import "fmt"

// Build-time variables set via ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Info holds structured version metadata.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
}

// Get returns the current version Info.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

// String returns a human-readable version string.
func String() string {
	return fmt.Sprintf("portwatch %s (commit=%s, built=%s)", Version, Commit, BuildDate)
}
