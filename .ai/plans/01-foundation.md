# Foundation Layer Implementation Plan

## Overview

This plan establishes the foundational structure for twiggit, implementing core domain entities, project structure, and basic infrastructure following TDD principles. The foundation layer provides the building blocks for all subsequent features.

## Project Structure

Based on [code-style-guide.md](../code-style-guide.md#project-structure):

```
cmd/                    # CLI entry points
├── create.go          # Create command implementation
├── delete.go          # Delete command implementation
├── list.go            # List command implementation
├── cd.go              # CD (change directory) command implementation
├── setup-shell.go     # Shell setup command implementation
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

## Implementation Steps

### Step 1: Initialize Go Module and Dependencies

**Files to create:**
- `go.mod`
- `go.sum` (generated)
- `.golangci.yml`

**Dependencies** (from [technology.md](../technology.md)):
```go
module github.com/yourorg/twiggit

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/go-git/go-git/v5 v5.11.0
    github.com/knadh/koanf/v2 v2.9.0
    github.com/stretchr/testify v1.8.4
    github.com/onsi/ginkgo/v2 v2.13.2
    github.com/onsi/gomega v1.30.0
    github.com/carapace-sh/carapace v3.1.2
)
```

### Step 2: Create Core Domain Entities

**Based on [design.md](../design.md) core entities: Project, Worktree, Context**

#### 2.1 Project Entity
**File:** `internal/domain/project.go`

**Tests first:** `internal/domain/project_test.go`

```go
// Tests following testing.md patterns
func TestProject_NewProject(t *testing.T) {
    testCases := []struct {
        name         string
        projectName  string
        gitRepo      string
        expectError  bool
        errorMessage string
    }{
        {
            name:        "valid project",
            projectName: "my-project",
            gitRepo:     "/path/to/repo",
            expectError: false,
        },
        {
            name:         "empty project name",
            projectName:  "",
            gitRepo:      "/path/to/repo",
            expectError:  true,
            errorMessage: "project name cannot be empty",
        },
    }
    // ... test implementation
}
```

**Implementation:**
```go
// Simple domain entity following code-style-guide.md
type Project struct {
    name     string
    gitRepo  string
    worktrees []*Worktree
}

// NewProject creates a new project instance
func NewProject(name, gitRepo string) (*Project, error) {
    if name == "" {
        return nil, fmt.Errorf("project name cannot be empty")
    }
    if gitRepo == "" {
        return nil, fmt.Errorf("git repository path cannot be empty")
    }
    
    return &Project{
        name:    name,
        gitRepo: gitRepo,
    }, nil
}

// Name returns the project name
func (p *Project) Name() string {
    return p.name
}

// GitRepoPath returns the git repository path
func (p *Project) GitRepoPath() string {
    return p.gitRepo
}
```

#### 2.2 Worktree Entity
**File:** `internal/domain/worktree.go`

**Tests first:** `internal/domain/worktree_test.go`

```go
func TestWorktree_NewWorktree(t *testing.T) {
    testCases := []struct {
        name         string
        path         string
        branch       string
        project      *Project
        expectError  bool
        errorMessage string
    }{
        {
            name:        "valid worktree",
            path:        "/home/user/Workspaces/project/feature",
            branch:      "feature",
            project:     &Project{name: "project", gitRepo: "/home/user/Projects/project"},
            expectError: false,
        },
        // ... more test cases
    }
    // ... test implementation
}
```

**Implementation:**
```go
type Worktree struct {
    path    string
    branch  string
    project *Project
}

func NewWorktree(path, branch string, project *Project) (*Worktree, error) {
    if path == "" {
        return nil, fmt.Errorf("worktree path cannot be empty")
    }
    if branch == "" {
        return nil, fmt.Errorf("worktree branch cannot be empty")
    }
    if project == nil {
        return nil, fmt.Errorf("worktree must belong to a project")
    }
    
    return &Worktree{
        path:    path,
        branch:  branch,
        project: project,
    }, nil
}

func (w *Worktree) Path() string {
    return w.path
}

func (w *Worktree) Branch() string {
    return w.branch
}

func (w *Worktree) Project() *Project {
    return w.project
}
```

#### 2.3 Context System
**File:** `internal/domain/context.go`

**Tests first:** `internal/domain/context_test.go`

```go
func TestContextDetector_DetectContext(t *testing.T) {
    testCases := []struct {
        name           string
        currentDir     string
        expectedType   ContextType
        expectedProject string
    }{
        {
            name:         "project context",
            currentDir:   "/home/user/Projects/myproject",
            expectedType: ContextProject,
            expectedProject: "myproject",
        },
        {
            name:         "workspace context", 
            currentDir:   "/home/user/Workspaces/myproject/feature",
            expectedType: ContextWorktree,
            expectedProject: "myproject",
        },
        {
            name:         "outside git context",
            currentDir:   "/tmp",
            expectedType: ContextOutsideGit,
            expectedProject: "",
        },
    }
    // ... test implementation
}
```

**Implementation:**
```go
type ContextType int

