## Purpose
Interface definitions for application layer contracts

## Core Interfaces
- **WorktreeService**: Worktree CRUD operations
- **ProjectService**: Project discovery/management
- **NavigationService**: Context-aware navigation
- **ShellService**: Shell wrapper management

See source files for full interface signatures.

## Request/Result Patterns

### Request Structure
```go
type CreateWorktreeRequest struct {
    ProjectName  string
    BranchName   string
    SourceBranch string
    Context      *domain.Context
}
```

### Result Structure
```go
type WorktreeInfo struct {
    Path   string
    Branch string
    Status string
}
```

## Dependency Injection Pattern
Interfaces injected via constructor:
```go
func NewWorktreeService(
    gitClient infrastructure.GitClient,
    projectRepo ProjectRepository,
    logger Logger,
) WorktreeService
```

## Testing
- **Unit tests**: Mock interfaces for testing
- **Integration tests**: Real implementations
- **Focus**: Interface contracts, behavior, not implementation
