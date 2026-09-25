# Spec Delta: cli-iostreams

## MODIFIED Requirements

### Requirement: Debug logger is a single shared instance bound to `IOStreams.ErrOut`

`IOStreams.Verbose bool` SHALL be set from the `-v` / `--verbose` flag and SHALL gate `Verbosef` output to stderr. `IOStreams.Logger *slog.Logger` SHALL be the single debug logger for the binary and SHALL be used for all debug diagnostics in any package. The system SHALL provide exactly one constructor, `iostreams.NewLogger(w io.Writer) *slog.Logger`, which: reads the `TWIGGIT_DEBUG` environment variable exactly once at construction time, returns a `slog.Logger` whose handler level is `slog.LevelDebug` when `TWIGGIT_DEBUG` is non-empty and `slog.LevelWarn` otherwise, and writes through the supplied `io.Writer`. `iostreams.System()` SHALL construct the singleton by passing `os.Stderr`; `iostreams.Test()` SHALL construct the singleton by passing `ios.ErrOut` so test buffers capture debug writes. `Factory.Logger` SHALL return the same `*slog.Logger` pointer that `IOStreams.Logger` holds; `main.go` SHALL call `slog.SetDefault` with that pointer exactly once before dispatch. Pointer equality between `Factory.Logger()` and `IOStreams.Logger` SHALL be asserted by a unit test. When `TWIGGIT_DEBUG` is unset, every debug write SHALL be discarded regardless of the underlying writer.

#### Scenario: TWIGGIT_DEBUG gates debug output
- **WHEN** `TWIGGIT_DEBUG=1` is set and a command calls `ios.Logger.Debug("msg", "key", "value")`
- **THEN** the structured log line lands on `ios.ErrOut`

#### Scenario: TWIGGIT_DEBUG unset suppresses debug output
- **WHEN** `TWIGGIT_DEBUG` is unset and a command calls `ios.Logger.Debug("msg")`
- **THEN** nothing is written to `ios.ErrOut`

#### Scenario: singleton shared across Factory and iostreams
- **WHEN** `main.go` constructs `iostreams.System()` and `cmdutil.NewFactory()` from the same `IOStreams`
- **THEN** `Factory.Logger()` and `IOStreams.Logger` are the same `*slog.Logger` pointer, and `slog.Default()` (after `main.go`'s `SetDefault` call) matches that same pointer

#### Scenario: test buffer captures debug output
- **WHEN** `TWIGGIT_DEBUG=1` is set in the test process and `iostreams.Test()` returns a logger
- **THEN** calling `ios.Logger.Debug("hello")` writes "hello" into the stderr `bytes.Buffer` returned by `iostreams.Test()`, not to the host process's `os.Stderr`
