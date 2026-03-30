# Type Safety

## Purpose

Ensure nil-safe operations throughout the codebase preventing panics in edge cases. All operations SHALL validate inputs before use.

## ADDED Requirements

### Requirement: Nil context SHALL be handled gracefully

The worktree service SHALL handle nil *Context in list operations without panic, returning a validation error instead.

#### Scenario: List with nil context returns error
- **WHEN** ListWorktrees receives nil context
- **THEN** ValidationError is returned indicating context required
- **AND** no panic occurs

#### Scenario: List with valid context proceeds normally
- **WHEN** ListWorktrees receives valid non-nil context
- **THEN** operation proceeds with context

### Requirement: Empty resolved path SHALL be validated

The delete command SHALL validate that ResolvedPath is not empty before attempting worktree operations.

#### Scenario: Delete with empty resolved path returns error
- **WHEN** delete receives empty resolved path from resolution
- **THEN** ValidationError is returned indicating path required
- **AND** operation does not proceed

#### Scenario: Delete with valid resolved path proceeds normally
- **WHEN** delete receives non-empty resolved path
- **THEN** operation proceeds with the resolved path
