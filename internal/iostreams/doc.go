// Package iostreams abstracts terminal I/O behind an IOStreams struct so
// commands never touch os.Stdout / os.Stderr directly.
//
// The struct carries In, Out, ErrOut streams plus TTY detection, the
// NO_COLOR-aware colorEnabled flag, a Quiet gate for --quiet, a Verbosef
// helper for --verbose, and a *slog.Logger channel for TWIGGIT_DEBUG.
// System() and Test() constructors cover production and unit-test paths.
package iostreams
