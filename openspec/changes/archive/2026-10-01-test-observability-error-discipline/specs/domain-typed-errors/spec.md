# Spec Delta: domain-typed-errors

## MODIFIED Requirements

### Requirement: Sentinel catalog

The `core` package SHALL export the following sentinel errors as package variables of type `error`:
| Sentinel | Message |
|---|---|
| `ErrGitRepoNotFound` | `"core: git repository not found"` |
| `ErrWorktreeNotFound` | `"core: worktree not found"` |
| `ErrProjectNotFound` | `"core: project not found"` |
| `ErrResolutionNotFound` | `"core: resolution target not found"` |
| `ErrShellAlreadyInstalled` | `"core: shell wrapper already installed"` |
| `ErrShellNotInstalled` | `"core: shell wrapper not installed"` |
| `ErrInvalidShellType` | `"core: invalid shell type"` |
| `ErrInferenceFailed` | `"core: could not infer shell type"` |
| `ErrDetectionFailed` | `"core: shell detection failed"` |

Each sentinel's `Error()` message SHALL be `"core: <resource> <state>"`. Callers SHALL identify these sentinels exclusively through `errors.Is`. Each shell sentinel SHALL be returned as the `Cause` (or wrapped via `*core.OperationError.Cause`) of a `*core.OperationError` whose `Op` field starts with the prefix `shell.` (specifically `shell.already_installed`, `shell.not_installed`, `shell.invalid_type`, `shell.inference`, `shell.detection`); callers SHALL detect the sentinel with `errors.Is(err, core.ErrShellAlreadyInstalled)` (and the corresponding sentinel for the other four cases) — string comparison against `OperationError.Op` SHALL NOT be used by callers.

#### Scenario: Sentinel catalog is exported and stable
- **WHEN** a caller imports the `core` package
- **THEN** the sentinel identifiers and their messages SHALL match the table above exactly
- **AND** no sentinel SHALL be unexported, renamed, or repurposed without a capability-level spec change

#### Scenario: shell sentinel match via errors.Is
- **WHEN** a `*core.OperationError` is returned with `Op = "shell.already_installed"` and `Cause = core.ErrShellAlreadyInstalled`
- **THEN** `errors.Is(err, core.ErrShellAlreadyInstalled)` returns `true` and `err.Error()` contains `"shell wrapper already installed"`

#### Scenario: shell sentinel match across wrap layers
- **WHEN** the shell-installation layer returns `fmt.Errorf("install wrapper for bash: %w", opErr)` where `opErr.Cause = core.ErrShellAlreadyInstalled`
- **THEN** `errors.Is(err, core.ErrShellAlreadyInstalled)` returns `true` at every layer of the chain

#### Scenario: not-found sentinels remain matched via errors.Is
- **WHEN** a caller encounters a `*core.NotFoundError{Entity: "worktree", Name: "feat/x"}` returned through any wrapper chain
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` returns `true`

#### Scenario: not-found sentinels excluded from shell-sentinel matching
- **WHEN** a caller matches `errors.Is(err, core.ErrShellAlreadyInstalled)` against a non-shell error
- **THEN** the result is `false`
