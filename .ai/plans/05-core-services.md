# Core Services Layer Implementation Plan

## Overview

This plan establishes the core services layer that orchestrates domain logic and provides the main business functionality for twiggit. The services layer coordinates between domain entities, infrastructure components, and CLI commands while maintaining clean separation of concerns.

**Context**: Foundation, configuration, context detection, and hybrid git layers are established. This layer provides the orchestration logic that ties everything together.

## Foundation Principles

### TDD Approach
- **Test First**: Write failing tests, then implement minimal code to pass
- **Red-Green-Refactor**: Follow strict TDD cycle for each service
- **Minimal Implementation**: Implement only what's needed to pass current tests
- **Interface Contracts**: Test service interfaces before implementation

### Functional Programming Principles
- **Pure Functions**: Service operations SHALL be pure functions without side effects
- **Immutability**: Request/response structures SHALL be immutable
- **Function Composition**: Complex operations SHALL be composed from smaller functions
- **Error Handling**: SHALL use Result/Either patterns for error handling

### Clean Architecture
- **Thin Orchestration**: Services coordinate, don't contain business logic
- **Dependency Injection**: All external dependencies SHALL be injected via interfaces
- **Context-Aware**: Behavior SHALL adapt based on ContextDetector results
- **Interface Segregation**: Services SHALL have focused, single-purpose interfaces

## Phase Boundaries

### Phase 5 Scope
- Service interfaces with method signatures
- Basic service implementations with dependency injection
- Unit testing for service contracts
- Quality assurance configuration
- Functional programming patterns

### Deferred to Later Phases
- CLI commands implementation (Phase 6)
- Performance optimization and caching (Phase 9)
- Advanced error recovery patterns (Phase 9)
- Integration/E2E testing (Phase 8)

## Project Structure

Phase 5 minimal structure following Go standards:

```
internal/
├── domain/
│   ├── service_requests.go     # Request/response types
│   ├── service_errors.go       # Service-specific errors
│   └── service_results.go      # Result/Either patterns
├── services/
│   ├── interfaces.go           # Service interfaces
│   ├── worktree_service.go     # WorktreeService implementation
│   ├── project_service.go      # ProjectService implementation
│   └── navigation_service.go   # NavigationService implementation
```

**Removed from Phase 5** (deferred to later phases):
- Advanced caching layers → Phase 9
- Performance monitoring → Phase 9
- Complex orchestration patterns → Phase 6
- Integration test fixtures → Phase 8

## Implementation Steps

### Step 1: Service Interfaces and Types

**Files to create:**
- `internal/domain/service_requests.go`
- `internal/domain/service_errors.go`
- `internal/services/interfaces.go`

**Tests first:** `internal/services/interfaces_test.go`

```go
func TestServiceInterfaces_ContractCompliance(t *testing.T) {
    testCases := []struct {
        name        string
        serviceType interface{}
        expectError bool
    }{
        {
            name:        "WorktreeService interface compliance",
            serviceType: (*WorktreeService)(nil),
            expectError: false,
        },
        {
            name:        "ProjectService interface compliance",
            serviceType: (*ProjectService)(nil),
            expectError: false,
        },
        {
            name:        "NavigationService interface compliance",
            serviceType: (*NavigationService)(nil),
            expectError: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Interface compliance tests
            assert.NotNil(t, tc.serviceType)
        })
    }
}
```

**Interface definitions:**
```go
// internal/services/interfaces.go
type WorktreeService interface {
    CreateWorktree(ctx context.Context, req *CreateWorktreeRequest) (*WorktreeInfo, error)
    DeleteWorktree(ctx context.Context, req *DeleteWorktreeRequest) error
    ListWorktrees(ctx context.Context, req *ListWorktreesRequest) ([]*WorktreeInfo, error)
    GetWorktreeStatus(ctx context.Context, worktreePath string) (*WorktreeStatus, error)
    ValidateWorktree(ctx context.Context, worktreePath string) error
}

type ProjectService interface {
    DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*ProjectInfo, error)
    ValidateProject(ctx context.Context, projectPath string) error
    ListProjects(ctx context.Context) ([]*ProjectInfo, error)
    GetProjectInfo(ctx context.Context, projectPath string) (*ProjectInfo, error)
}

type NavigationService interface {
    ResolvePath(ctx context.Context, req *ResolvePathRequest) (*domain.ResolutionResult, error)
    ValidatePath(ctx context.Context, path string) error
    GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error)
}
```

### Step 2: WorktreeService Implementation

**Tests first:** `internal/services/worktree_service_test.go`

