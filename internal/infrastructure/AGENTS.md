## Infrastructure Layer
Layer: External integrations (git, config, CLI execution)

## Error Handling

### Infrastructure Layer Error Wrapping

**Rule**: Infrastructure layer SHALL always return domain error types, not plain errors or fmt.Errorf

| Operation | Error Type | Pattern |
|-----------|-------------|----------|
| Repository operations | `domain.NewGitRepositoryError(path, message, cause)` | GoGit operations |
| Worktree operations | `domain.NewGitWorktreeError(worktreePath, branchName, message, cause)` | CLI operations |
| Context detection | `domain.NewContextDetectionError(path, message, cause)` | Path validation failures |

**Examples**:
```go
// Wrong: Plain error
return nil, fmt.Errorf("failed to open repository: %w", err)

// Right: Domain error type
return nil, domain.NewGitRepositoryError(path, "failed to open repository", err)
```

### Error Chain Preservation

- All error wrapping uses `%w` verb to preserve error chain
- Enables `errors.As()` and `errors.Is()` for error type checking in upper layers
- Domain error types implement `Unwrap()` method

### CLI Error Parsing

**Exception**: String-based checks are appropriate for parsing external CLI output
- `strings.Contains(result.Stderr, "not found")` is acceptable
- This is parsing external output, not internal error type checking
- Do not use string matching for internal error detection (use `errors.As()` instead)

### Silent Error Handling

**Guideline**: Silent error degradation is acceptable for non-critical paths
- Suggestion methods may silently return empty lists on errors
- Document why errors are ignored with comments
- Example: completion suggestions in `ContextResolver`

## Git Client Routing Strategy

Two implementations routed deterministically:

| Operation | GoGitClient | CLIClient | Rationale |
|-----------|-------------|-----------|-----------|
| Open repository | ✅ | ❌ | GoGit is portable, deterministic |
| List branches | ✅ | ❌ | GoGit is portable, deterministic |
| Branch exists | ✅ | ❌ | GoGit is portable, deterministic |
| Get repository status | ✅ | ❌ | GoGit is portable, deterministic |
| Validate repository | ✅ | ❌ | GoGit is portable, deterministic |
| Get repository info | ✅ | ❌ | GoGit is portable, deterministic |
| List remotes | ✅ | ❌ | GoGit is portable, deterministic |
| Get commit info | ✅ | ❌ | GoGit is portable, deterministic |
| Create worktree | ❌ | ✅ | go-git lacks worktree support |
| Delete worktree | ❌ | ✅ | go-git lacks worktree support |
| List worktrees | ❌ | ✅ | go-git lacks worktree support |
| Prune worktrees | ❌ | ✅ | go-git lacks worktree support |
| Is branch merged | ❌ | ✅ | go-git lacks merge status support |

**Composite client** routes to operation-specific implementation (no fallback logic).

## GoGitClient Interface

```go
type GoGitClient interface {
    OpenRepository(path string) (*git.Repository, error)
    ListBranches(ctx context.Context, repoPath string) ([]domain.BranchInfo, error)
    BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)
    GetRepositoryStatus(ctx context.Context, repoPath string) (domain.RepositoryStatus, error)
    ValidateRepository(path string) error
    GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error)
    ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error)
    GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error)
}
```

**Methods SHALL be idempotent and thread-safe.**

## CLIClient Interface

```go
type CLIClient interface {
    CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error
    DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error
    ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error)
    PruneWorktrees(ctx context.Context, repoPath string) error
    IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error)
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
