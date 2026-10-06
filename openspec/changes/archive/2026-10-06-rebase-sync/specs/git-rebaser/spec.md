# Spec Delta

## Purpose

Defines the two consumer-side role interfaces that the rebase and sync
commands consume: `Rebaser` for the rebase lifecycle and the fetch
primitive, and `BaseTracker` for reading and writing the per-worktree
tracked-base git config. Both interfaces live in `internal/core/git.go`
alongside the existing six role interfaces, are satisfied by the
composite `*git.Client` via embedded promotion, and contribute to the
single drift sentinel that enforces method-set invariants.

## ADDED Requirements

### Requirement: Rebaser exposes rebase lifecycle and fetch

The system SHALL expose a `Rebaser` role interface with exactly four
context-aware methods: `Rebase(ctx context.Context, wtPath, onto
string) (core.RebaseOutcome, error)`, `Abort(ctx context.Context,
wtPath string) error`, `Continue(ctx context.Context, wtPath string)
(core.RebaseOutcome, error)`, and `Fetch(ctx context.Context, repoPath,
remote, ref string) error`. The role SHALL be defined consumer-side and
SHALL NOT include branch listing, remote listing, or worktree lifecycle
methods.

#### Scenario: Rebaser declares four methods

- **WHEN** the role is read from `internal/core/git.go`
- **THEN** the role SHALL declare exactly `Rebase`, `Abort`, `Continue`,
  and `Fetch`
- **AND** the role SHALL NOT declare `CreateWorktree`, `ListBranches`,
  or any other method

#### Scenario: Composite *git.Client satisfies Rebaser

- **WHEN** the composite `*git.Client` is type-asserted against
  `Rebaser`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.Rebaser = (*git.Client)(nil)` in `internal/git/client.go`

### Requirement: Continue reports the post-continue outcome

`Rebaser.Continue` SHALL return `(core.RebaseOutcome, error)` so the
caller can branch on the post-continue state without a separate progress
probe.

#### Scenario: Clean continue returns Clean

- **WHEN** `git rebase --continue` completes without further conflicts
- **THEN** the returned `RebaseOutcome` SHALL be `RebaseOutcomeClean`
- **AND** the returned `error` SHALL be nil

#### Scenario: Still-conflict continue returns Conflicted

- **WHEN** `git rebase --continue` returns a non-zero exit and the
  worktree remains in mid-rebase state
- **THEN** the returned `RebaseOutcome` SHALL be `RebaseOutcomeConflicted`
- **AND** the returned `error` SHALL wrap `core.ErrRebaseConflict`

#### Scenario: Continue with no rebase in progress returns Aborted and ErrRebaseInProgress

- **WHEN** the caller invokes `Continue` against a worktree with no
  rebase in progress
- **THEN** the returned `RebaseOutcome` SHALL be `RebaseOutcomeAborted`
- **AND** the returned `error` SHALL wrap `core.ErrRebaseInProgress`

### Requirement: BaseTracker exposes tracked-base read and write

The system SHALL expose a `BaseTracker` role interface with exactly two
context-aware methods: `SetTrackedBase(ctx context.Context, wtPath,
base string) error` and `GetTrackedBase(ctx context.Context, wtPath
string) (string, error)`. The role SHALL be defined consumer-side.
`GetTrackedBase` SHALL return `(empty, nil)` (not an error) when the
worktree has no `twiggit.tracked-base` entry; callers decide whether
to treat empty as missing or as a fallback signal.

#### Scenario: BaseTracker is the only role that reads worktree config

- **WHEN** a caller requires reading or writing the tracked-base config
- **THEN** the role SHALL be `BaseTracker`
- **AND** no other role SHALL expose either method

#### Scenario: Composite *git.Client satisfies BaseTracker

- **WHEN** the composite `*git.Client` is type-asserted against
  `BaseTracker`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.BaseTracker = (*git.Client)(nil)` in `internal/git/client.go`

### Requirement: Rebaser and BaseTracker routing table pins CLI implementation

The composite `git.Client` SHALL route every `Rebaser` and
`BaseTracker` method to the CLI adapter (`*cliClient`). The routing
table in `git-client` SHALL pin this allocation; `go-git` SHALL NOT be
used for rebase, abort, continue, fetch, or per-worktree config
read/write. The rationale is the same as `WorktreeWriter` and
`BranchWriter`: go-git lacks support for these operations.

#### Scenario: Rebaser and BaseTracker routes are CLI-backed

- **WHEN** the routing table in `git-client` is read
- **THEN** the rows for `Rebase`, `Abort`, `Continue`, `Fetch`,
  `SetTrackedBase`, and `GetTrackedBase` SHALL each list the CLI client
  as `Concrete`

### Requirement: Drift sentinel extended to cover the new roles

The drift sentinel in `internal/git/client_test.go` SHALL embed
`core.Rebaser` and `core.BaseTracker` alongside the existing six roles.
Renaming or removing any `Rebaser` or `BaseTracker` method on the
composite SHALL cause `go build ./...` to fail with a type-mismatch
error referencing the affected role.

#### Scenario: Drift sentinel embeds the new roles

- **WHEN** `go build ./internal/git/...` runs
- **THEN** the `_DriftCheck` sentinel SHALL include fields typed as
  `core.Rebaser` and `core.BaseTracker`

#### Scenario: Rebaser or BaseTracker drift fails the build

- **WHEN** any `Rebaser` or `BaseTracker` method on `*cliClient` is
  renamed or removed
- **THEN** `go build ./...` SHALL fail with a type-mismatch error
  naming the affected role

### Requirement: Rebaser and BaseTracker respect the Tier 2 depguard

Both new role interfaces SHALL be importable from
`twiggit/internal/core` without importing any package other than the
Go standard library. `Rebaser` SHALL return `core.RebaseOutcome` (a
core-defined enum) so the role signature stays depguard-clean.

#### Scenario: New roles import only stdlib and core types

- **WHEN** the import graph of `internal/core/git.go` is computed
- **THEN** every import SHALL be a `$gostd` package
- **AND** the `Rebaser` and `BaseTracker` declarations SHALL reference
  no type from `internal/git/`