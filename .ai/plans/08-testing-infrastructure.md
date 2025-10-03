# Testing Infrastructure Implementation Plan

## Overview

This plan establishes the comprehensive testing infrastructure for twiggit, implementing the pragmatic TDD approach with >80% coverage and three-tier testing strategy as defined in [testing.md](../testing.md).

> "Testing philosophy: Pragmatic TDD, 80%+ coverage, three-tier approach" - testing.md:3

## Architecture

### Testing Tiers

1. **Unit Tests** - Testify suites, fast, isolated
2. **Integration Tests** - Real git operations, temporary repos
3. **E2E Tests** - Ginkgo/Gomega, built CLI binaries

### Directory Structure

```
test/
├── unit/                    # Unit tests with Testify
│   ├── config/
│   ├── git/
│   ├── cli/
│   └── services/
├── integration/             # Integration tests with real git
│   ├── git/
│   ├── hybrid/
│   └── performance/
├── e2e/                     # E2E tests with Ginkgo
│   ├── commands/
│   ├── shell/
│   └── workflows/
├── mocks/                   # Centralized mock implementations
│   ├── git/
│   ├── config/
│   └── services/
├── fixtures/                # Test data and repositories
│   ├── repos/
│   ├── configs/
│   └── scripts/
└── helpers/                 # Testing utilities and helpers
    ├── git.go
    ├── repo.go
    ├── shell.go
    └── performance.go
```

## Implementation Steps

### Phase 1: Foundation Setup

#### 1.1 Testing Framework Configuration

**File**: `test/framework_test.go`
```go
//go:build !integration && !e2e

package test

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

// BaseTestSuite provides common functionality for all unit tests
type BaseTestSuite struct {
    suite.Suite
    TempDir string
    Cleanup func()
}

func (s *BaseTestSuite) SetupTest() {
    s.TempDir = s.T().TempDir()
    s.Cleanup = func() {}
}

func (s *BaseTestSuite) TearDownTest() {
    s.Cleanup()
}
```

#### 1.2 Build Tags Configuration

**File**: `test/build_tags.go`
```go
//go:build integration

package test

const IntegrationBuild = true

//go:build e2e

package test

const E2EBuild = true
```

#### 1.3 Test Registry

**File**: `test/registry.go`
```go
package test

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

// TestRegistry manages test suite registration
type TestRegistry struct {
    suites map[string]func(t *testing.T)
}

func NewTestRegistry() *TestRegistry {
    return &TestRegistry{
        suites: make(map[string]func(t *testing.T)),
    }
}

func (r *TestRegistry) RegisterSuite(name string, suiteFunc func(t *testing.T)) {
    r.suites[name] = suiteFunc
}

func (r *TestRegistry) RunAll(t *testing.T) {
    for name, suiteFunc := range r.suites {
        t.Run(name, suiteFunc)
    }
}
```

### Phase 2: Mock Infrastructure

#### 2.1 Centralized Git Mock

**File**: `test/mocks/git/mock_repository.go`
```go
package mocks

import (
    "github.com/stretchr/testify/mock"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Worktrees() (*git.WorktreeContext, error) {
    args := m.Called()
    return args.Get(0).(*git.WorktreeContext), args.Error(1)
}

func (m *MockRepository) Worktree() (*git.Worktree, error) {
    args := m.Called()
    return args.Get(0).(*git.Worktree), args.Error(1)
}

func (m *MockRepository) Branch(name string) (*plumbing.Reference, error) {
    args := m.Called(name)
    return args.Get(0).(*plumbing.Reference), args.Error(1)
}

// Add more methods as needed...
```

#### 2.2 Config Mock

**File**: `test/mocks/config/mock_config.go`
```go
package mocks

import (
    "github.com/stretchr/testify/mock"
    "github.com/twiggit/twiggit/internal/config"
)

type MockConfig struct {
    mock.Mock
}

func (m *MockConfig) Get(key string) interface{} {
    args := m.Called(key)
    return args.Get(0)
}

func (m *MockConfig) Set(key string, value interface{}) error {
    args := m.Called(key, value)
    return args.Error(0)
}

func (m *MockConfig) GetString(key string) string {
    args := m.Called(key)
    return args.String(0)
}

// Add more methods as needed...
```

#### 2.3 Service Mocks

**File**: `test/mocks/services/mock_worktree_service.go`
```go
package mocks

import (
    "github.com/stretchr/testify/mock"
    "github.com/twiggit/twiggit/internal/services"
)

type MockWorktreeService struct {
    mock.Mock
}

func (m *MockWorktreeService) List() ([]*services.WorktreeInfo, error) {
    args := m.Called()
    return args.Get(0).([]*services.WorktreeInfo), args.Error(1)
}

func (m *MockWorktreeService) Create(name, branch string) (*services.WorktreeInfo, error) {
    args := m.Called(name, branch)
    return args.Get(0).(*services.WorktreeInfo), args.Error(1)
}

// Add more methods as needed...
```

### Phase 3: Test Helpers and Utilities

#### 3.1 Git Repository Helper

