# CLI Commands Layer Implementation Plan

## Overview

This plan implements the user-facing CLI commands layer using the Cobra framework with strict TDD principles and functional programming patterns. The commands layer provides pure function wrappers around existing services while maintaining context awareness and consistent user experience.

**Context**: Foundation, configuration, context detection, hybrid git, and core services layers are established. This layer provides the thin orchestration that connects user input to service operations.

## Foundation Principles

### TDD Approach
- **Test First**: Write failing tests, then implement minimal code to pass
- **Red-Green-Refactor**: Follow strict TDD cycle for each command
- **Minimal Implementation**: Implement only what's needed to pass current tests
- **Command Contracts**: Test command interfaces before implementation

### Functional Programming Principles
- **Pure Functions**: Command operations SHALL be pure functions without side effects
- **Immutability**: Command configuration SHALL be immutable
- **Function Composition**: Complex commands SHALL be composed from smaller functions
- **Error Handling**: SHALL use explicit error returns for predictable flow

### Clean Architecture
- **Thin Wrappers**: Commands delegate to services, don't contain business logic
- **Dependency Injection**: All services SHALL be injected via interfaces
- **Context-Aware**: Behavior SHALL adapt based on detected context
- **Command Segregation**: Each command SHALL have a single, focused responsibility

## Phase Boundaries

### Phase 6 Scope
- Command interfaces with Cobra integration
- Basic command implementations with service injection
- Unit testing for command contracts
- Quality assurance configuration
- Functional programming patterns

### Deferred to Later Phases
- Shell integration commands (Phase 7)
- Performance optimization and caching (Phase 9)
- Advanced error recovery patterns (Phase 9)
- Integration/E2E testing (Phase 8)

## Project Structure

Phase 6 minimal structure following Go standards:

```
cmd/
├── root.go              # Root command and global configuration
├── list.go              # list command implementation
├── create.go            # create command implementation
├── delete.go            # delete command implementation
├── cd.go                # cd command implementation
└── command_test.go      # Command contract tests
```

**Removed from Phase 6** (deferred to later phases):
- Shell integration commands → Phase 7
- Auto-completion with Carapace → Phase 7
- Advanced help customization → Phase 9
- Integration test fixtures → Phase 8

## Implementation Steps

### Step 1: Command Interfaces and Contracts

**Files to create:**
- `cmd/root.go` (extend existing)
- `cmd/command_test.go`

**Tests first:** `cmd/command_test.go`

```go
func TestCommandInterfaces_ContractCompliance(t *testing.T) {
    testCases := []struct {
        name        string
        command     *cobra.Command
        expectError bool
        setupFunc   func() *cobra.Command
    }{
        {
            name:      "list command interface compliance",
            setupFunc: setupListCommand,
            expectError: false,
        },
        {
            name:      "create command interface compliance", 
            setupFunc: setupCreateCommand,
            expectError: false,
        },
        {
            name:      "delete command interface compliance",
            setupFunc: setupDeleteCommand,
            expectError: false,
        },
        {
            name:      "cd command interface compliance",
            setupFunc: setupCDCommand,
            expectError: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := tc.setupFunc()
            assert.NotNil(t, cmd)
            assert.NotEmpty(t, cmd.Use)
            assert.NotEmpty(t, cmd.Short)
            assert.NotNil(t, cmd.RunE)
        })
    }
}
```

**Interface definitions:**
```go
// cmd/root.go (extend existing)
type CommandConfig struct {
    Services *ServiceContainer
    Config   *domain.Config
}

type ServiceContainer struct {
    WorktreeService  services.WorktreeService
    ProjectService   services.ProjectService
    NavigationService services.NavigationService
    ContextService   services.ContextService
}

func NewRootCommand(config *CommandConfig) *cobra.Command
func initializeContext(cmd *cobra.Command, config *CommandConfig) error
```

### Step 2: List Command Implementation

**Tests first:** `cmd/list_test.go`

