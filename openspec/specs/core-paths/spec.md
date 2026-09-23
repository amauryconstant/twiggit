# Capability: Paths

## Purpose

Documents the path-utility helpers in their new home at `internal/core/`, used by the context resolver and project discovery. The legacy `infrastructure-path-utils` spec is left intact per the deferred-migration non-goal in the proposal.

## Requirements

### Requirement: ExtractProjectFromWorktreePath returns (project string, err error)

`core.ExtractProjectFromWorktreePath(worktreePath, worktreesDir)` SHALL return the project name segment from a path under `{worktreesDir}/{project}/{branch}/...`. The function SHALL return `(project, nil)` on success.

#### Scenario: Extract project from a typical worktree path

- **WHEN** `worktreePath = "$HOME/Worktrees/myapp/feat/foo"` and `worktreesDir = "$HOME/Worktrees"`
- **THEN** the function returns `("myapp", nil)`

#### Scenario: Extract project fails when path is outside worktreesDir

- **WHEN** `worktreePath = "$HOME/Projects/myapp"` and `worktreesDir = "$HOME/Worktrees"`
- **THEN** the function returns `("", error)` describing the mismatch

### Requirement: NormalizePath returns the absolute symlink-resolved path

`core.NormalizePath(path)` SHALL return the absolute, symlink-resolved form of `path`. When `filepath.EvalSymlinks` fails, the function SHALL fall back to `filepath.Abs(path)` and return the absolute (non-symlink-resolved) form along with the evaluation error.

#### Scenario: Symlink resolves to target

- **WHEN** `path = "/tmp/link"` and `/tmp/link` is a symlink to `/tmp/real`
- **THEN** `NormalizePath` returns `("/tmp/real", nil)` (or equivalent absolute resolved path)

#### Scenario: Fallback when symlink resolution fails

- **WHEN** `path = "/tmp/missing"` and `EvalSymlinks` returns an error
- **THEN** `NormalizePath` returns the absolute form of `/tmp/missing` and a non-nil error indicating the resolution failure

### Requirement: IsPathUnder returns (bool, error) after symlink resolution

`core.IsPathUnder(base, target)` SHALL return `(true, nil)` when `target` is under `base` after symlink resolution, `(false, nil)` when it is not, and `(false, error)` on `filepath.Abs` failure. The function SHALL reject `..` traversal by symlink resolution.

#### Scenario: target is under base

- **WHEN** `base = "/tmp/base"` and `target = "/tmp/base/sub/file"`
- **THEN** `IsPathUnder` returns `(true, nil)`

#### Scenario: target escapes base via symlink

- **WHEN** `base = "/tmp/base"` and `target = "/tmp/sneaky"` where `/tmp/sneaky` is a symlink outside `/tmp/base`
- **THEN** after symlink resolution `IsPathUnder` returns `(false, nil)`; the `..` traversal is rejected

### Requirement: All failures returned as plain error (no domain wrapper)

`core.ExtractProjectFromWorktreePath`, `core.NormalizePath`, and `core.IsPathUnder` SHALL return plain `error` values (no domain wrapper). The `infrastructure-path-utils` rule that wrapped failures via `domain.NewContextDetectionError` is removed; callers in `internal/git/context_resolver.go` wrap as needed.

#### Scenario: Helper returns plain error

- **WHEN** `NormalizePath` fails on a non-existent path
- **THEN** the returned error's type is `*os.PathError` (or similar stdlib error), NOT a `*core.OperationError`
