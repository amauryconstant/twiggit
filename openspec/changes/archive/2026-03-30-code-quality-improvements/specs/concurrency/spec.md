# Concurrency

## Purpose

Ensure concurrent operations are race-condition-free. Prune operations modifying shared result structures SHALL use proper synchronization.

## ADDED Requirements

### Requirement: Prune result modifications SHALL be synchronized

The PruneMergedWorktrees function SHALL use mutex protection when modifying the shared result structure from multiple goroutines.

#### Scenario: Concurrent prune operations complete without race conditions
- **WHEN** multiple goroutines prune worktrees concurrently on the same project
- **THEN** no race conditions are detected by go test -race
- **AND** all results are correctly accumulated in the shared result

#### Scenario: PruneProjectWorktrees synchronization
- **WHEN** pruneProjectWorktrees is called from multiple goroutines
- **THEN** result modifications are protected by mutex
- **AND** no data races occur

## MODIFIED Requirements

### Requirement: Concurrent list operations are safe

The existing requirement in specs/concurrent-operations/spec.md is strengthened by implementation with mutex protection for prune operations.

#### Scenario: Concurrent list on same project
- **WHEN** multiple goroutines call list on the same project simultaneously
- **THEN** all operations complete without race detector warnings
- **AND** results are consistent

#### Scenario: Prune while listing
- **WHEN** prune operation runs during list operation on same project
- **THEN** no race conditions occur
- **AND** operations complete with consistent results