```go
func TestWorktreeService_CreateWorktree_Success(t *testing.T) {
    testCases := []struct {
        name         string
        request      *CreateWorktreeRequest
        expectError  bool
        errorMessage string
    }{
        {
            name: "valid worktree creation",
            request: &CreateWorktreeRequest{
                ProjectName:  "test-project",
                BranchName:   "feature-branch",
                SourceBranch: "main",
                Context: &domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "test-project",
                },
            },
            expectError: false,
        },
        {
            name: "empty branch name",
            request: &CreateWorktreeRequest{
                ProjectName:  "test-project",
                BranchName:   "",
                SourceBranch: "main",
                Context: &domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "test-project",
                },
            },
            expectError:  true,
            errorMessage: "branch name cannot be empty",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := setupTestWorktreeService()
            result, err := service.CreateWorktree(context.Background(), tc.request)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                assert.Equal(t, tc.request.BranchName, result.Branch)
            }
        })
    }
}
```

**Implementation sketch:**
```go
// internal/services/worktree_service.go
func NewWorktreeService(
    gitClient infrastructure.GitClient,
    projectService ProjectService,
    config *domain.Config,
) WorktreeService {
    return &worktreeService{
        gitClient:      gitClient,
        projectService: projectService,
        config:         config,
    }
}

func (s *worktreeService) CreateWorktree(ctx context.Context, req *CreateWorktreeRequest) (*WorktreeInfo, error) {
    // Validate request
    if err := s.validateCreateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Resolve project
    project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve project: %w", err)
    }
    
    // Create worktree
    worktreePath := s.calculateWorktreePath(project.Name, req.BranchName)
    err = s.gitClient.CreateWorktree(ctx, project.GitRepoPath, req.BranchName, req.SourceBranch, worktreePath)
    if err != nil {
        return nil, fmt.Errorf("failed to create worktree: %w", err)
    }
    
    return &WorktreeInfo{
        Path:   worktreePath,
        Branch: req.BranchName,
    }, nil
}
```

### Step 3: ProjectService Implementation

**Tests first:** `internal/services/project_service_test.go`

```go
func TestProjectService_DiscoverProject_Success(t *testing.T) {
    testCases := []struct {
        name         string
        projectName  string
        context      *domain.Context
        expectError  bool
        errorMessage string
    }{
        {
            name:        "valid project discovery",
            projectName: "test-project",
            context: &domain.Context{
                Type: domain.ContextOutsideGit,
            },
            expectError: false,
        },
        {
            name:        "empty project name outside context",
            projectName: "",
            context: &domain.Context{
                Type: domain.ContextOutsideGit,
            },
            expectError:  true,
            errorMessage: "project name required when outside git context",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := setupTestProjectService()
            result, err := service.DiscoverProject(context.Background(), tc.projectName, tc.context)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                assert.Equal(t, tc.projectName, result.Name)
            }
        })
    }
}
```

**Implementation sketch:**
```go
// internal/services/project_service.go
func NewProjectService(
    gitClient infrastructure.GitClient,
    contextService *domain.ContextService,
    config *domain.Config,
) ProjectService {
    return &projectService{
        gitClient:      gitClient,
        contextService: contextService,
        config:         config,
    }
}

func (s *projectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*ProjectInfo, error) {
    if projectName != "" {
        return s.discoverProjectByName(ctx, projectName)
    }
    
    if context != nil {
        return s.discoverProjectFromContext(ctx, context)
    }
    
    return nil, fmt.Errorf("project name required when outside git context")
}
```

### Step 4: NavigationService Implementation

**Tests first:** `internal/services/navigation_service_test.go`

```go
func TestNavigationService_ResolvePath_Success(t *testing.T) {
    testCases := []struct {
        name         string
        request      *ResolvePathRequest
        expectError  bool
        errorMessage string
    }{
        {
            name: "valid branch resolution from project context",
            request: &ResolvePathRequest{
                Target: "feature-branch",
                Context: &domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "test-project",
                },
            },
            expectError: false,
        },
        {
            name: "empty target",
            request: &ResolvePathRequest{
                Target: "",
                Context: &domain.Context{
                    Type: domain.ContextOutsideGit,
                },
            },
            expectError:  true,
            errorMessage: "target cannot be empty",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := setupTestNavigationService()
            result, err := service.ResolvePath(context.Background(), tc.request)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

**Implementation sketch:**
```go
// internal/services/navigation_service.go
func NewNavigationService(
    projectService ProjectService,
    contextService *domain.ContextService,
    config *domain.Config,
) NavigationService {
    return &navigationService{
        projectService: projectService,
        contextService: contextService,
        config:         config,
    }
}

