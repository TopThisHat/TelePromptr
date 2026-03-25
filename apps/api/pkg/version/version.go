// Package version provides build and version information for the TelePromptr API.
package version

// Build-time version constants. These values are intended to be overridden
// via -ldflags during CI builds (e.g., -ldflags "-X ...version.Version=v1.2.3").
var (
	// Version is the semantic version of the API server.
	Version = "0.1.0-dev"

	// GitCommit is the short SHA of the git commit used for the build.
	GitCommit = "unknown"

	// BuildTime is the UTC timestamp of the build.
	BuildTime = "unknown"
)
