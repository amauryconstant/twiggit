# Spec Delta

## Purpose

Defines the 0/1/2 exit-code contract for the CLI binary so scripts that key on exit codes remain portable across this change.

## ADDED Requirements

### Requirement: ExitOK = 0, ExitError = 1, ExitUsage = 2

The `cmdutil` package SHALL export three exit-code constants: `ExitOK` (0), `ExitError` (1), `ExitUsage` (2). The cmd layer SHALL NOT define additional exit-code constants.

#### Scenario: Successful command exits 0
- **WHEN** a command completes without error
- **THEN** `main` returns `ExitOK` (0) to the OS

#### Scenario: Validation error exits 1
- **WHEN** a command returns a `*core.ValidationError`
- **THEN** `main` returns `ExitError` (1) to the OS

#### Scenario: Usage error exits 2
- **WHEN** a command returns a `*core.UsageError`
- **THEN** `main` returns `ExitUsage` (2) to the OS

### Requirement: ExitCodeFor dispatches on error type via errors.As

`cmdutil.ExitCodeFor(err error) int` SHALL dispatch first via `errors.As(err, &*core.UsageError{})`, returning `ExitUsage` for any match; otherwise returning `ExitError` for any non-nil error and `ExitOK` for nil.

#### Scenario: errors.As walk reaches UsageError through wrapping
- **WHEN** the error chain is `fmt.Errorf("wrapped: %w", core.NewUsageError("bad arg"))`
- **THEN** `ExitCodeFor` returns `ExitUsage` (2)

#### Scenario: ValidationError does not dispatch to Usage
- **WHEN** the error is `core.NewValidationError(...)`
- **THEN** `ExitCodeFor` returns `ExitError` (1), not `ExitUsage` (2)

### Requirement: ValidationError and NotFoundError return ExitError (not ExitUsage)

The dispatch table SHALL NOT classify validation errors or not-found errors as usage errors. Both SHALL exit 1.

#### Scenario: NotFoundError exits 1
- **WHEN** the user requests a worktree that does not exist
- **THEN** `core.NewNotFoundError(...)` flows through `ExitCodeFor` and returns `ExitError` (1), not `ExitUsage` (2)
