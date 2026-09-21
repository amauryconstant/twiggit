# Spec Delta: Application Worktree Management

## MODIFIED Requirements

### Requirement: WorktreeService constructor dependencies

The `WorktreeService` constructor SHALL accept the
`application.GoGitClient`, `application.CLIClient`, `application.ProjectService`,
and `*domain.Config` as injected dependencies. No globals, no `init()`.
See `application-service-interfaces`.

#### Scenario: Constructor signature reflects two role interfaces

- **WHEN** `NewWorktreeService` is invoked
- **THEN** its parameter list SHALL include separate `GoGitClient` and
  `CLIClient` arguments
- **AND** it SHALL NOT accept a single composite `GitClient` argument

#### Scenario: Service stores two role clients as unexported fields

- **WHEN** the service implementation is inspected
- **THEN** it SHALL hold `goGit application.GoGitClient` and
  `cli application.CLIClient` as separate unexported fields
- **AND** read operations (BranchExists, GetRepositoryStatus) SHALL be
  dispatched through the GoGit field
- **AND** worktree and branch mutation operations SHALL be dispatched
  through the CLI field

## REMOVED Requirements

### Requirement: Per-project worktree mutation mutex

**Reason**: The current implementation carries a single
`sync.Mutex` field on `worktreeService` that serializes every
mutation globally. Spec compliance would require a per-project
mutex map, which the current change does not implement and which
the project policy defers to a future quality change. The
requirement is therefore removed rather than left unfulfilled.

**Migration**: Concurrent callers MUST continue to serialize
their own calls if they require project-scoped ordering. A future
change SHALL reintroduce a per-project mutex when concurrent
mutations become a real requirement.