**File**: `test/helpers/git.go`
```go
package helpers

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

// GitTestHelper provides utilities for git testing
type GitTestHelper struct {
    t       *testing.T
    baseDir string
}

func NewGitTestHelper(t *testing.T) *GitTestHelper {
    return &GitTestHelper{
        t:       t,
        baseDir: t.TempDir(),
    }
}

// CreateBareRepo creates a bare repository for testing
func (h *GitTestHelper) CreateBareRepo() string {
    repoPath := filepath.Join(h.baseDir, "bare.git")
    _, err := git.PlainInit(repoPath, true)
    if err != nil {
        h.t.Fatalf("Failed to create bare repo: %v", err)
    }
    return repoPath
}

// CreateRepoWithCommits creates a repository with initial commits
func (h *GitTestHelper) CreateRepoWithCommits(commits int) string {
    repoPath := filepath.Join(h.baseDir, "repo")
    repo, err := git.PlainInit(repoPath, false)
    if err != nil {
        h.t.Fatalf("Failed to create repo: %v", err)
    }

    wt, err := repo.Worktree()
    if err != nil {
        h.t.Fatalf("Failed to get worktree: %v", err)
    }

    // Create initial commits
    for i := 0; i < commits; i++ {
        filename := filepath.Join(repoPath, "file.txt")
        content := []byte(fmt.Sprintf("Content %d\n", i))
        
        if err := os.WriteFile(filename, content, 0644); err != nil {
            h.t.Fatalf("Failed to write file: %v", err)
        }

        _, err = wt.Add("file.txt")
        if err != nil {
            h.t.Fatalf("Failed to add file: %v", err)
        }

        commit := &object.Commit{
            Message: fmt.Sprintf("Commit %d", i),
            Author: object.Signature{
                Name:  "Test User",
                Email: "test@example.com",
            },
        }

        _, err = wt.Commit(commit.Message, &git.CommitOptions{
            Author: &commit.Author,
        })
        if err != nil {
            h.t.Fatalf("Failed to commit: %v", err)
        }
    }

    return repoPath
}

// CreateBranch creates a new branch in the repository
func (h *GitTestHelper) CreateBranch(repoPath, branchName string) {
    repo, err := git.PlainOpen(repoPath)
    if err != nil {
        h.t.Fatalf("Failed to open repo: %v", err)
    }

    head, err := repo.Head()
    if err != nil {
        h.t.Fatalf("Failed to get HEAD: %v", err)
    }

    ref := plumbing.NewBranchReferenceName(branchName)
    err = repo.Storer.SetReference(plumbing.NewHashReference(ref, head.Hash()))
    if err != nil {
        h.t.Fatalf("Failed to create branch: %v", err)
    }
}
```

#### 3.2 Repository Management Helper

**File**: `test/helpers/repo.go`
```go
package helpers

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/twiggit/twiggit/internal/config"
)

// RepoTestHelper manages test repositories
type RepoTestHelper struct {
    t       *testing.T
    tempDir string
    repos   map[string]string
}

func NewRepoTestHelper(t *testing.T) *RepoTestHelper {
    return &RepoTestHelper{
        t:       t,
        tempDir: t.TempDir(),
        repos:   make(map[string]string),
    }
}

// SetupTestRepo creates a test repository with configuration
func (h *RepoTestHelper) SetupTestRepo(name string) string {
    repoPath := filepath.Join(h.tempDir, name)
    err := os.MkdirAll(repoPath, 0755)
    if err != nil {
        h.t.Fatalf("Failed to create repo dir: %v", err)
    }

    // Initialize git repo
    gitHelper := NewGitTestHelper(h.t)
    repoPath = gitHelper.CreateRepoWithCommits(3)

    // Create twiggit config
    configPath := filepath.Join(repoPath, ".twiggit.yml")
    configContent := `
default_branch: main
worktree_prefix: "wt-"
git:
  implementation: "go-git"
  timeout: "30s"
`
    err = os.WriteFile(configPath, []byte(configContent), 0644)
    if err != nil {
        h.t.Fatalf("Failed to write config: %v", err)
    }

    h.repos[name] = repoPath
    return repoPath
}

// Cleanup removes all test repositories
func (h *RepoTestHelper) Cleanup() {
    for _, path := range h.repos {
        os.RemoveAll(path)
    }
    h.repos = make(map[string]string)
}
```

#### 3.3 Shell Integration Helper

**File**: `test/helpers/shell.go`
```go
package helpers

import (
    "os"
    "os/exec"
    "strings"
    "testing"
)

// ShellTestHelper provides shell testing utilities
type ShellTestHelper struct {
    t *testing.T
}

func NewShellTestHelper(t *testing.T) *ShellTestHelper {
    return &ShellTestHelper{t: t}
}

// RunShellCommand executes a command in the specified shell
func (h *ShellTestHelper) RunShellCommand(shell, command string) (string, error) {
    var cmd *exec.Cmd
    
    switch shell {
    case "bash":
        cmd = exec.Command("bash", "-c", command)
    case "zsh":
        cmd = exec.Command("zsh", "-c", command)
    case "fish":
        cmd = exec.Command("fish", "-c", command)
    default:
        cmd = exec.Command("sh", "-c", command)
    }

    output, err := cmd.CombinedOutput()
    return string(output), err
}

// SetupShellEnvironment prepares shell environment for testing
func (h *ShellTestHelper) SetupShellEnvironment() {
    // Set up environment variables for testing
    os.Setenv("TWIGGIT_TEST_MODE", "1")
    os.Setenv("TWIGGIT_CONFIG_DIR", h.t.TempDir())
}

// ValidateShellScript checks if a shell script is valid
func (h *ShellTestHelper) ValidateShellScript(shell, script string) error {
    var cmd *exec.Cmd
    
    switch shell {
    case "bash":
        cmd = exec.Command("bash", "-n", script)
    case "zsh":
        cmd = exec.Command("zsh", "-n", script)
    case "fish":
        cmd = exec.Command("fish", "-n", script)
    default:
        cmd = exec.Command("sh", "-n", script)
    }

    return cmd.Run()
}
```

#### 3.4 Performance Testing Helper

**File**: `test/helpers/performance.go`
```go
package helpers

import (
    "testing"
    "time"
    "github.com/stretchr/testify/require"
)

// PerformanceTestHelper provides performance testing utilities
type PerformanceTestHelper struct {
    t *testing.T
}

func NewPerformanceTestHelper(t *testing.T) *PerformanceTestHelper {
    return &PerformanceTestHelper{t: t}
}

// BenchmarkOperation measures operation performance
func (h *PerformanceTestHelper) BenchmarkOperation(name string, operation func() error) time.Duration {
    start := time.Now()
    err := operation()
    duration := time.Since(start)
    
    require.NoError(h.t, err, "Operation %s failed", name)
    h.t.Logf("Operation %s took %v", name, duration)
    
    return duration
}

// CreateLargeRepo creates a repository with many worktrees for performance testing
func (h *PerformanceTestHelper) CreateLargeRepo(worktreeCount int) string {
    gitHelper := NewGitTestHelper(h.t)
    repoPath := gitHelper.CreateRepoWithCommits(10)
    
    // Create multiple branches
    for i := 0; i < worktreeCount; i++ {
        branchName := fmt.Sprintf("feature-%d", i)
        gitHelper.CreateBranch(repoPath, branchName)
    }
    
    return repoPath
}

// AssertPerformance asserts that an operation meets performance requirements
func (h *PerformanceTestHelper) AssertPerformance(operation func() error, maxDuration time.Duration) {
    duration := h.BenchmarkOperation("performance_test", operation)
    require.Less(h.t, duration, maxDuration, 
        "Operation took %v, expected less than %v", duration, maxDuration)
}
```

