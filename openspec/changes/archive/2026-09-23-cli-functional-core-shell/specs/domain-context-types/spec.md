# Spec Delta

## MODIFIED Requirements

### Requirement: `ContextType` enum

The `core` package SHALL define `ContextType` with constants:

| Constant | String | Meaning |
|---|---|---|
| `ContextUnknown` | `unknown` | Default; detection not run |
| `ContextProject` | `project` | Inside a project's main repository |
| `ContextWorktree` | `worktree` | Inside a worktree (subdirectory under `Worktrees/`) |
| `ContextOutsideGit` | `outside-git` | Not in any git repository |

`String()` SHALL return the lowercased constant name. The previous `domain.ContextType` is removed; the package qualifier is `core` for every call site.

#### Scenario: String() returns lowercased name
- **WHEN** the value is `core.ContextProject`
- **THEN** `String()` returns `"project"`

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `Context` struct

The `core` package SHALL define `core.Context` with the same field surface as the previous `domain.Context`. The previous `domain.Context` is removed; the package qualifier is `core`.

#### Scenario: Project context
- **WHEN** user is inside `~/Projects/myapp/`
- **THEN** `Context` SHALL have `Type = ContextProject`, `ProjectName = "myapp"`, `BranchName = ""`, `Path = ~/Projects/myapp`, and an explanation like "project 'myapp' detected via .git directory"

#### Scenario: Worktree context
- **WHEN** user is inside `~/Worktrees/myapp/feature/`
- **THEN** `Context` SHALL have `Type = ContextWorktree`, `ProjectName = "myapp"`, `BranchName = "feature"`, `Path = ~/Worktrees/myapp/feature`

#### Scenario: Outside git
- **WHEN** user is in a directory that is not a git repo
- **THEN** `Context` SHALL have `Type = ContextOutsideGit`
- **AND** all other fields SHALL be empty

### Requirement: `PathType` enum

The `core` package SHALL define `PathType` with constants `PathTypeUnknown`, `PathTypeProject`, `PathTypeWorktree`, `PathTypeInvalid`. The `iota` zero value SHALL be `PathTypeUnknown`. `String()` SHALL return `"unknown"`, `"project"`, `"worktree"`, `"invalid"` respectively. The previous `domain.PathType` is removed.

#### Scenario: Zero-value PathType is Unknown
- **WHEN** a `core.PathType` variable is declared without explicit initialization
- **THEN** `var p PathType` SHALL equal `PathTypeUnknown`
- **AND** `p.String()` SHALL return `"unknown"`

#### Scenario: Existing constants retain their strings
- **WHEN** `PathTypeProject`, `PathTypeWorktree`, `PathTypeInvalid` are rendered via `String()`
- **THEN** the result SHALL equal `"project"`, `"worktree"`, `"invalid"` respectively

#### Scenario: Numeric shift documented for callers
- **WHEN** downstream code compares `PathType` against an integer literal
- **THEN** the comparison SHALL use the named constant (`core.PathTypeProject`, etc.) rather than the underlying integer value, because the integer values shift with the addition of `PathTypeUnknown = 0`

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `ResolutionResult` struct

The `core` package SHALL provide a `ResolutionResult` struct with the same field surface as the previous `domain.ResolutionResult`. The previous `domain.ResolutionResult` is removed.

#### Scenario: ResolutionResult fields carry resolver output
- **WHEN** the context resolver returns a populated `*core.ResolutionResult`
- **THEN** every field SHALL carry the resolver's value

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `ResolutionSuggestion` struct (live fields)

The `core` package SHALL provide `ResolutionSuggestion` with the same field surface as the previous `domain.ResolutionSuggestion`. The `Remote` and `StyleHint` fields are kept on the struct to avoid breaking future consumers but SHALL be treated as zero-valued until a caller populates them. The previous `domain.ResolutionSuggestion` is removed.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `SuggestionOption`

`core.SuggestionOption` SHALL be a functional option applied to suggestion generation. The built-in `core.WithExistingOnly()` SHALL filter suggestions to materialized worktrees (no remote-only branches, no stale entries). The previous `domain.SuggestionOption` and `domain.WithExistingOnly()` are removed.

#### Scenario: WithExistingOnly filters remote-only branches
- **WHEN** suggestion generation runs with `core.WithExistingOnly()` applied and the candidates include both local-worktree branches and remote-only branches
- **THEN** the returned suggestions SHALL include only the local-worktree branches

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
