# Spec Delta: Application Service Interfaces

## MODIFIED Requirements

### Requirement: Constructor injection

Every service SHALL receive its dependencies via constructor parameters.
There SHALL be NO package-level globals for services, NO `init()`
wiring, and NO environment-driven service construction at runtime.

`ProjectService` SHALL additionally receive its project-discovery
dependency via a constructor-injected `RepoLocator` interface so that
production code walks the filesystem while tests substitute a
deterministic locator without filesystem coupling.

#### Scenario: WorktreeService construction

- **WHEN** `WorktreeService` is built
- **THEN** its constructor SHALL take `GitClient`, `ProjectService`,
  and `*domain.Config` as explicit arguments
- **AND** SHALL store them as unexported fields

#### Scenario: ProjectService construction

- **WHEN** `ProjectService` is built
- **THEN** its constructor SHALL take `GoGitClient`, `CLIClient`,
  `RepoLocator`, and `*domain.Config` as explicit arguments
- **AND** SHALL store them as unexported fields

#### Scenario: Production wiring uses the filesystem locator

- **WHEN** the production wiring constructs `ProjectService`
- **THEN** it SHALL pass an `application.RepoLocator` whose
  implementation walks `Config.ProjectsDirectory` and validates each
  entry with `GoGitClient.ValidateRepository`

#### Scenario: Test wiring substitutes an in-memory locator

- **WHEN** a service test constructs `ProjectService`
- **THEN** it SHALL pass a mock `RepoLocator` whose `FindGitRepositories`
  returns the test's prepared `[]domain.GitDir` directly
- **AND** the test SHALL NOT touch the real filesystem

#### Scenario: No globals

- **WHEN** searching the codebase for `var ... = NewWorktreeService(...)`
- **THEN** such declarations SHALL NOT exist
- **AND** all wiring SHALL happen in the presentation layer
