# Go Basics Patterns

Go-specific patterns for package naming, error handling, and interface design in codebase auditing.

## Package Naming

### Conventions

- **Singular package names**: `service`, `application`, `domain`, `infrastructure`, `version`
- **Lowercase package names**: All lowercase, single words, no underscores
- **Domain package**: No dependencies on other internal packages
- **Application package**: Depends only on `domain`
- **Infrastructure package**: Depends only on `domain`
- **Services package**: Depends on `application`, `domain`, `infrastructure`

### Anti-Patterns

- **Plural package names**: `services` should be `service`, `repositories` should be `repository`
- **Mixed case**: Package names with uppercase letters or underscores

### Common Patterns

**Note**: Examples below illustrate patterns found during a specific codebase audit. For general use, see audit-areas.md for pattern descriptions.

- **Interfaces**: `Interface` suffix (e.g., `WorktreeService`, `GitClient`)
- **Structs**: Concrete type without suffix (e.g., `worktreeService`, `gitClient`)
- **Constructors**: `New` prefix (e.g., `NewWorktreeService`, `NewConfigManager`)
- **Private types**: Lowercase first letter (e.g., `worktreeService`, `gitClient`)
- **Exported types**: PascalCase (e.g., `CreateWorktree`, `ListProjects`)
- **Getters**: `Get` prefix (e.g., `GetName`, `GetPath`)
- **Setters**: `Set` prefix (e.g., `SetName`, `SetPath`)
- **Booleans**: `Is` or `Has` prefix (e.g., `IsValid`, `HasContext`)
- **Converters**: `To` prefix (e.g., `ToString`, `ToContext`)

### Anti-Patterns

- **Abbreviations**: Unclear abbreviations (`Mgt`, `Svc`, `Cfg`)
- **Hungarian notation**: Type prefixes in names (`strName`, `intCount`)
- **Inconsistent casing**: Mixing camelCase and PascalCase for similar concepts

## Error Handling

### Idioms

- **Error wrapping**: `fmt.Errorf("operation failed: %w", err)` to preserve error chain
- **Domain error types**: Use specific error types for different error categories
- **Error interfaces**: Implement `error` interface for custom errors
- **Sentinels**: Define sentinel errors for common cases (e.g., `ErrNotFound`, `ErrInvalid`)

### Common Patterns

```go
// Infrastructure layer - return domain error types
return domain.NewGitRepositoryError(path, "operation", "message", err)

// Services layer - wrap with context
return fmt.Errorf("validation failed: %w", err)

// Constructors - use domain errors
return nil, domain.NewValidationError("Type", "field", "message")

// Check error types with errors.As
var gitErr *domain.GitRepositoryError
if errors.As(err, &gitErr) {
    // Handle specific error type
}
```

### Anti-Patterns

- **Panic in production**: `panic()` should only be used in initialization or unrecoverable conditions
- **String matching**: `strings.Contains(err.Error(), "text")` instead of type checking
- **Error swallowing**: Ignoring errors without handling or logging
- **String-based error detection** (CRITICAL): Using string matching on error messages
  ```go
  // BAD: Fragile, breaks if error message changes
  if strings.Contains(err.Error(), "worktree not found") {
      return fmt.Errorf("worktree not found")
  }

  // GOOD: Type-based detection
  var worktreeErr *domain.WorktreeServiceError
  if errors.As(err, &worktreeErr) {
      return fmt.Errorf("worktree not found: %s", worktreeErr.WorktreePath)
  }
  ```

## Interface Design

### Patterns

- **Small interfaces**: Prefer many small interfaces over few large ones
- **Accept interfaces**: Return interfaces rather than concrete types
- **Behavior over data**: Define what to do, not what something is

### Common Interface Shapes

```go
// Service interface (application layer)
type WorktreeService interface {
    CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error)
    DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error
}

// Infrastructure interface (infrastructure layer)
type GitClient interface {
    CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error
    ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error)
}
```

### Anti-Patterns

- **God interfaces**: Too many methods (>10) with mixed concerns
- **Interface pollution**: Defining interfaces that are never implemented
- **Unused interface methods**: Methods defined but never called

## Audit-Specific Patterns

### For Package Structure Audits

- Check that `domain/` has no internal dependencies
- Verify all service interfaces are in `application/interfaces.go`
- Ensure infrastructure depends only on `domain/`
- Verify services depends on appropriate layers (application, domain, infrastructure)

### For Duplicate Code Audits

- Look for identical function signatures across files
- Check for similar logic with minor variations
- Identify template-like code (copy-paste with small changes)
- Find duplicate struct definitions

### For Interface Compliance Audits

- Verify all interface methods have implementations
- Check for unused interface definitions
- Ensure interface methods match signatures exactly
- Look for interface methods only called in tests

### For Documentation Accuracy Audits

- Compare AGENTS.md struct definitions with actual code
- Check interface method signatures match implementations
- Look for missing field documentation
- Find undocumented types and methods

### For Import Consistency Audits

- Verify import ordering: stdlib (alphabetical), third-party (alphabetical), internal (alphabetical)
- Check for circular dependencies (try `go build`)
- Look for unused imports (run `golangci-lint`)
- Verify layer dependencies are correct
