# Spec Delta

## Purpose

Lets users rebase worktree branches onto their tracked base branch via a
single CLI verb, with offline-by-default operation and stop-on-first-conflict
semantics for fan-out. Surfaces a per-worktree tracked base persisted in git
per-worktree config so rebase targets survive worktree moves and are
discoverable with stock git tooling.

## ADDED Requirements

### Requirement: Rebase current worktree onto its tracked base

The system SHALL rebase the current worktree's branch onto its tracked base
branch when the user runs `twiggit rebase` from inside a worktree and the
tracked base is set.

#### Scenario: Rebase from inside a worktree succeeds

- **WHEN** user runs `twiggit rebase` from a worktree whose tracked base
  resolves to `main`
- **THEN** the system SHALL run `git rebase main` in the worktree
- **AND** SHALL exit with status 0 on a conflict-free rebase
- **AND** SHALL print a success summary to stderr

#### Scenario: Rebase from inside a worktree hits a conflict

- **WHEN** user runs `twiggit rebase` and the worktree branch has conflicts
  with the tracked base
- **THEN** the system SHALL leave the worktree in mid-rebase state
- **AND** SHALL exit with status 1
- **AND** SHALL print a conflict message naming the worktree and the
  conflicting files

### Requirement: Rebase a specific worktree by project/branch

The system SHALL rebase the worktree identified by `[project/branch]` when
the user supplies that argument.

#### Scenario: Named worktree rebase

- **WHEN** user runs `twiggit rebase myproject/feature`
- **THEN** the system SHALL resolve `myproject/feature` to a worktree
- **AND** SHALL rebase that worktree onto its tracked base
- **AND** SHALL NOT touch other worktrees

### Requirement: Project fan-out rebase stops on first conflict

The system SHALL rebase every worktree in the current project when
`--all` is set, halting on the first conflict and reporting remaining
worktrees as skipped with reason "prior conflict".

#### Scenario: Fan-out hits a conflict mid-iteration

- **WHEN** user runs `twiggit rebase --all` and the third worktree
  produces a conflict
- **THEN** the system SHALL rebase the first two worktrees
- **AND** SHALL leave the third in mid-rebase state
- **AND** SHALL report the fourth and subsequent worktrees as skipped with
  reason "prior conflict"
- **AND** SHALL exit with status 1

#### Scenario: Fan-out finishes clean

- **WHEN** user runs `twiggit rebase --all` and no worktree produces a
  conflict
- **THEN** the system SHALL rebase every worktree
- **AND** SHALL exit with status 0
- **AND** SHALL print a per-worktree summary to stderr

### Requirement: Fetch the tracked base before rebasing

The system SHALL run `git fetch <remote> <tracked-base>` in the project
main repository before rebasing when `--fetch` is set. The system SHALL NOT
fetch a remote by default; `twiggit rebase` is offline unless `--fetch`
appears.

#### Scenario: --fetch triggers a fetch

- **WHEN** user runs `twiggit rebase --fetch` and the tracked base is
  `main` and the remote is `origin`
- **THEN** the system SHALL run `git fetch origin main` in the project
  main repo
- **AND** SHALL rebase the worktree on the freshly fetched base

#### Scenario: Default rebase does not touch the network

- **WHEN** user runs `twiggit rebase` without `--fetch`
- **THEN** the system SHALL NOT invoke `git fetch`
- **AND** SHALL rebase against the locally tracked commit of the base

### Requirement: Resume a paused rebase

The system SHALL resume an in-progress rebase when `--continue` is
supplied.

#### Scenario: --continue after conflict resolution

- **WHEN** user runs `twiggit rebase --continue myproject/feature` after
  fixing a conflict and committing
- **THEN** the system SHALL run `git rebase --continue` in the worktree
- **AND** SHALL exit with status 0 on clean resume

#### Scenario: --continue without a paused rebase

- **WHEN** user runs `twiggit rebase --continue myproject/feature` and
  no rebase is in progress
- **THEN** the system SHALL exit with status 1
- **AND** SHALL print an error message identifying the absent rebase

### Requirement: Abort a paused rebase

The system SHALL abort an in-progress rebase when `--abort` is supplied,
restoring the worktree to its pre-rebase state.

#### Scenario: --abort restores the worktree

- **WHEN** user runs `twiggit rebase --abort myproject/feature` with a
  paused rebase in the worktree
- **THEN** the system SHALL run `git rebase --abort`
- **AND** SHALL exit with status 0 on success
- **AND** SHALL leave no `.git/rebase-merge/` or `.git/rebase-apply/`
  state

### Requirement: Set the tracked base without rebasing

The system SHALL update a worktree's tracked base branch in git
per-worktree config when `--set-base <branch>` is supplied, without
performing a rebase.

#### Scenario: --set-base persists to per-worktree config

- **WHEN** user runs `twiggit rebase --set-base develop` from a worktree
- **THEN** the system SHALL run `git -C <worktree> config --worktree
  twiggit.tracked-base develop`
- **AND** SHALL NOT run `git rebase`
- **AND** SHALL exit with status 0

#### Scenario: Subsequent rebase uses the new base

- **WHEN** user runs `twiggit rebase --set-base develop` and then
  `twiggit rebase`
- **THEN** the second command SHALL rebase onto `develop`

### Requirement: Refuse to rebase a dirty worktree

The system SHALL refuse to rebase a worktree with uncommitted changes
unless an explicit force flag is set.

#### Scenario: Dirty worktree refused

- **WHEN** user runs `twiggit rebase` in a worktree with uncommitted
  changes
- **THEN** the system SHALL exit with status 1
- **AND** SHALL print an actionable message suggesting the user stash,
  commit, or pop before rebasing

### Requirement: Fall back to a protected branch when tracked base is missing

The system SHALL fall back to the first entry of
`Config.Validation.ProtectedBranches` when a worktree's tracked base is
not set in per-worktree config.

#### Scenario: Missing tracked base falls back to protected branches

- **WHEN** user runs `twiggit rebase` in a worktree that has no
  `twiggit.tracked-base` entry
- **THEN** the system SHALL resolve the rebase target to
  `Config.Validation.ProtectedBranches[0]`
- **AND** SHALL print a notice indicating the fallback

### Requirement: Usage error outside any git context without explicit project

The system SHALL return a `UsageError` (exit 2) when the user runs
`twiggit rebase` from outside any git context and no `[project/branch]`
positional argument is supplied.

#### Scenario: Outside git, no argument

- **WHEN** user runs `twiggit rebase` from a directory that is not inside
  any git context
- **THEN** the system SHALL exit with status 2
- **AND** SHALL print a usage error suggesting an explicit
  `[project/branch]` argument

### Requirement: Single-target rebase prints navigation path for the shell wrapper

The system SHALL print the worktree's absolute path to stdout on a
successful single-target rebase so the shell wrapper can `cd` into it.

#### Scenario: Single-target rebase prints path

- **WHEN** user runs `twiggit rebase myproject/feature` and the rebase
  finishes clean
- **THEN** the system SHALL print the worktree's absolute path to stdout
  as the last line
- **AND** SHALL NOT print any other path-like output

### Requirement: --all and a positional argument are mutually exclusive

The system SHALL return a `UsageError` (exit 2) when the user supplies
both `--all` and a `[project/branch]` positional argument; the two
forms describe disjoint invocations and combining them is ambiguous.

#### Scenario: --all with positional returns UsageError

- **WHEN** user runs `twiggit rebase --all myproject/feature`
- **THEN** the system SHALL exit with status 2
- **AND** SHALL print a usage error naming the conflict between
  `--all` and the positional argument