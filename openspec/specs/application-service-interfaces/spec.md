# Capability: Service Interfaces and Dependency Injection

## Purpose

Define the application-layer service interfaces and the DI rules that
all service implementations SHALL follow. No `ServiceContainer` global,
no `init()` wiring, no implicit dependencies.

## Requirements

### Requirement: Service interfaces in `internal/application/`

The application layer SHALL own all service interfaces:
`WorktreeService`, `ProjectService`, `NavigationService`,
`ContextService`, `ShellService`. The service layer SHALL provide the
implementations and SHALL be the only consumer of those interfaces
apart from the cmd layer.

#### Scenario: Interface location

- **WHEN** a service interface is added or changed
- **THEN** it SHALL live in `internal/application/`
- **AND** the implementation SHALL live in `internal/service/`

### Requirement: Constructor injection

Every service SHALL receive its dependencies via constructor parameters.
There SHALL be NO package-level globals for services, NO `init()`
wiring, and NO environment-driven service construction at runtime.

#### Scenario: WorktreeService construction

- **WHEN** `WorktreeService` is built
- **THEN** its constructor SHALL take `GitClient`, `ProjectService`,
  and `*domain.Config` as explicit arguments
- **AND** SHALL store them as unexported fields

#### Scenario: No globals

- **WHEN** searching the codebase for `var ... = NewWorktreeService(...)`
- **THEN** such declarations SHALL NOT exist
- **AND** all wiring SHALL happen in `cmd/root.go`

### Requirement: Compile-time interface check

Every infrastructure or service implementation SHALL include a
compile-time assertion `var _ Interface = (*Implementation)(nil)` at
package scope so that breaking changes to interfaces surface at build
time rather than at runtime.

#### Scenario: Git client assertion

- **WHEN** the GoGitClient interface changes
- **THEN** `var _ application.GoGitClient = (*GoGitClientImpl)(nil)` SHALL fail
  to compile until the implementation is updated

### Requirement: Layer dependency direction

Imports SHALL flow downward: `cmd → application → domain`,
`service → application → domain`,
`infrastructure → application → domain` (and `infrastructure → domain`
for error types only). Reverse imports SHALL NOT compile.

#### Scenario: Domain has no cross-layer deps

- **WHEN** `internal/domain/` is inspected
- **THEN** it SHALL NOT import `internal/application/`, `internal/service/`,
  `internal/infrastructure/`, or `cmd/`

### Requirement: ContextService contract

The `ContextService` interface SHALL expose:

- `GetCurrentContext() (*Context, error)`
- `DetectContextFromPath(path) (*Context, error)`
- `ResolveIdentifier(identifier) (*ResolutionResult, error)`
- `ResolveIdentifierFromContext(ctx, identifier) (*ResolutionResult, error)`
- `GetCompletionSuggestions(partial string, opts ...) ([]*ResolutionSuggestion, error)`
- `GetCompletionSuggestionsFromContext(ctx, partial, opts ...) ([]*ResolutionSuggestion, error)`

The context-detection priority (worktree > project > outside git) is
owned by `infrastructure-context-resolver`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