### Phase 4: Integration Test Framework

#### 4.1 Integration Test Base

**File**: `test/integration/integration_test.go`
```go
//go:build integration

package integration

import (
    "testing"
    "github.com/stretchr/testify/suite"
    "github.com/twiggit/twiggit/test/helpers"
)

type IntegrationTestSuite struct {
    suite.Suite
    gitHelper *helpers.GitTestHelper
    repoHelper *helpers.RepoTestHelper
}

func (s *IntegrationTestSuite) SetupSuite() {
    s.gitHelper = helpers.NewGitTestHelper(s.T())
    s.repoHelper = helpers.NewRepoTestHelper(s.T())
}

func (s *IntegrationTestSuite) TearDownSuite() {
    s.repoHelper.Cleanup()
}

func TestIntegrationSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}
```

#### 4.2 Hybrid Git Integration Tests

**File**: `test/integration/hybrid/hybrid_test.go`
```go
//go:build integration

package hybrid

import (
    "testing"
    "github.com/stretchr/testify/suite"
    "github.com/twiggit/twiggit/test/integration"
)

> "Hybrid git testing: validate both implementations work identically" - testing.md:89

type HybridGitTestSuite struct {
    integration.IntegrationTestSuite
}

func (s *HybridGitTestSuite) TestBothImplementationsIdentical() {
    // Test with go-git implementation
    s.testWithImplementation("go-git")
    
    // Test with libgit2 implementation  
    s.testWithImplementation("libgit2")
    
    // Compare results
    s.compareResults()
}

func (s *HybridGitTestSuite) testWithImplementation(impl string) {
    repoPath := s.repoHelper.SetupTestRepo("hybrid-" + impl)
    
    // Configure implementation
    config := config.New()
    config.Set("git.implementation", impl)
    
    // Test operations
    s.testListWorktrees(repoPath)
    s.testCreateWorktree(repoPath)
    s.testDeleteWorktree(repoPath)
}

func (s *HybridGitTestSuite) compareResults() {
    // Compare results from both implementations
    // Ensure they work identically
}

func TestHybridGitSuite(t *testing.T) {
    suite.Run(t, new(HybridGitTestSuite))
}
```

#### 4.3 Performance Integration Tests

**File**: `test/integration/performance/performance_test.go`
```go
//go:build integration

package performance

import (
    "testing"
    "time"
    "github.com/stretchr/testify/suite"
    "github.com/twiggit/twiggit/test/integration"
    "github.com/twiggit/twiggit/test/helpers"
)

> "Performance testing: large repositories with 100+ worktrees" - testing.md:95

type PerformanceTestSuite struct {
    integration.IntegrationTestSuite
    perfHelper *helpers.PerformanceTestHelper
}

func (s *PerformanceTestSuite) SetupSuite() {
    s.IntegrationTestSuite.SetupSuite()
    s.perfHelper = helpers.NewPerformanceTestHelper(s.T())
}

func (s *PerformanceTestSuite) TestLargeRepositoryPerformance() {
    // Create repository with 100+ worktrees
    repoPath := s.perfHelper.CreateLargeRepo(100)
    
    // Test list performance
    s.perfHelper.AssertPerformance(func() error {
        return s.listWorktrees(repoPath)
    }, 5*time.Second)
    
    // Test create performance
    s.perfHelper.AssertPerformance(func() error {
        return s.createWorktree(repoPath, "perf-test")
    }, 10*time.Second)
}

func (s *PerformanceTestSuite) TestMemoryUsage() {
    // Test memory usage with large repositories
    // Ensure no memory leaks
}

func TestPerformanceSuite(t *testing.T) {
    suite.Run(t, new(PerformanceTestSuite))
}
```

### Phase 5: E2E Test Framework

#### 5.1 E2E Test Base

**File**: `test/e2e/e2e_test.go`
```go
//go:build e2e

package e2e

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
)

var (
    binaryPath string
    tempDir    string
)

func TestE2E(t *testing.T) {
    gomega.RegisterFailHandler(ginkgo.Fail)
    ginkgo.RunSpecs(t, "E2E Suite")
}

var _ = ginkgo.BeforeSuite(func() {
    // Build the CLI binary for testing
    buildBinary()
    
    // Set up temporary directory
    tempDir = ginkgo.GinkgoT().TempDir()
})

var _ = ginkgo.AfterSuite(func() {
    // Clean up
    if binaryPath != "" {
        os.Remove(binaryPath)
    }
})

func buildBinary() {
    cmd := exec.Command("go", "build", "-o", "twiggit-test", "./cmd/twiggit")
    err := cmd.Run()
    gomega.Expect(err).NotTo(gomega.HaveOccurred())
    
    binaryPath, err = filepath.Abs("twiggit-test")
    gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func runTwiggitCommand(args ...string) *exec.Cmd {
    cmd := exec.Command(binaryPath, args...)
    cmd.Dir = tempDir
    return cmd
}
```

#### 5.2 Command E2E Tests

**File**: `test/e2e/commands/commands_test.go`
```go
//go:build e2e

package commands

import (
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
    "github.com/twiggit/twiggit/test/e2e"
)

var _ = ginkgo.Describe("CLI Commands", func() {
    ginkgo.Context("list command", func() {
        ginkgo.It("should list worktrees", func() {
            cmd := e2e.runTwiggitCommand("list")
            output, err := cmd.CombinedOutput()
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
            gomega.Expect(string(output)).To(gomega.ContainSubstring("worktrees"))
        })
    })
    
    ginkgo.Context("create command", func() {
        ginkgo.It("should create a new worktree", func() {
            cmd := e2e.runTwiggitCommand("create", "test-feature", "-b", "feature/test")
            output, err := cmd.CombinedOutput()
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
            gomega.Expect(string(output)).To(gomega.ContainSubstring("Created"))
        })
    })
    
    ginkgo.Context("delete command", func() {
        ginkgo.It("should delete a worktree", func() {
            // First create a worktree
            createCmd := e2e.runTwiggitCommand("create", "temp-feature")
            _, err := createCmd.CombinedOutput()
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
            
            // Then delete it
            deleteCmd := e2e.runTwiggitCommand("delete", "temp-feature")
            output, err := deleteCmd.CombinedOutput()
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
            gomega.Expect(string(output)).To(gomega.ContainSubstring("Deleted"))
        })
    })
})
```

