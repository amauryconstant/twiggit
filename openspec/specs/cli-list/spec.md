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

#### Scenario: Empty result

- **WHEN** user runs `twiggit list` and no worktrees exist
- **THEN** system SHALL display "No worktrees found"
- **AND** SHALL exit cleanly

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

The system SHALL display `(modified)` for dirty worktrees, `(detached)`
for detached HEAD, and no indicator for clean attached worktrees.

#### Scenario: Dirty worktree

- **WHEN** a worktree has uncommitted changes
- **THEN** system SHALL display `(modified)` next to its branch name
- **AND** SHALL still show the path and branch

#### Scenario: Detached HEAD

- **WHEN** a worktree's HEAD is detached
- **THEN** system SHALL display `(detached)` next to its branch name

#### Scenario: Clean attached worktree

- **WHEN** a worktree is clean and HEAD is attached
- **THEN** system SHALL display no status indicator
- **AND** SHALL show only branch name and path

### Requirement: JSON output via `--output`

The system SHALL accept `--output=json`/`-o json` to emit a JSON object
suitable for scripting. Data SHALL go to stdout; errors and progress
SHALL go to stderr.

#### Scenario: JSON output

- **WHEN** user runs `twiggit list -o json` from a project context
- **THEN** system SHALL emit JSON with shape:
  `{"worktrees": [{"branch": "...", "path": "...", "status": "clean|modified|detached"}]}`
- **AND** SHALL exit with status 0

#### Scenario: JSON `--all`

- **WHEN** user runs `twiggit list --all -o json`
- **THEN** system SHALL emit JSON with project context per worktree
- **AND** main worktrees SHALL be excluded
