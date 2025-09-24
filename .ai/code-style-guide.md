# Twiggit Code Style Guide

## Go Conventions

### Project Structure

Follow standard Go project structure with domain-driven design:

```
cmd/                    # CLI entry points
├── create.go          # Create command implementation
├── delete.go          # Delete command implementation
├── list.go            # List command implementation
├── switch.go          # Switch command implementation
└── root.go            # Root command and CLI setup

internal/              # Private application code
├── di/                # Dependency injection
├── domain/            # Business logic and entities
├── infrastructure/    # External integrations (git, config, mise)
└── services/          # Application services and use cases

test/                  # Test files
├── e2e/               # End-to-end tests
├── fixtures/          # Test data and fixtures
├── helpers/           # Test utilities
├── integration/       # Integration tests
└── mocks/             # Mock implementations
```

### File Organization

**Good:**
```go
// internal/domain/project.go
package domain

import "fmt"

// Project represents a git worktree project
type Project struct {
    name     string
    path     string
    worktree *Worktree
}

// NewProject creates a new project instance
func NewProject(name, path string) (*Project, error) {
    if name == "" {
        return nil, fmt.Errorf("project name cannot be empty")
    }
    
    return &Project{
        name: name,
        path: path,
    }, nil
}
```

**Bad:**
```go
// internal/domain/project.go
package domain

// Bad: Mixed responsibilities
type Project struct {
    Name     string
    Path     string
    GitRepo  *git.Repository // Should be in infrastructure layer
    Config   *Config         // Should be injected, not embedded
}
```

## Error Handling Patterns

### Good Error Handling

```go
// Good: Wrap errors with context
func (s *WorktreeService) CreateWorktree(path string, ref string) error {
    if err := validatePath(path); err != nil {
        return fmt.Errorf("failed to validate path: %w", err)
    }
    
    worktree, err := s.gitClient.CreateWorktree(path, ref)
    if err != nil {
        return fmt.Errorf("failed to create worktree: %w", err)
    }
    
    s.worktrees = append(s.worktrees, worktree)
    return nil
}

// Good: Custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}

func validateProjectName(name string) error {
    if name == "" {
        return &ValidationError{
            Field:   "name",
            Message: "project name cannot be empty",
        }
    }
    return nil
}
```

### Bad Error Handling

```go
// Bad: Naked returns
func (s *WorktreeService) CreateWorktree(path string, ref string) error {
    if err := validatePath(path); err != nil {
        return err // No context
    }
    
    worktree, err := s.gitClient.CreateWorktree(path, ref)
    if err != nil {
        return err // No context
    }
    
    return nil
}

// Bad: Panics in production code
func GetProject(name string) *Project {
    if name == "" {
        panic("project name cannot be empty") // Never panic in production
    }
    return &Project{name: name}
}
```

## Interface Design

### Good Interface Design

```go
// Good: Clear, focused interfaces
package infrastructure

import "github.com/go-git/go-git/v5/plumbing"

// GitClient defines the contract for git operations
type GitClient interface {
    // CreateWorktree creates a new git worktree
    CreateWorktree(path string, ref plumbing.ReferenceName) (*git.Worktree, error)
    
    // ListWorktrees returns all existing worktrees
    ListWorktrees() ([]*git.Worktree, error)
    
    // DeleteWorktree removes a worktree
    DeleteWorktree(path string) error
    
    // GetCurrentBranch returns the current branch name
    GetCurrentBranch() (plumbing.ReferenceName, error)
}

// ConfigManager defines configuration operations
type ConfigManager interface {
    Load(configPath string) (*Config, error)
    Save(config *Config, configPath string) error
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}
```

### Bad Interface Design

```go
// Bad: God interface with too many responsibilities
type Repository interface {
    CreateWorktree(path string, ref plumbing.ReferenceName) (*git.Worktree, error)
    ListWorktrees() ([]*git.Worktree, error)
    DeleteWorktree(path string) error
    GetCurrentBranch() (plumbing.ReferenceName, error)
    LoadConfig(path string) (*Config, error)
    SaveConfig(config *Config, path string) error
    ValidateProject(project *Project) error
    SendNotification(message string) error
    LogError(err error) error
    // ... 20 more methods
}
```

