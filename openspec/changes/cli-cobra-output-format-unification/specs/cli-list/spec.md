# Spec Delta

## REMOVED Requirements

### Requirement: JSON output via --output

**Reason**: The implementation drops the envelope shape (`{"worktrees":[…]}`) and emits a bare JSON array (`[…]`) for collections. The shape pin lives in `cli-output` and `cli-output-formats`; per-command JSON output declares the array shape directly.

**Migration**: See ADDED Requirement "JSON array output via --output". Scripts consuming the envelope shape must switch from `jq '.worktrees[]'` to `jq '.[]'`.

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: JSON array output via --output

The system SHALL accept `--output=json`/`-o json` to emit a bare JSON array suitable for scripting. Collections SHALL be emitted as a bare JSON array on stdout; the empty case yields `[]`. Data SHALL go to stdout; errors and progress SHALL go to stderr.

#### Scenario: JSON list output is a bare array

- **WHEN** the user runs `twiggit list -o json` from a project context
- **THEN** the system SHALL emit a bare JSON array on stdout with shape:
  `[{branch, path, status}, ...]`
- **AND** SHALL exit with status 0

#### Scenario: JSON list with --all is a bare array

- **WHEN** the user runs `twiggit list --all -o json`
- **THEN** the system SHALL emit a bare JSON array with project context per worktree
- **AND** main worktrees SHALL be excluded

#### Scenario: Empty worktrees yields empty array

- **WHEN** the user runs `twiggit list -o json` and no worktrees exist
- **THEN** the stdout payload is `[]` (an empty JSON array)