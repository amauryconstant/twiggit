## Infrastructure Layer
Layer: External integrations (git, config, CLI execution)

## Error Wrapping

**Rule:** Return domain error types, not plain `fmt.Errorf`.

| Operation | Pattern |
|-----------|---------|
| Repository | `domain.NewGitRepositoryError(path, message, cause)` |
| Worktree | `domain.NewGitWorktreeError(worktreePath, branchName, message, cause)` |
| Context | `domain.NewContextDetectionError(path, message, cause)` |
| Config | `domain.NewConfigError(path, message, cause)` |

**Exception:** String-based checks OK for parsing external CLI output.

## Git Client Routing

| Operation | GoGitClient | CLIClient | Rationale |
|-----------|:-----------:|:---------:|-----------|
| Open repo, List branches, Branch exists | ✅ | ❌ | Portable, deterministic |
| Get status, Validate repo, Get info | ✅ | ❌ | Portable, deterministic |
| List remotes, Get commit info | ✅ | ❌ | Portable, deterministic |
| Create/Delete/List worktree, Prune | ❌ | ✅ | go-git lacks support |
| Is branch merged, Delete branch | ❌ | ✅ | go-git limitations |

## GoGitClient

```go
type GoGitClient interface {
    OpenRepository(path string) (*git.Repository, error)
    ListBranches(ctx, repoPath) ([]domain.BranchInfo, error)
    BranchExists(ctx, repoPath, branchName) (bool, error)
    GetRepositoryStatus(ctx, repoPath) (domain.RepositoryStatus, error)
    ValidateRepository(path string) error
    GetRepositoryInfo(ctx, repoPath) (*domain.GitRepository, error)
    ListRemotes(ctx, repoPath) ([]domain.RemoteInfo, error)
    GetCommitInfo(ctx, repoPath, hash) (*domain.CommitInfo, error)
}
```

**Cache:** LRU cache (default 25 repos) prevents memory leak.
```go
NewGoGitClient(cacheEnabled...)           // default size 25
NewGoGitClientWithSize(cacheSize, ...)    // custom size
```

## CLIClient

```go
type CLIClient interface {
    CreateWorktree(ctx, repoPath, branch, source, worktreePath) error
    DeleteWorktree(ctx, repoPath, worktreePath, force bool) error
    ListWorktrees(ctx, repoPath) ([]domain.WorktreeInfo, error)
    PruneWorktrees(ctx, repoPath) error
    IsBranchMerged(ctx, repoPath, branchName) (bool, error)
    DeleteBranch(ctx, repoPath, branchName) error
}
```

## Path Utilities

| Function | Purpose |
|----------|---------|
| `IsPathUnder(base, target)` | Check target under base, resolves symlinks |
| `ExtractProjectFromWorktreePath(path, worktreesDir)` | Get project name from `{worktreesDir}/{project}/{branch}/...` |
| `NormalizePath(path)` | Absolute path, symlinks resolved |

## Context Detection

**Priority:** Worktree folder → Project folder (`.git/` found) → Outside git

**Worktree pattern:** `$HOME/Worktrees/<project>/<branch>/` with valid `.git` file

## Context Resolution

| Context | `<branch>` | `<project>` | `<project>/<branch>` |
|---------|------------|-------------|----------------------|
| Project | Current project worktree | Different project main | Cross-project worktree |
| Worktree | Different worktree, same project | Different project main | Cross-project worktree |
| Outside | - | Project main | Cross-project worktree |

## ShellInfrastructure

| Method | Purpose |
|--------|---------|
| GenerateWrapper | Shell-specific wrapper |
| DetectConfigFile | Find shell config |
| InstallWrapper | Add wrapper to config |
| ValidateInstallation | Check installed |
| ComposeWrapper | Template → final wrapper |

| Shell | Config Files (preference order) |
|-------|--------------------------------|
| Bash | `.bashrc`, `.bash_profile`, `.profile` |
| Zsh | `.zshrc`, `.zprofile`, `.profile` |
| Fish | `.config/fish/config.fish`, `config.fish`, `.fishrc` |

## Configuration

**Location:** `$HOME/.config/twiggit/config.toml` (XDG)
**Priority:** defaults → config file → env vars (`TWIGGIT_*`) → flags
