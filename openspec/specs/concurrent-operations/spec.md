# Concurrent Operations

## Purpose

The system handles concurrent operations safely without race conditions, ensuring data consistency when multiple operations occur simultaneously.

## Requirements

### Requirement: Concurrent list operations are safe
The system SHALL handle concurrent list operations on the same project without race conditions.

#### Scenario: Concurrent list on same project
- **WHEN** multiple goroutines call list on the same project simultaneously
- **THEN** all operations complete without race detector warnings
- **AND** results are consistent

#### Scenario: List during worktree creation
- **WHEN** list operation runs during worktree creation on same project
- **THEN** no race conditions occur
- **AND** list reflects consistent state

### Requirement: Concurrent worktree operations are safe
The system SHALL handle concurrent create and delete operations on different worktrees without race conditions.

#### Scenario: Concurrent create different worktrees
- **WHEN** multiple goroutines create different worktrees on same project
- **THEN** all operations complete successfully
- **AND** no race conditions detected

#### Scenario: Concurrent delete different worktrees
- **WHEN** multiple goroutines delete different worktrees on same project
- **THEN** all operations complete successfully
- **AND** no race conditions detected

#### Scenario: Create and delete different worktrees concurrently
- **WHEN** goroutine A creates worktree-1 while goroutine B deletes worktree-2
- **THEN** both operations complete without race conditions

### Requirement: Prune during list is safe
The system SHALL handle prune operations running concurrently with list operations.

#### Scenario: Prune while listing
- **WHEN** prune operation runs during list operation on same project
- **THEN** no race conditions occur
- **AND** operations complete with consistent results
