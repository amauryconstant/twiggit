# Spec Delta

## ADDED Requirements

### Requirement: Rebase sentinels live in internal/core/sentinels.go

The system SHALL expose three package-level sentinels for rebase and
base-tracking error conditions: `core.ErrRebaseConflict`,
`core.ErrRebaseInProgress`, and `core.ErrBaseNotSet`. Each sentinel
SHALL be declared in `internal/core/sentinels.go` (the single
canonical home for `core.Err*` sentinels per the
"Error types follow naming and sentinel-zero-value rules"
requirement). Each sentinel's message SHALL be lowercase without
trailing punctuation.

#### Scenario: Sentinels live in sentinels.go

- **WHEN** the source tree for `internal/core/` is searched for
  sentinel declarations
- **THEN** `ErrRebaseConflict`, `ErrRebaseInProgress`, and
  `ErrBaseNotSet` SHALL be declared in `sentinels.go`
- **AND** SHALL NOT be redeclared in any other file

#### Scenario: Adapter wraps via %w so errors.Is walks the chain

- **WHEN** the rebase adapter constructs an error for a conflict, an
  absent rebase, or a missing tracked base
- **THEN** the returned error SHALL wrap the appropriate sentinel via
  `fmt.Errorf("{context}: %w", core.Err...)`
- **AND** `errors.Is(err, core.ErrRebaseConflict)` (or the relevant
  matching sentinel) SHALL return true

### Requirement: Per-entity dispatch extended for rebase and base-tracking

The `core.OperationError.Is` and `core.NotFoundError.Is` dispatch
SHALL walk `ErrRebaseConflict` and `ErrBaseNotSet` when `Op` carries
the corresponding prefix (`rebase.*` and `base.*`) so callers can
use `errors.Is(err, Sentinel)` regardless of which constructor
produced the error.

#### Scenario: OperationError walking matches ErrRebaseConflict

- **WHEN** an adapter returns a `*core.OperationError{Op: "rebase.worktree"}`
  whose `Cause` walks to `ErrRebaseConflict`
- **THEN** `errors.Is(err, core.ErrRebaseConflict)` SHALL return true