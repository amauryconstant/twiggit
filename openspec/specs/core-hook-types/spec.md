# Capability: Hook Types

## Purpose

Defines the hook-domain value objects used by the post-create hook runner and the rebase/sync hook lifecycle. Types live in `internal/core/hook_types.go`; the runner that consumes them lives in `git-hook-runner`.

## Requirements

### Requirement: HookConfig

`core.HookConfig` SHALL carry a list of `HookDefinition` entries grouped by `HookType` (currently `HookTypePostCreate`, `HookTypePreRebase`, `HookTypePostRebase`, and `HookTypePostSync`). The config loads from `[hooks.<lifecycle>]` blocks in `.twiggit.toml`.

#### Scenario: HookConfig decodes TOML block

- **WHEN** `.twiggit.toml` contains:
  ```toml
  [hooks.post-create]
  commands = ["mise trust", "npm install"]
  ```
- **THEN** `HookConfig.PostCreate` SHALL be `[]core.HookDefinition{{Command: "mise trust"}, {Command: "npm install"}}`

### Requirement: HookDefinition

`core.HookDefinition` SHALL be a struct with `Command string` (the shell command to execute), `WorkingDirectory string` (optional; defaults to the new worktree path), and `TimeoutSeconds int` (optional; defaults to `Config.Shell.HookTimeout`).

#### Scenario: Default working directory and timeout

- **WHEN** a `HookDefinition` is constructed with only `Command = "mise trust"`
- **THEN** `WorkingDirectory == ""` and `TimeoutSeconds == 0`
- **AND** the runner SHALL fall back to the new worktree path for `WorkingDirectory`
- **AND** SHALL fall back to `Config.Shell.HookTimeout` for the per-command deadline

### Requirement: HookType enum

`core.HookType` SHALL be an `int` type with an explicit `HookTypeUnknown` sentinel at `iota` position 0. The enum SHALL declare `HookTypePostCreate`, `HookTypePreRebase`, `HookTypePostRebase`, and `HookTypePostSync` as named variants in numeric order after the unknown sentinel. New variants added for future hook lifecycle stages SHALL be appended without renumbering.

#### Scenario: HookType unknown sentinel

- **WHEN** `var t core.HookType` is declared
- **THEN** `t == core.HookTypeUnknown` (iota 0)

#### Scenario: HookType exposes rebase and sync variants

- **WHEN** the `core.HookType` enum is read
- **THEN** `HookTypePreRebase`, `HookTypePostRebase`, and `HookTypePostSync` SHALL be defined as distinct constants
- **AND** no existing constant SHALL be renumbered

### Requirement: HookFailure

`core.HookFailure` SHALL carry `Command string`, `ExitCode int`, `Output string`, and `TimedOut bool`. The runner populates one `HookFailure` per non-zero exit. `TimedOut == true` distinguishes a process killed by `context.WithTimeout` from a process that exited non-zero on its own.

#### Scenario: Timed-out hook

- **WHEN** a hook command exceeds `Config.Shell.HookTimeout`
- **THEN** the runner SHALL record `HookFailure{TimedOut: true}` for that command
- **AND** SHALL continue with the next hook

### Requirement: HookResult uses HasExecuted and IsSuccessful

`core.HookResult` SHALL carry `HasExecuted bool` (true when at least one hook ran), `IsSuccessful bool` (true when every hook exited 0 or no hooks ran), and `Failures []core.HookFailure` (empty when no failures). The cmd layer reads these fields directly without string-keyed access.

#### Scenario: No hooks configured

- **WHEN** `.twiggit.toml` has no `post-create` block
- **THEN** `HookResult{HasExecuted: false, IsSuccessful: true, Failures: nil}` SHALL be returned

#### Scenario: One of three hooks fails

- **WHEN** `mise trust` exits 1, `npm install` exits 0, `bundle install` exits 0
- **THEN** `HookResult.HasExecuted == true`, `IsSuccessful == false`, `len(Failures) == 1`, `Failures[0].Command == "mise trust"`

### Requirement: HookRunRequest carries optional rebase and sync fields

`core.HookRunRequest` SHALL be extended with six optional string fields: `RebaseBase`, `RebaseOldTip`, `RebaseNewTip`, `RebaseResult`, `SyncRemote`, and `SyncBranch`. Each field SHALL default to the empty string when not set; the runner SHALL emit the corresponding environment variable only when the field is non-empty and the `HookType` matches the field's lifecycle stage. The seven existing fields SHALL remain unchanged; existing post-create callers SHALL compile without change.

#### Scenario: PostRebase request populates rebase fields

- **WHEN** a `HookRunRequest` is constructed with `HookType = HookTypePostRebase`, `RebaseBase = "main"`, `RebaseOldTip = "abc"`, `RebaseNewTip = "def"`, `RebaseResult = "clean"`
- **THEN** the runner SHALL set `TWIGGIT_REBASE_BASE`, `TWIGGIT_REBASE_OLD_TIP`, `TWIGGIT_REBASE_NEW_TIP`, and `TWIGGIT_REBASE_RESULT` in the hook process environment
- **AND** SHALL set `TWIGGIT_WORKTREE_PATH`, `TWIGGIT_PROJECT_NAME`, `TWIGGIT_BRANCH_NAME`, and `TWIGGIT_SOURCE_BRANCH` as for any other hook type