const (
    ContextUnknown ContextType = iota
    ContextProject
    ContextWorktree
    ContextOutsideGit
)

type Context struct {
    Type     ContextType
    Project  *Project
    Worktree *Worktree
    Path     string
}

type ContextDetector struct {
    projectsDir  string
    workspacesDir string
}

func NewContextDetector(projectsDir, workspacesDir string) *ContextDetector {
    return &ContextDetector{
        projectsDir:  projectsDir,
        workspacesDir: workspacesDir,
    }
}

func (cd *ContextDetector) Detect(currentPath string) (*Context, error) {
    // Implementation following design.md context detection rules
    // Priority: workspace > project > outside git
}
```

### Step 3: Infrastructure Interfaces

**File:** `internal/infrastructure/interfaces.go`

Following [code-style-guide.md](../code-style-guide.md#interface-design):

```go
package infrastructure

import "github.com/go-git/go-git/v5/plumbing"

// GitClient defines the contract for git operations
type GitClient interface {
    // CreateWorktree creates a new git worktree
    CreateWorktree(path string, ref plumbing.ReferenceName) error
    
    // ListWorktrees returns all existing worktrees
    ListWorktrees(repoPath string) ([]string, error)
    
    // DeleteWorktree removes a worktree
    DeleteWorktree(path string) error
    
    // GetCurrentBranch returns the current branch name
    GetCurrentBranch(repoPath string) (plumbing.ReferenceName, error)
    
    // ValidateRepository checks if path contains valid git repository
    ValidateRepository(path string) error
}

// ConfigManager defines configuration operations
type ConfigManager interface {
    Load(configPath string) (*Config, error)
    Save(config *Config, configPath string) error
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}

// FileSystem defines file system operations for testing
type FileSystem interface {
    Exists(path string) bool
    CreateDir(path string) error
    RemoveDir(path string) error
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte) error
}
```

### Step 4: Basic Configuration System

**File:** `internal/infrastructure/config.go`

**Tests first:** `internal/infrastructure/config_test.go`

```go
func TestConfigManager_Load(t *testing.T) {
    testCases := []struct {
        name        string
        configData  string
        expectError bool
        expected    *Config
    }{
        {
            name: "valid config",
            configData: `
projects_dir = "/custom/projects"
workspaces_dir = "/custom/workspaces"
default_source_branch = "main"
`,
            expectError: false,
            expected: &Config{
                ProjectsDir:        "/custom/projects",
                WorkspacesDir:      "/custom/workspaces",
                DefaultSourceBranch: "main",
            },
        },
        // ... more test cases
    }
    // ... test implementation
}
```

**Implementation:**
```go
type Config struct {
    ProjectsDir        string
    WorkspacesDir      string
    DefaultSourceBranch string
}

type ConfigManager struct {
    koanf *koanf.Koanf
}

func NewConfigManager() *ConfigManager {
    k := koanf.New(".")
    return &ConfigManager{koanf: k}
}

func (cm *ConfigManager) Load(configPath string) (*Config, error) {
    // Implementation using Koanf with TOML provider
    // Following implementation.md configuration requirements
}
```

### Step 5: Dependency Injection Setup

**File:** `internal/di/container.go`

**Tests first:** `internal/di/container_test.go`

```go
func TestContainer_ResolveGitClient(t *testing.T) {
    container := NewContainer()
    
    gitClient, err := container.ResolveGitClient()
    assert.NoError(t, err)
    assert.NotNil(t, gitClient)
    
    // Test singleton behavior
    gitClient2, err := container.ResolveGitClient()
    assert.NoError(t, err)
    assert.Equal(t, gitClient, gitClient2)
}
```

**Implementation:**
```go
type Container struct {
    gitClient     infrastructure.GitClient
    configManager infrastructure.ConfigManager
    fileSystem    infrastructure.FileSystem
}

func NewContainer() *Container {
    return &Container{}
}

func (c *Container) ResolveGitClient() (infrastructure.GitClient, error) {
    if c.gitClient == nil {
        c.gitClient = infrastructure.NewGitClient()
    }
    return c.gitClient, nil
}

func (c *Container) ResolveConfigManager() (infrastructure.ConfigManager, error) {
    if c.configManager == nil {
        c.configManager = infrastructure.NewConfigManager()
    }
    return c.configManager, nil
}
```

### Step 6: Basic CLI Structure

**File:** `cmd/root.go`

**Tests first:** `test/e2e/root_test.go` (E2E tests for CLI as per [testing.md](../testing.md))

```go
//go:build e2e
// +build e2e

