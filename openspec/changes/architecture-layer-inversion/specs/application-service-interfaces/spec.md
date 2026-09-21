# Spec Delta: Application Service Interfaces

## ADDED Requirements

### Requirement: Project discovery isolation

`ProjectService` SHALL receive its project-discovery dependency via a
constructor-injected `RepoLocator` interface so that production code
walks the filesystem while tests substitute a deterministic locator
without filesystem coupling.

#### Scenario: Production wiring uses the filesystem locator

- **WHEN** `main.go` constructs `ProjectService`
- **THEN** it SHALL pass an `application.RepoLocator` whose
  implementation walks `Config.ProjectsDirectory` and validates each
  entry with `GoGitClient.ValidateRepository`

#### Scenario: Test wiring substitutes an in-memory locator

- **WHEN** a service test constructs `ProjectService`
- **THEN** it SHALL pass a mock `RepoLocator` whose `FindGitRepositories`
  returns the test's prepared `[]domain.GitDir` directly
- **AND** the test SHALL NOT touch the real filesystem