#### 5.3 Shell Integration E2E Tests

**File**: `test/e2e/shell/shell_test.go`
```go
//go:build e2e

package shell

import (
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
    "github.com/twiggit/twiggit/test/e2e"
    "github.com/twiggit/twiggit/test/helpers"
)

> "Shell integration testing: bash, zsh, fish compatibility" - testing.md:98

var _ = ginkgo.Describe("Shell Integration", func() {
    var shellHelper *helpers.ShellTestHelper
    
    ginkgo.BeforeEach(func() {
        shellHelper = helpers.NewShellTestHelper(ginkgo.GinkgoT())
        shellHelper.SetupShellEnvironment()
    })
    
    ginkgo.Context("bash integration", func() {
        ginkgo.It("should work with bash completion", func() {
            output, err := shellHelper.RunShellCommand("bash", 
                "source <(twiggit setup-shell bash) && compgen -W 'list create delete' twiggit")
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
            gomega.Expect(output).To(gomega.ContainSubstring("list"))
            gomega.Expect(output).To(gomega.ContainSubstring("create"))
            gomega.Expect(output).To(gomega.ContainSubstring("delete"))
        })
    })
    
    ginkgo.Context("zsh integration", func() {
        ginkgo.It("should work with zsh completion", func() {
            output, err := shellHelper.RunShellCommand("zsh", 
                "source <(twiggit setup-shell zsh) && compadd -W 'list create delete' -- twiggit")
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
        })
    })
    
    ginkgo.Context("fish integration", func() {
        ginkgo.It("should work with fish completion", func() {
            output, err := shellHelper.RunShellCommand("fish", 
                "twiggit setup-shell fish | source && complete -C'twiggit '")
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
        })
    })
})
```

### Phase 6: Coverage and CI Integration

#### 6.1 Coverage Configuration

**File**: `test/coverage.go`
```go
package test

import (
    "os"
    "os/exec"
    "path/filepath"
)

// CoverageHelper manages test coverage
type CoverageHelper struct {
    coverageDir string
    profileFile string
}

func NewCoverageHelper() *CoverageHelper {
    return &CoverageHelper{
        coverageDir: "coverage",
        profileFile: "coverage.out",
    }
}

// RunCoverage runs tests with coverage
func (c *CoverageHelper) RunCoverage() error {
    // Create coverage directory
    os.MkdirAll(c.coverageDir, 0755)
    
    // Run tests with coverage
    cmd := exec.Command("go", "test", "-v", "-coverprofile="+c.profileFile, "./...")
    return cmd.Run()
}

// GenerateHTMLReport generates HTML coverage report
func (c *CoverageHelper) GenerateHTMLReport() error {
    cmd := exec.Command("go", "tool", "cover", "-html="+c.profileFile, "-o", filepath.Join(c.coverageDir, "coverage.html"))
    return cmd.Run()
}

// CheckCoverageThreshold checks if coverage meets threshold
func (c *CoverageHelper) CheckCoverageThreshold(threshold float64) (float64, error) {
    cmd := exec.Command("go", "tool", "cover", "-func="+c.profileFile)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return 0, err
    }
    
    // Parse coverage percentage from output
    // Implementation depends on go tool cover output format
    // Return actual coverage percentage
    
    return 85.0, nil // Example value
}
```

#### 6.2 CI Integration Script

**File**: `.mise/tasks/ci/test.sh`
```bash
#!/bin/bash

set -e

echo "Running comprehensive test suite..."

# Unit tests
echo "Running unit tests..."
go test -v -race ./test/unit/...

# Integration tests
echo "Running integration tests..."
go test -v -tags=integration ./test/integration/...

# E2E tests
echo "Running E2E tests..."
go test -v -tags=e2e ./test/e2e/...

# Coverage
echo "Generating coverage report..."
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage below 80% threshold"
    exit 1
fi

echo "All tests passed!"
```

## Testing Patterns

### Unit Test Pattern

```go
func TestWorktreeService_Create(t *testing.T) {
    suite.Run(t, new(WorktreeServiceTestSuite))
}

type WorktreeServiceTestSuite struct {
    suite.Suite
    service *services.WorktreeService
    mockRepo *mocks.MockRepository
}

func (s *WorktreeServiceTestSuite) SetupTest() {
    s.mockRepo = new(mocks.MockRepository)
    s.service = services.NewWorktreeService(s.mockRepo)
}

func (s *WorktreeServiceTestSuite) TestCreate_Success() {
    tests := []struct {
        name     string
        branch   string
        setup    func()
        expected *services.WorktreeInfo
        err      error
    }{
        {
            name:   "valid worktree creation",
            branch: "feature/test",
            setup: func() {
                s.mockRepo.On("Worktree").Return(nil, nil)
            },
            expected: &services.WorktreeInfo{Name: "test", Branch: "feature/test"},
            err: nil,
        },
    }
    
    for _, tt := range tests {
        s.Run(tt.name, func() {
            tt.setup()
            
            result, err := s.service.Create("test", tt.branch)
            
            if tt.err != nil {
                s.Error(err)
                s.Equal(tt.err, err)
            } else {
                s.NoError(err)
                s.Equal(tt.expected, result)
            }
            
            s.mockRepo.AssertExpectations(s.T())
        })
    }
}
```

### Integration Test Pattern

