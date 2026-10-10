# Spec Delta

## Purpose

Owns the `core.WorktreeStatus` value type that every worktree diagnostic view reads, and the read path that produces it. Defines the `WorktreeStatusReader` role interface that `cmd/status`, `cmd/delete`, and any future consumer compose against, plus the per-worktree best-effort error contract that lets one failed worktree never abort the rest of the walk. Cross-references `core-types` (the existing data type) and `core-git` (the role-interface pattern); does not restate either.

## ADDED Requirements

### Requirement: `core.WorktreeStatus` carries the full diagnostic projection

`core.WorktreeStatus` SHALL expose the fields `Worktree *core.Worktree`, `RepositoryStatus *core.RepositoryStatus`, `LastChecked time.Time`, `IsClean bool`, `HasUncommittedChanges bool`, `BranchStatus string`, `Base string`, `IsMerged bool`, `IsStale bool`, `LastCommitDate time.Time`, `IsSkipped bool`, and `SkipReason string`.

#### Scenario: Field set is complete

- **WHEN** a caller constructs a `core.WorktreeStatus` with one value per named field
- **THEN** the type compiles and the values round-trip through field access

#### Scenario: `RepositoryStatus.Ahead` and `Behind` reflect the local counts

- **WHEN** the adapter populates `RepositoryStatus.Ahead` and `RepositoryStatus.Behind` from `git rev-list --count` against the worktree's tracked base
- **THEN** the populated `core.WorktreeStatus.RepositoryStatus` carries those counts
- **AND** `BranchStatus` is `"up-to-date"` when both are zero, `"ahead"` when only `Ahead` is positive, `"behind"` when only `Behind` is positive, `"diverged"` when both are positive

### Requirement: Every exported field carries a JSON struct tag

Every exported field on `core.WorktreeStatus` SHALL carry a JSON struct tag so consumers can pipe the value through `encoding/json` directly.

#### Scenario: JSON round-trip on a populated value

- **WHEN** a caller marshals a populated `core.WorktreeStatus` to JSON and decodes the result back into the same type
- **THEN** every populated field survives the round-trip

### Requirement: Zero value of `core.WorktreeStatus` is safe to use

The zero value of `core.WorktreeStatus` SHALL be safe: no nil-deref panics on field access, no panics on JSON marshal, no panics on a JSON unmarshal into a fresh value.

#### Scenario: Zero value marshals to `{}` without panic

- **WHEN** a caller marshals a zero-valued `core.WorktreeStatus` to JSON
- **THEN** the output is a valid JSON object with zero-valued fields
- **AND** the marshal call returns no error

### Requirement: `WorktreeStatusReader` is a one-method role

`core.WorktreeStatusReader` SHALL be a single-method interface: `ReadWorktreeStatus(ctx context.Context, repoPath, wtPath, base string) (core.WorktreeStatus, error)`. The interface SHALL be declared in `internal/core` per `core-git` consumer-side segregation rules.

#### Scenario: Interface declares the read verb only

- **WHEN** a reader inspects `internal/core/git.go` for the `WorktreeStatusReader` interface
- **THEN** the interface declares exactly one method with the signature above

### Requirement: Compile-time drift sentinel enforces `WorktreeStatusReader` satisfaction

The composite `*git.Client` SHALL satisfy `core.WorktreeStatusReader`. The satisfaction SHALL be enforced by a compile-time drift sentinel of the form `var _ core.WorktreeStatusReader = (*Client)(nil)` placed next to the existing drift sentinels in `internal/git/client.go`.

#### Scenario: Adding a method signature drift breaks the build

- **WHEN** the `ReadWorktreeStatus` method on `*Client` changes its signature
- **THEN** the drift sentinel fails to compile
- **AND** the change cannot land without updating the interface

### Requirement: `Base` resolution follows the rebase tracked-base chain

The `Base` field SHALL be resolved using the priority chain already owned by `cli-rebase`: explicit per-worktree `twiggit.tracked-base` config first, then the first entry of `Config.Validation.ProtectedBranches`. When neither source yields a value, the adapter SHALL still return a populated `core.WorktreeStatus` with `Base = ""` and SHALL record a per-worktree `SkipReason` naming the absent base, so the walk continues across other worktrees.

#### Scenario: Worktree with a tracked base

- **WHEN** the worktree has `twiggit.tracked-base` set to `develop`
- **THEN** the returned `core.WorktreeStatus.Base` is `develop`

#### Scenario: Worktree without a tracked base, fallback present

- **WHEN** the worktree has no `twiggit.tracked-base`
- **AND** `Config.Validation.ProtectedBranches` is `["main", "master"]`
- **THEN** the returned `core.WorktreeStatus.Base` is `main`

#### Scenario: Worktree without a tracked base, no fallback

- **WHEN** the worktree has no `twiggit.tracked-base`
- **AND** `Config.Validation.ProtectedBranches` is empty
- **THEN** the returned `core.WorktreeStatus` carries `Base = ""` and `IsSkipped = true`
- **AND** `SkipReason` names the absent base
- **AND** no error is returned (the walk continues)