func (s *navigationService) ResolvePath(ctx context.Context, req *ResolvePathRequest) (*domain.ResolutionResult, error) {
    if req.Target == "" {
        return nil, fmt.Errorf("target cannot be empty")
    }
    
    // Delegate to ContextResolver for consistency
    return s.contextService.ResolveIdentifierFromContext(req.Context, req.Target)
}
```

## Testing Strategy

Phase 5 focuses exclusively on unit testing for service contracts.

### Unit Tests Only
- **Framework**: Testify with table-driven tests
- **Coverage**: >80% for service logic (realistic for orchestration layer)
- **Location**: `*_test.go` files alongside implementation
- **Focus**: Interface contracts, error handling, context integration

### Test Organization
- **Interface Tests**: Test all service interfaces before implementation
- **Contract Tests**: Test service behavior with various inputs
- **Error Path Tests**: Test all error scenarios and edge cases
- **Context Tests**: Test context-aware behavior

### Deferred Testing Types
- **Integration Tests**: Phase 8 (when real coordination exists)
- **E2E Tests**: Phase 8 (when CLI commands exist)
- **Performance Tests**: Phase 9

## Quality Gates

### Pre-commit Requirements
- All tests pass: `go test ./...`
- Linting passes: `golangci-lint run`
- Coverage >80%: `go test -cover ./...`
- Interface compliance tests pass

### CI Requirements
- Unit tests pass
- Linting passes
- Build succeeds on target platforms
- Functional programming principles verified

## Key Principles

### TDD Approach
- **Write failing test first**
- **Implement minimal code to pass**
- **Refactor while keeping tests green**
- **Repeat for next service**

### Functional Programming
- **Pure functions**: No side effects in service operations
- **Immutability**: Immutable request/response structures
- **Composition**: Build complex operations from simple functions
- **Error handling**: Use Result patterns for predictable error flow

### Clean Code
- **Interface segregation**: Small, focused interfaces
- **Dependency injection**: All dependencies injected
- **Single responsibility**: Each service has one clear purpose
- **Consistent error handling**: Same error pattern throughout

## Configuration Integration

### Service Configuration Extensions

Since Phase 02 (Configuration) is already implemented, the following service-specific configuration SHALL be added in Phase 05:

```toml
# config.toml additions for services
[services]
cache_enabled = true
cache_ttl = "5m"
concurrent_operations = true
max_concurrent = 4

[services.validation]
strict_branch_names = true
require_clean_worktree = true
allow_force_delete = false

[services.navigation]
enable_suggestions = true
max_suggestions = 10
fuzzy_matching = false
```

### Configuration Implementation
**File:** `internal/domain/config.go` (extend existing)

```go
// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
    CacheEnabled         bool          `koanf:"cache_enabled"`
    CacheTTL            time.Duration `koanf:"cache_ttl"`
    ConcurrentOps       bool          `koanf:"concurrent_operations"`
    MaxConcurrent       int           `koanf:"max_concurrent"`
}

type ValidationConfig struct {
    StrictBranchNames    bool `koanf:"strict_branch_names"`
    RequireCleanWorktree bool `koanf:"require_clean_worktree"`
    AllowForceDelete     bool `koanf:"allow_force_delete"`
}

type NavigationConfig struct {
    EnableSuggestions bool `koanf:"enable_suggestions"`
    MaxSuggestions    int  `koanf:"max_suggestions"`
    FuzzyMatching     bool `koanf:"fuzzy_matching"`
}
```

## Success Criteria

1. ✅ Service interfaces (WorktreeService, ProjectService, NavigationService) with comprehensive contracts
2. ✅ Service implementations with dependency injection and functional patterns
3. ✅ Service configuration extensions integrated with existing config system
4. ✅ Unit tests for service contracts pass with >80% coverage
5. ✅ Basic linting passes without errors
6. ✅ Clean service structure following Go standards
7. ✅ Quality gates enforce functional programming principles

## Incremental Development Strategy

Phase 5 follows strict incremental development:

1. **Write Test**: Create failing test for service interface
2. **Define Interface**: Add interface with method signatures
3. **Implement**: Add minimal code to make test pass
4. **Refactor**: Apply functional programming patterns while keeping tests green
5. **Repeat**: Move to next service

**No detailed implementation, no premature optimization, no future-proofing.** Each service builds only what's needed for that phase.

## Next Phases

Phase 5 provides the service orchestration needed for sequential development:

1. **Phase 6**: CLI commands implementation using service interfaces
2. **Phase 7**: Shell integration features
3. **Phase 8**: Comprehensive testing infrastructure
4. **Phase 9**: Performance optimization and advanced caching
5. **Phase 10**: Final integration and validation

This services layer provides the essential orchestration needed for building CLI features while following true TDD principles, functional programming patterns, and maintaining clean phase boundaries.
