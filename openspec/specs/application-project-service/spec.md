# Capability: Project Service

## Purpose

Service-layer orchestration for project discovery, listing, and
detailed info lookup. Projects are top-level directories under
`Config.ProjectsDirectory` containing a `.git` folder.

## Requirements

### Requirement: Service contract surface

The `ProjectService` interface SHALL expose:

- `DiscoverProject(ctx, projectName string, context *domain.Context) (*ProjectInfo, error)`
- `ValidateProject(ctx, projectPath string) error`
- `ListProjects(ctx) ([]*ProjectInfo, error)`
- `ListProjectSummaries(ctx) ([]*ProjectSummary, error)`
- `GetProjectInfo(ctx, projectPath string) (*ProjectInfo, error)`

Implementation lives in `internal/service/`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Discovery by name or context

`DiscoverProject` SHALL resolve a project by name, by current context,
or both. When `context` is inside a project, that project SHALL be
returned. When `projectName = ""`, the context SHALL be required.

#### Scenario: Discover by name

- **WHEN** user calls `DiscoverProject(ctx, "myapp", nil)`
- **AND** `myapp` exists under `Config.ProjectsDirectory`
- **THEN** system SHALL return its `ProjectInfo`

#### Scenario: Discover from context

- **WHEN** user calls `DiscoverProject(ctx, "", currentCtx)`
- **AND** `currentCtx.Type == ContextProject` and `currentCtx.ProjectName = "myapp"`
- **THEN** system SHALL return the `myapp` `ProjectInfo`

#### Scenario: Case-insensitive match

- **WHEN** user calls `DiscoverProject(ctx, "MyApp", nil)`
- **AND** the actual folder is `myapp`
- **THEN** system SHALL match case-insensitively and return `myapp`

#### Scenario: Not found

- **WHEN** the project does not exist
- **THEN** system SHALL return a not-found error (exit code 6)

### Requirement: List all projects

`ListProjects` SHALL return the full `ProjectInfo` (with branches,
remotes, default branch, worktrees) for every project on disk.

#### Scenario: Multiple projects

- **WHEN** multiple projects exist under `Config.ProjectsDirectory`
- **THEN** `ListProjects` SHALL return one `ProjectInfo` per project

### Requirement: Lightweight summary listing

`ListProjectSummaries` SHALL return `ProjectSummary` (name + path
only, no branches/worktrees) for fast listing when full info is
unnecessary.

#### Scenario: Summary faster than full list

- **WHEN** caller needs only project names and paths
- **THEN** `ListProjectSummaries` SHALL be used instead of `ListProjects`

### Requirement: Project validation

`ValidateProject` SHALL return nil if `projectPath` is a real project
directory (with `.git`), or a `domain.ValidationError` otherwise.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Detailed project info

`GetProjectInfo(ctx, projectPath)` SHALL load the full `ProjectInfo`
(branches, remotes, worktrees, default branch) for a single project.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
