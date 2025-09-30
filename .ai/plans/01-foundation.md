# Foundation Layer Implementation Plan

## Overview

This plan establishes the minimal viable foundation for twiggit, implementing core domain entities and basic project structure following true TDD principles. The foundation layer focuses on essential building blocks only, deferring all non-essential functionality to appropriate later phases.

## Foundation Principles

### Incremental TDD Approach
- **Test First**: Write failing tests, then implement minimal code to pass
- **Red-Green-Refactor**: Follow strict TDD cycle for each feature
- **Minimal Implementation**: Implement only what's needed to pass current tests
- **No Future-Proofing**: Avoid implementing features for future phases

### Simple Domain Model
- **Basic Entities Only**: Project and Worktree with simple validation
- **No Complex Relationships**: Defer entity relationships to service layer (Phase 5)
- **Validation Focus**: Domain entities handle input validation only
- **Pure Functions**: Keep domain logic simple and testable

### Error Handling Strategy
- **Simple wrapped errors**: Use `fmt.Errorf("operation: %w", err)` consistently
- **Contextual messages**: Always include operation context in error messages
- **No custom error types**: Keep error handling simple for now

## Phase Boundaries

### Phase 1 Scope
- Domain entities with basic validation (Project, Worktree)
- Basic project structure following Go standards
- Unit testing setup for domain entities
- Quality assurance configuration

### Deferred to Later Phases
- CLI commands (Phase 6)
- Configuration system (Phase 2)
- Git operations (Phase 4)
- Context detection (Phase 3)
- Service layer (Phase 5)
- Infrastructure interfaces (Phase 4)
- Dependency injection (Phase 5)
- Integration/E2E testing (Phase 8)

## Project Structure

Phase 1 minimal structure following Go standards:

```
cmd/
├── root.go          # Basic root command only
└── main.go          # Entry point

internal/
├── domain/          # Core entities only
│   ├── project.go
│   └── worktree.go

go.mod
go.sum
.golangci.yml
.gitignore
```

**Removed from Phase 1** (deferred to later phases):
- Individual command files (create.go, delete.go, etc.) → Phase 6
- infrastructure/ directory → Phase 4
- services/ directory → Phase 5
- di/ directory → Phase 5
- test/ directory structure → Phase 8

## Implementation Steps

### Step 1: Initialize Go Module and Dependencies

**Files to create:**
- `go.mod`
- `go.sum` (generated)
- `.gitignore`
- `.golangci.yml`

**Dependencies** (minimal for Phase 1):
```bash
go mod init twiggit
go get github.com/stretchr/testify@latest
```

**Note**: Additional dependencies will be added in appropriate phases as needed

### Step 2: Create Core Domain Entities

#### 2.1 Project Entity
**File:** `internal/domain/project.go`

**Tests first:** `internal/domain/project_test.go`

```go
func TestProject_NewProject(t *testing.T) {
    testCases := []struct {
        name         string
        projectName  string
        path         string
        expectError  bool
        errorMessage string
    }{
        {
            name:        "valid project",
            projectName: "my-project",
            path:        "/path/to/project",
            expectError: false,
        },
        {
            name:         "empty project name",
            projectName:  "",
            path:         "/path/to/project",
            expectError:  true,
            errorMessage: "new project: name cannot be empty",
        },
        {
            name:         "empty project path",
            projectName:  "my-project",
            path:         "",
            expectError:  true,
            errorMessage: "new project: path cannot be empty",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            project, err := NewProject(tc.projectName, tc.path)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMessage)
                assert.Nil(t, project)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, project)
                assert.Equal(t, tc.projectName, project.Name())
                assert.Equal(t, tc.path, project.Path())
            }
        })
    }
}
```

**Implementation:**
```go
package domain

import "fmt"

// Project represents a git project with basic validation
type Project struct {
    name string
    path string
}

// NewProject creates a new project with validation
func NewProject(name, path string) (*Project, error) {
    if name == "" {
        return nil, fmt.Errorf("new project: name cannot be empty")
    }
    if path == "" {
        return nil, fmt.Errorf("new project: path cannot be empty")
    }
    
    return &Project{
        name: name,
        path: path,
    }, nil
}

// Name returns the project name
func (p *Project) Name() string {
    return p.name
}

// Path returns the project filesystem path
func (p *Project) Path() string {
    return p.path
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
        expectError  bool
        errorMessage string
    }{
        {
            name:        "valid worktree",
            path:        "/home/user/Workspaces/project/feature",
            branch:      "feature",
            expectError: false,
        },
        {
            name:         "empty path",
            path:         "",
            branch:       "feature",
            expectError:  true,
            errorMessage: "new worktree: path cannot be empty",
        },
        {
            name:         "empty branch",
            path:         "/home/user/Workspaces/project/feature",
            branch:       "",
            expectError:  true,
            errorMessage: "new worktree: branch cannot be empty",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            worktree, err := NewWorktree(tc.path, tc.branch)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMessage)
                assert.Nil(t, worktree)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, worktree)
                assert.Equal(t, tc.path, worktree.Path())
                assert.Equal(t, tc.branch, worktree.Branch())
            }
        })
    }
}
```