```go
func (s *GitIntegrationTestSuite) TestRealGitOperations() {
    repoPath := s.repoHelper.SetupTestRepo("real-git-test")
    
    // Test actual git operations
    worktrees, err := s.service.List()
    s.NoError(err)
    s.Empty(worktrees) // Initially empty
    
    // Create worktree
    info, err := s.service.Create("feature-branch", "feature/test")
    s.NoError(err)
    s.NotNil(info)
    
    // Verify worktree exists
    worktrees, err = s.service.List()
    s.NoError(err)
    s.Len(worktrees, 1)
    s.Equal("feature-branch", worktrees[0].Name)
}
```

### E2E Test Pattern

```go
var _ = ginkgo.Describe("Worktree Management", func() {
    ginkgo.Context("when creating worktrees", func() {
        ginkgo.BeforeEach(func() {
            // Set up test repository
            cmd := e2e.runTwiggitCommand("setup", "--test")
            _, err := cmd.CombinedOutput()
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
        })
        
        ginkgo.It("should create worktree with correct branch", func() {
            cmd := e2e.runTwiggitCommand("create", "test-feature", "-b", "feature/test")
            output, err := cmd.CombinedOutput()
            
            gomega.Expect(err).NotTo(gomega.HaveOccurred())
            gomega.Expect(string(output)).To(gomega.ContainSubstring("Created worktree"))
            gomega.Expect(string(output)).To(gomega.ContainSubstring("test-feature"))
        })
    })
})
```

## Implementation Checklist

### Phase 1: Foundation Setup
- [ ] Testing framework configuration
- [ ] Build tags setup
- [ ] Test registry implementation

### Phase 2: Mock Infrastructure
- [ ] Git repository mocks
- [ ] Configuration mocks
- [ ] Service mocks

### Phase 3: Test Helpers
- [ ] Git repository helper
- [ ] Repository management helper
- [ ] Shell integration helper
- [ ] Performance testing helper

### Phase 4: Integration Tests
- [ ] Integration test base
- [ ] Hybrid git tests
- [ ] Performance integration tests

### Phase 5: E2E Tests
- [ ] E2E test framework
- [ ] Command E2E tests
- [ ] Shell integration E2E tests

### Phase 6: Coverage and CI
- [ ] Coverage helper
- [ ] CI integration script
- [ ] Coverage threshold enforcement

## Service Layer Testing

### Integration Testing Patterns

Service layer integration tests SHALL verify coordination between services:

```go
// test/integration/services_test.go
package integration

import (
    "context"
    "testing"
    "github.com/stretchr/testify/suite"
    "github.com/twiggit/twiggit/internal/services"
    "github.com/twiggit/twiggit/test/helpers"
)

type ServiceIntegrationSuite struct {
    suite.Suite
    tempDir        string
    gitHelper      *helpers.GitTestHelper
    worktreeService services.WorktreeService
    projectService  services.ProjectService
    navService      services.NavigationService
}

func (s *ServiceIntegrationSuite) SetupSuite() {
    var err error
    s.tempDir = s.T().TempDir()
    s.gitHelper = helpers.NewGitTestHelper(s.T())
    
    // Setup test repository
    repoPath := s.gitHelper.CreateTestRepository("test-project")
    
    // Initialize services with real dependencies
    s.worktreeService = setupWorktreeService(repoPath)
    s.projectService = setupProjectService(repoPath)
    s.navService = setupNavigationService(repoPath)
}

func (s *ServiceIntegrationSuite) TestWorktreeCreationFlow() {
    // Test complete worktree creation flow
    req := &CreateWorktreeRequest{
        ProjectName:  "test-project",
        BranchName:   "feature-branch",
        SourceBranch: "main",
        Context: &domain.Context{
            Type:       domain.ContextOutsideGit,
            ProjectName: "test-project",
        },
    }
    
    // Create worktree
    worktree, err := s.worktreeService.CreateWorktree(context.Background(), req)
    s.Require().NoError(err)
    s.Equal("feature-branch", worktree.Branch)
    
    // Verify worktree exists
    status, err := s.worktreeService.GetWorktreeStatus(context.Background(), worktree.Path)
    s.Require().NoError(err)
    s.Equal("feature-branch", status.Branch)
    
    // Test navigation to worktree
    navReq := &ResolvePathRequest{
        Target:  "feature-branch",
        Context: req.Context,
    }
    
    resolution, err := s.navService.ResolvePath(context.Background(), navReq)
    s.Require().NoError(err)
    s.Equal(worktree.Path, resolution.ResolvedPath)
}
```

### Error Scenario Testing

Service layer SHALL include comprehensive error scenario testing:

```go
// test/integration/service_errors_test.go
package integration

func (s *ServiceIntegrationSuite) TestWorktreeCreationErrors() {
    testCases := []struct {
        name         string
        request      *CreateWorktreeRequest
        expectError  string
        errorType    error
    }{
        {
            name: "duplicate worktree",
            request: &CreateWorktreeRequest{
                ProjectName: "test-project",
                BranchName:  "main", // Already exists
                Context: &domain.Context{
                    Type: domain.ContextOutsideGit,
                },
            },
            expectError: "worktree already exists",
            errorType:   &domain.WorktreeExistsError{},
        },
        {
            name: "invalid project",
            request: &CreateWorktreeRequest{
                ProjectName: "nonexistent-project",
                BranchName:  "feature",
                Context: &domain.Context{
                    Type: domain.ContextOutsideGit,
                },
            },
            expectError: "project not found",
            errorType:   &domain.ProjectNotFoundError{},
        },
        {
            name: "unsafe operation - dirty worktree",
            request: &DeleteWorktreeRequest{
                ProjectName:  "test-project",
                BranchName:   "feature-branch",
                WorktreePath: s.tempDir + "/worktrees/feature-branch",
                Force:        false,
                Context: &domain.Context{
                    Type: domain.ContextWorktree,
                },
            },
            expectError: "worktree has uncommitted changes",
            errorType:   &domain.UnsafeOperationError{},
        },
    }
    
    for _, tc := range testCases {
        s.Run(tc.name, func() {
            // Setup dirty worktree for delete test
            if tc.request.WorktreePath != "" {
                s.gitHelper.CreateDirtyWorktree(tc.request.WorktreePath)
            }
            
            var err error
            switch req := tc.request.(type) {
            case *CreateWorktreeRequest:
                _, err = s.worktreeService.CreateWorktree(context.Background(), req)
            case *DeleteWorktreeRequest:
                err = s.worktreeService.DeleteWorktree(context.Background(), req)
            }
            
            s.Error(err)
            s.Contains(err.Error(), tc.expectError)
            
            // Check error type
            s.IsType(tc.errorType, err)
        })
    }
}
```