```go
func TestListCommand_Execute(t *testing.T) {
    testCases := []struct {
        name         string
        args         []string
        flags        map[string]string
        setupContext func(*services.MockWorktreeService, *services.MockContextService)
        expectError  bool
        errorMessage string
        validateOut  func(string) bool
    }{
        {
            name: "list worktrees in project context",
            args: []string{},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "test-project",
                }, nil)
                mockWS.EXPECT().ListWorktrees(gomock.Any(), &services.ListWorktreesRequest{
                    ProjectName: "test-project",
                    All:         false,
                }).Return([]*domain.WorktreeInfo{
                    {Path: "/home/user/Worktrees/test-project/main", Branch: "main"},
                    {Path: "/home/user/Worktrees/test-project/feature", Branch: "feature"},
                }, nil)
            },
            expectError: false,
            validateOut: func(output string) bool {
                return strings.Contains(output, "main") && strings.Contains(output, "feature")
            },
        },
        {
            name: "list all worktrees with --all flag",
            args: []string{"--all"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type: domain.ContextOutsideGit,
                }, nil)
                mockWS.EXPECT().ListWorktrees(gomock.Any(), &services.ListWorktreesRequest{
                    All: true,
                }).Return([]*domain.WorktreeInfo{}, nil)
            },
            expectError: false,
        },
        {
            name: "context detection failure",
            args: []string{},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(nil, fmt.Errorf("detection failed"))
            },
            expectError:  true,
            errorMessage: "context detection failed",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Setup mocks and execute command
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockWS := services.NewMockWorktreeService(ctrl)
            mockCS := services.NewMockContextService(ctrl)
            
            tc.setupContext(mockWS, mockCS)
            
            config := &CommandConfig{
                Services: &ServiceContainer{
                    WorktreeService: mockWS,
                    ContextService:  mockCS,
                },
            }
            
            cmd := NewListCommand(config)
            cmd.SetArgs(tc.args)
            
            var buf bytes.Buffer
            cmd.SetOut(&buf)
            err := cmd.Execute()
            
            // Validate results
            if tc.expectError {
                assert.Error(t, err)
                if tc.errorMessage != "" {
                    assert.Contains(t, err.Error(), tc.errorMessage)
                }
            } else {
                assert.NoError(t, err)
                if tc.validateOut != nil {
                    assert.True(t, tc.validateOut(buf.String()))
                }
            }
        })
    }
}
```

**Implementation sketch:**
```go
// cmd/list.go
func NewListCommand(config *CommandConfig) *cobra.Command
func executeList(config *CommandConfig, all bool) error
func displayWorktrees(worktrees []*domain.WorktreeInfo)
```

### Step 3: Create Command Implementation

**Tests first:** `cmd/create_test.go`

