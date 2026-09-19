# Capability: Navigation Service

## Purpose

Service-layer orchestration for path resolution and validation, used
by the `cd` command and completion engines.

## Requirements

### Requirement: Service contract surface

The `NavigationService` interface SHALL expose:

- `ResolvePath(ctx, *ResolvePathRequest) (*ResolutionResult, error)`
- `ValidatePath(ctx, path string) error`
- `GetNavigationSuggestions(ctx, *domain.Context, partial string) ([]*ResolutionSuggestion, error)`

Implementation lives in `internal/service/`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Resolve identifier

`ResolvePath` SHALL resolve a partial or full identifier
(`branch`, `project`, `project/branch`) against the supplied context
or against the current CWD context if none is provided.

#### Scenario: Resolve branch

- **WHEN** `ResolvePath` is called with `Identifier = "feature"` from a
  project context with branch `feature`
- **THEN** system SHALL return a `ResolutionResult` with
  `Type = PathTypeWorktree` and `ResolvedPath` set to the worktree path

#### Scenario: Resolve cross-project

- **WHEN** `ResolvePath` is called with `Identifier = "other/feature"`
- **THEN** system SHALL return the path to `~/Worktrees/other/feature`

#### Scenario: Resolve project main

- **WHEN** `ResolvePath` is called with `Identifier = "other"` (no slash)
- **THEN** system SHALL return the path to `~/Projects/other`

### Requirement: Existing-only filter

When the `ResolvePathRequest.Search = true` or when `WithExistingOnly()`
is supplied, the system SHALL filter to materialized worktrees only.

#### Scenario: Existing-only filter

- **WHEN** `ResolvePath` is called with `WithExistingOnly()`
- **THEN** the result SHALL exclude branches without a worktree

### Requirement: Path validation

`ValidatePath` SHALL return nil if `path` is a valid absolute path
under one of the configured roots (`ProjectsDirectory` or
`WorktreesDirectory`), or a `domain.ContextDetectionError` otherwise.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Navigation suggestions

`GetNavigationSuggestions` SHALL return completion suggestions for
navigation. The resolver contract is owned by
`infrastructure-context-resolver`.

#### Scenario: Suggest from project context

- **WHEN** user is inside a project
- **AND** `GetNavigationSuggestions(ctx, "fe")` is called
- **THEN** system SHALL return branch suggestions matching `fe*` for
  the current project

#### Scenario: Suggest from outside git

- **WHEN** user is outside any git context
- **AND** `GetNavigationSuggestions(ctx, "fe")` is called
- **THEN** system SHALL return project suggestions matching `fe*`
