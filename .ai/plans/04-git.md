# Git Operations & Context Integration Plan (TDD-First)

## Overview

This plan implements git operations while completing the remaining context detection integration points. The approach uses go-git for all supported operations and CLI only for worktree management (which go-git explicitly doesn't support), integrating git operations directly into the context detection system for intelligent completion and validation.

**Context**: Structural foundation (domain, infrastructure, context detection) is 90% complete, but git operations layer is completely missing. This phase builds the entire git operations foundation from scratch while completing the remaining TODO comments and adding comprehensive git operations with deterministic implementation selection.

## Git Strategy Principles

### Deterministic Implementation Selection
- **go-git usage**: Repository operations, branch management, status checks, remote operations, configuration  
- **CLI usage**: Worktree creation/deletion, worktree listing, worktree status (operations not supported by go-git)
- **No fallback logic**: Deterministic implementation selection based on capabilities

### TDD-First Approach
- **Test First**: Write failing tests, then implement minimal code to pass
- **Red-Green-Refactor**: Follow strict TDD cycle for each git operation
- **Minimal Implementation**: Implement only what's needed to pass current tests
- **Interface-Driven**: Define interfaces first, then implement

## Phase Boundaries

### Phase 4 Scope
- Complete context detection TODO comments with git integration
- Git service interfaces and deterministic routing
- GoGit client implementation (80% coverage)
- CLI client implementation for worktree operations (20% coverage)
- Unit and integration testing for git operations

### Deferred to Later Phases
- CLI commands integration (Phase 6)
- Performance optimization (Phase 9)
- E2E testing (Phase 8)
- Advanced git features (future phases)

## Project Structure

Phase 4 additions to existing structure:

```
internal/
├── service/
│   └── git_service.go          # Git service interfaces and routing
├── infrastructure/
│   ├── gogit_client.go         # go-git implementation  
│   ├── cli_client.go           # CLI implementation
│   └── command_executor.go     # Command execution infrastructure
└── domain/
    └── git_types.go             # Shared git types
    └── errors.go                # Git-specific error types
```

### Phase 04 as Critical Foundation

Phase 04 git operations are directly consumed by subsequent phases:
- **Phase 05 (Core Services)**: WorktreeService requires HybridGitClient interface, ProjectService needs repository operations
- **Phase 06 (CLI Commands)**: All commands (list, create, delete, cd) depend on git operations from this phase
- **Phase 07 (Shell Integration)**: Completion suggestions require git-powered branch and worktree discovery
- **Phase 09 (Performance)**: Optimization layer wraps git services implemented in this phase

**Success Criteria**: Phase 04 MUST provide clean, testable interfaces that enable easy dependency injection and mocking for later phases.

## Implementation Steps

### Prerequisites

#### Dependency Management
- Add `github.com/go-git/go-git/v5` to go.mod
- Add any required git utility libraries
- Update import paths in existing files

#### Infrastructure Setup
- Create `internal/domain/git_types.go` for shared git types
- Create `internal/infrastructure/command_executor.go` for CLI operations
- Verify existing context detection can integrate with git operations

### Step 1: Complete Context Detection TODOs with Git Integration

#### 1.1 Resolution Suggestions with Git Discovery
**File:** `internal/infrastructure/context_resolver.go`

**Tests first:** `internal/infrastructure/context_resolver_test.go`

```go
func TestContextResolver_GetProjectContextSuggestions_WithGitIntegration(t *testing.T) {
    // Setup temporary git repository with worktrees
    tempDir := t.TempDir()
    repoPath := setupTestRepo(t, tempDir)
    setupTestWorktrees(t, repoPath, []string{"feature-1", "feature-2"})
    
    resolver := NewContextResolver(config)
    ctx := &domain.Context{ProjectName: "test-project", GitRepoPath: repoPath}
    
    suggestions := resolver.getProjectContextSuggestions(ctx, "feat")
    
    assert.NotEmpty(t, suggestions)
    assert.Contains(t, suggestions, "feature-1")
    assert.Contains(t, suggestions, "feature-2")
    assert.Contains(t, suggestions, "main")
}
```

**Implementation:**
Replace TODO comments at lines 97,145,170 with git-powered discovery:
- Use git operations to scan actual worktrees
- Integrate with GitService for branch discovery
- Provide intelligent suggestions based on git state

#### 1.2 Enhanced Caching with Git Integration
**File:** `internal/infrastructure/context_detector.go`

**Tests first:** `internal/infrastructure/context_detector_test.go`

```go
func TestContextDetector_CacheWithTTL(t *testing.T) {
    detector := NewContextDetector(config)
    
    // First call should cache result
    ctx1, err1 := detector.DetectContext("/test/path")
    assert.NoError(t, err1)
    
    // Second call should use cache
    ctx2, err2 := detector.DetectContext("/test/path")
    assert.NoError(t, err2)
    assert.Equal(t, ctx1, ctx2)
}
```

**Implementation:**
Add TTL-based cache entry with git repository validation:
- Cache git repository state along with path
- Invalidate cache when git repository changes
- Integrate with git operations for validation

### Step 2: Define Git Service Interfaces

#### 2.1 Core Interfaces
**File:** `internal/service/git_service.go`

**Tests first:** `internal/service/git_service_test.go`

```go
func TestGitService_InterfaceDefinitions(t *testing.T) {
    var _ GoGitClient = (*goGitClient)(nil)
    var _ CLIClient = (*cliClient)(nil)
    var _ GitService = (*gitService)(nil)
    
    // Interface compliance tests
}
```

**Implementation:**
```go
// GoGitClient defines go-git operations (deterministic routing - no CLI fallback)
// All methods SHALL be idempotent and thread-safe
type GoGitClient interface {
    // OpenRepository opens git repository (pure function, idempotent)
    OpenRepository(path string) (*git.Repository, error)
    
    // ListBranches lists all branches in repository (idempotent)
    ListBranches(ctx context.Context, repoPath string) ([]BranchInfo, error)
    
    // BranchExists checks if branch exists (idempotent)
    BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)
    
    // GetRepositoryStatus returns repository status (idempotent)
    GetRepositoryStatus(ctx context.Context, repoPath string) (RepositoryStatus, error)
    
    // ValidateRepository checks if path contains valid git repository (pure function)
    ValidateRepository(path string) error
}

// CLIClient defines CLI operations for worktree management ONLY
// All methods SHALL be idempotent and thread-safe
type CLIClient interface {
    // CreateWorktree creates new worktree using git CLI (idempotent)
    CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error
    
    // DeleteWorktree removes worktree using git CLI (idempotent, no-op if already deleted)
    DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error
    
    // ListWorktrees lists all worktrees using git CLI (idempotent)
    ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
}

// GitService provides unified git operations with deterministic routing
// No fallback logic - operations use predetermined implementation
type GitService interface {
    GoGitClient
    CLIClient
}
```

### Step 3: Implement GoGit Client (TDD)

#### 3.1 Repository Operations
**File:** `internal/infrastructure/gogit_client.go`

**Tests first:** `internal/infrastructure/gogit_client_test.go`

```go
func TestGoGitClient_ListBranches(t *testing.T) {
    client := NewGoGitClient()
    
    branches, err := client.ListBranches(context.Background(), "/test/repo")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, branches)
    assert.Contains(t, branches, BranchInfo{Name: "main"})
}

func TestGoGitClient_BranchExists(t *testing.T) {
    client := NewGoGitClient()
    
    exists, err := client.BranchExists(context.Background(), "/test/repo", "main")
    
    assert.NoError(t, err)
    assert.True(t, exists)
}
```

**Implementation:**
Implement repository operations using go-git library with caching.

### Step 4: Implement CLI Client (TDD)

#### 4.1 Worktree Operations
**File:** `internal/infrastructure/cli_client.go`

**Tests first:** `internal/infrastructure/cli_client_test.go`

```go
func TestCLIClient_CreateWorktree(t *testing.T) {
    mockExecutor := &MockCommandExecutor{}
    client := NewCLIClient(mockExecutor)
    
    mockExecutor.On("Execute", mock.Anything, "/test/repo", 
        "worktree", "add", "-b", "feature", "/path/to/worktree", "main").
        Return(&CommandResult{ExitCode: 0}, nil)
    
    err := client.CreateWorktree(context.Background(), "/test/repo", "feature", "main", "/path/to/worktree")
    
    assert.NoError(t, err)
    mockExecutor.AssertExpectations(t)
}
```

**Implementation:**
Implement worktree operations using git CLI with proper error handling.

### Step 5: Integrate Git Service

#### 5.1 Deterministic Routing
**File:** `internal/service/git_service.go`

**Tests first:** `internal/service/git_service_test.go`

```go
func TestGitService_DeterministicRouting(t *testing.T) {
    mockGoGit := &MockGoGitClient{}
    mockCLI := &MockCLIClient{}
    service := NewGitService(mockGoGit, mockCLI)
    
    // Branch operations should use GoGit only
    mockGoGit.On("ListBranches", mock.Anything, "/test/repo").Return([]BranchInfo{}, nil)
    
    branches, err := service.ListBranches(context.Background(), "/test/repo")
    
    assert.NoError(t, err)
    mockGoGit.AssertExpectations(t)
    mockCLI.AssertNotCalled(t, "ListBranches")
}
```

**Implementation:**
Create service with deterministic routing (no fallback logic).

## Testing Strategy

### Unit Tests (Primary)
- **Framework**: Testify suites with table-driven tests
- **Focus**: Interface compliance, error handling, deterministic routing
- **Coverage**: >80% for git operations
- **Mocking**: 
  - Centralized mocks from `test/mocks/` for service layer testing
  - Inline mocks for domain layer testing
  - Command execution for CLI client, repository objects for GoGit client

### Integration Tests (Secondary)
- **Framework**: Testify suites with build tags (`//go:build integration`)
- **Focus**: Real git operations, end-to-end git workflows
- **Setup**: Temporary repositories with real git commands
- **Scope**: Critical paths only (worktree creation, branch listing)
- **Skip in short mode**: Use `testing.Short()` for CI optimization

### E2E Tests Preparation
- **Framework**: Ginkgo/Gomega (prepared for Phase 08)
- **Focus**: CLI command workflows (prepared for Phase 06)
- **Pattern**: Build actual binary and execute commands

### Deferred Testing
- **E2E Tests**: Phase 8 (CLI commands)
- **Performance Tests**: Phase 9

## Quality Gates

### Pre-commit Requirements
- All tests pass: `go test ./...`
- Linting passes: `golangci-lint run`
- Coverage >80%: `go test -cover ./...`
- Integration tests pass: `go test -tags=integration ./...`

### CI Requirements
- Unit tests pass on all platforms
- Integration tests pass with real git
- Build succeeds with go-git and CLI dependencies

## Configuration

```toml
[context_detection]
cache_ttl = "5m"
git_operation_timeout = "30s"
enable_git_validation = true

[git]
cli_timeout = "30s"
cache_enabled = true
```

### Error Handling Requirements

#### Error Types and Wrapping
- All git operations SHALL wrap errors with context using `fmt.Errorf`
- Custom error types SHALL be defined for git-specific failures
- Error messages SHALL include actionable guidance per design.md requirements

#### Error Categories
```go
// Git-specific error types
type GitRepositoryError struct {
    Path    string
    Message string
    Cause   error
}

type GitWorktreeError struct {
    WorktreePath string
    BranchName   string
    Message      string
    Cause        error
}
```

#### Consistent Error Patterns
- Repository operations: Include repository path in error context
- Worktree operations: Include worktree path and branch name
- Validation errors: Include specific validation failure details

## Success Criteria

1. ✅ All context detection TODO comments resolved with git-powered discovery
2. ✅ GoGitClient handles repository and branch operations deterministically (no CLI fallback)
3. ✅ CLIClient handles worktree operations only (go-git limitations)
4. ✅ GitService provides deterministic routing without fallback logic
5. ✅ Unit tests for all git operations pass with >80% coverage using Testify suites
6. ✅ Integration tests pass with real git repositories using build tags
7. ✅ ContextService fully integrated with GitService for intelligent suggestions
8. ✅ Resolution suggestions powered by actual git discovery (not hardcoded)
9. ✅ Error handling follows consistent patterns with context wrapping
10. ✅ Interface design enables easy dependency injection for Phase 05

## Key Principles

### TDD Approach
- **Write failing test first**
- **Implement minimal code to pass**
- **Refactor while keeping tests green**
- **Repeat for next git operation**

### Deterministic Behavior
- **No fallback logic** - predictable implementation selection
- **Clear separation** - go-git vs CLI responsibilities
- **Consistent errors** - same error types across implementations

### Clean Architecture
- **Interface-driven** - define contracts before implementation
- **Dependency injection** - testable and modular
- **Single responsibility** - each client has clear scope

## Next Phases

Phase 4 provides the critical git operations foundation:

1. **Phase 5**: Service layer and dependency injection
   - WorktreeService directly uses HybridGitClient interface
   - ProjectService depends on repository operations from Phase 4
   
2. **Phase 6**: CLI commands implementation
   - All commands (list, create, delete, cd) use git operations from Phase 4
   - Context-aware behavior powered by Phase 4 integration
   
3. **Phase 7**: Shell integration features
   - Completion suggestions use git-powered discovery from Phase 4
   - Target resolution depends on git repository analysis
   
4. **Phase 8**: Comprehensive testing infrastructure
   - Interface design enables comprehensive mocking and testing
   - Test patterns established in Phase 4 extend to E2E testing
   
5. **Phase 9**: Performance optimization
   - Caching layer wraps git services implemented in Phase 4
   - Monitoring hooks built into Phase 4 interface design

This approach delivers robust git operations while maintaining clean TDD principles and architectural boundaries, providing the critical foundation for all subsequent implementation phases.