## Import Organization

### Good Import Organization

```go
package services

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/spf13/cobra"
    
    "github.com/yourorg/twiggit/internal/domain"
    "github.com/yourorg/twiggit/internal/infrastructure"
)
```

### Bad Import Organization

```go
package services

import "fmt"
import "os"
import "path/filepath"
import "github.com/go-git/go-git/v5"
import "github.com/go-git/go-git/v5/plumbing"
import "github.com/spf13/cobra"
import "github.com/yourorg/twiggit/internal/domain"
import "github.com/yourorg/twiggit/internal/infrastructure"
```

## Testing Patterns

For comprehensive testing philosophy, framework usage, and command reference, see [testing.md](./testing.md). This section focuses on concrete Go code patterns for implementing tests.

### Unit Test Structure

```go
// Good: Table-driven tests
package domain

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProject_Validate(t *testing.T) {
    tests := []struct {
        name    string
        project *Project
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid project",
            project: &Project{name: "my-project", path: "/path/to/project"},
            wantErr: false,
        },
        {
            name:    "empty name",
            project: &Project{name: "", path: "/path/to/project"},
            wantErr: true,
            errMsg:  "project name cannot be empty",
        },
        {
            name:    "empty path",
            project: &Project{name: "my-project", path: ""},
            wantErr: true,
            errMsg:  "project path cannot be empty",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.project.Validate()
            
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Mock Usage

```go
// Good: Mock interfaces for testing
package services_test

import (
    "testing"
    
    "github.com/stretchr/testify/mock"
    
    "github.com/yourorg/twiggit/internal/infrastructure"
)

// MockGitClient implements GitClient interface for testing
type MockGitClient struct {
    mock.Mock
}

func (m *MockGitClient) CreateWorktree(path string, ref plumbing.ReferenceName) (*git.Worktree, error) {
    args := m.Called(path, ref)
    return args.Get(0).(*git.Worktree), args.Error(1)
}

func (m *MockGitClient) ListWorktrees() ([]*git.Worktree, error) {
    args := m.Called()
    return args.Get(0).([]*git.Worktree), args.Error(1)
}

func TestWorktreeService_CreateWorktree(t *testing.T) {
    mockGitClient := new(MockGitClient)
    service := NewWorktreeService(mockGitClient)
    
    // Setup mock expectations
    mockGitClient.On("CreateWorktree", "/path/to/worktree", plumbing.ReferenceName("main")).
        Return(&git.Worktree{}, nil)
    
    err := service.CreateWorktree("/path/to/worktree", "main")
    
    require.NoError(t, err)
    mockGitClient.AssertExpectations(t)
}
```

### Integration Test Structure

```go
//go:build integration

package integration

import (
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/require"
)

func TestWorktreeCreator_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Create temporary directory for test
    tempDir := t.TempDir()
    
    // Initialize a real git repository
    repoPath := filepath.Join(tempDir, "repo")
    err := os.MkdirAll(repoPath, 0755)
    require.NoError(t, err)
    
    // Test with real git operations
    creator := NewWorktreeCreator()
    worktreePath := filepath.Join(tempDir, "worktree")
    
    err = creator.CreateWorktree(repoPath, worktreePath, "main")
    require.NoError(t, err)
    
    // Verify worktree was created
    _, err = os.Stat(filepath.Join(worktreePath, ".git"))
    require.NoError(t, err)
}
```

## Naming Conventions

### Good Naming

```go
// Good: Clear, descriptive names
type ProjectValidator interface {
    ValidateProject(project *domain.Project) error
    ValidateProjectName(name string) error
    ValidateProjectPath(path string) error
}

type WorktreeCreator struct {
    gitClient infrastructure.GitClient
    logger    Logger
}

