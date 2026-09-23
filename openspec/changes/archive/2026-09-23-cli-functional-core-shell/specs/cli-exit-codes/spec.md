# Spec Delta

## Purpose

Defines the 0/1/2 exit-code contract for the CLI binary so scripts that key on exit codes remain portable across this change. The constants and the dispatch helper live in `internal/cmdutil/`; the contract is owned by `cli-error-formatting`.

## ADDED Requirements

### Requirement: ExitCode type with ExitOK, ExitError, ExitUsage

The `cmdutil` package SHALL define `type ExitCode int` and SHALL export three constants of that type: `ExitOK` (`ExitCode` = 0), `ExitError` (`ExitCode` = 1), `ExitUsage` (`ExitCode` = 2). The cmd layer SHALL NOT define additional exit-code constants. See `cli-error-formatting` for the canonical contract and `core-errors` for the error types.

#### Scenario: Successful command exits 0
- **WHEN** a command completes without error
- **THEN** `main` returns `int(ExitOK)` (0) to the OS via `os.Exit(int(cmdutil.ExitCodeFor(err)))`

#### Scenario: Validation error exits 1
- **WHEN** a command returns a `*core.ValidationError`
- **THEN** `ExitCodeFor` returns `ExitError` and `main` exits 1

#### Scenario: Usage error exits 2
- **WHEN** a command returns a `*core.UsageError`
- **THEN** `ExitCodeFor` returns `ExitUsage` and `main` exits 2

### Requirement: ExitCodeFor dispatches on error type via errors.As

`cmdutil.ExitCodeFor(err error) ExitCode` SHALL dispatch first via `errors.As(err, &*core.UsageError{})`, returning `ExitUsage` for any match; otherwise returning `ExitError` for any non-nil error and `ExitOK` for nil. The previous `GetExitCodeForError` name is not used.

#### Scenario: errors.As walk reaches UsageError through wrapping
- **WHEN** the error chain is `fmt.Errorf("wrapped: %w", core.NewUsageError("bad arg"))`
- **THEN** `ExitCodeFor` returns `ExitUsage` (2)

#### Scenario: ValidationError does not dispatch to Usage
- **WHEN** the error is `core.NewValidationError(...)`
- **THEN** `ExitCodeFor` returns `ExitError` (1), not `ExitUsage` (2)

#### Scenario: NotFoundError does not dispatch to Usage
- **WHEN** the error is `core.NewNotFoundError("worktree", "feat/missing")` wrapping `core.ErrWorktreeNotFound`
- **THEN** `ExitCodeFor` returns `ExitError` (1), not `ExitUsage` (2)

### Requirement: ValidationError and NotFoundError return ExitError (not ExitUsage)

The dispatch table SHALL NOT classify validation errors or not-found errors as usage errors. Both SHALL exit 1 via `ExitError`.

#### Scenario: NotFoundError exits 1
- **WHEN** the user requests a worktree that does not exist
- **THEN** `core.NewNotFoundError(...)` flows through `ExitCodeFor` and returns `ExitError` (1), not `ExitUsage` (2)

### Requirement: SIGINT and SIGTERM bypass ExitCodeFor

When `signal.NotifyContext` cancels the in-flight command via SIGINT or SIGTERM, `main.go` SHALL exit with `130` (SIGINT) or `143` (SIGTERM) regardless of the error returned by the command. The signal-exit code is computed in `main.go` from `ctx.Err()` and bypasses `ExitCodeFor`.

#### Scenario: SIGINT exits 130
- **WHEN** the user sends SIGINT to the binary during a long-running command
- **THEN** `main` detects the cancellation via `ctx.Err() != nil` and calls `os.Exit(130)`

#### Scenario: SIGTERM exits 143
- **WHEN** the user sends SIGTERM to the binary
- **THEN** `main` exits 143
