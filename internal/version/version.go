package version

import "fmt"

var (
	// Version is the current release version.
	Version = "0.1.0"
	// Commit is the git commit hash set at build time.
	Commit = "none"
	// Date is the build date set at build time.
	Date = "unknown"
)

// String returns a human-readable version string.
func String() string {
	return fmt.Sprintf("portwatch v%s (commit=%s, built=%s)", Version, Commit, Date)
}
