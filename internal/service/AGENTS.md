## Service Architecture
Layer: Application services orchestrate domain logic + infrastructure

**Services:**
- ContextService: Context detection and identifier resolution
- WorktreeService: Worktree CRUD operations
- ProjectService: Project discovery/management
- NavigationService: Context-aware navigation
- ShellService: Shell wrapper management

## Service Implementation Pattern

```go
type worktreeService struct {
    gitService     infrastructure.GitClient
    projectService application.ProjectService
    config         *domain.Config
}

func NewWorktreeService(
    gitService infrastructure.GitClient,
    projectService application.ProjectService,
    config *domain.Config,
) application.WorktreeService {
    return &worktreeService{
        gitService:     gitService,
        projectService: projectService,
        config:         config,
    }
}

func (s *worktreeService) CreateWorktree(
    ctx context.Context,
    req *domain.CreateWorktreeRequest,
) (*domain.WorktreeInfo, error) {
    // ValidationError returned directly, not wrapped
    if err := s.validateCreateRequest(req); err != nil {
        return nil, err
    }
    project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve project: %w", err)
    }
    worktree, err := s.gitService.CreateWorktree(ctx, project.GitRepoPath, req.BranchName, req.SourceBranch, worktreePath)
    if err != nil {
        return nil, domain.NewWorktreeServiceError(worktreePath, req.BranchName, "CreateWorktree", "failed to create worktree", err)
    }
    return &domain.WorktreeInfo{Path: worktreePath, Branch: req.BranchName}, nil
}
```

## Error Handling

### Error Wrapping Rules

1. **ValidationError MUST be returned directly, not wrapped**
   - ValidationError already provides context (request type, field, message, suggestions)
   - Wrapping adds unnecessary nesting and breaks error type checking

   ```go
   // Wrong: Wrapping ValidationError
   if err := s.validateRequest(req); err != nil {
       return nil, fmt.Errorf("validation failed: %w", err)
   }

   // Right: Return ValidationError directly
   if err := s.validateRequest(req); err != nil {
       return nil, err
   }
   ```

2. **Wrap non-validation errors with domain-specific ServiceError types**
   - Use `domain.NewWorktreeServiceError()` for worktree operations
   - Use `domain.NewProjectServiceError()` for project operations
   - Use `domain.NewNavigationServiceError()` for navigation operations

   ```go
   // Wrap infrastructure errors with service context
   err = s.gitService.CreateWorktree(ctx, repoPath, branch, sourceBranch, worktreePath)
   if err != nil {
       return nil, domain.NewWorktreeServiceError(worktreePath, branchName, "CreateWorktree", "failed to create worktree", err)
   }
   ```

3. **Use fmt.Errorf("operation failed: %w", err) only when no domain error type applies**
   - For generic operation failures that don't have a specific domain error type
   - Always use `%w` verb to preserve error chain
   - Provide meaningful context in the error message

   ```go
    // For non-domain-specific failures
    project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve project: %w", err)
    }
```

### Error Type Checking

**Use `errors.As()` for type checking instead of string matching**:
```go
// Wrong: String-based error detection
if strings.Contains(err.Error(), "worktree not found") {
    return nil
}

// Right: Type-based error detection
var worktreeErr *domain.WorktreeServiceError
if errors.As(err, &worktreeErr) && worktreeErr.Message == "worktree not found in any project" {
    return nil
}
```

**Exception**: String-based checks are acceptable for parsing external CLI output (infrastructure layer only).

## Testing
- **Unit tests**: Testify suites with mocks from `test/mocks/`
- **Integration tests**: Real git repos in temp dirs
- **Build tags**: `//go:build integration`
- **Skip in short mode**: `if testing.Short() { t.Skip() }`

## Service-Specific Patterns

### WorktreeService
- Validate project name, branch name before git operations
- Use GitClient for worktree operations
- Handle worktree existence errors with specific error messages
- Methods: `BranchExists`, `IsBranchMerged`, `GetWorktreeByPath` (added for cmd layer isolation)

### ProjectService
- Discover projects by name or from context
- Validate project directories contain valid git repos
- Use ContextDetector for context-aware discovery
- Method: `ListProjectSummaries` for lightweight listings without expensive git data

### ContextService
- Detect context from current working directory
- Resolve identifiers based on current context
- Provide completion suggestions for context-aware operations
- Delegate to ContextDetector and ContextResolver (infrastructure layer)

### NavigationService
- Delegate to ContextResolver for identifier resolution
- Provide navigation suggestions based on context
- Validate paths before returning

### ShellService
- Generate shell-specific wrapper functions (delegates to ShellInfrastructure)
- Include escape hatch for builtin cd
- Write to shell-specific config file
- Auto-detect shell from SHELL environment variable when not specified
- Auto-detect config file location when not specified
- Validate shell types using `domain.IsValidShellType()` (domain-level validation)

## Quality Requirements
- All golangci-lint checks SHALL pass
- Business logic separated from infrastructure
- Clear interfaces for external dependencies
- Exported functions and structs documented
- Error messages include actionable guidance
