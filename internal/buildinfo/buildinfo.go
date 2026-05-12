// Package buildinfo holds compile-time metadata injected via ldflags.
package buildinfo

import "runtime"

var (
	// Version is the semantic version of the binary.
	Version = "dev"
	// Commit is the git commit hash embedded at link time.
	Commit = "none"
	// Date is the build timestamp embedded at link time.
	Date = "unknown"
	// GoVersion is the Go toolchain used to compile the binary.
	GoVersion = "unknown"
)

func init() {
	if GoVersion == "unknown" {
		GoVersion = runtime.Version()
	}
}

// Info holds build metadata injected via ldflags (except GoVersion fallback).
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

// Get returns current build information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
	}
}
