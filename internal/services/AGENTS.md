## Service Architecture
Layer: Application services orchestrate domain logic + infrastructure

**Services:**
- WorktreeService: Worktree CRUD operations
- ProjectService: Project discovery/management
- NavigationService: Context-aware navigation
- ShellService: Shell wrapper management

## Service Implementation Pattern

```go
type WorktreeService struct {
    gitClient   infrastructure.GitClient
    projectRepo domain.ProjectRepository
    config      *domain.Config
    logger      Logger
}

func (s *WorktreeService) CreateWorktree(
    ctx context.Context,
    req *service_requests.CreateWorktreeRequest,
) (*service_results.WorktreeInfo, error) {
    if err := validateCreateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    project, err := s.projectRepo.FindByName(ctx, req.ProjectName)
    if err != nil {
        return nil, fmt.Errorf("project not found: %w", err)
    }
    worktree, err := s.gitClient.CreateWorktree(ctx, project.GitRepoPath, req.BranchName, req.SourceBranch, worktreePath)
    if err != nil {
        return nil, fmt.Errorf("create failed: %w", err)
    }
    return &service_results.WorktreeInfo{Path: worktreePath, Branch: req.BranchName}, nil
}
```

## Error Handling
- Wrap errors with context: `fmt.Errorf("action failed: %w", err)`
- Use domain error types: `ValidationError`, `WorktreeServiceError`
- Never panic in production code

## Testing
- **Unit tests**: Testify suites with mocks from `test/mocks/services/`
- **Integration tests**: Real git repos in temp dirs
- **Build tags**: `//go:build integration`
- **Skip in short mode**: `if testing.Short() { t.Skip() }`

## Service-Specific Patterns

### WorktreeService
- Validate project name, branch name before git operations
- Use GitClient for worktree operations
- Handle worktree existence errors with specific error messages

### ProjectService
- Discover projects by name or from context
- Validate project directories contain valid git repos
- Use ContextDetector for context-aware discovery

### NavigationService
- Delegate to ContextResolver for identifier resolution
- Provide navigation suggestions based on context
- Validate paths before returning

### ShellService
- Generate shell-specific wrapper functions
- Include escape hatch for builtin cd
- Write to shell-specific config file

## Quality Requirements
- All golangci-lint checks SHALL pass
- Business logic separated from infrastructure
- Clear interfaces for external dependencies
- Exported functions and structs documented
- Error messages include actionable guidance