### Service Coordination Tests

Tests SHALL verify services work together correctly:

```go
// test/integration/service_coordination_test.go
package integration

func (s *ServiceIntegrationSuite) TestProjectWorktreeCoordination() {
    // Create project with multiple worktrees
    project, err := s.projectService.DiscoverProject(
        context.Background(), 
        "test-project", 
        &domain.Context{Type: domain.ContextOutsideGit},
    )
    s.Require().NoError(err)
    
    // Create multiple worktrees
    branches := []string{"feature-a", "feature-b", "feature-c"}
    for _, branch := range branches {
        req := &CreateWorktreeRequest{
            ProjectName:  project.Name,
            BranchName:   branch,
            SourceBranch: "main",
            Context: &domain.Context{
                Type:       domain.ContextProject,
                ProjectName: project.Name,
            },
        }
        
        _, err := s.worktreeService.CreateWorktree(context.Background(), req)
        s.Require().NoError(err)
    }
    
    // List all worktrees for project
    listReq := &ListWorktreesRequest{
        ProjectName: project.Name,
        AllProjects: false,
        Context: &domain.Context{
            Type:       domain.ContextProject,
            ProjectName: project.Name,
        },
    }
    
    worktrees, err := s.worktreeService.ListWorktrees(context.Background(), listReq)
    s.Require().NoError(err)
    s.Len(worktrees, len(branches)+1) // +1 for main branch
    
    // Test navigation to each worktree
    for _, worktree := range worktrees {
        navReq := &ResolvePathRequest{
            Target:  worktree.Branch,
            Context: &domain.Context{
                Type:       domain.ContextProject,
                ProjectName: project.Name,
            },
        }
        
        resolution, err := s.navService.ResolvePath(context.Background(), navReq)
        s.Require().NoError(err)
        s.Equal(worktree.Path, resolution.ResolvedPath)
    }
}
```

### Mock Service Implementations

Mock services SHALL be provided for unit testing:

```go
// test/mocks/services/worktree_service_mock.go
package mocks

import (
    "context"
    "github.com/stretchr/testify/mock"
    "github.com/twiggit/twiggit/internal/domain"
    "github.com/twiggit/twiggit/internal/services"
)

type MockWorktreeService struct {
    mock.Mock
}

func (m *MockWorktreeService) CreateWorktree(ctx context.Context, req *CreateWorktreeRequest) (*WorktreeInfo, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*WorktreeInfo), args.Error(1)
}

func (m *MockWorktreeService) DeleteWorktree(ctx context.Context, req *DeleteWorktreeRequest) error {
    args := m.Called(ctx, req)
    return args.Error(0)
}

func (m *MockWorktreeService) ListWorktrees(ctx context.Context, req *ListWorktreesRequest) ([]*WorktreeInfo, error) {
    args := m.Called(ctx, req)
    return args.Get(0).([]*WorktreeInfo), args.Error(1)
}

func (m *MockWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*WorktreeStatus, error) {
    args := m.Called(ctx, worktreePath)
    return args.Get(0).(*WorktreeStatus), args.Error(1)
}

func (m *MockWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
    args := m.Called(ctx, worktreePath)
    return args.Error(0)
}

// Mock setup helpers
func (m *MockWorktreeService) SetupCreateWorktreeSuccess(projectName, branchName string) {
    worktree := &WorktreeInfo{
        Path:   "/test/worktrees/" + projectName + "/" + branchName,
        Branch: branchName,
    }
    
    m.On("CreateWorktree", mock.Anything, mock.MatchedBy(func(req *CreateWorktreeRequest) bool {
        return req.ProjectName == projectName && req.BranchName == branchName
    })).Return(worktree, nil)
}

func (m *MockWorktreeService) SetupCreateWorktreeError(projectName, branchName, errorMsg string) {
    m.On("CreateWorktree", mock.Anything, mock.MatchedBy(func(req *CreateWorktreeRequest) bool {
        return req.ProjectName == projectName && req.BranchName == branchName
    })).Return(nil, fmt.Errorf(errorMsg))
}
```

## Quality Gates

> ">80% coverage enforced in CI" - testing.md:85

1. **Coverage**: Minimum 80% code coverage
2. **Performance**: Operations complete within specified time limits
3. **Service Integration**: All service coordination tests pass
4. **Error Scenarios**: Comprehensive error scenario coverage
5. **Mock Coverage**: All service interfaces have mock implementations
3. **Compatibility**: All supported shells work correctly
4. **Hybrid Testing**: Both git implementations produce identical results
5. **CI Integration**: All tests pass in CI environment

## Shell Integration Testing (Deferred from Phase 7)

### Integration Tests for Shell Integration

#### 4.1 Shell Detection Integration Tests

**File**: `test/integration/shell_detection_test.go`

```go
//go:build integration
// +build integration

package integration

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/twiggit/twiggit/internal/infrastructure/shell"
    "github.com/twiggit/twiggit/test/helpers"
)

func TestShellDetection_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    testCases := []struct {
        name         string
        shellEnv     string
        expectType   shell.ShellType
        expectError  bool
    }{
        {
            name:       "detect bash in real environment",
            shellEnv:   "/bin/bash",
            expectType: shell.ShellBash,
        },
        {
            name:       "detect zsh in real environment",
            shellEnv:   "/usr/bin/zsh",
            expectType: shell.ShellZsh,
        },
        {
            name:        "invalid shell path",
            shellEnv:    "/nonexistent/shell",
            expectError: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            helper := helpers.NewShellTestHelper(t)
            helper.SetupShellEnvironment()

            // Set environment variable
            oldShell := os.Getenv("SHELL")
            defer os.Setenv("SHELL", oldShell)
            os.Setenv("SHELL", tc.shellEnv)

            detector := shell.NewShellDetector()
            detectedShell, err := detector.DetectCurrentShell()

            if tc.expectError {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tc.expectType, detectedShell.Type())
                assert.Equal(t, tc.shellEnv, detectedShell.Path())
            }
        })
    }
}
```

