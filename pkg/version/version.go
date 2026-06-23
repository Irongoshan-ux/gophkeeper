// Package version exposes build metadata injected via ldflags.
package version

// BuildVersion is the release version of the binary.
var BuildVersion = "dev"

// BuildDate is the build timestamp.
var BuildDate = "unknown"

// BuildCommit is the source commit hash.
var BuildCommit = "unknown"

// Info returns formatted build metadata.
func Info() string {
	return "Build version: " + BuildVersion + "\nBuild date: " + BuildDate + "\nBuild commit: " + BuildCommit
}
