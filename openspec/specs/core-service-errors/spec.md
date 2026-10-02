# Capability: Service Errors

## Purpose

Defines the legacy error constructors in `internal/core/errors_legacy.go` that wrap canonical `core.Error` subtypes with resource-specific constructors. These constructors are retained for backwards-compatible error construction at service-layer call sites but are NOT promoted to canonical error types; the canonical taxonomy remains the four `core.Error` subtypes (`ValidationError`, `NotFoundError`, `OperationError`, `UsageError`) owned by `core-errors`.

## Requirements

### Requirement: Service-layer constructors return *core.OperationError

The constructors `core.NewConfigError`, `core.NewContextDetectionError`, `core.NewGitCommandError`, `core.NewGitRepositoryError`, and `core.NewGitWorktreeError` SHALL return `*core.OperationError` with `Op` set to the operation name (`config.load`, `context.detect`, `git.command`, `git.repository`, `git.worktree`). Each constructor SHALL populate `Entity`, `Message`, `Cause` (via wrapping with `%w`), and `Suggestions` (where applicable).

#### Scenario: ConfigError wraps koanf parse failure

- **WHEN** the koanf loader fails to parse `config.toml`
- **THEN** `err := core.NewConfigError(path, "parse failure", koanfErr)`
- **SHALL** return `*core.OperationError{Op: "config.load", Entity: "config", Message: path + ": parse failure", Cause: koanfErr}`
- **AND** `errors.AsType[*core.OperationError](err)` SHALL return a non-nil result

### Requirement: Error constructors follow naming rules

Every constructor SHALL use the `New` prefix (`NewConfigError`, not `ConfigError`); sentinel matches SHALL use the `Err` prefix; error-typed values SHALL use the `Error` suffix. The `core.NewXxxError` family SHALL NOT depend on `internal/git/`, `internal/cmdutil/`, `internal/output/`, or `cmd/` packages.

#### Scenario: core.NewConfigError has no I/O dependency

- **WHEN** `internal/core/errors_legacy.go` is compiled
- **THEN** it SHALL NOT import any package from `internal/git/`, `internal/cmdutil/`, `internal/output/`, or `cmd/`
- **AND** it SHALL return the canonical `*core.OperationError` subtype

### Requirement: Git error constructors delegate to adapter

`core.NewGitCommandError`, `core.NewGitRepositoryError`, and `core.NewGitWorktreeError` SHALL be convenience constructors for service-layer code paths that have not yet migrated to the `git.NewRepoError` / `git.NewWorktreeError` / `git.NewCommandError` family (owned by `git-client`). Both families produce `*core.OperationError` walks; new code SHOULD use the `git.New*` family per `git-client`'s "All failures wrapped via git.New*Error" requirement.

#### Scenario: Both error families walk to *core.OperationError

- **WHEN** a caller invokes `git.NewRepoError(path, msg, cause)` (the adapter family)
- **OR** `core.NewGitRepositoryError(path, msg, cause)` (the legacy service family)
- **THEN** `errors.AsType[*core.OperationError](err)` SHALL return a non-nil result in both cases
- **AND** `err.Error()` SHALL be lowercase without trailing punctuation per the `errors-lowercase-no-punct` rule