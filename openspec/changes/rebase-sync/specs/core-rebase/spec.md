# Spec Delta

## Purpose

Defines the pure value objects used by the rebase and sync commands:
request shapes, result shapes, outcome enum, and the per-operation
sentinels. These types carry no I/O and live in `internal/core/` so
that the cmd layer can compose them with the role interface and the
adapter layer can construct them from `git` CLI output.

## ADDED Requirements

### Requirement: RebaseRequest carries all rebase inputs

`core.RebaseRequest` SHALL carry `ProjectName string`, `BranchName string`,
`WorktreePath string`, `OntoBranch string`, `IsForce bool`,
`IsFetch bool`, `IsAll bool`, `IsContinue bool`, `IsAbort bool`,
`IsSetBase bool`, and `NewBaseBranch string`. Empty optional fields
indicate "not set"; the cmd layer resolves them before dispatch.

#### Scenario: RebaseRequest round-trips fields

- **WHEN** a `RebaseRequest` is constructed with all flags set
- **THEN** every field SHALL be readable through field accessors without
  string-keyed lookups
- **AND** the type SHALL NOT carry method-set state (no methods beyond
  constructors)

### Requirement: RebaseOutcome enum classifies per-worktree rebase results

`core.RebaseOutcome` SHALL be an `int` enum with `RebaseOutcomeUnknown`
at `iota` position 0 and the variants `RebaseOutcomeClean`,
`RebaseOutcomeConflicted`, `RebaseOutcomeAborted`, and
`RebaseOutcomeNothingToDo`.

#### Scenario: RebaseOutcome zero value is the unknown sentinel

- **WHEN** `var r core.RebaseOutcome` is declared
- **THEN** `r == core.RebaseOutcomeUnknown` (iota 0)

#### Scenario: RebaseOutcomeString maps every variant

- **WHEN** `RebaseOutcome.String()` is called for each variant
- **THEN** it SHALL return a lowercase identifier per variant:
  `unknown`, `clean`, `conflicted`, `aborted`, `nothing-to-do`

### Requirement: RebasedWorktree carries per-worktree rebase state

`core.RebasedWorktree` SHALL carry `ProjectName string`, `BranchName
string`, `WorktreePath string`, `TrackedBase string`, `Outcome
RebaseOutcome`, `SkipReason string`, and `Error error`. `Error` carries
the wrapped adapter error when the worktree's rebase fails; cmd layer
reads the field directly without string-keyed access.

#### Scenario: RebasedWorktree zero value is safe

- **WHEN** `var w core.RebasedWorktree` is declared
- **THEN** every field SHALL be the type's zero value
- **AND** `Outcome == core.RebaseOutcomeUnknown`

### Requirement: RebaseResult aggregates a rebase walk

`core.RebaseResult` SHALL carry `RebasedWorktrees []*RebasedWorktree`,
`SkippedWorktrees []*RebasedWorktree`, `TotalRebased int`,
`TotalSkipped int`, `TotalConflicts int`, and `NavigationPath string`.
`NavigationPath` SHALL be non-empty only for single-target rebase
successes, mirroring `core.PruneWorktreesResult`.

#### Scenario: Empty rebase walk produces an empty result

- **WHEN** the rebase walk visits zero worktrees
- **THEN** `RebaseResult` SHALL be a non-nil zero-valued result
- **AND** the cmd layer SHALL print "nothing to do" or the equivalent
  user-facing message

### Requirement: SyncRequest carries sync inputs

`core.SyncRequest` SHALL carry `ProjectName string`, `BranchName string`,
`Remote string`, `IsAll bool`, `IsFetchOnly bool`, and `IsRebase bool`.
Empty `Remote` resolves to `Config.Sync.DefaultRemote` or `origin`.

#### Scenario: SyncRequest with --rebase sets the flag

- **WHEN** user runs `twiggit sync --rebase`
- **THEN** `SyncRequest.IsRebase == true`

### Requirement: SyncedBranch records before- and after-tip

`core.SyncedBranch` SHALL carry `ProjectName string`, `BranchName
string`, `RemoteName string`, `OldTip string`, and `NewTip string`.
`OldTip` and `NewTip` SHALL be commit sha strings; identical values
indicate the branch was already up to date.

#### Scenario: SyncedBranch records the tip delta

- **WHEN** a sync fetches and updates `main` from `abc123` to `def456`
- **THEN** `SyncedBranch{OldTip: "abc123", NewTip: "def456"}` SHALL be
  added to `SyncResult.SyncedBranches`

### Requirement: SyncResult aggregates a sync walk

`core.SyncResult` SHALL carry `SyncedBranches []*SyncedBranch`,
`RebasedBranches []*RebasedWorktree` (populated only when
`SyncRequest.IsRebase` is true), `TotalSynced int`, `TotalRebased int`,
`TotalConflicts int`, and `NavigationPath string`.

#### Scenario: --fetch-only leaves RebasedBranches empty

- **WHEN** user runs `twiggit sync --fetch-only`
- **THEN** `SyncResult.RebasedBranches` SHALL be empty
- **AND** `SyncResult.TotalRebased == 0`

### Requirement: Rebase sentinels participate in errors.Is

The system SHALL expose three package-level sentinels in
`internal/core/sentinels.go`: `ErrRebaseConflict`, `ErrRebaseInProgress`,
and `ErrBaseNotSet`. Adapter and cmd code SHALL walk these via
`errors.Is(err, core.ErrRebaseConflict)`, `errors.Is(err,
core.ErrRebaseInProgress)`, and `errors.Is(err, core.ErrBaseNotSet)`.
Sentinel strings SHALL be lowercase without trailing punctuation.

#### Scenario: Rebase conflict wraps ErrRebaseConflict

- **WHEN** a rebase hits a conflict
- **THEN** the returned error SHALL walk to `core.ErrRebaseConflict` via
  `errors.Is`
- **AND** the error message SHALL be lowercase without trailing
  punctuation

#### Scenario: ErrBaseNotSet surfaces when tracked base is missing and no fallback exists

- **WHEN** `twiggit rebase` runs and the worktree has no tracked base
  AND `Config.Validation.ProtectedBranches` is empty
- **THEN** the system SHALL return an error wrapping `core.ErrBaseNotSet`