func (w *WorktreeCreator) CreateWorktree(basePath string, worktreePath string, branch string) error {
    // Implementation
}

// Good: Function names describe what they do
func validateProjectConfiguration(config *Config) error {
    // Implementation
}

func getWorktreeListFromRepository(repoPath string) ([]*git.Worktree, error) {
    // Implementation
}
```

### Bad Naming

```go
// Bad: Unclear or abbreviated names
type PVdr interface {
    Validate(p *domain.Project) error
    ValidateName(n string) error
    ValidatePath(p string) error
}

type WC struct {
    gc infrastructure.GitClient
    l  Logger
}

func (w *WC) Create(basePath string, wtPath string, b string) error {
    // Implementation
}

// Bad: Function names don't describe purpose
func doStuff(config *Config) error {
    // Implementation
}

func getWts(repoPath string) ([]*git.Worktree, error) {
    // Implementation
}
```

## Constants and Variables

### Good Constants

```go
// Good: Clear, descriptive constants
const (
    DefaultProjectName    = "default"
    DefaultBranchName     = "main"
    MaxProjectNameLength  = 50
    ConfigFileName        = "twiggit.yaml"
    ConfigEnvPrefix       = "TWIGGIT_"
    
    // Error messages
    ErrProjectNotFound    = "project not found"
    ErrInvalidProjectName = "invalid project name"
    ErrWorktreeExists     = "worktree already exists"
)

// Good: Group related constants
type ExitCode int

const (
    ExitSuccess ExitCode = 0
    ExitError   ExitCode = 1
    ExitUsage   ExitCode = 2
)
```

### Bad Constants

```go
// Bad: Magic numbers and unclear names
const (
    dn = "default"
    mb = "main"
    mx = 50
    cf = "twiggit.yaml"
    ep = "TWIGGIT_"
    
    // Unclear error codes
    e1 = 0
    e2 = 1
    e3 = 2
)
```

## Dependency Injection

### Good Dependency Injection

```go
// Good: Constructor injection with interfaces
type WorktreeService struct {
    gitClient     infrastructure.GitClient
    configManager infrastructure.ConfigManager
    validator     domain.Validator
    logger        Logger
}

func NewWorktreeService(
    gitClient infrastructure.GitClient,
    configManager infrastructure.ConfigManager,
    validator domain.Validator,
    logger Logger,
) *WorktreeService {
    return &WorktreeService{
        gitClient:     gitClient,
        configManager: configManager,
        validator:     validator,
        logger:        logger,
    }
}

// Usage in main.go
func main() {
    gitClient := infrastructure.NewGitClient()
    configManager := infrastructure.NewConfigManager()
    validator := domain.NewValidator()
    logger := NewLogger()
    
    worktreeService := NewWorktreeService(gitClient, configManager, validator, logger)
}
```

### Bad Dependency Injection

```go
// Bad: Global dependencies and tight coupling
type WorktreeService struct{}

func (w *WorktreeService) CreateWorktree(path string, ref string) error {
    // Bad: Direct dependency on global state
    gitClient := infrastructure.GetGlobalGitClient() // Global dependency
    config := infrastructure.GetGlobalConfig()        // Global dependency
    
    // Implementation
}

// Bad: No dependency injection
var worktreeService *WorktreeService

func init() {
    worktreeService = &WorktreeService{}
}
```

## Summary

This code style guide provides concrete examples and patterns for writing clean, maintainable Go code in the twiggit project. Following these patterns ensures consistency across the codebase and makes it easier for AI agents to understand and contribute to the project.

Key principles:
- **Clear separation of concerns**: Domain logic separate from infrastructure
- **Dependency injection**: Use interfaces and constructor injection
- **Error handling**: Wrap errors with context, avoid panics
- **Testing**: Use table-driven tests and proper mocking
- **Naming**: Use descriptive, clear names for types, functions, and variables
- **Project structure**: Follow standard Go conventions with domain-driven design