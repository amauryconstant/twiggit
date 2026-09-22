# Spec Delta: Domain Path Utilities

## Purpose

Symlink-aware filesystem path helpers used by context detection and
project discovery. The helpers return plain `error` rather than typed
domain errors so that callers in `internal/service/` wrap at the service
boundary into the appropriate `domain.*ServiceError`. See
`domain-typed-errors`.

## ADDED Requirements

### Requirement: `ExtractProjectFromWorktreePath`

`ExtractProjectFromWorktreePath(worktreePath, worktreesDir)` SHALL
return the project name from a path under
`{worktreesDir}/{project}/{branch}/...`. On invalid input, the function
SHALL return a non-nil plain `error`.

#### Scenario: Valid worktree path

- **WHEN** `worktreePath = /home/u/Worktrees/myapp/feature`
- **AND** `worktreesDir = /home/u/Worktrees`
- **THEN** the system SHALL return the string `"myapp"`
- **AND** the returned error SHALL be `nil`

#### Scenario: Path not under worktrees dir

- **WHEN** `worktreePath` does not start with `worktreesDir + /`
- **THEN** the system SHALL return an empty string
- **AND** the returned error SHALL be `nil`

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `NormalizePath`

`NormalizePath(path)` SHALL return the absolute, symlink-resolved form
of `path`. When symlink resolution fails, the function SHALL fall back
to the absolute path without returning an error.

#### Scenario: Symlink resolution

- **WHEN** `path` is a symlink to `/real/path`
- **THEN** `NormalizePath(path)` SHALL return `/real/path`
- **AND** the returned error SHALL be `nil`

#### Scenario: Symlink resolution failure

- **WHEN** `filepath.EvalSymlinks(abs)` fails (broken link)
- **THEN** the system SHALL return the absolute path anyway
- **AND** SHALL NOT return an error

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `IsPathUnder`

`IsPathUnder(base, target)` SHALL return `(true, nil)` when `target`
is under `base` after symlink resolution, `(false, nil)` otherwise. The
function SHALL reject `..` traversal by symlink resolution and return
`(false, error)` on `filepath.Abs` failure. The error SHALL be a plain
`error` (no typed chain); callers wrap at the service boundary.

#### Scenario: Target is under base

- **WHEN** `base = /home/u/Projects` and `target = /home/u/Projects/myapp`
- **THEN** the system SHALL return `(true, nil)`

#### Scenario: Target escapes base

- **WHEN** `base = /home/u/Projects` and `target = /etc/passwd`
- **THEN** the system SHALL return `(false, nil)`

#### Scenario: Empty path rejected

- **WHEN** either `base` or `target` is empty
- **THEN** the system SHALL return `(false, error)`
- **AND** the error SHALL NOT be a typed domain error (callers wrap at
  the service boundary into `domain.New*ServiceError`)

#### Scenario: Symlink escape blocked

- **WHEN** `target` is a symlink that resolves outside `base`
- **THEN** the system SHALL return `(false, nil)` after symlink
  resolution

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
