# Spec Delta: cli-iostreams

## MODIFIED Requirements

### Requirement: Verbose and Logger are distinct channels

`IOStreams.Verbose bool` SHALL be set from the `-v` / `--verbose` flag and SHALL gate `Verbosef` output to stderr. `IOStreams.Logger *slog.Logger` SHALL be the single debug logger for the binary and SHALL be used for all debug diagnostics in any package. The system SHALL provide exactly one constructor, `iostreams.NewLogger(w io.Writer) *slog.Logger`, which: reads the `TWIGGIT_DEBUG` environment variable exactly once at construction time, returns a `slog.Logger` whose handler level is `slog.LevelDebug` when `TWIGGIT_DEBUG` is non-empty and `slog.LevelWarn` otherwise, and writes through the supplied `io.Writer`. `iostreams.System()` SHALL construct the singleton by passing `os.Stderr`. The same `*slog.Logger` pointer SHALL be reachable from both `IOStreams.Logger` and the cmdutil command factory's logger accessor; the binary's entry point SHALL call `slog.SetDefault` with that pointer exactly once before dispatch. Pointer equality between the command factory's logger accessor and `IOStreams.Logger` SHALL be asserted by a unit test. When `TWIGGIT_DEBUG` is unset, every debug write SHALL be discarded regardless of the underlying writer.

#### Scenario: Logger emits only when TWIGGIT_DEBUG is set
- **WHEN** `TWIGGIT_DEBUG=1` is set and a command calls `ios.Logger.Debug("msg", "key", "value")`
- **THEN** the structured log line lands on `ios.ErrOut`

#### Scenario: Logger is silent without TWIGGIT_DEBUG
- **WHEN** `TWIGGIT_DEBUG` is unset and a command calls `ios.Logger.Debug("msg")`
- **THEN** nothing is written to `ios.ErrOut`

#### Scenario: singleton shared across the command factory and iostreams
- **WHEN** the binary's entry point constructs `iostreams.System()` and `cmdutil.NewFactory()` from the same `IOStreams`
- **THEN** the command factory's logger accessor and `IOStreams.Logger` are the same `*slog.Logger` pointer, and `slog.Default()` (after the binary entry point's `SetDefault` call) matches that same pointer

### Requirement: Test() returns buffers for command tests

`iostreams.Test()` SHALL return `(*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer)` where the second return is stdin, third is stdout, fourth is stderr. The returned IOStreams SHALL have `Verbose=false`, `Quiet=false`, `colorEnabled=false`. The returned IOStreams SHALL expose a `*slog.Logger` constructed by `iostreams.NewLogger(errOutBuffer)` so `TWIGGIT_DEBUG=1` writes land in the test stderr buffer rather than the host process's `os.Stderr`. The test path SHALL use the same `*slog.Logger` pointer as the production path (the same singleton guarantee that applies to `iostreams.System()`).

#### Scenario: Test buffers capture both streams independently
- **WHEN** a test calls `ios, stdin, stdout, stderr := iostreams.Test()` and runs a command against it
- **THEN** the test asserts on `stdout.String()` and `stderr.String()` independently

#### Scenario: test buffer captures debug output
- **WHEN** `TWIGGIT_DEBUG=1` is set in the test process and `iostreams.Test()` is called
- **THEN** calling `ios.Logger.Debug("hello")` writes "hello" into the stderr `bytes.Buffer` returned by `iostreams.Test()`, not to the host process's `os.Stderr`