#### Scenario: PostSync request populates sync fields

- **WHEN** a `HookRunRequest` is constructed with `HookType = HookTypePostSync`, `SyncRemote = "origin"`, `SyncBranch = "main"`
- **THEN** the runner SHALL set `TWIGGIT_SYNC_REMOTE` and `TWIGGIT_SYNC_BRANCH` in the hook process environment

#### Scenario: PostCreate request compiles without rebase fields

- **WHEN** an existing call site constructs a `HookRunRequest` with `HookType = HookTypePostCreate` and does not set any rebase or sync fields
- **THEN** the call site SHALL compile unchanged
- **AND** the runner SHALL NOT emit any `TWIGGIT_REBASE_*` or `TWIGGIT_SYNC_*` environment variables

### Requirement: Hook runner sets rebase and sync environment variables

When the runner executes a `HookTypePreRebase`, `HookTypePostRebase`, or `HookTypePostSync` hook, it SHALL set the corresponding `TWIGGIT_REBASE_*` or `TWIGGIT_SYNC_*` environment variables in the hook process from the populated fields of the `HookRunRequest`. The existing five environment variables (`TWIGGIT_WORKTREE_PATH`, `TWIGGIT_PROJECT_NAME`, `TWIGGIT_BRANCH_NAME`, `TWIGGIT_SOURCE_BRANCH`, `TWIGGIT_MAIN_REPO_PATH`) SHALL be set for every hook lifecycle stage.

#### Scenario: PreRebase hook receives the base variable

- **WHEN** a `HookTypePreRebase` hook fires with `RebaseBase = "main"`
- **THEN** the hook process SHALL observe `TWIGGIT_REBASE_BASE=main`
- **AND** the hook process SHALL observe `TWIGGIT_WORKTREE_PATH=<worktree path>`

#### Scenario: PostRebase hook receives result and tip variables

- **WHEN** a `HookTypePostRebase` hook fires with `RebaseResult = "clean"`, `RebaseOldTip = "abc"`, `RebaseNewTip = "def"`
- **THEN** the hook process SHALL observe `TWIGGIT_REBASE_RESULT=clean`, `TWIGGIT_REBASE_OLD_TIP=abc`, `TWIGGIT_REBASE_NEW_TIP=def`

### Requirement: PreRebase hook failure aborts the rebase

The runner SHALL treat a non-zero exit code from any `HookTypePreRebase` hook as a hard failure: the rebase SHALL NOT proceed; the returned `HookResult` SHALL record the failure; the cmd layer SHALL surface the failure and exit with status 1. This is the opposite of `HookTypePostCreate` and `HookTypePostRebase`, whose failures are warnings.

#### Scenario: PreRebase hook exit code 1 stops the rebase

- **WHEN** a `HookTypePreRebase` hook exits with status 1
- **THEN** `git rebase` SHALL NOT be invoked
- **AND** `HookResult.IsSuccessful == false`
- **AND** the cmd layer SHALL exit with status 1

#### Scenario: PostRebase hook exit code 1 is a warning

- **WHEN** a `HookTypePostRebase` hook exits with status 1 after a successful rebase
- **THEN** the rebase outcome SHALL remain `RebaseOutcomeClean`
- **AND** `HookResult.IsSuccessful == false`
- **AND** the cmd layer SHALL print a warning but exit with status 0

### Requirement: HookConfig exposes pre-rebase, post-rebase, post-sync hook slices

`core.HookConfig` SHALL be extended with three new `[]HookDefinition` fields: `PreRebase`, `PostRebase`, and `PostSync`. Each field SHALL use a kebab-case `koanf` tag (`pre-rebase`, `post-rebase`, `post-sync`). The existing `PostCreate` field SHALL remain unchanged; existing post-create configs SHALL load without change. The four hook slices SHALL be mutually independent; a config may declare any subset including none.

#### Scenario: HookConfig exposes new fields

- **WHEN** the `core.HookConfig` struct is read
- **THEN** it SHALL declare `PreRebase`, `PostRebase`, and `PostSync` fields of type `[]HookDefinition`
- **AND** the `PostCreate` field SHALL be unchanged

#### Scenario: HookConfig loads pre-rebase from TOML

- **WHEN** a `.twiggit.toml` config file declares:
  ```toml
  [hooks]
  pre-rebase = [{ command = "git stash" }]
  ```
- **THEN** `core.HookConfig.PreRebase` SHALL contain that single definition

#### Scenario: HookConfig loads post-rebase from TOML

- **WHEN** a `.twiggit.toml` config file declares:
  ```toml
  [hooks]
  post-rebase = [{ command = "go build ./..." }]
  ```
- **THEN** `core.HookConfig.PostRebase` SHALL contain that single definition

#### Scenario: HookConfig loads post-sync from TOML

- **WHEN** a `.twiggit.toml` config file declares:
  ```toml
  [hooks]
  post-sync = [{ command = "echo synced" }]
  ```
- **THEN** `core.HookConfig.PostSync` SHALL contain that single definition