# Capability: List Worktrees

## Purpose

Display worktrees for the current project (default) or all projects
(`--all`), with branch name, path, and status indicators. JSON output
is supported via `--output` for scripting. Excludes main worktrees by
default.

## Requirements

### Requirement: List worktrees from current context

The system SHALL list worktrees for the current project when invoked
without `--all`. Main worktrees SHALL be excluded from the output.

#### Scenario: List from project directory

- **WHEN** user runs `twiggit list` from inside a project directory or worktree
- **AND** worktrees exist for that project
- **THEN** system SHALL display branch name, path, and status per worktree
- **AND** SHALL exclude the main worktree from output

### Requirement: List all projects with `--all`

The system SHALL list worktrees from all projects when `--all`/`-a` is
provided. Main worktrees SHALL still be excluded.

#### Scenario: List across projects

- **WHEN** user runs `twiggit list --all`
- **AND** multiple projects have worktrees
- **THEN** system SHALL show project context (branch, path, status)
  per worktree
- **AND** main worktrees SHALL be excluded

### Requirement: Outside git without `--all`

The system SHALL return a usage error when the user runs `twiggit list`
from outside any git context without `--all`, suggesting the
`--all` flag or running from within a project.

#### Scenario: Outside git context

- **WHEN** user runs `twiggit list` from outside any git context
- **AND** `--all` is NOT passed
- **THEN** system SHALL return error indicating a project name is required
- **AND** SHALL suggest `--all` or running from a project context

### Requirement: Status indicators

When `--output` is not supplied (the empty string), `list` SHALL render each worktree on stdout as `BRANCH -> PATH` plus conditional suffixes: ` (modified)` if the worktree has uncommitted changes, ` (detached)` if HEAD is detached, no suffix if clean attached. Main worktree SHALL be excluded. One line per worktree; the system SHALL emit no lines when no worktrees exist for the project.

#### Scenario: Dirty worktree

- **WHEN** a worktree has uncommitted changes
- **THEN** system SHALL display `branch_name -> /path/to/worktree (modified)`
- **AND** SHALL still show the path and branch

#### Scenario: Detached HEAD

- **WHEN** a worktree's HEAD is detached
- **THEN** system SHALL display `branch_name -> /path/to/worktree (detached)`

#### Scenario: Clean attached worktree

- **WHEN** a worktree is clean and HEAD is attached
- **THEN** system SHALL display `branch_name -> /path/to/worktree` with no status indicator

#### Scenario: Empty project renders no lines

- **WHEN** the user runs `twiggit list` and no worktrees exist for the project
- **THEN** the system emits no lines (silent) rather than the legacy `"No worktrees found"` text

### Requirement: JSON array output via --output

The system SHALL accept `--output=json` (long form only; no `-o`
short on `list` per `cli-command-options-pattern`) to emit a bare
JSON array suitable for scripting. Collections SHALL be emitted as a
bare JSON array on stdout; the empty case yields `[]`. Data SHALL
go to stdout; errors and progress SHALL go to stderr. Emitted errors SHALL be lowercase without
trailing punctuation. Every git query SHALL execute under a `context.WithTimeout`.

#### Scenario: JSON list output is a bare array

- **WHEN** the user runs `twiggit list --output=json` from a project context
- **THEN** the system SHALL emit a bare JSON array on stdout with shape:
  `[{branch, path, status}, ...]`
- **AND** SHALL exit with status 0

#### Scenario: JSON list with --all is a bare array

- **WHEN** the user runs `twiggit list --all --output=json`
- **THEN** the system SHALL emit a bare JSON array with project context per worktree
- **AND** main worktrees SHALL be excluded

#### Scenario: Empty worktrees yields empty array

- **WHEN** the user runs `twiggit list -o json` and no worktrees exist
- **THEN** the stdout payload is `[]` (an empty JSON array)