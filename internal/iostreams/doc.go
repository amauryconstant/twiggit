// Package iostreams abstracts terminal I/O behind an IOStreams struct so
// commands never touch os.Stdout / os.Stderr directly.
//
// The struct carries In, Out, ErrOut streams plus TTY detection, the
// NO_COLOR-aware isColorEnabled flag, a Quiet gate for --quiet, a Verbosef
// helper for --verbose, and a *slog.Logger channel for TWIGGIT_DEBUG.
// System() and Test() constructors cover production and unit-test paths.
//
// # Logger cache lifecycle
//
// NewLogger caches the *slog.Logger it builds per writer in a package-level
// map. The cache is CLI-lifetime only: it is initialized at package load and
// is never cleared during the process. This delivers the TestLoggerPointer
// singleton guarantee — repeated calls with the same writer return the same
// pointer, so System() / Test() / Factory.Logger all share one instance
// and main.go's slog.SetDefault binds to the same one. The cache grows
// monotonically for the lifetime of the CLI invocation; since each CLI
// process issues a fixed, finite number of distinct writers (typically
// os.Stderr for production plus the test path's *bytes.Buffer), the
// unbounded map is safe in practice. Tests that exercise many distinct
// writers must not depend on the cache being reset between cases.
package iostreams
