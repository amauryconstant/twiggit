# Spec Delta: core-git

## MODIFIED Requirements

### Requirement: RepositoryReader exposes GetRepositoryStatus, GetRepositoryInfo, and GetCommitInfo

The system SHALL expose a `RepositoryReader` role interface with exactly
three context-aware methods: `RepositoryStatus(ctx context.Context, repoPath string) (core.RepositoryStatus, error)`, `Repository(ctx context.Context, repoPath string) (*core.Repository, error)`, and `Commit(ctx context.Context, repoPath, commitHash string) (*core.Commit, error)`. The role SHALL be defined consumer-side. Method names SHALL NOT carry a `Get` prefix; return types SHALL be the renamed data types from `core-types` (`Repository` not `GitRepository`, `Commit` not `CommitInfo`).

#### Scenario: RepositoryReader groups repository inspection methods

- **WHEN** a caller requires `RepositoryStatus`, `Repository`, and `Commit`
- **THEN** the role SHALL be `RepositoryReader` and SHALL NOT include
  branch listing, remote listing, or repository opening
- **AND** none of the three methods SHALL begin with `Get`

#### Scenario: Composite *git.Client satisfies RepositoryReader

- **WHEN** the composite `*git.Client` is type-asserted against
  `RepositoryReader`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.RepositoryReader = (*git.Client)(nil)`

## ADDED Requirements

### Requirement: Role method names follow the noun-only convention

Every method on the six role interfaces SHALL be named after its return
type, minus the `Get` prefix. Read methods (`RepositoryStatus`,
`Repository`, `Commit`, `ListBranches`, `BranchExists`, `ListRemotes`,
`ValidateRepository`) SHALL be callable without a verb prefix. Mutating
methods (`CreateWorktree`, `DeleteWorktree`, `ListWorktrees`,
`PruneWorktrees`, `DeleteBranch`, `IsBranchMerged`) MAY keep imperative
verbs because their return is `error`, not a noun.

#### Scenario: No Get-prefixed methods on role interfaces

- **WHEN** the role interface declarations in the core package are
  read for method names
- **THEN** no method SHALL begin with `Get`
- **AND** the declaration surface SHALL NOT carry a `//nolint`
  directive suppressing the rule

## REMOVED Requirements

### Requirement: Factory per-role fields are additive to composite

**Reason**: Per-role fields were speculative — they had zero external
callers (only `factory_test.go` and `Factory.Init()` touched them) and
added 30 lines of byte-identical wrapper code. `*git.Client` returned
by `Factory.GitClient()` already satisfies every role through embedded
promotion, so callers needing role-narrowed access can assign the
client to a local role-typed variable with one line.

**Migration**: Replace `f.BranchReader()` calls with
`client, err := f.GitClient(); if err != nil { return err }; ro := core.BranchReader(client)`.
Errors propagate per the single-handling rule (no `_` discard). Update
`Factory.Init()` to drop the six per-role touch calls.

### Requirement: Factory lazy-field errors are propagated, not swallowed

**Reason**: Folded into the broader "errors pass through the Factory
unchanged" pattern; the previous requirement covered only per-role
errors and the per-role fields no longer exist.

**Migration**: Behavior preserved — `Factory.GitClient()` continues to
return errors unwrapped from `git.NewClient()`. See `git-client` spec
for the error contract.