**Implementation:**
```go
package domain

import "fmt"

// Worktree represents a git worktree with basic validation
type Worktree struct {
    path   string
    branch string
}

// NewWorktree creates a new worktree with validation
func NewWorktree(path, branch string) (*Worktree, error) {
    if path == "" {
        return nil, fmt.Errorf("new worktree: path cannot be empty")
    }
    if branch == "" {
        return nil, fmt.Errorf("new worktree: branch cannot be empty")
    }
    
    return &Worktree{
        path:   path,
        branch: branch,
    }, nil
}

// Path returns the worktree filesystem path
func (w *Worktree) Path() string {
    return w.path
}

// Branch returns the worktree branch name
func (w *Worktree) Branch() string {
    return w.branch
}
```

### Step 3: Basic Project Structure

#### 3.1 Main Entry Point
**File:** `cmd/main.go`

```go
package main

import "twiggit/cmd"

func main() {
    cmd.Execute()
}
```

#### 3.2 Basic Root Command
**File:** `cmd/root.go`

**Tests first:** `cmd/root_test.go`

```go
func TestRootCommand_BasicProperties(t *testing.T) {
    assert.Equal(t, "twiggit", rootCmd.Use)
    assert.Equal(t, "Pragmatic git worktree management tool", rootCmd.Short)
    assert.Contains(t, rootCmd.Long, "git worktrees")
}
```

**Implementation:**
```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
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
```

### Step 4: Quality Assurance Setup

#### 4.1 Git Ignore
**File:** `.gitignore`

```gitignore
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db
```

#### 4.2 Linting Configuration
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
    - ineffassign
    - typecheck
    - gosec
    - misspell
    - unconvert
    - goconst

linters-settings:
  goconst:
    min-len: 3
    min-occurrences: 3
```

## Implementation Order

1. **Step 1**: Go module, dependencies, and quality setup
2. **Step 2**: Core domain entities (Project, Worktree) with TDD
3. **Step 3**: Basic CLI structure (root command only)
4. **Step 4**: Quality assurance configuration

## Testing Strategy

Phase 1 focuses exclusively on unit testing for domain entities.

### Unit Tests Only
- **Framework**: Testify with table-driven tests
- **Coverage**: >60% for domain entities (realistic for basic validation)
- **Location**: `*_test.go` files alongside implementation
- **Focus**: Input validation, basic entity behavior

### Deferred Testing Types
- **Integration Tests**: Phase 8 (when real git operations exist)
- **E2E Tests**: Phase 8 (when CLI commands exist)
- **Performance Tests**: Phase 9

## Quality Gates

### Pre-commit Requirements
- All tests pass: `go test ./...`
- Linting passes: `golangci-lint run`
- Coverage >60%: `go test -cover ./...`

### CI Requirements
- Unit tests pass
- Linting passes
- Build succeeds on target platforms

## Key Principles

### TDD Approach
- **Write failing test first**
- **Implement minimal code to pass**
- **Refactor while keeping tests green**
- **Repeat for next feature**

### Simplicity First
- **YAGNI**: Implement only what's needed now
- **KISS**: Keep solutions simple and direct
- **No premature optimization**: Defer optimizations to Phase 9

### Clean Code
- **Clear naming**: Use descriptive names
- **Small functions**: Keep functions focused and short
- **Consistent error handling**: Use same error pattern throughout

## Success Criteria

1. ✅ Domain entities (Project, Worktree) with comprehensive validation
2. ✅ Project compiles and runs `twiggit --help`
3. ✅ Unit tests for domain entities pass with >60% coverage
4. ✅ Basic linting passes without errors
5. ✅ Clean project structure following Go standards
6. ✅ Quality gates enforce basic code standards

## Incremental Development Strategy

Phase 1 follows strict incremental development:

1. **Write Test**: Create failing test for specific feature
2. **Implement**: Add minimal code to make test pass
3. **Refactor**: Improve code while keeping tests green
4. **Repeat**: Move to next small feature

**No stubs, no scaffolding, no future-proofing.** Each phase builds only what's needed for that phase.

## Next Phases

Phase 1 provides the minimal foundation needed for sequential development:

1. **Phase 2**: Configuration system (Koanf/TOML)
2. **Phase 3**: Context detection system
3. **Phase 4**: Infrastructure layer (git operations)
4. **Phase 5**: Service layer and dependency injection
5. **Phase 6**: CLI commands implementation
6. **Phase 7**: Shell integration features
7. **Phase 8**: Comprehensive testing infrastructure
8. **Phase 9**: Performance optimization
9. **Phase 10**: Final integration and validation

This foundation provides the essential base needed for building robust features while following true TDD principles and maintaining clean phase boundaries.