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

## Path Validation Helper

**Function**: `validatePathUnder(base, target, targetType, baseDesc string) error`
**Purpose**: Validates that a target path is under a base directory, preventing path traversal attacks
**Returns**: Error if validation fails or path is outside base
**Usage**: Used consistently across ContextResolver for path validation

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

**Note**: ContextResolver uses `ProjectRef` type internally for lightweight project references (name + path), distinct from `domain.ProjectInfo` which contains comprehensive details (worktrees, branches, remotes).

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

## ShellInfrastructure Interface

**Location:** `shell_infra.go`

```go
type ShellInfrastructure interface {
    GenerateWrapper(shellType domain.ShellType) (string, error)
    DetectConfigFile(shellType domain.ShellType) (string, error)
    InstallWrapper(shellType domain.ShellType, wrapper, configFile string, force bool) error
    ValidateInstallation(shellType domain.ShellType, configFile string) error
}
```

**Operations:**
- Generate shell-specific wrapper functions (bash, zsh, fish)
- Detect config file location based on shell type
- Install wrapper to shell config file (idempotent with force flag)
- Validate wrapper installation

**Supported Shells:**
- Bash: `.bashrc`, `.bash_profile`, `.profile`
- Zsh: `.zshrc`, `.zprofile`, `.profile`
- Fish: `.config/fish/config.fish`, `config.fish`, `.fishrc`

**Methods SHALL be idempotent.**

## Testing
- **Unit tests**: Testify suites with mocks
- **Integration tests**: Real git repos in temp dirs
- **Build tags**: `//go:build integration`
- **Skip in short mode**: `if testing.Short() { t.Skip() }`
