# Capability: Error Formatting

## Purpose

User-friendly error rendering on stderr with actionable hints, exit-code
dispatch, and panic recovery. This spec is the cmd-side surface;
the error-type taxonomy itself is owned by `core-errors`.

## Requirements

### Requirement: Exit code mapping (3-code canonical)

The system SHALL map error categories to exit codes via the `cmdutil.ExitCodeFor(err) ExitCode` helper returning `type ExitCode int`. The mapping SHALL be exactly:

| Exit code | Constant | Meaning |
|---|---|---|
| 0 | `ExitOK` | Success |
| 1 | `ExitError` | General unclassified error, runtime failure, or recovered panic |
| 2 | `ExitUsage` | Usage error (invalid command syntax) |

The cmd layer SHALL NOT define additional exit-code constants. See `core-errors` for the canonical definitions; per-resource NotFound categories are distinguished in user-facing output by the per-resource NotFound hints requirement, not by exit code. `ExitCodeFor` SHALL dispatch first via `errors.AsType[*core.UsageError](err)`, returning `ExitUsage`; otherwise returning `ExitError` for any non-nil error and `ExitOK` for nil.

#### Scenario: Validation error → ExitCodeError

- **WHEN** a `core.ValidationError` reaches the formatter
- **THEN** system SHALL exit with code `ExitError` (1)

#### Scenario: Git error → ExitCodeError

- **WHEN** a `git.ExternalError` (whose embedded core type walks to `*core.OperationError`) reaches the formatter
- **THEN** system SHALL exit with code `ExitError` (1)

#### Scenario: NotFound → ExitCodeError

- **WHEN** an error matching `core.ErrGitRepoNotFound`, `core.ErrWorktreeNotFound`, `core.ErrProjectNotFound`, or `core.ErrResolutionNotFound` via `errors.Is` reaches the formatter
- **THEN** system SHALL exit with code `ExitError` (1)
- **AND** the formatter SHALL append the resource-specific hint per the Actionable hints requirement

#### Scenario: Success → ExitCodeSuccess

- **WHEN** the command returns no error
- **THEN** system SHALL exit with code `ExitOK` (0)

#### Scenario: Cobra usage error → ExitCodeUsage

- **WHEN** an error matching `*core.UsageError` via `errors.AsType[*core.UsageError](err)` reaches the formatter
- **THEN** system SHALL exit with code `ExitUsage` (2)

### Requirement: User-friendly messages

The system SHALL format error messages for users, omitting internal operation names and including actionable hints.

#### Scenario: Format ValidationError

- **WHEN** `core.ValidationError` is rendered
- **THEN** output SHALL be "Error: <message>"
- **AND** SHALL include any `Suggestions` lines

#### Scenario: Format WorktreeServiceError

- **WHEN** an error whose chain reaches `*core.OperationError` with `Op = "worktree.create"` (formerly `domain.WorktreeServiceError`) is rendered
- **THEN** output SHALL be "Error: <message>" with the worktree path and operation removed from the user-visible text

### Requirement: Type-matched dispatch

The system SHALL dispatch formatters via an explicit matcher strategy using `errors.AsType[T](err)` (Go 1.27 generic API) rather than reflection. Registration order SHALL place more specific matchers before generic ones. Formatters SHALL extract typed errors via a single `errors.AsType[*T]` call. The registry SHALL register formatters for the four core types in this order: `*core.ValidationError`, `*core.NotFoundError`, `*core.OperationError`, `*core.UsageError`. A generic fallback renders unknown error types without dereferencing a nil pointer. `output.FormatError` SHALL return the error without logging so the cmd boundary is the sole logging site. `FormatError` SHALL return a `*core.OperationError` or `*core.UsageError` for every expected failure; it SHALL never `panic`.

#### Scenario: Specific matcher wins

- **WHEN** `errors.AsType[*core.ValidationError](err)` returns a non-nil result
- **AND** `errors.AsType[*core.OperationError](err)` also returns a non-nil result
- **THEN** the ValidationError formatter SHALL win because it is registered first

#### Scenario: asType helper guards nil dereference

- **WHEN** a formatter receives an error that does not contain a matching typed error
- **THEN** the formatter SHALL fall back to the generic message rather than dereferencing a nil pointer

### Requirement: Actionable hints

The system SHALL attach hints to common error cases, pointing the user at remediation. For errors matching a per-resource NotFound sentinel via `errors.Is`, the system SHALL attach a resource-specific hint as documented in the table below.

