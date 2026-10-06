# Spec Delta

## Purpose

Lets users refresh a project's default branch (or a named branch) against
its remote tracking ref in one CLI verb, with optional follow-up rebase
of every worktree onto the updated base. Closes the loop that
`twiggit create` and `twiggit rebase` start by ensuring the base itself
is fresh before any worktree attempts to track it.

## ADDED Requirements

### Requirement: Sync the current project's default branch

The system SHALL fetch the remote tracking ref for the project's default
branch into the project main repository when the user runs
`twiggit sync`.

#### Scenario: Sync fetches origin/default-branch

- **WHEN** user runs `twiggit sync` from a project whose default branch
  is `main` and whose default remote is `origin`
- **THEN** the system SHALL run `git fetch origin` in the project main
  repo
- **AND** SHALL update the local `main` tracking ref to match the
  upstream tip
- **AND** SHALL print a per-branch summary to stderr

### Requirement: Sync a named branch instead of the default

The system SHALL fetch the remote tracking ref for the named branch when
`--branch <name>` is supplied.

#### Scenario: --branch fetches a non-default branch

- **WHEN** user runs `twiggit sync --branch develop`
- **THEN** the system SHALL run `git fetch origin develop`
- **AND** SHALL update the local `develop` tracking ref

### Requirement: Sync uses a custom remote name

The system SHALL use the supplied remote name when `--remote <name>` is
set; otherwise it SHALL use `Config.Sync.DefaultRemote` and fall back to
`origin`.

#### Scenario: --remote uses the named remote

- **WHEN** user runs `twiggit sync --remote upstream`
- **THEN** the system SHALL run `git fetch upstream` rather than
  `git fetch origin`

#### Scenario: Default remote when no flag

- **WHEN** user runs `twiggit sync` and `Config.Sync.DefaultRemote` is
  unset
- **THEN** the system SHALL use `origin`

### Requirement: Fan-out sync across project worktrees

The system SHALL sync the project's default branch when `--all` is set,
hitting every project in the workspace if a project is named or every
project in the configured projects directory otherwise.

#### Scenario: --all hits every project

- **WHEN** user runs `twiggit sync --all`
- **THEN** the system SHALL iterate the configured projects directory
- **AND** SHALL fetch the default branch for each project
- **AND** SHALL print a per-project summary

### Requirement: Follow sync with rebase when --rebase is set

The system SHALL run the same rebase walk as `twiggit rebase --all`
when `--rebase` is supplied, after the fetch completes.

#### Scenario: --rebase fetches and rebases every worktree

- **WHEN** user runs `twiggit sync --rebase`
- **THEN** the system SHALL fetch the project's default branch
- **AND** SHALL rebase every worktree in the project onto the
  refreshed base
- **AND** SHALL stop on the first conflict
- **AND** SHALL exit with status 1 on conflict, status 0 on clean run

#### Scenario: --rebase omitted leaves worktrees untouched

- **WHEN** user runs `twiggit sync` without `--rebase`
- **THEN** the system SHALL fetch and update tracking refs
- **AND** SHALL NOT touch any worktree branch

### Requirement: --fetch-only skips the rebase step

The system SHALL accept `--fetch-only` as a no-op-equivalent safety
guard, ensuring no rebase runs even if `Config.Sync.RebaseAfterSync` is
true.

#### Scenario: --fetch-only overrides config

- **WHEN** user runs `twiggit sync --fetch-only` and the config sets
  `Sync.RebaseAfterSync = true`
- **THEN** the system SHALL fetch only
- **AND** SHALL NOT run any rebase

### Requirement: Usage error on sync outside any git context

The system SHALL return a `UsageError` (exit 2) when the user runs
`twiggit sync` from outside any git context without a project
positional argument and without an explicit `--all`.

#### Scenario: Outside git, no argument, no --all

- **WHEN** user runs `twiggit sync` from a directory outside any git
  context
- **THEN** the system SHALL exit with status 2
- **AND** SHALL print a usage error