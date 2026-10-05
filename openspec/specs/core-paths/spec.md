# Capability: Paths

## Purpose

Documents the path-utility helpers in `internal/core/pathutils.go`, used by the context resolver and project discovery.

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

`core.ExtractProjectFromWorktreePath`, `core.NormalizePath`, and `core.IsPathUnder` SHALL return plain `error` values (no `domain.*` wrapper — the `domain` package does not exist in the Tier 2 layout). Callers in `internal/git/context_resolver.go` wrap as needed. Returned error strings SHALL be lowercase without trailing punctuation. Any error wrapper a caller adds SHALL use the `Error` suffix.

#### Scenario: Helper returns plain error

- **WHEN** `NormalizePath` fails on a non-existent path
- **THEN** the returned error's type is `*os.PathError` (or similar stdlib error), NOT a `*core.OperationError`

### Requirement: core package performs no I/O

The `internal/core/` package SHALL NOT perform filesystem `Stat`/`Open`/`Read`/`Write`/`os.Stat` calls. Path helpers in `core-paths` accept paths as strings and return strings/errors; the on-disk inspection lives in `internal/git/repo_inspector.go` (`IsMainRepo`, `FindMainRepoByTraversal`), which the git adapter consumes via a `RepoInspector` role interface when it needs to classify a path.

#### Scenario: Path helper does not touch the filesystem

- **WHEN** a caller invokes `core.NormalizePath("/tmp/missing")`
- **THEN** the call SHALL NOT call `os.Stat`, `os.Lstat`, `filepath.EvalSymlinks`, or any other filesystem-touching function
- **AND** SHALL return the absolute form plus an error when symlink resolution cannot run

#### Scenario: Repo introspection lives in the git adapter

- **WHEN** the cmd layer needs to know whether a path is a main repo
- **THEN** it SHALL route through `internal/git/repo_inspector.go`, not through `internal/core/`
- **AND** `internal/core/` SHALL NOT export any `Is*Repo`, `Find*`, or filesystem-probing helper
