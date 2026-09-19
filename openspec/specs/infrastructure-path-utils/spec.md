# Capability: Path Utilities

## Purpose

Low-level filesystem path helpers used by context detection and
project discovery. Symlink-aware, cross-platform-safe.

## Requirements

### Requirement: `ExtractProjectFromWorktreePath`

`ExtractProjectFromWorktreePath(worktreePath, worktreesDir)` SHALL
return the project name from a path under
`{worktreesDir}/{project}/{branch}/...`.

#### Scenario: Valid worktree path

- **WHEN** `worktreePath = /home/u/Worktrees/myapp/feature`
- **AND** `worktreesDir = /home/u/Worktrees`
- **THEN** system SHALL return `"myapp"`

#### Scenario: Path not under worktrees dir

- **WHEN** `worktreePath` does not start with `worktreesDir + /`
- **THEN** system SHALL return `""` (no error)

### Requirement: `NormalizePath`

`NormalizePath(path)` SHALL return the absolute, symlink-resolved
form of `path`. When symlink resolution fails, the system SHALL
fall back to the absolute path.

#### Scenario: Symlink resolution

- **WHEN** `path` is a symlink to `/real/path`
- **THEN** `NormalizePath(path)` SHALL return `/real/path`

#### Scenario: Symlink resolution failure

- **WHEN** `filepath.EvalSymlinks(abs)` fails (broken link)
- **THEN** system SHALL return the absolute path anyway
- **AND** SHALL NOT error

### Requirement: `IsPathUnder`

`IsPathUnder(base, target)` SHALL return `(true, nil)` when `target`
is under `base` after symlink resolution, `(false, nil)` otherwise.
The function SHALL reject `..` traversal by symlink resolution and
return `(false, error)` on `filepath.Abs` failure.

#### Scenario: Target is under base

- **WHEN** `base = /home/u/Projects` and `target = /home/u/Projects/myapp`
- **THEN** system SHALL return `(true, nil)`

#### Scenario: Target escapes base

- **WHEN** `base = /home/u/Projects` and `target = /etc/passwd`
- **THEN** system SHALL return `(false, nil)`

#### Scenario: Empty path rejected

- **WHEN** either `base` or `target` is empty
- **THEN** system SHALL return `(false, *domain.ContextDetectionError)`

#### Scenario: Symlink escape blocked

- **WHEN** `target` is a symlink that resolves outside `base`
- **THEN** system SHALL return `(false, nil)` after symlink resolution

### Requirement: Error wrapping

All failures SHALL be wrapped via `domain.NewContextDetectionError`
per the infrastructure AGENTS error-wrapping rules.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