#### 4.2 Shell Wrapper Installation Tests

**File**: `test/integration/shell_wrapper_test.go`

```go
//go:build integration
// +build integration

package integration

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/twiggit/twiggit/internal/infrastructure/shell"
    "github.com/twiggit/twiggit/test/helpers"
)

func TestShellWrapperInstallation_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    testCases := []struct {
        name      string
        shellType shell.ShellType
        configExt string
    }{
        {
            name:      "bash wrapper installation",
            shellType: shell.ShellBash,
            configExt: "bashrc",
        },
        {
            name:      "zsh wrapper installation",
            shellType: shell.ShellZsh,
            configExt: "zshrc",
        },
        {
            name:      "fish wrapper installation",
            shellType: shell.ShellFish,
            configExt: "config.fish",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            helper := helpers.NewShellTestHelper(t)
            tempDir := t.TempDir()
            
            // Create temporary config file
            configPath := filepath.Join(tempDir, tc.configExt)
            initialContent := "# Initial shell configuration\n"
            require.NoError(t, os.WriteFile(configPath, []byte(initialContent), 0644))

            // Create mock shell with temporary config
            mockShell := &mockShell{
                shellType:   tc.shellType,
                configFiles: []string{configPath},
            }

            detector := shell.NewShellDetector()
            integration := shell.NewShellIntegrationService(detector)

            // Generate wrapper
            wrapper, err := integration.GenerateWrapper(mockShell)
            require.NoError(t, err)
            assert.NotEmpty(t, wrapper)

            // Validate wrapper syntax
            err = helper.ValidateShellScript(string(tc.shellType), wrapper)
            assert.NoError(t, err)

            // Install wrapper
            err = integration.InstallWrapper(mockShell, wrapper)
            require.NoError(t, err)

            // Verify installation
            content, err := os.ReadFile(configPath)
            require.NoError(t, err)
            
            contentStr := string(content)
            assert.Contains(t, contentStr, "# Twiggit shell wrapper")
            assert.Contains(t, contentStr, "twiggit()")
            assert.Contains(t, contentStr, initialContent) // Original content preserved

            // Test duplicate installation prevention
            err = integration.InstallWrapper(mockShell, wrapper)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), "already installed")
        })
    }
}
```

#### 4.3 Shell Configuration File Detection Tests

**File**: `test/integration/shell_config_test.go`

```go
//go:build integration
// +build integration

package integration

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/twiggit/twiggit/internal/infrastructure/shell"
)

func TestShellConfigDetection_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    testCases := []struct {
        name          string
        shellType     shell.ShellType
        createFiles   []string
        expectedFile  string
    }{
        {
            name:      "bash prefers .bashrc over .profile",
            shellType: shell.ShellBash,
            createFiles: []string{
                ".profile",
                ".bashrc",
            },
            expectedFile: ".bashrc",
        },
        {
            name:      "zsh prefers .zshrc over .zprofile",
            shellType: shell.ShellZsh,
            createFiles: []string{
                ".zprofile",
                ".zshrc",
            },
            expectedFile: ".zshrc",
        },
        {
            name:      "fish uses config.fish in .config/fish",
            shellType: shell.ShellFish,
            createFiles: []string{
                ".config/fish/config.fish",
            },
            expectedFile: ".config/fish/config.fish",
        },
        {
            name:         "bash falls back to .profile when .bashrc missing",
            shellType:    shell.ShellBash,
            createFiles:  []string{".profile"},
            expectedFile: ".profile",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            tempDir := t.TempDir()
            
            // Create test files
            for _, file := range tc.createFiles {
                fullPath := filepath.Join(tempDir, file)
                dir := filepath.Dir(fullPath)
                require.NoError(t, os.MkdirAll(dir, 0755))
                require.NoError(t, os.WriteFile(fullPath, []byte("# test config"), 0644))
            }

            // Create mock shell
            mockShell := &mockShell{
                shellType:   tc.shellType,
                configFiles: tc.createFiles,
            }

            detector := shell.NewShellDetector()
            integration := shell.NewShellIntegrationService(detector)

            // Detect config file
            detectedPath, err := integration.DetectConfigFile(mockShell)
            require.NoError(t, err)

            expectedPath := filepath.Join(tempDir, tc.expectedFile)
            assert.Equal(t, expectedPath, detectedPath)
        })
    }
}
```

### E2E Tests for Shell Integration

#### 4.4 Setup-Shell Command E2E Tests

**File**: `test/e2e/shell_setup_test.go`

```go
//go:build e2e
// +build e2e

package e2e

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
    "github.com/onsi/gomega/gexec"
)

var _ = ginkgo.Describe("Setup-Shell Command", func() {
    var (
        binPath    string
        tempDir    string
        configPath string
    )

    ginkgo.BeforeEach(func() {
        var err error
        tempDir = ginkgo.GinkgoT().TempDir()
        
        // Build test binary
        binPath, err = gexec.Build("github.com/twiggit/twiggit")
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
        
        // Create test config file
        configPath = filepath.Join(tempDir, ".bashrc")
        err = os.WriteFile(configPath, []byte("# Initial config\n"), 0644)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
    })

    ginkgo.AfterEach(func() {
        gexec.CleanupBuildArtifacts()
    })

    ginkgo.Context("when setting up bash shell", func() {
        ginkgo.It("should install wrapper successfully", func() {
            session := gexec.Start(gexec.Command(binPath, "setup-shell", "--dry-run"), 
                gexec.NewBuffer(), gexec.NewBuffer())
            gomega.Eventually(session).Should(gexec.Exit(0))
            
            output := string(session.Out.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring("Dry run completed"))
            gomega.Expect(output).To(gomega.ContainSubstring("twiggit() {"))
        })

        ginkgo.It("should detect current shell", func() {
            env := []string{"SHELL=/bin/bash"}
            session := gexec.Start(gexec.Command(binPath, "setup-shell", "--dry-run"), 
                gexec.NewBuffer(), gexec.NewBuffer(), env...)
            gomega.Eventually(session).Should(gexec.Exit(0))
            
            output := string(session.Out.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring("bash"))
        })

        ginkgo.It("should handle unsupported shell", func() {
            env := []string{"SHELL=/bin/sh"}
            session := gexec.Start(gexec.Command(binPath, "setup-shell"), 
                gexec.NewBuffer(), gexec.NewBuffer(), env...)
            gomega.Eventually(session).ShouldNot(gexec.Exit(0))
            
            output := string(session.Err.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring("unsupported shell"))
        })
    })

    ginkgo.Context("when using force flag", func() {
        ginkgo.It("should reinstall wrapper", func() {
            // First installation
            env := []string{"SHELL=/bin/bash", "HOME=" + tempDir}
            session := gexec.Start(gexec.Command(binPath, "setup-shell"), 
                gexec.NewBuffer(), gexec.NewBuffer(), env...)
            gomega.Eventually(session).Should(gexec.Exit(0))

            // Force reinstall
            session = gexec.Start(gexec.Command(binPath, "setup-shell", "--force"), 
                gexec.NewBuffer(), gexec.NewBuffer(), env...)
            gomega.Eventually(session).Should(gexec.Exit(0))
            
            output := string(session.Out.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring("installed successfully"))
        })
    })
})
```

