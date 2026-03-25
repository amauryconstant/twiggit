## Purpose

All interface contracts are defined here. Infrastructure implementations MUST implement these interfaces.

## Design Rationale

**Why centralize interfaces?**
- Service layer depends only on `application/` for contracts
- Infrastructure implementations satisfy `application/` interfaces
- Enables mocking in tests (see `test/mocks/AGENTS.md`)

**Dependency Rule**:
```
service/ imports application/ ✓
application/ imports domain/ ✓
application/ imports infrastructure/ ✗
```

## Infrastructure Contracts

These interfaces define contracts between service layer and infrastructure implementations:

| Interface | Purpose | Implementation |
|-----------|---------|----------------|
| `ConfigManager` | Configuration loading | `infrastructure/` |
| `ContextDetector` | Git context detection | `infrastructure/` |
| `ContextResolver` | Identifier resolution | `infrastructure/` |
| `GitClient` | Unified git operations | `infrastructure/` |
| `GoGitClient` | go-git operations | `infrastructure/` |
| `CLIClient` | CLI git operations | `infrastructure/` |
| `HookRunner` | Hook execution | `infrastructure/` |
| `ShellInfrastructure` | Shell integration | `infrastructure/` |

### ConfigManager
- `Load() (*domain.Config, error)` - Load from defaults + config file
- `GetConfig() *domain.Config` - Returns immutable config after Load

### ContextDetector
- `DetectContext(dir string) (*domain.Context, error)` - Detect from directory

### ContextResolver
- `ResolveIdentifier(ctx, identifier) (*domain.ResolutionResult, error)`
- `GetResolutionSuggestions(ctx, partial, opts...) ([]*domain.ResolutionSuggestion, error)`

### GitClient (Composite)
- Combines `GoGitClient` + `CLIClient` for unified operations

### GoGitClient
- `OpenRepository(path) (*git.Repository, error)`
- `ListBranches(ctx, repoPath) ([]domain.BranchInfo, error)`
- `BranchExists(ctx, repoPath, branchName) (bool, error)`
- `GetRepositoryStatus(ctx, repoPath) (domain.RepositoryStatus, error)`
- `ValidateRepository(path) error`
- `GetRepositoryInfo(ctx, repoPath) (*domain.GitRepository, error)`
- `ListRemotes(ctx, repoPath) ([]domain.RemoteInfo, error)`
- `GetCommitInfo(ctx, repoPath, hash) (*domain.CommitInfo, error)`

### CLIClient
- `CreateWorktree(ctx, repoPath, branch, source, worktreePath) error`
- `DeleteWorktree(ctx, repoPath, worktreePath, force) error`
- `ListWorktrees(ctx, repoPath) ([]domain.WorktreeInfo, error)`
- `PruneWorktrees(ctx, repoPath) error`
- `IsBranchMerged(ctx, repoPath, branchName) (bool, error)`
- `DeleteBranch(ctx, repoPath, branchName) error`

### HookRunner
- `Run(ctx, *HookRunRequest) (*domain.HookResult, error)`
- Hook types: `post-create`
- Env vars: `TWIGGIT_WORKTREE_PATH`, `TWIGGIT_PROJECT_NAME`, `TWIGGIT_BRANCH_NAME`, `TWIGGIT_SOURCE_BRANCH`, `TWIGGIT_MAIN_REPO_PATH`

### ShellInfrastructure
- `GenerateWrapper(shellType) (string, error)`
- `ComposeWrapper(template, shellType) string`
- `DetectConfigFile(shellType) (string, error)`
- `InstallWrapper(shellType, wrapper, configFile, force) error`
- `ValidateInstallation(shellType, configFile) error`

## Service Contracts

### ContextService
- `GetCurrentContext() (*domain.Context, error)`
- `DetectContextFromPath(path) (*domain.Context, error)`
- `ResolveIdentifier(identifier) (*domain.ResolutionResult, error)`
- `ResolveIdentifierFromContext(ctx, identifier) (*domain.ResolutionResult, error)`
- `GetCompletionSuggestions(partial) ([]*domain.ResolutionSuggestion, error)`
- `GetCompletionSuggestionsFromContext(ctx, partial) ([]*domain.ResolutionSuggestion, error)`

### WorktreeService
- `CreateWorktree(ctx, *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error)`
- `DeleteWorktree(ctx, *domain.DeleteWorktreeRequest) error`
- `ListWorktrees(ctx, *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error)`
- `GetWorktreeStatus(ctx, worktreePath) (*domain.WorktreeStatus, error)`
- `ValidateWorktree(ctx, worktreePath) error`
- `PruneMergedWorktrees(ctx, *domain.PruneWorktreesRequest) (*domain.PruneWorktreesResult, error)`
- `BranchExists(ctx, projectPath, branchName) (bool, error)`
- `IsBranchMerged(ctx, worktreePath, branchName) (bool, error)`
- `GetWorktreeByPath(ctx, projectPath, worktreePath) (*domain.WorktreeInfo, error)`

### ProjectService
- `DiscoverProject(ctx, projectName, context) (*domain.ProjectInfo, error)`
- `ValidateProject(ctx, projectPath) error`
- `ListProjects(ctx) ([]*domain.ProjectInfo, error)`
- `ListProjectSummaries(ctx) ([]*domain.ProjectSummary, error)`
- `GetProjectInfo(ctx, projectPath) (*domain.ProjectInfo, error)`

### NavigationService
- `ResolvePath(ctx, *domain.ResolvePathRequest) (*domain.ResolutionResult, error)`
- `ValidatePath(ctx, path) error`
- `GetNavigationSuggestions(ctx, context, partial) ([]*domain.ResolutionSuggestion, error)`

### ShellService
- `SetupShell(ctx, *domain.SetupShellRequest) (*domain.SetupShellResult, error)`
- `ValidateInstallation(ctx, *domain.ValidateInstallationRequest) (*domain.ValidateInstallationResult, error)`
- `GenerateWrapper(ctx, *domain.GenerateWrapperRequest) (*domain.GenerateWrapperResult, error)`

## Request Types

```go
type CreateWorktreeRequest struct {
    ProjectName, BranchName, SourceBranch string
    Context *domain.Context
    Force   bool
}
```

**Result types:** See `internal/domain/AGENTS.md` (WorktreeInfo, PruneWorktreesResult)

## Dependency Injection

```go
func NewWorktreeService(
    gitClient application.GitClient,
    projectService application.ProjectService,
    config *domain.Config,
) application.WorktreeService
```

**Compile-time checks:** Infrastructure implementations include `var _ Interface = (*Implementation)(nil)` to verify interface compliance.
