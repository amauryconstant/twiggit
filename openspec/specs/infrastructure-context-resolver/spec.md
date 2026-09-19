# Capability: Context Resolver

## Purpose

Detect the current CWD context (worktree, project, outside git) and
resolve identifiers to filesystem paths. Owns the canonical
context-priority chain.

## Requirements

### Requirement: Detection priority

The system SHALL detect context with the following priority
(first match wins):

1. **Worktree** — CWD is under
   `<Config.WorktreesDirectory>/<project>/<branch>/` and contains a
   `.git` file (worktree pointer).
2. **Project** — CWD is under `<Config.ProjectsDirectory>/<project>/`
   and contains a `.git` directory.
3. **Outside git** — none of the above.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Worktree detection

`DetectContext(dir)` SHALL recognize a worktree by the presence of
`.git` as a *file* (not a directory) at `dir/.git`. The file content
points back to the main repo's worktree metadata.

#### Scenario: Inside a worktree

- **WHEN** CWD is `~/Worktrees/myapp/feature/`
- **AND** `.git` is a regular file in that directory
- **THEN** `Context.Type = ContextWorktree`, `ProjectName = "myapp"`,
  `BranchName = "feature"`

### Requirement: Project detection

`DetectContext(dir)` SHALL recognize a project by the presence of a
`.git` *directory* at `dir/.git` and a parent directory matching
`Config.ProjectsDirectory/<name>`.

#### Scenario: Inside a project main

- **WHEN** CWD is `~/Projects/myapp/`
- **AND** `.git` is a directory in that directory
- **THEN** `Context.Type = ContextProject`, `ProjectName = "myapp"`,
  `BranchName = ""`

### Requirement: Outside git

When neither worktree nor project patterns match, the system SHALL
return `ContextOutsideGit`.

#### Scenario: Plain home directory

- **WHEN** CWD is `~/` and contains no `.git`
- **THEN** `Context.Type = ContextOutsideGit`

### Requirement: Identifier resolution

`ResolveIdentifier(ctx, identifier)` SHALL resolve:

| Input | From project/worktree ctx | From outside-git ctx |
|---|---|---|
| `<branch>` | worktree of same project | error |
| `<project>` | main path of that project | main path of that project |
| `<project>/<branch>` | cross-project worktree | cross-project worktree |

#### Scenario: Resolve branch

- **WHEN** `ResolveIdentifier` is called with `"feature"` from a project
  context with worktree `feature`
- **THEN** system SHALL return `Type = PathTypeWorktree`,
  `ResolvedPath = ~/Worktrees/<project>/feature`

#### Scenario: Resolve cross-project

- **WHEN** `ResolveIdentifier` is called with `"myapp/feature"` from
  any context
- **THEN** system SHALL return `Type = PathTypeWorktree`,
  `ResolvedPath = ~/Worktrees/myapp/feature`

#### Scenario: Invalid identifier

- **WHEN** `ResolveIdentifier` is called with an unparseable identifier
- **THEN** system SHALL return a `ResolutionError` with `Suggestions`
  listing close matches

### Requirement: Search mode (`Search = true`)

When `ResolvePathRequest.Search = true` or `WithExistingOnly()` is
supplied, the resolver SHALL restrict suggestions to materialized
worktrees. The owning service is `application-navigation-service`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Completion suggestions

`GetResolutionSuggestions(ctx, partial, opts...)` SHALL return
`[]*ResolutionSuggestion` covering projects, branches, and worktrees
matching `partial`, applying `NavigationConfig.FuzzyMatching` and
`CompletionConfig.ExcludeBranches` / `CompletionConfig.ExcludeProjects`.
See `cli-completion` for sort and display rules.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Path-traversal rejection

`ResolveIdentifier` SHALL reject identifiers that contain `..`
segments or absolute paths, returning a `domain.ContextDetectionError`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Symlink-aware comparison

When checking whether `dir` is under `Config.ProjectsDirectory` or
`Config.WorktreesDirectory`, the resolver SHALL resolve symlinks
(via `filepath.EvalSymlinks`) before comparing, to prevent
symlink-based path-traversal bypasses. See `infrastructure-path-utils`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