### Requirement: `IsMerged` reflects the branch's merge state into the resolved base

The `IsMerged` field SHALL be `true` when the branch is merged into the resolved `Base` per the same `IsBranchMerged` call that `cli-prune` uses, and `false` otherwise. When the `IsBranchMerged` call fails, the adapter SHALL set `IsSkipped = true` and `SkipReason` to the failure reason; the walk SHALL continue.

#### Scenario: Branch merged into base

- **WHEN** the worktree's branch is merged into the resolved `Base`
- **THEN** the returned `core.WorktreeStatus.IsMerged` is `true`

#### Scenario: Branch not merged

- **WHEN** the worktree's branch is not merged into the resolved `Base`
- **THEN** the returned `core.WorktreeStatus.IsMerged` is `false`

### Requirement: `LastCommitDate` reflects the worktree branch tip

The `LastCommitDate` field SHALL be set to the author date of the commit that the worktree's branch currently points at, taken from `core.Branch.Date` via `core.BranchReader.ListBranches`. When the lookup fails or yields no matching branch, the field SHALL be the zero `time.Time` and the row SHALL still be reported (no `IsSkipped` unless the upstream base resolution also failed).

#### Scenario: Worktree branch present in `ListBranches`

- **WHEN** `ListBranches` returns a `core.Branch` whose `Name` matches the worktree's branch
- **THEN** `LastCommitDate` equals that `core.Branch.Date`

#### Scenario: Worktree branch not in `ListBranches`

- **WHEN** `ListBranches` returns no matching `core.Branch`
- **THEN** `LastCommitDate` is the zero `time.Time`
- **AND** `IsSkipped` remains `false` when the base resolution succeeded

### Requirement: `IsStale` derives from `Behind` and `LastCommitDate` against configured thresholds

The `IsStale` field SHALL be `true` when EITHER `Behind >= Config.Status.StaleBehind` (when `StaleBehind > 0`) OR the duration `time.Since(LastCommitDate)` exceeds `Config.Status.StaleDays * 24h` (when `StaleDays > 0` and `LastCommitDate` is not the zero time). When both thresholds are zero, `IsStale` SHALL be `false`. The derivation SHALL be a pure function on the populated value and the configuration; tests SHALL table-drive every combination.

#### Scenario: Behind at or above the threshold

- **WHEN** `Behind = 25` and `Config.Status.StaleBehind = 20`
- **THEN** `IsStale` is `true`

#### Scenario: Age above the threshold

- **WHEN** `LastCommitDate` is 40 days ago and `Config.Status.StaleDays = 30`
- **THEN** `IsStale` is `true`

#### Scenario: Both thresholds zero

- **WHEN** `Config.Status.StaleBehind = 0` and `Config.Status.StaleDays = 0`
- **THEN** `IsStale` is `false` for every populated value

#### Scenario: Behind below threshold, age below threshold

- **WHEN** `Behind = 5`, `LastCommitDate` is 3 days ago, `StaleBehind = 20`, `StaleDays = 30`
- **THEN** `IsStale` is `false`

### Requirement: Per-worktree read failures populate the skip fields, not an error

The adapter SHALL return a `core.WorktreeStatus` for every worktree it visits. A failure on any of the per-worktree reads (ahead/behind, merged, base resolution, branch lookup) SHALL NOT abort the walk; the adapter SHALL set `IsSkipped = true` and `SkipReason` to a lowercase, no-trailing-punctuation reason, leave the unpopulated fields at their zero values, and continue. The walk SHALL return a non-nil error only when a precondition fails (config load, context resolution, project discovery).

#### Scenario: Ahead/behind read fails

- **WHEN** the per-worktree ahead/behind call exits non-zero
- **THEN** the returned `core.WorktreeStatus` carries `IsSkipped = true`
- **AND** `SkipReason` is a lowercase message naming the failure
- **AND** `Ahead` and `Behind` are zero
- **AND** the other populated fields (`IsMerged`, `IsDirty`, `Base`, `LastCommitDate`) keep their values when their reads succeeded independently

#### Scenario: Walk completes after several per-worktree failures

- **WHEN** the walk visits N worktrees and M of them fail individual reads
- **THEN** the returned slice has N entries
- **AND** the M failures carry `IsSkipped = true` and a `SkipReason`
- **AND** the slice is the only return value besides a nil error

### Requirement: `Config.Status` carries the stale heuristic thresholds

`core.Config` SHALL expose a `Status StatusConfig` field. `StatusConfig` SHALL carry `StaleBehind int` and `StaleDays int`. The default values SHALL be `StaleBehind = 20` and `StaleDays = 30` per `git-config`. The field SHALL load via the existing koanf path (`git-config`); the TOML key prefix is `[status]`.

#### Scenario: Default config

- **WHEN** `Config` is built from `core.DefaultConfig()`
- **THEN** `Config.Status.StaleBehind` is `20`
- **AND** `Config.Status.StaleDays` is `30`

#### Scenario: TOML override

- **WHEN** the user sets `[status] stale_behind = 50` in the config file
- **THEN** `Config.Status.StaleBehind` is `50` after `Config.Validate()` and the koanf load
