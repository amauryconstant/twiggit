# Capability: Context Types

## Purpose

Domain types describing the CWD context for context-aware commands.
Detection rules and resolution live in `infrastructure-context-resolver`;
this spec owns the type surface.

## Requirements

### Requirement: `ContextType` enum

The system SHALL define `ContextType` with constants:

| Constant | String | Meaning |
|---|---|---|
| `ContextUnknown` | `unknown` | Default; detection not run |
| `ContextProject` | `project` | Inside a project's main repository |
| `ContextWorktree` | `worktree` | Inside a worktree (subdirectory under `Worktrees/`) |
| `ContextOutsideGit` | `outside-git` | Not in any git repository |

`String()` SHALL return the lowercased name.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `Context` struct

The system SHALL define:

```go
type Context struct {
    Type        ContextType
    ProjectName string
    BranchName  string  // only for ContextWorktree
    Path        string  // absolute path to context root
    Explanation string  // human-readable explanation of detection
}
```

#### Scenario: Project context

- **WHEN** user is inside `~/Projects/myapp/`
- **THEN** `Context` SHALL have `Type = ContextProject`,
  `ProjectName = "myapp"`, `BranchName = ""`,
  `Path = ~/Projects/myapp`, and an explanation like
  "project 'myapp' detected via .git directory"

#### Scenario: Worktree context

- **WHEN** user is inside `~/Worktrees/myapp/feature/`
- **THEN** `Context` SHALL have `Type = ContextWorktree`,
  `ProjectName = "myapp"`, `BranchName = "feature"`,
  `Path = ~/Worktrees/myapp/feature`

#### Scenario: Outside git

- **WHEN** user is in a directory that is not a git repo
- **THEN** `Context` SHALL have `Type = ContextOutsideGit`
- **AND** all other fields SHALL be empty

### Requirement: `PathType` enum

The system SHALL define `PathType` with constants `PathTypeProject`,
`PathTypeWorktree`, `PathTypeInvalid`. `String()` SHALL return
`"project"`, `"worktree"`, `"invalid"` respectively.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ResolutionResult` struct

The system SHALL provide a `ResolutionResult` struct with the
following shape:

```go
type ResolutionResult struct {
    ResolvedPath string
    Type         PathType
    ProjectName  string
    BranchName   string
    Explanation  string
}
```



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ResolutionSuggestion` struct (live fields)

```go
type ResolutionSuggestion struct {
    Text        string   // displayed in completion menu
    Description string   // shown as the secondary line
    Type        PathType
    ProjectName string
    BranchName  string
    IsCurrent   bool     // populated for current-worktree sorting/prioritizing
    IsDirty     bool     // populated for the "modified" indicator
    // The following fields exist in the struct but are NOT populated anywhere
    // in the current codebase. See openspec/dead-code.md.
    Remote      string
    StyleHint   string
}
```

The `Remote` and `StyleHint` fields are kept on the struct to avoid
breaking future consumers but SHALL be treated as zero-valued until a
caller populates them.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `SuggestionOption`

`SuggestionOption` SHALL be a functional option applied to suggestion
generation. The built-in `WithExistingOnly()` SHALL filter suggestions
to materialized worktrees (no remote-only branches, no stale entries).


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
