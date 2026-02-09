## Purpose
Test utilities for unit, integration, and E2E tests

## Helpers Overview
- **Repository helper**: Functional API for test repo creation
- **Git helper**: Git operations for test setup
- **Shell helper**: Shell-related test utilities
- **Worktree helper**: Worktree test utilities

## Repository Helper (`repo.go`)

Functional API for test repo creation:

```go
repo := helpers.NewTestRepo(t, "/tmp/test-repo")
repo.CommitFile("README.md", "# Test")
repo.CreateBranch("feature-1")
repo.CreateBranch("feature-2")
repoPath := repo.Path()
```

**Features:** Idempotent operations, auto-cleanup via t.TempDir(), functional API, cross-platform paths.

## Git Helper (`git.go`)

```go
git.Init(t, repoPath)
git.SetConfig(t, repoPath, "user.email", "test@twiggit.dev")
git.Commit(t, repoPath, "Initial commit")
git.CreateBranch(t, repoPath, "feature-1")
git.CreateWorktree(t, repoPath, worktreePath, "feature-1")
worktrees := git.ListWorktrees(t, repoPath)
```

## Shell Helper (`shell.go`)

```go
output, err := helpers.ExecuteShellCommand(t, "echo", "test")
shellPath := helpers.GetShellPath(t)
configPath := helpers.CreateTempShellConfig(t, "bash")
```

## Worktree Helper (`worktree.go`)

```go
worktreePath := helpers.CreateTestWorktree(t, repoPath, "feature-1")
helpers.ValidateWorktree(t, worktreePath, "feature-1")
helpers.CleanupWorktree(t, worktreePath)
```

## Integration Test Repo

```go
repo := helpers.NewIntegrationTestRepo(t)
repo.CreateBranch("feature-1")
repo.CreateWorktree("feature-1")
repo.CommitInWorktree("feature-1", "change.txt", "content")
```

**Auto-cleanup:** Repository deleted when test completes.

## Cross-Platform Paths

All helpers use `filepath` package for proper path separators, absolute path conversion, symlink resolution, relative path calculation.

## Error Handling

Helpers fail tests on errors. Use `t.Fatal()` for fatal errors, `t.Error()` for non-fatal errors. Include context in error messages.
