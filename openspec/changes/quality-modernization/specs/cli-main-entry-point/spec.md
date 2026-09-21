# Spec Delta

## MODIFIED Requirements

### Requirement: Panic recovery logs via slog with stack trace

The system SHALL recover from any panic in the command tree and
log the panic value plus `debug.Stack()` via `slog.Error`
(severity ERROR) with the panic value as a structured field. The
recovery handler SHALL also write a one-line user-facing message
to stderr that does not include the stack trace (the stack trace
is operator-facing via the log aggregator, not user-facing on the
terminal).

#### Scenario: Panic in a RunE logs structured error
- **WHEN** a panic occurs during command execution
- **THEN** `slog.Error` receives the panic value as a field
  named `panic` and the stack trace as a field named `stack`;
  the user sees a single line "Internal error: <panic value>"
  on stderr

#### Scenario: Debug flag enables user-visible stack trace
- **WHEN** the user sets `TWIGGIT_DEBUG=1`
- **THEN** the user-facing stderr line also includes the stack
  trace so a developer can diagnose locally

### Requirement: Main consumes HandleCLIError exit code

The main entry point SHALL use the exit code returned by the
error-formatter dispatch (`HandleCLIErrorWithCommand`) instead of
hardcoding `os.Exit(1)`. Config errors SHALL exit with code 3,
validation errors with code 5, not-found errors with code 6.

#### Scenario: Config error exits with code 3
- **WHEN** the configuration file is missing or invalid
- **THEN** the process exits with status 3 (config error) and
  the error message names the config file path

#### Scenario: Validation error exits with code 5
- **WHEN** a CLI argument fails validation (e.g., invalid
  branch name)
- **THEN** the process exits with status 5 (validation error)

#### Scenario: Not-found error exits with code 6
- **WHEN** a resource (worktree, project, branch) does not exist
  and the error type is in the NotFound sentinel set
- **THEN** the process exits with status 6 (not found)

## ADDED Requirements

### Requirement: Signal-aware context

The main entry point SHALL wrap the cobra command tree with
`signal.NotifyContext` (catching SIGINT and SIGTERM) so that
`cmd.Context()` reflects cancellation. Leaf commands SHALL call
`cmd.Context()` instead of `context.Background()` to receive the
signal. See `cli-signal-handling` for the propagation rules.

#### Scenario: SIGINT cancels the running command
- **WHEN** the user runs `twiggit prune --all` and presses
  Ctrl-C
- **THEN** the cobra context is cancelled and the prune stops
  cleanly without leaving partial state

### Requirement: Slog-based observability

The main entry point SHALL initialize `slog.Default()` with a
handler that writes to `c.ErrOrStderr()` (resolved at first
log call). Log level SHALL be `INFO` by default and `DEBUG`
when `TWIGGIT_DEBUG=1` is set. The logger SHALL be passed via
context (via `slog.NewContext` / `slog.FromContext`) to all
service and infrastructure call sites that emit logs.