#### 4.5 Shell Wrapper Functionality Tests

**File**: `test/e2e/shell_wrapper_test.go`

```go
//go:build e2e
// +build e2e

package e2e

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/onsi/ginkgo/v2"
    "github.com/onsi/gomega"
    "github.com/onsi/gomega/gexec"
)

var _ = ginkgo.Describe("Shell Wrapper Functionality", func() {
    var (
        binPath    string
        tempDir    string
        worktreeDir string
    )

    ginkgo.BeforeEach(func() {
        var err error
        tempDir = ginkgo.GinkgoT().TempDir()
        worktreeDir = filepath.Join(tempDir, "worktrees", "test-project", "feature-branch")
        
        // Create test worktree directory
        err = os.MkdirAll(worktreeDir, 0755)
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
        
        // Build test binary
        binPath, err = gexec.Build("github.com/twiggit/twiggit")
        gomega.Expect(err).NotTo(gomega.HaveOccurred())
    })

    ginkgo.AfterEach(func() {
        gexec.CleanupBuildArtifacts()
    })

    ginkgo.Context("when using twiggit cd command", func() {
        ginkgo.It("should output worktree path", func() {
            session := gexec.Start(gexec.Command(binPath, "cd", "feature-branch"), 
                gexec.NewBuffer(), gexec.NewBuffer())
            gomega.Eventually(session).Should(gexec.Exit(0))
            
            output := string(session.Out.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring(worktreeDir))
        })

        ginkgo.It("should handle non-existent worktree", func() {
            session := gexec.Start(gexec.Command(binPath, "cd", "nonexistent"), 
                gexec.NewBuffer(), gexec.NewBuffer())
            gomega.Eventually(session).ShouldNot(gexec.Exit(0))
            
            output := string(session.Err.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring("worktree not found"))
        })
    })

    ginkgo.Context("when using shell wrapper", func() {
        ginkgo.It("should intercept cd commands", func() {
            // Create a test script that simulates the wrapper
            wrapperScript := filepath.Join(tempDir, "wrapper_test.sh")
            scriptContent := `#!/bin/bash
twiggit() {
    if [[ "$1" == "cd" ]]; then
        local target_dir
        target_dir=$(command ` + binPath + ` "${@:2}")
        if [[ $? -eq 0 && -n "$target_dir" ]]; then
            echo "Would change to: $target_dir"
        else
            return $?
        fi
    else
        command ` + binPath + ` "$@"
    fi
}

# Test the wrapper
twiggit cd feature-branch
`
            
            err := os.WriteFile(wrapperScript, []byte(scriptContent), 0755)
            gomega.Expect(err).NotTo(gomega.HaveOccurred())

            session := gexec.Start(gexec.Command("bash", wrapperScript), 
                gexec.NewBuffer(), gexec.NewBuffer())
            gomega.Eventually(session).Should(gexec.Exit(0))
            
            output := string(session.Out.Contents())
            gomega.Expect(output).To(gomega.ContainSubstring("Would change to:"))
            gomega.Expect(output).To(gomega.ContainSubstring(worktreeDir))
        })
    })
})
```

### Shell Integration Test Fixtures

#### 4.6 Shell Test Fixtures

**File**: `test/fixtures/shell/`

```
test/fixtures/shell/
├── configs/
│   ├── bashrc.example          # Example bash configuration
│   ├── zshrc.example           # Example zsh configuration
│   └── fish.config.example     # Example fish configuration
├── scripts/
│   ├── wrapper_bash.sh         # Bash wrapper template
│   ├── wrapper_zsh.sh          # Zsh wrapper template
│   └── wrapper_fish.sh         # Fish wrapper template
└── environments/
    ├── bash_env.sh             # Bash test environment
    ├── zsh_env.sh              # Zsh test environment
    └── fish_env.sh             # Fish test environment
```

**Example fixture content**:

```bash
# test/fixtures/shell/configs/bashrc.example
# Example bash configuration for testing
export PS1="\u@\h:\w$ "
alias ll="ls -la"

# Space for twiggit wrapper installation
```

```bash
# test/fixtures/shell/scripts/wrapper_bash.sh
# Twiggit bash wrapper template for testing
twiggit() {
    if [[ "$1" == "cd" ]]; then
        local target_dir
        target_dir=$(command twiggit "${@:2}")
        if [[ $? -eq 0 && -n "$target_dir" ]]; then
            builtin cd "$target_dir"
        else
            return $?
        fi
    elif [[ "$1" == "cd" && "$2" == "--help" ]]; then
        command twiggit "$@"
    else
        command twiggit "$@"
    fi
}
```

This comprehensive shell integration testing infrastructure ensures that the shell wrapper functionality works correctly across all supported shells and integrates seamlessly with the existing twiggit workflow.

This comprehensive testing infrastructure ensures twiggit meets the highest quality standards while maintaining the pragmatic TDD approach outlined in the testing philosophy.