```go
func TestCreateCommand_Execute(t *testing.T) {
    testCases := []struct {
        name         string
        args         []string
        flags        map[string]string
        setupContext func(*services.MockWorktreeService, *services.MockContextService, *services.MockProjectService)
        expectError  bool
        errorMessage string
        validateOut  func(string) bool
    }{
        {
            name: "create worktree with project/branch",
            args: []string{"test-project/feature-branch"},
            flags: map[string]string{"source": "main"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService, mockPS *services.MockProjectService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type: domain.ContextOutsideGit,
                }, nil)
                mockPS.EXPECT().DiscoverProject(gomock.Any(), "test-project", gomock.Any()).Return(&domain.ProjectInfo{
                    Name:        "test-project",
                    GitRepoPath: "/home/user/Projects/test-project",
                }, nil)
                mockWS.EXPECT().CreateWorktree(gomock.Any(), &services.CreateWorktreeRequest{
                    ProjectName:  "test-project",
                    BranchName:   "feature-branch",
                    SourceBranch: "main",
                    Context:      gomock.Any(),
                }).Return(&domain.WorktreeInfo{
                    Path:   "/home/user/Worktrees/test-project/feature-branch",
                    Branch: "feature-branch",
                }, nil)
            },
            expectError: false,
            validateOut: func(output string) bool {
                return strings.Contains(output, "Created worktree")
            },
        },
        {
            name: "infer project from context",
            args: []string{"feature-branch"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService, mockPS *services.MockProjectService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "current-project",
                }, nil)
                mockPS.EXPECT().DiscoverProject(gomock.Any(), "current-project", gomock.Any()).Return(&domain.ProjectInfo{
                    Name:        "current-project",
                    GitRepoPath: "/home/user/Projects/current-project",
                }, nil)
                mockWS.EXPECT().CreateWorktree(gomock.Any(), &services.CreateWorktreeRequest{
                    ProjectName:  "current-project",
                    BranchName:   "feature-branch",
                    SourceBranch: "main", // default
                    Context:      gomock.Any(),
                }).Return(&domain.WorktreeInfo{}, nil)
            },
            expectError: false,
        },
        {
            name: "invalid project/branch format",
            args: []string{"invalid-format"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService, mockPS *services.MockProjectService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type: domain.ContextOutsideGit,
                }, nil)
            },
            expectError:  true,
            errorMessage: "invalid format: expected <project>/<branch>",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Setup mocks and execute command
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockWS := services.NewMockWorktreeService(ctrl)
            mockCS := services.NewMockContextService(ctrl)
            mockPS := services.NewMockProjectService(ctrl)
            
            tc.setupContext(mockWS, mockCS, mockPS)
            
            config := &CommandConfig{
                Services: &ServiceContainer{
                    WorktreeService: mockWS,
                    ContextService:  mockCS,
                    ProjectService:  mockPS,
                },
            }
            
            cmd := NewCreateCommand(config)
            cmd.SetArgs(tc.args)
            
            // Set flags
            for flag, value := range tc.flags {
                cmd.Flags().Set(flag, value)
            }
            
            var buf bytes.Buffer
            cmd.SetOut(&buf)
            err := cmd.Execute()
            
            // Validate results
            if tc.expectError {
                assert.Error(t, err)
                if tc.errorMessage != "" {
                    assert.Contains(t, err.Error(), tc.errorMessage)
                }
            } else {
                assert.NoError(t, err)
                if tc.validateOut != nil {
                    assert.True(t, tc.validateOut(buf.String()))
                }
            }
        })
    }
}
```

**Implementation sketch:**
```go
// cmd/create.go
func NewCreateCommand(config *CommandConfig) *cobra.Command
func executeCreate(config *CommandConfig, spec, source, cdFlag string) error
func parseProjectBranch(spec string, ctx *domain.Context) (string, string, error)
```

### Step 4: Delete Command Implementation

**Tests first:** `cmd/delete_test.go`

