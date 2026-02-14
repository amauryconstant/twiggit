## Purpose
Interface definitions for application layer contracts

## Core Interfaces

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
    gitService infrastructure.GitClient,
    projectService application.ProjectService,
    config *domain.Config,
) application.WorktreeService
```
