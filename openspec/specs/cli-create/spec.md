# Capability: Create Worktree

## Purpose

Create a new git worktree for a project/branch, with branch-name
validation, source-branch selection, context-aware project inference,
and optional post-create hook execution.

## Requirements

### Requirement: Create with explicit project/branch

The system SHALL create a worktree at `~/Worktrees/<project>/<branch>/`
when the user provides an explicit `project/branch` argument.

#### Scenario: Basic create

- **WHEN** user runs `twiggit create myproject/feature-branch`
- **AND** source branch exists (default `main`)
- **THEN** system SHALL create worktree at `~/Worktrees/myproject/feature-branch/`
- **AND** success message SHALL display worktree path and branch name
- **AND** parent directories SHALL be created automatically if missing

### Requirement: Infer project from context

When the user runs `twiggit create <branch>` without a project
prefix, the system SHALL infer the project from CWD via
`application.ContextService`. See `infrastructure-context-resolver`.

#### Scenario: Infer project

- **WHEN** user runs `twiggit create feature-branch` from inside a project
- **AND** a project is detected from CWD
- **THEN** system SHALL create the worktree for the inferred project
- **AND** SHALL use `Config.DefaultSourceBranch` (default `main`) as the source

### Requirement: Source branch selection

The system SHALL accept `--source <branch>` (no short flag) to override
the default source branch.

#### Scenario: Custom source branch

- **WHEN** user runs `twiggit create myproject/feature --source develop`
- **AND** `develop` exists in the project
- **THEN** system SHALL create the worktree from `develop`
- **AND** success message SHALL reflect the feature branch name

#### Scenario: Missing source branch

- **WHEN** user runs `twiggit create myproject/feature --source missing`
- **AND** `missing` does not exist
- **THEN** system SHALL return an error naming the missing branch
- **AND** SHALL NOT attempt creation

### Requirement: Branch name validation

The system SHALL validate the branch name before attempting creation
and return a `domain.ValidationError` on invalid input.

#### Scenario: Invalid branch name

- **WHEN** user provides an invalid branch name (empty, too long, or
  contains forbidden characters)
- **THEN** system SHALL return validation error before any git operation
- **AND** error message SHALL explain why the branch name is invalid
- **AND** error message SHALL suggest the valid branch name format

### Requirement: Outside git without explicit project

The system SHALL return a usage error when the user runs `twiggit create <branch>`
from outside any git context without a `project/` prefix.

#### Scenario: No context, no project

- **WHEN** user runs `twiggit create feature` from outside any git context
- **THEN** system SHALL return "cannot infer project: not in a project
  context and no project specified"
- **AND** SHALL NOT attempt creation

### Requirement: Duplicate worktree detection

The system SHALL detect attempts to create a worktree that already
exists and SHALL return a clear error.

#### Scenario: Worktree exists

- **WHEN** user runs `twiggit create myproject/feature`
- **AND** `~/Worktrees/myproject/feature/` already exists
- **THEN** system SHALL return error naming the existing path
- **AND** SHALL NOT modify the filesystem

#### Scenario: Branch exists, no worktree

- **WHEN** user runs `twiggit create myproject/feature`
- **AND** branch exists in the repo
- **AND** no worktree exists for it
- **THEN** system SHALL create the worktree on the existing branch

### Requirement: Post-create hooks

When the project has `.twiggit.toml` with `[hooks.post-create]`
configured, the system SHALL execute the configured commands after a
successful create. Failures SHALL be collected and reported but SHALL
NOT roll back the worktree. The hook runner contract is owned by
`infrastructure-hook-runner`.

#### Scenario: Hooks configured, all succeed

- **WHEN** worktree creation succeeds
- **AND** `.twiggit.toml` defines `[hooks.post-create].commands`
- **THEN** system SHALL execute all commands in order
- **AND** `CreateWorktreeResult.HookResult.Success` SHALL be `true`
- **AND** `CreateWorktreeResult.HookResult.Failures` SHALL be empty

#### Scenario: Hooks configured, one fails

- **WHEN** a hook command exits non-zero
- **THEN** system SHALL continue running remaining hooks
- **AND** SHALL collect the failure in `HookResult.Failures`
- **AND** SHALL display a warning that worktree is ready but setup may
  be incomplete

#### Scenario: No hooks configured

- **WHEN** worktree creation succeeds
- **AND** `.twiggit.toml` does not exist or has no `post-create` block
- **THEN** `CreateWorktreeResult.HookResult` SHALL be `nil`
- **AND** worktree info SHALL still be returned

### Requirement: `-C` / `--cd` navigation

When `-C`/`--cd` is passed, the system SHALL print the new worktree's
absolute path to stdout after success, for the shell wrapper to
consume.

#### Scenario: `-C` after create

- **WHEN** user runs `twiggit create -C myproject/feature`
- **AND** creation succeeds
- **THEN** system SHALL print the new worktree path to stdout
- **AND** success message SHALL go to stderr