```go
func TestDeleteCommand_Execute(t *testing.T) {
    testCases := []struct {
        name         string
        args         []string
        flags        map[string]string
        setupContext func(*services.MockWorktreeService, *services.MockContextService)
        expectError  bool
        errorMessage string
    }{
        {
            name: "delete worktree with safety checks",
            args: []string{"test-project/feature-branch"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{}, nil)
                mockCS.EXPECT().ResolveIdentifier(gomock.Any(), "test-project/feature-branch").Return(&domain.Resolution{
                    WorktreePath: "/home/user/Worktrees/test-project/feature-branch",
                }, nil)
                mockWS.EXPECT().GetWorktreeStatus(gomock.Any(), "/home/user/Worktrees/test-project/feature-branch").Return(&domain.WorktreeStatus{
                    Clean:   true,
                    Current: false,
                }, nil)
                mockWS.EXPECT().DeleteWorktree(gomock.Any(), &services.DeleteWorktreeRequest{
                    WorktreePath: "/home/user/Worktrees/test-project/feature-branch",
                    KeepBranch:   false,
                    Force:        false,
                }).Return(nil)
            },
            expectError: false,
        },
        {
            name: "abort on dirty worktree",
            args: []string{"test-project/feature-branch"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{}, nil)
                mockCS.EXPECT().ResolveIdentifier(gomock.Any(), "test-project/feature-branch").Return(&domain.Resolution{}, nil)
                mockWS.EXPECT().GetWorktreeStatus(gomock.Any(), gomock.Any()).Return(&domain.WorktreeStatus{
                    Clean: false,
                }, nil)
            },
            expectError:  true,
            errorMessage: "worktree has uncommitted changes",
        },
        {
            name: "force delete dirty worktree",
            args: []string{"--force", "test-project/feature-branch"},
            setupContext: func(mockWS *services.MockWorktreeService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{}, nil)
                mockCS.EXPECT().ResolveIdentifier(gomock.Any(), "test-project/feature-branch").Return(&domain.Resolution{}, nil)
                mockWS.EXPECT().DeleteWorktree(gomock.Any(), &services.DeleteWorktreeRequest{
                    WorktreePath: gomock.Any(),
                    KeepBranch:   false,
                    Force:        true,
                }).Return(nil)
            },
            expectError: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Setup mocks and execute command
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockWS := services.NewMockWorktreeService(ctrl)
            mockCS := services.NewMockContextService(ctrl)
            
            tc.setupContext(mockWS, mockCS)
            
            config := &CommandConfig{
                Services: &ServiceContainer{
                    WorktreeService: mockWS,
                    ContextService:  mockCS,
                },
            }
            
            cmd := NewDeleteCommand(config)
            cmd.SetArgs(tc.args)
            
            // Set flags
            for flag, value := range tc.flags {
                cmd.Flags().Set(flag, value)
            }
            
            err := cmd.Execute()
            
            // Validate results
            if tc.expectError {
                assert.Error(t, err)
                if tc.errorMessage != "" {
                    assert.Contains(t, err.Error(), tc.errorMessage)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Implementation sketch:**
```go
// cmd/delete.go
func NewDeleteCommand(config *CommandConfig) *cobra.Command
func executeDelete(config *CommandConfig, target string, force, keepBranch bool) error
```

### Step 5: CD Command Implementation

**Tests first:** `cmd/cd_test.go`

```go
func TestCDCommand_Execute(t *testing.T) {
    testCases := []struct {
        name         string
        args         []string
        setupContext func(*services.MockNavigationService, *services.MockContextService)
        expectError  bool
        errorMessage string
        expectedPath string
    }{
        {
            name: "cd to worktree with branch name",
            args: []string{"feature-branch"},
            setupContext: func(mockNS *services.MockNavigationService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "test-project",
                }, nil)
                mockNS.EXPECT().ResolvePath(gomock.Any(), &services.ResolvePathRequest{
                    Target:  "feature-branch",
                    Context: gomock.Any(),
                }).Return(&domain.ResolutionResult{
                    ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
                }, nil)
            },
            expectError:  false,
            expectedPath: "/home/user/Worktrees/test-project/feature-branch",
        },
        {
            name: "cd to default worktree",
            args: []string{},
            setupContext: func(mockNS *services.MockNavigationService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type:            domain.ContextWorktree,
                    ProjectName:     "test-project",
                    DefaultWorktree: "main",
                }, nil)
                mockNS.EXPECT().ResolvePath(gomock.Any(), &services.ResolvePathRequest{
                    Target:  "main",
                    Context: gomock.Any(),
                }).Return(&domain.ResolutionResult{
                    ResolvedPath: "/home/user/Worktrees/test-project/main",
                }, nil)
            },
            expectError:  false,
            expectedPath: "/home/user/Worktrees/test-project/main",
        },
        {
            name: "no target and no default",
            args: []string{},
            setupContext: func(mockNS *services.MockNavigationService, mockCS *services.MockContextService) {
                mockCS.EXPECT().DetectContext().Return(&domain.Context{
                    Type: domain.ContextOutsideGit,
                }, nil)
            },
            expectError:  true,
            errorMessage: "no target specified and no default worktree in context",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Setup mocks and execute command
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockNS := services.NewMockNavigationService(ctrl)
            mockCS := services.NewMockContextService(ctrl)
            
            tc.setupContext(mockNS, mockCS)
            
            config := &CommandConfig{
                Services: &ServiceContainer{
                    NavigationService: mockNS,
                    ContextService:    mockCS,
                },
            }
            
            cmd := NewCDCommand(config)
            cmd.SetArgs(tc.args)
            
            var buf bytes.Buffer
            cmd.SetOut(&buf)
            err := cmd.Execute()
            
            // Validate results
            if tc.expectError {
                assert.Error(t, err)
                if tc.errorMessage != "" {
                    assert.Contains(t, err.Error(), tc.errorMessage)
                }
            } else {
                assert.NoError(t, err)
                if tc.expectedPath != "" {
                    assert.Equal(t, tc.expectedPath, strings.TrimSpace(buf.String()))
                }
            }
        })
    }
}
```

**Implementation sketch:**
```go
// cmd/cd.go
func NewCDCommand(config *CommandConfig) *cobra.Command
func executeCD(config *CommandConfig, target string) error
```

## Testing Strategy

Phase 6 focuses exclusively on unit testing for command contracts.

### Unit Tests Only
- **Framework**: Testify with table-driven tests and gomock for service mocking
- **Coverage**: >80% for command logic (realistic for thin wrapper layer)
- **Location**: `*_test.go` files alongside implementation
- **Focus**: Command contracts, flag parsing, error handling, service integration

### Test Organization
- **Interface Tests**: Test all command interfaces before implementation
- **Contract Tests**: Test command behavior with various inputs and contexts
- **Error Path Tests**: Test all error scenarios and service failure modes
- **Context Tests**: Test context-aware behavior and parameter inference

### Deferred Testing Types
- **Integration Tests**: Phase 8 (when real CLI execution exists)
- **E2E Tests**: Phase 8 (when binary is built)
- **Shell Integration Tests**: Phase 7

## Quality Gates

### Pre-commit Requirements
- All tests pass: `go test ./cmd/...`
- Linting passes: `golangci-lint run ./cmd/`
- Coverage >80%: `go test -cover ./cmd/...`
- Command contract tests pass

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
- **Repeat for next command**

### Functional Programming
- **Pure functions**: No side effects in command operations
- **Immutability**: Immutable command configuration
- **Composition**: Build complex commands from simple functions
- **Error handling**: Use explicit error returns for predictable flow

### Clean Code
- **Command segregation**: Each command has one clear purpose
- **Dependency injection**: All services injected via interfaces
- **Single responsibility**: Commands coordinate, don't contain logic
- **Consistent error handling**: Same error pattern throughout

## Service Integration Patterns

### Dependency Injection Pattern
```go
type CommandConfig struct {
    Services *ServiceContainer
    Config   *domain.Config
}

