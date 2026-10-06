# Spec Delta

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

### Requirement: Composite *git.Client satisfies all eight roles via embedded promotion

The composite `git.Client` SHALL satisfy `Rebaser` and `BaseTracker`
in addition to the existing six role interfaces through embedded
promotion of `*reader` and `*cliClient`. Each satisfaction SHALL be
asserted at compile time via a `var _ core.Role = (*git.Client)(nil)`
declaration placed in `internal/git/client.go`. The eight declarations
SHALL be co-located at the bottom of that file.

#### Scenario: Compile-time checks cover the two new roles

- **WHEN** `internal/git/client.go` is read
- **THEN** it SHALL contain `var _ core.Rebaser = (*git.Client)(nil)`
  and `var _ core.BaseTracker = (*git.Client)(nil)`
- **AND** the existing six `var _` declarations SHALL remain

### Requirement: Drift sentinel embeds the new roles

The drift sentinel in `internal/git/client_test.go` SHALL embed
`core.Rebaser` and `core.BaseTracker` alongside the existing six
roles. Renaming or removing any `Rebaser` or `BaseTracker` method on
the composite SHALL cause `go build ./...` to fail with a
type-mismatch error referencing the affected role.

#### Scenario: Drift sentinel embeds the new roles

- **WHEN** `go build ./internal/git/...` runs
- **THEN** the `_DriftCheck` sentinel SHALL include fields typed as
  `core.Rebaser` and `core.BaseTracker`

#### Scenario: Rebaser or BaseTracker drift fails the build

- **WHEN** any `Rebaser` or `BaseTracker` method on `*cliClient` is
  renamed or removed
- **THEN** `go build ./...` SHALL fail with a type-mismatch error
  naming the affected role