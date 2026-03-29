// Package buildinfo holds ldflags-injected metadata (shared by cmd and internal/tui).
package buildinfo

// These are overridden at link time via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
