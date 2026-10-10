# Spec Delta

## Purpose

Lets users see the diagnostic view of every worktree in the current project (or every project with `--all`): ahead/behind counts against the tracked base, merge readiness, dirty state, last-commit date, and a stale flag. Output is human-readable by default and pipeable via `--output json|table|plain` per `cli-output-formats`. The command is read-only; no side effects, no hooks.

## ADDED Requirements

### Requirement: `status` shows one row per worktree in the detected project

The system SHALL display the per-worktree diagnostic projection for every worktree in the current project context. The default output (no `--output` flag) SHALL render one row per worktree on stdout with the columns `BRANCH`, `PATH`, `AHEAD`, `BEHIND`, `BASE`, `MERGED`, `DIRTY`, `STALE` separated by tab-aligned padding, and SHALL exit 0. The main worktree SHALL be excluded.

#### Scenario: Single project, several worktrees

- **WHEN** the user runs `twiggit status` from inside a project that owns three worktrees (`feature/a`, `feature/b`, `feature/c`)
- **THEN** stdout contains three rows, one per worktree
- **AND** stdout does not contain the main worktree

#### Scenario: Empty project renders no lines

- **WHEN** the user runs `twiggit status` and no worktrees exist for the project
- **THEN** stdout is empty (no rows, no banner)

### Requirement: `--all` shows worktrees from every project

The system SHALL accept `--all` / `-a` to list worktrees from every project under the configured `ProjectsDirectory`. When `--all` is set, the output SHALL prepend each row's project name as the leading column. The main worktree SHALL still be excluded per project. Without `--all`, the project name is not surfaced (single-project mode omits the column to keep the output narrow).

#### Scenario: Multiple projects, worktrees present

- **WHEN** the user runs `twiggit status --all` and two projects each have worktrees
- **THEN** stdout contains rows for every worktree
- **AND** each row carries the project name as the leading column

#### Scenario: Single project under --all still shows the leading column

- **WHEN** the user runs `twiggit status --all` and only one project has worktrees
- **THEN** stdout contains the project name as the leading column on every row

### Requirement: Outside git without positional or `--all` returns a usage error

The system SHALL return a `core.UsageError` (exit code 2) when the user runs `twiggit status` from outside any git context without a positional `[project]` argument and without `--all`. The error SHALL suggest `--all` or running from within a project.

#### Scenario: Outside git, no flags

- **WHEN** the user runs `twiggit status` from a directory that is not inside any git repository
- **AND** no positional argument is supplied
- **AND** `--all` is not set
- **THEN** the system SHALL emit a usage error to stderr naming both options (`--all` and positional project)
- **AND** SHALL exit 2

### Requirement: `[project]` positional selects a single project

The system SHALL accept an optional `[project]` positional argument. When supplied, the system SHALL list worktrees only for that project. The argument SHALL be a `core.ProjectName` value; an unknown project name SHALL return a `*core.NotFoundError` (exit 1).

#### Scenario: Named project, worktrees present

- **WHEN** the user runs `twiggit status myproject`
- **AND** the project `myproject` has worktrees
- **THEN** stdout contains only the worktrees of `myproject`

#### Scenario: Unknown project

- **WHEN** the user runs `twiggit status no-such-project`
- **THEN** the system SHALL emit a `*core.NotFoundError` naming the missing project
- **AND** SHALL exit 1

### Requirement: `--output` accepts `json`, `table`, `plain`

The system SHALL accept `--output` / `-o` with values `json`, `table`, `plain`. The value SHALL be resolved before the walk begins; an unknown value SHALL return a `core.UsageError` and exit 2. The vocabulary, stream discipline, and shell-completion contract are owned by `cli-output-formats`.

#### Scenario: `--output json` emits a bare array

- **WHEN** the user runs `twiggit status --output json` and three worktrees exist
- **THEN** stdout contains a single JSON document: a bare array with one object per worktree
- **AND** the array's objects SHALL carry the fields `branch`, `path`, `base`, `ahead`, `behind`, `merged`, `dirty`, `last_commit_date`, `stale`, `skipped`, `skip_reason` per `core-worktree-status`
- **AND** the system SHALL exit 0

#### Scenario: Unknown output value

