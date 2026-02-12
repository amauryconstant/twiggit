# Go Code Patterns

Go-specific heuristics, conventions, and anti-patterns for codebase auditing.

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

## Testing

### Frameworks

- **Testify**: `github.com/stretchr/testify` - most common
  - `assert` package for assertions
  - `require` package for fatal assertions
  - `suite` package for test suites
- **Table-driven tests**: Subtests with `t.Run()` for different scenarios
- **Mocking**: Functional mocks (function fields) or `mock.Mock` from testify

### Test Organization

- **Test alongside source**: `file_test.go` next to `file.go`
- **Package tests**: `_test.go` suffix
- **Build tags**: `//go:build integration` for integration tests
- **Short mode**: `if testing.Short() { t.Skip() }`

### Common Patterns

```go
// Table-driven test
func TestCreateWorktree(t *testing.T) {
    testCases := []struct {
        name string
        input *domain.CreateWorktreeRequest
        wantErr bool
    }{
        {
            name: "valid request",
            input: &domain.CreateWorktreeRequest{...},
            wantErr: false,
        },
        {
            name: "invalid project name",
            input: &domain.CreateWorktreeRequest{ProjectName: ""},
            wantErr: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := service.CreateWorktree(context.Background(), tc.input)
            if tc.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

// Mock setup
func setupTestWorktreeService() application.WorktreeService {
    gitClient := mocks.NewMockGitClient()
    projectService := mocks.NewMockProjectService()
    config := domain.DefaultConfig()
    return NewWorktreeService(gitClient, projectService, config)
}
```

### Anti-Patterns

- **Inline mocks**: Mocks defined in test files instead of centralized `test/mocks/`
- **Test logic in tests**: Complex setup/teardown that should be helpers
- **Brittle assertions**: Checking exact error messages instead of error types

## Context Usage

### Patterns

- **First parameter**: `context.Context` as first parameter in exported functions
- **Background context**: Use `context.Background()` or `context.TODO()` when no parent context
- **Pass through**: Pass context through call chain, don't create new at each level
- **Timeouts**: Use `context.WithTimeout()` or `context.WithDeadline()` for time-bound operations

### Common Patterns

```go
// Service method with context
func (s *worktreeService) CreateWorktree(
    ctx context.Context,
    req *domain.CreateWorktreeRequest,
) (*domain.WorktreeInfo, error) {
    // Use ctx in all operations
    branch, err := s.gitClient.BranchExists(ctx, repoPath, req.BranchName)
    if err != nil {
        return nil, fmt.Errorf("failed to check branch: %w", err)
    }
    // ...
}

// HTTP client with context timeout
client := &http.Client{Timeout: 30 * time.Second}
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

### Anti-Patterns

- **Ignoring context**: Passing `context.Background()` when parent context available
- **Discarding context**: Not passing context through call chain
- **Context leaks**: Creating contexts without cancellation or timeout

## Concurrency

### Patterns

- **Goroutine lifecycle**: Ensure goroutines can be cancelled via context
- **Mutex for state**: Protect shared mutable state with `sync.Mutex`
- **Channels for coordination**: Use channels for goroutine communication
- **WaitGroups**: Use `sync.WaitGroup` for coordinating multiple goroutines

### Common Patterns

```go
// Goroutine with context
go func() {
    for {
        select {
        case <-ctx.Done():
            return // Cancel on context cancellation
        case <-ch:
            // Process work
        }
    }
}()

// Protecting shared state with mutex
type Service struct {
    mu sync.Mutex
    cache map[string]string
}

func (s *Service) Get(key string) string {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.cache[key]
}
```

### Anti-Patterns

- **Data races**: Concurrent access to shared mutable state without synchronization
- **Goroutine leaks**: Goroutines created but never terminated
- **Mutex contention**: Holding locks for too long in critical sections

## Configuration

### Patterns

- **Environment variables**: Use `os.Getenv()` with defaults
- **Config files**: Use well-known locations (`~/.config/app/config.toml`)
- **Flags**: Use flag package for CLI configuration
- **Priority order**: Defaults → config file → environment variables → flags

### Common Patterns

```go
// Load configuration with defaults
type Config struct {
    ProjectsDirectory string `toml:"projects_dir"`
    WorktreesDirectory string `toml:"worktrees_dir"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{
        ProjectsDirectory: "~/projects",  // Default
        WorktreesDirectory: "~/worktrees",
    }

    // Load from file
    if err := koanf.UnmarshalWithConf(toml.Parser(), &cfg); err != nil {
        return nil, err
    }

    // Override with environment
    if env := os.Getenv("PROJECTS_DIR"); env != "" {
        cfg.ProjectsDirectory = env
    }

    return cfg, nil
}
```

### Anti-Patterns

- **Hardcoded paths**: Configuration values embedded in code
- **Multiple config systems**: Mixing file-based, env vars, and flags without clear priority
- **Global config**: Mutable global configuration variable

## Dependency Management

### Go Module Practices

- **Semantic versioning**: Use semantic versions in go.mod (v1.2.3)
- **Minimal dependencies**: Prefer stdlib and well-maintained packages
- **Indirect dependencies**: Keep indirect dependencies minimal
- **Vendor directory**: Avoid vendoring unless necessary

### Common Patterns

```
module github.com/user/project

go 1.21

require (
    github.com/gorilla/mux v1.8.0
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/gorilla/mux v1.8.0 // Direct
    github.com/stretchr/testify v1.8.4 // Indirect
)
```

### Anti-Patterns

- **Unnecessary dependencies**: Adding packages for functionality in stdlib
- **Outdated dependencies**: Using old versions with known vulnerabilities
- **Pinning to master**: Using `master` branch instead of semantic version tags

## Common Anti-Patterns

### Code Smells

- **Long parameter lists**: Functions with >5 parameters indicate need for struct
- **Deeply nested code**: >3 levels of nesting indicates need for extraction
- **God functions**: Functions that do too many things (>50 lines, multiple switch statements)
- **Magic numbers**: Unexplained numeric literals (use constants)
- **Flag parameters**: Boolean parameters that change function behavior (should be separate functions)

### Security Anti-Patterns

- **SQL injection**: String concatenation in database queries
- **Path traversal**: Not validating user-provided file paths
- **Command injection**: Passing user input directly to exec.Command
- **Hardcoded secrets**: API keys, passwords in source code

### Performance Anti-Patterns

- **N+1 in loops**: Querying database inside loop instead of single query with IN clause
- **String concatenation in loops**: Using `+` instead of strings.Builder
- **Unnecessary allocations**: Creating slices/strings in loops instead of pre-allocating

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

### For Test Pattern Audits

- Verify all test files follow `<source>_test.go` naming
- Check for inline mocks that should be in `test/mocks/`
- Identify duplicate mock implementations
- Find files without corresponding tests (if logic exists)

### For Documentation Accuracy Audits

- Compare AGENTS.md struct definitions with actual code
- Check interface method signatures match implementations
- Look for missing field documentation
- Find undocumented types and methods

### For Import Consistency Audits

- Verify import ordering: stdlib (alphabetical), third-party (alphabetical), internal (alphabetical)
- Check for circular dependencies (try `go build ./...`)
- Look for unused imports (run `golangci-lint`)
- Verify layer dependencies are correct
