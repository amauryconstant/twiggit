package version

// Version holds the version string, injected at build time by GoReleaser.
// Default value is "dev" for development builds.
var Version = "dev"

// Commit holds the full commit hash, injected at build time by GoReleaser.
// Default value is empty for development builds.
var Commit = ""

// Date holds the build date, injected at build time by GoReleaser.
// Default value is empty for development builds.
var Date = ""

// String returns a formatted version string.
// The format is "<version> (<commit>) <date>" with the following variations:
//   - When commit is empty: "<version> () "
//   - When date is empty but commit is present: "<version> (<commit>) "
//   - When both are present: "<version> (<commit>) <date>"
//
// Note: String() does NOT include the "twiggit " prefix - the command layer prepends it.
func String() string {
	if Commit == "" {
		// Dev build or no commit info: "<version> () "
		return Version + " () "
	}
	if Date == "" {
		// Build with commit but no date: "<version> (<commit>) "
		return Version + " (" + Commit + ") "
	}
	// Complete information: "<version> (<commit>) <date>"
	return Version + " (" + Commit + ") " + Date
}
