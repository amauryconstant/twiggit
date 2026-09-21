# Spec Delta

## ADDED Requirements

### Requirement: PruneWorktrees errors logged before branch delete

`DeleteWorktree` (the underlying call behind `prune --merged`)
SHALL invoke `gitService.PruneWorktrees(ctx, repoPath)` before
the branch delete step. When `PruneWorktrees` returns a
non-nil error, the service SHALL log the error via `slog.Error`
and continue to the branch delete step. The error SHALL be
captured in the result struct so the cmd layer can surface it.

#### Scenario: PruneWorktrees fails but branch delete proceeds
- **WHEN** `PruneWorktrees` returns an error and the target
  branch is otherwise eligible for delete
- **THEN** `slog.Error` records the prune failure with the
  repo path; the branch is deleted; the result struct's
  `PruneError` field is populated; the cmd layer formats
  the partial-failure state for the user

### Requirement: Per-project mutex requirement removed

The previous requirement that per-project worktree mutations
SHALL be protected by a per-project mutex is REMOVED. The
implementation does not currently provide this protection, and
sequential operation is the observed runtime behavior. A future
change may reintroduce per-project mutexes with a real
implementation.

#### Scenario: Concurrent operations on different projects
- **WHEN** two goroutines call `DeleteWorktree` on different
  project paths simultaneously
- **THEN** both operations proceed concurrently without
  blocking; the contract does not require a per-project
  mutex

#### Scenario: Concurrent operations on the same project
- **WHEN** two goroutines call `DeleteWorktree` on the same
  project path simultaneously
- **THEN** both operations are serialized by the file system
  (git's own atomicity); the contract does not require
  application-level locking
