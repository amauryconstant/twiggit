# Error Handling

## Purpose

Ensure consistent error handling following project conventions. All service errors SHALL have IsNotFound() method for reliable error categorization. ValidationError SHALL be returned directly without wrapping.

## ADDED Requirements

### Requirement: Service errors SHALL include IsNotFound method

ProjectServiceError, NavigationServiceError, and ResolutionError SHALL implement IsNotFound() bool method for consistent error categorization via errors.As().

#### Scenario: ProjectServiceError IsNotFound returns true for not-found errors
- **WHEN** ProjectServiceError is created with NotFound kind
- **THEN** IsNotFound() returns true

#### Scenario: ProjectServiceError IsNotFound returns false for other errors
- **WHEN** ProjectServiceError is created with other kinds (CreateFailed, DeleteFailed)
- **THEN** IsNotFound() returns false

#### Scenario: NavigationServiceError IsNotFound returns true for not-found errors
- **WHEN** NavigationServiceError is created with NotFound kind
- **THEN** IsNotFound() returns true

#### Scenario: ResolutionError IsNotFound returns true for not-found path resolution
- **WHEN** ResolutionError is created with NotFound kind for path resolution failures
- **THEN** IsNotFound() returns true

### Requirement: ValidationError SHALL be returned directly

The cmd layer SHALL return domain.ValidationError directly without wrapping with fmt.Errorf or other error types.

#### Scenario: Create with invalid source branch returns ValidationError directly
- **WHEN** user provides source branch that does not exist
- **THEN** ValidationError is returned directly
- **AND** error message includes the invalid branch name

#### Scenario: Delete with non-existent worktree returns ValidationError directly
- **WHEN** user attempts to delete a worktree that does not exist
- **THEN** ValidationError is returned directly with worktree path context

#### Scenario: Cd with invalid path returns ValidationError directly
- **WHEN** user provides path that cannot be resolved
- **THEN** ValidationError is returned directly with path context

### Requirement: Error wrapping uses domain types

Service layer SHALL wrap infrastructure errors with appropriate domain error types using fmt.Errorf pattern "operation: %w".

#### Scenario: Worktree service wraps project resolution errors
- **WHEN** worktree operation fails due to project resolution failure
- **THEN** error is wrapped with ProjectServiceError
- **AND** original error is accessible via Unwrap()
