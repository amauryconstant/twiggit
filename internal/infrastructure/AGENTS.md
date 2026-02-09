## Infrastructure Layer
Layer: External integrations (git, config, CLI execution)

## Git Client Routing Strategy

Two implementations routed deterministically:

| Operation | GoGitClient | CLIClient | Rationale |
|-----------|-------------|-----------|-----------|
| Create worktree | ❌ | ✅ | go-git lacks worktree support |
| List worktrees | ✅ | ❌ | GoGit is portable, deterministic |
| Delete worktree | ❌ | ✅ | go-git lacks worktree support |
| Get current branch | ✅ | ❌ | GoGit is portable, deterministic |
| Validate repo | ✅ | ❌ | GoGit is portable, deterministic |

**Composite client** routes to operation-specific implementation (no fallback logic).

## GoGitClient Interface

```go
type GoGitClient interface {
    OpenRepository(path string) (*git.Repository, error)
    ListBranches(ctx context.Context, repoPath string) ([]BranchInfo, error)
    BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)
    GetRepositoryStatus(ctx context.Context, repoPath string) (RepositoryStatus, error)
    ValidateRepository(path string) error
}
```

**Methods SHALL be idempotent and thread-safe.**

## CLIClient Interface

```go
type CLIClient interface {
    CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch, worktreePath string) error
    DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error
    ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
}
```

**Methods SHALL be idempotent (no-op if already deleted).**

## Context Detection Rules

**Canonical location** for context detection logic.

**Context Types:**
- **Project context**: Inside `.git/` directory (found in current/parent dirs)
- **Worktree context**: Path matches `$HOME/Worktrees/<project>/<branch>/`
- **Outside git**: Neither condition met

**Detection Priority:**
1. **Worktree folder**: Path matches `$HOME/Worktrees/<project>/<branch>/` AND contains valid worktree
2. **Project folder**: `.git/` found in current/parent directories AND path doesn't match worktree pattern
3. **Outside git**: Neither condition met

**Implementation:**
- Use `filepath` package (cross-platform compatibility)
- Traverse up directory tree for `.git/` detection
- Validate worktree `.git` file contains "gitdir:" marker

## Context Resolution

**Identifier Resolution:**

| Context | Input | Target |
|---------|-------|--------|
| Project | `<branch>` | Current project worktree |
| Project | `<project>` | Different project main |
| Project | `<project>/<branch>` | Cross-project worktree |
| Worktree | `<branch>` | Different worktree same project |
| Worktree | `main` | Current project main |
| Worktree | `<project>` | Different project main |
| Outside | `<project>` | Project main directory |
| Outside | `<project>/<branch>` | Cross-project worktree |

## Context Resolution

**Identifier Resolution:**

| Context | Input | Target |
|---------|-------|--------|
| Project | `<branch>` | Current project worktree |
| Project | `<project>` | Different project main |
| Project | `<project>/<branch>` | Cross-project worktree |
| Worktree | `<branch>` | Different worktree same project |
| Worktree | `main` | Current project main |
| Worktree | `<project>` | Different project main |
| Outside | `<project>` | Project main directory |
| Outside | `<project>/<branch>` | Cross-project worktree |

## Configuration Management

**Location:** `$HOME/.config/twiggit/config.toml` (XDG standard)
**Format:** TOML (Koanf library)
**Priority:** defaults → config file → env vars → flags
**Env var prefix:** `TWIGGIT_`

```toml
projects_dir = "/custom/path/to/projects"
worktrees_dir = "/custom/path/to/worktrees"
default_source_branch = "main"
```

## Command Execution

`command_executor.go` handles external CLI calls:
- Timeout support (configurable)
- Output capture
- Error handling with exit codes

## Testing
- **Unit tests**: Testify suites with mocks
- **Integration tests**: Real git repos in temp dirs
- **Build tags**: `//go:build integration`
- **Skip in short mode**: `if testing.Short() { t.Skip() }`