- **WHEN** the user runs `twiggit status --output xml`
- **THEN** the system SHALL return a `core.UsageError` naming the accepted values
- **AND** SHALL exit 2

#### Scenario: Empty collection under `--output json` yields `[]`

- **WHEN** the user runs `twiggit status --output json` and no worktrees exist
- **THEN** stdout is exactly `[]`

### Requirement: `--stale-behind` and `--stale-days` override the stale heuristic

The system SHALL accept `--stale-behind N` and `--stale-days N` integer flags. The flag values SHALL override the configuration defaults from `git-config` (the `Config.Status` sub-struct owned by `git-config`) for the duration of the invocation. The `IsStale` column SHALL be `true` when the worktree's `Behind` count is at or above `--stale-behind` OR the worktree's last commit is older than `--stale-days` from the current time.

#### Scenario: Behind threshold trips the stale column

- **WHEN** the user runs `twiggit status --stale-behind 5`
- **AND** a worktree is 6 commits behind its base
- **THEN** that worktree's `STALE` column is `true`

#### Scenario: Age threshold trips the stale column

- **WHEN** the user runs `twiggit status --stale-days 7`
- **AND** a worktree's last commit is 10 days old
- **THEN** that worktree's `STALE` column is `true`

#### Scenario: Both thresholds zero disables the heuristic

- **WHEN** the user runs `twiggit status --stale-behind 0 --stale-days 0`
- **THEN** the `STALE` column is `false` for every worktree

### Requirement: Per-worktree read failures are best-effort, never fatal

The system SHALL treat each per-worktree read as best-effort. A read failure on one worktree SHALL set the row's `IsSkipped` to `true` and populate `SkipReason` with a lowercase, no-trailing-punctuation reason; the remaining rows SHALL still be produced. The walk SHALL return no error after the loop completes successfully; a `*core.OperationError` SHALL be returned only when the walk cannot run at all (config load, context resolution, project discovery).

#### Scenario: One worktree's read fails

- **WHEN** the walk visits 5 worktrees and one of them fails the ahead/behind read
- **THEN** stdout contains 5 rows
- **AND** the failed worktree's row carries `IsSkipped=true` and a non-empty `SkipReason`
- **AND** the remaining 4 rows are complete
- **AND** a warning line is emitted on stderr for the failed worktree
- **AND** the system exits 0

#### Scenario: All worktrees fail

- **WHEN** the walk visits 3 worktrees and all three fail the per-worktree read
- **THEN** stdout contains 3 skipped rows
- **AND** the system exits 0
- **AND** each row's `SkipReason` names the failure

### Requirement: Per-worktree warnings land on stderr, not stdout

The system SHALL emit per-worktree skip warnings on stderr and SHALL suppress them under `--quiet` / `-q`. The data output (the table, JSON, or plain projection) SHALL land on stdout untouched so `twiggit status --output json | jq .` works.

#### Scenario: `--output json` pipeline

- **WHEN** the user runs `twiggit status --output json | jq '.[].branch'`
- **THEN** `jq` receives a valid bare array
- **AND** no warning text is interleaved into the JSON stream

### Requirement: `status` joins the `core` command group

The system SHALL register the `status` subcommand under the `core` command group alongside `list`, `create`, `delete`, and `prune`. `twiggit --help` SHALL list `status` in the `Core:` section. No alias SHALL be registered; the verb is short and discoverable.

#### Scenario: Help text groups `status` under `Core:`

- **WHEN** the user runs `twiggit --help`
- **THEN** the help output's `Core:` section includes `status`
- **AND** the `status` line carries its one-line description

### Requirement: Exit-code contract for `status`

The system SHALL exit 0 on success (including the all-skipped case), 1 on a `*core.OperationError` or `*core.NotFoundError`, 2 on a `*core.UsageError`. The signal-context bypass (130 / 143) and the formatter dispatch are owned by `cli-error-formatting`.

#### Scenario: Outside git without flags exits 2

- **WHEN** the user runs `twiggit status` from outside any git context with no flags
- **THEN** the system SHALL exit 2

#### Scenario: Unknown project exits 1

- **WHEN** the user runs `twiggit status missing-project`
- **THEN** the system SHALL exit 1