package e2e

import (
    "testing"
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
)

func TestRootCommand(t *testing.T) {
    gomega.RegisterFailHandler(ginkgo.Fail)
    ginkgo.RunSpecs(t, "Root Command Suite")
}

var _ = ginkgo.Describe("Root Command", func() {
    ginkgo.Context("when executed without arguments", func() {
        ginkgo.It("should display help", func() {
            // E2E test implementation using gexec
        })
    })
})
```

**Implementation:**
```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/yourorg/twiggit/internal/di"
)

var rootCmd = &cobra.Command{
    Use:   "twiggit",
    Short: "Pragmatic git worktree management tool",
    Long: `Twiggit is a pragmatic tool for managing git worktrees 
with a focus on rebase workflows.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
}

func initConfig() {
    // Initialize configuration and dependencies
    container := di.NewContainer()
    // Store container for use by subcommands
}
```

### Step 7: Testing Infrastructure Setup

**File:** `test/helpers/testutil.go`

```go
package helpers

import (
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/require"
)

// SetupTestRepo creates a temporary git repository for testing
func SetupTestRepo(t *testing.T) string {
    tempDir := t.TempDir()
    repoPath := filepath.Join(tempDir, "repo")
    
    err := os.MkdirAll(repoPath, 0755)
    require.NoError(t, err)
    
    // Initialize git repository
    // ... git init and initial commit setup
    
    return repoPath
}

// CleanupTestRepo removes test repository
func CleanupTestRepo(t *testing.T, repoPath string) {
    err := os.RemoveAll(repoPath)
    require.NoError(t, err)
}
```

### Step 8: Quality Assurance Setup

**File:** `.golangci.yml`

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - misspell
    - unconvert
    - dupl
    - goconst
    - gocyclo

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3
```

## Implementation Order

1. **Step 1**: Go module and dependencies
2. **Step 2**: Core domain entities (Project, Worktree, Context)
3. **Step 3**: Infrastructure interfaces
4. **Step 4**: Configuration system
5. **Step 5**: Dependency injection
6. **Step 6**: Basic CLI structure
7. **Step 7**: Testing infrastructure
8. **Step 8**: Quality assurance setup

## Testing Strategy

Following [testing.md](../testing.md) requirements:

### Unit Tests
- **Framework**: Testify suites
- **Coverage**: >80% for business logic
- **Location**: `*_test.go` files alongside implementation
- **Focus**: Domain entities, validation, business rules

### Integration Tests
- **Framework**: Testify suites with build tags
- **Coverage**: Real git operations in temp directories
- **Location**: `test/integration/`
- **Focus**: Service interactions, git operations

### E2E Tests
- **Framework**: Ginkgo/Gomega with gexec
- **Coverage**: CLI command workflows
- **Location**: `test/e2e/`
- **Focus**: User-facing functionality

## Quality Gates

### Pre-commit Requirements
- All tests pass: `mise run test`
- Linting passes: `mise run lint:fix`
- Coverage >80%: `mise run test:coverage`

### CI Requirements
- All test types pass (unit, integration, E2E)
- Race condition detection: `mise run test:race`
- Cross-platform builds succeed
- Security scans pass

## Key Principles

### From [implementation.md](../implementation.md):
- **TDD Approach**: Write tests BEFORE implementation
- **Simple Domain Entities**: Keep business logic focused and clean
- **Dependency Injection**: Use interfaces for all external dependencies
- **Error Handling**: Wrap errors with context using `fmt.Errorf`

### From [code-style-guide.md](../code-style-guide.md):
- **Clear Separation**: Domain logic separate from infrastructure
- **Constructor Injection**: Use interfaces and constructor injection
- **Descriptive Naming**: Use clear, descriptive names
- **Table-driven Tests**: Use table-driven test patterns

### From [testing.md](../testing.md):
- **Pragmatic TDD**: Tests serve as safety net, not bureaucracy
- **Integration Focus**: Test how components work together
- **E2E for CLI**: Test complete user workflows
- **Coverage Threshold**: Enforce >80% coverage in CI

## Success Criteria

1. ✅ All domain entities have comprehensive tests
2. ✅ Project structure follows Go standards
3. ✅ Dependency injection container resolves all dependencies
4. ✅ Configuration system loads and validates settings
5. ✅ CLI structure supports subcommands
6. ✅ Test infrastructure supports all test types
7. ✅ Quality gates enforce code standards
8. ✅ Coverage exceeds 80% threshold

## Next Steps

After foundation implementation:
1. Implement git operations infrastructure
2. Build service layer for worktree management
3. Create CLI commands for core functionality
4. Add shell integration features
5. Implement context-aware navigation

This foundation provides the solid base needed for building robust, maintainable features while following established patterns and quality standards.