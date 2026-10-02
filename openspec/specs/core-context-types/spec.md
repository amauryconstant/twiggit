# Capability: Context Types

## Purpose

Defines the `core.ContextType`, `core.PathType`, `core.Context`, `core.ResolutionResult`, and `core.ResolutionSuggestion` value objects that drive context detection and identifier resolution. Pure: no filesystem I/O lives here. Both `ContextType` and `PathType` declare explicit `Unknown` sentinels at `iota` position 0.

## Requirements

### Requirement: ContextType enum

`core.ContextType` SHALL be an `int` type with the following variants, in order:

| Position | Constant | Meaning |
|---|---|---|
| 0 | `ContextUnknown` | Uninitialized; never a valid detected state |
| 1 | `ContextProject` | Inside a project's main repository |
| 2 | `ContextWorktree` | Inside a worktree under `Config.WorktreesDirectory` |
| 3 | `ContextOutsideGit` | Not in any git repository |

`ContextType.String()` SHALL return `"unknown"` for `ContextUnknown`, `"project"` for `ContextProject`, `"worktree"` for `ContextWorktree`, and `"outside-git"` for `ContextOutsideGit`.

#### Scenario: Unknown sentinel at iota 0

- **WHEN** the package user reads `ContextType` constants
- **THEN** `ContextUnknown` SHALL be at `iota` position 0
- **AND** `ContextType(0).String()` SHALL return `"unknown"`

### Requirement: PathType enum

`core.PathType` SHALL be an `int` type with the following variants:

| Position | Constant | Meaning |
|---|---|---|
| 0 | `PathTypeUnknown` | Uninitialized |
| 1 | `PathTypeProject` | Path resolves to a project |
| 2 | `PathTypeWorktree` | Path resolves to a worktree |
| 3 | `PathTypeInvalid` | Path is not under any configured directory |

`PathType.String()` SHALL return `"unknown"`, `"project"`, `"worktree"`, or `"invalid"` per the variant.

#### Scenario: PathType zero-value prints unknown

- **WHEN** `var p core.PathType` is declared
- **THEN** `p.String()` SHALL return `"unknown"`

### Requirement: Context struct carries detection result

`core.Context` SHALL carry `Type ContextType`, `ProjectName string`, `BranchName string` (only meaningful for `ContextWorktree`), `Path string` (absolute path to the context root), and `Explanation string` (human-readable description of why this context was detected). No mutation methods SHALL exist on `Context`.

#### Scenario: ContextWorktree carries branch name

- **WHEN** `git.NewContextDetector(cfg).DetectContext("$HOME/Worktrees/myapp/feat/foo")` returns a context
- **THEN** `ctx.Type == core.ContextWorktree`
- **AND** `ctx.ProjectName == "myapp"`
- **AND** `ctx.BranchName == "feat/foo"`

### Requirement: ResolutionResult and ResolutionSuggestion

`core.ResolutionResult` SHALL carry `ResolvedPath string`, `Type PathType`, `ProjectName string`, `BranchName string`, and `Explanation string`. `core.ResolutionSuggestion` SHALL carry `Text string`, `Description string`, `Type PathType`, `ProjectName string`, `BranchName string`, `IsCurrent bool`, and `IsDirty bool`. The `IsCurrent` and `IsDirty` flags support completion-candidate sorting and visual indication per `cli-completion`.

#### Scenario: ResolutionSuggestion marks the current branch

- **WHEN** `GetResolutionSuggestions(...)` returns three candidates
- **AND** one of them matches the user's current branch
- **THEN** the matching suggestion SHALL have `IsCurrent == true`
- **AND** the other two SHALL have `IsCurrent == false`

### Requirement: SuggestionOption is a functional-option type

`core.SuggestionOption` SHALL be declared as `func(any)` (or a more specific type when the consumer-side interface is widened). The function type lives where it is consumed; the current placeholder SHALL be replaced by a concrete type when the suggestion-consumer interface is widened.

#### Scenario: SuggestionOption applies at construction

- **WHEN** `NewResolutionSuggestion(...)` is invoked with no `SuggestionOption`
- **THEN** the returned suggestion SHALL have zero-value defaults for all fields
- **AND** when `WithCurrent(true)` is passed, `IsCurrent == true` on the returned suggestion