func NewCommand(config *CommandConfig) *cobra.Command
```

### Context Detection Pattern
```go
func executeWithContext(config *CommandConfig, operation func(*domain.Context) error) error
```

### Error Handling Pattern
```go
func executeCommand(config *CommandConfig, req interface{}) error
```

## Success Criteria

1. ✅ Command interfaces (list, create, delete, cd) with comprehensive contracts
2. ✅ Command implementations with service injection and functional patterns
3. ✅ Context-aware parameter inference following design specifications
4. ✅ Unit tests for command contracts pass with >80% coverage
5. ✅ Basic linting passes without errors
6. ✅ Clean command structure following Go and Cobra standards
7. ✅ Quality gates enforce functional programming principles

## Incremental Development Strategy

Phase 6 follows strict incremental development:

1. **Write Test**: Create failing test for command interface
2. **Define Command**: Add Cobra command with proper structure
3. **Implement**: Add minimal code to make test pass
4. **Refactor**: Apply functional programming patterns while keeping tests green
5. **Repeat**: Move to next command

**No detailed implementation, no premature optimization, no future-proofing.** Each command builds only what's needed for that phase.

## Next Phases

Phase 6 provides the CLI foundation needed for sequential development:

1. **Phase 7**: Shell integration commands and completion
2. **Phase 8**: Comprehensive testing infrastructure with E2E tests
3. **Phase 9**: Performance optimization and advanced features
4. **Phase 10**: Final integration and validation

This CLI layer provides the essential user interface needed for interacting with service operations while following true TDD principles, functional programming patterns, and maintaining clean phase boundaries.