| Sentinel | Hint |
|---|---|
| `core.ErrProjectNotFound` | "Use 'twiggit list --all' to see available projects" |
| `core.ErrWorktreeNotFound` | "Use 'twiggit list' to see available worktrees" |
| `core.ErrResolutionNotFound` | "Use 'twiggit list' to see available navigation targets" |
| `core.ErrGitRepoNotFound` | "Verify the repository path" |

For non-NotFound errors the system SHALL retain a generic hint ("Check your configuration and try again" or equivalent).

#### Scenario: Project not found

- **WHEN** an error matching `core.ErrProjectNotFound` via `errors.Is` is rendered
- **THEN** system SHALL append the project-specific hint from the hint table

#### Scenario: Worktree not found

- **WHEN** an error matching `core.ErrWorktreeNotFound` via `errors.Is` is rendered
- **THEN** system SHALL append the worktree-specific hint from the hint table

#### Scenario: OperationError suggestions render after the message

- **WHEN** `core.OperationError` carries `Suggestions = []string{"run 'twiggit init' first"}`
- **THEN** `FormatError` renders the message followed by a hint line containing the suggestion

#### Scenario: Quiet mode strips hints

- **WHEN** the formatter is configured for quiet mode
- **THEN** the hint lines SHALL be omitted from the rendered output
- **AND** quiet-mode behavior SHALL match `cli-quiet-mode`

### Requirement: Panic recovery

The system SHALL recover from panics in `main.go` via `defer`/`recover`, display "Internal error: <panic value>" to stderr, and exit with code 1. When `TWIGGIT_DEBUG=1` is set, the system SHALL append the full stack trace after the panic value.

#### Scenario: Recover from panic

- **WHEN** a panic occurs anywhere in the command tree
- **THEN** system SHALL write "Internal error: ..." to stderr
- **AND** SHALL exit with code 1

#### Scenario: Debug mode shows stack

- **WHEN** `TWIGGIT_DEBUG=1` is set and a panic occurs
- **THEN** system SHALL print the full stack trace in addition to the panic value

### Requirement: Debug mode details

The system SHALL honor `TWIGGIT_DEBUG=1` to surface internal operation names and error context normally hidden from users. `FormatError` SHALL print the user-facing message first, then a separator, then every entry of the wrapped error chain (`fmt.Errorf("%w", ...`) in order. The debug chain output goes to stderr.

#### Scenario: Debug shows internals

- **WHEN** `TWIGGIT_DEBUG=1` and a `core.OperationError` with `Op = "git.open"` is rendered
- **THEN** the rendered output SHALL include the operation name (normally hidden)

#### Scenario: Debug mode unwraps the chain

- **WHEN** `TWIGGIT_DEBUG=1` and the error chain is `fmt.Errorf("twiggit list: %w", core.NewOperationError("git.open", msg, io.EOF))`
- **THEN** `FormatError` writes the user message plus "twiggit list: <op error>" plus "EOF" to stderr in order

### Requirement: Signal exit codes bypass ExitCodeFor

Signal cancellation (SIGINT, SIGTERM) SHALL exit with conventional shell exit codes (130 for SIGINT, 143 for SIGTERM) and SHALL bypass `cmdutil.ExitCodeFor`. The detection point is `main.go`: when `ctx.Err() != nil` after `cobra.ExecuteContext(ctx)` returns, the binary maps the cause to 130/143 rather than routing the wrapped `*core.OperationError` through the formatter dispatch. This contract is referenced by `cli-main-entry-point` (signal-context setup); this requirement constrains only the exit-code mapping. The formatter SHALL NOT be invoked for signal-cancelled runs.

#### Scenario: SIGINT exits 130

- **WHEN** SIGINT is delivered while a command is running and `ctx.Err()` reports cancellation
- **THEN** `main.go` SHALL call `os.Exit(130)` directly
- **AND** `cmdutil.ExitCodeFor` SHALL NOT be invoked for the cancelled error

#### Scenario: SIGTERM exits 143

- **WHEN** SIGTERM is delivered while a command is running and `ctx.Err()` reports cancellation
- **THEN** `main.go` SHALL call `os.Exit(143)` directly
- **AND** `cmdutil.ExitCodeFor` SHALL NOT be invoked for the cancelled error

#### Scenario: Non-signal cancellation still routes through ExitCodeFor

- **WHEN** a command returns `*core.OperationError` wrapping `context.Canceled` without an OS signal
- **THEN** `cmdutil.ExitCodeFor` SHALL map it to `ExitError` (1)
- **AND** the SIGINT/SIGTERM branch SHALL NOT trigger
