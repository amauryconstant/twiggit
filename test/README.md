# Testing Infrastructure

This directory contains the testing infrastructure for the twiggit project.

## Overview

The testing infrastructure provides organized utilities and patterns for testing across all levels:

- **Structured Organization**: Clear separation between helpers, fixtures, integration tests, and E2E tests
- **Comprehensive Coverage**: Support for unit, integration, and end-to-end testing
- **Utility-Focused**: Provides essential helpers and fixtures without over-engineering

For detailed testing philosophy, quality standards, and framework strategy, see [../.ai/testing.md](../.ai/testing.md).

## Directory Structure

```
test/
├── e2e/                # End-to-end CLI tests
│   ├── create_test.go       # Create command E2E tests
│   ├── delete_test.go       # Delete command E2E tests
│   ├── e2e_suite_test.go    # E2E test suite setup
│   ├── global_cli_test.go   # Global CLI behavior tests
│   ├── help_test.go         # Help command tests
│   ├── list_test.go         # List command tests
│   └── cd_test.go           # CD command tests
├── fixtures/           # Test data and test cases
│   ├── domain_cases.go      # Domain layer test cases
│   └── infrastructure_cases.go # Infrastructure layer test cases
├── helpers/            # Test utilities and helpers
│   ├── cli.go              # CLI testing utilities for E2E tests
│   ├── git.go              # Git repository testing utilities
│   ├── mise.go             # Mise integration testing utilities
│   ├── temp.go             # Temporary directory management
│   └── assertion_helpers.go # Custom assertion helpers
├── integration/        # Integration tests
│   ├── mise_integration_test.go
│   └── worktree_integration_test.go
└── mocks/              # Mock implementations
    └── git_client_mock.go
```

## Key Components

### Helpers Package (`test/helpers/`)

The helpers package provides essential testing utilities:

#### Git Utilities (`git.go`)
- `GitRepo`: Test git repository with automatic cleanup
- `NewGitRepo()`: Creates a new git repository with initial commit
- `NewGitRepoWithBranches()`: Creates a repository with multiple branches
- `CreateBranch()`: Creates a new branch with content
- `AddMiseConfig()`: Adds mise configuration to repository

#### Mise Utilities (`mise.go`)
- `MiseIntegration`: Mise tool integration testing
- `DetectConfigFiles()`: Detects mise configuration files
- `CopyConfigFiles()`: Copies configuration between directories
- `IsAvailable()`: Checks if mise is available

#### Temporary Directory Management (`temp.go`)
- `TempDir()`: Creates temporary directories with automatic cleanup
- `TempFile()`: Creates temporary files with automatic cleanup

#### CLI Utilities (`cli.go`)
- `TwiggitCLI`: CLI wrapper for end-to-end testing
- `NewTwiggitCLI()`: Creates new CLI test wrapper
- `Run()`: Executes CLI commands with gexec session
- `RunWithDir()`: Executes CLI commands in specific directory



### Fixtures Package (`test/fixtures/`)

The fixtures package contains test data and test cases:

- `domain_cases.go`: Test cases for domain layer validation
- `infrastructure_cases.go`: Test cases for infrastructure layer components

### E2E Tests (`test/e2e/`)

End-to-end tests verify CLI functionality from user perspective:

- `create_test.go`: Create command E2E tests
- `delete_test.go`: Delete command E2E tests
- `e2e_suite_test.go`: E2E test suite setup and configuration
- `global_cli_test.go`: Global CLI behavior tests
- `help_test.go`: Help command tests
- `list_test.go`: List command tests
- `cd_test.go`: CD command tests

### Integration Tests (`test/integration/`)

Integration tests verify component interactions:

- `mise_integration_test.go`: Mise integration tests
- `worktree_integration_test.go`: Worktree lifecycle integration tests

### Mocks (`test/mocks/`)

Mock implementations for testing:

- `git_client_mock.go`: Mock git client for unit tests

## Usage Patterns

For detailed testing patterns, framework usage, and examples, see [../.ai/testing.md](../.ai/testing.md). This section focuses on infrastructure-specific usage.

### Git Repository Testing

```go
func TestGitOperation(t *testing.T) {
    // Create test repository
    repo := helpers.NewGitRepo(t, "test-repo-*")
    defer repo.Cleanup()

    // Create branches for testing
    repo.CreateBranch(t, "feature-branch")
    
    // Test your git operations
    err := YourGitOperation(repo.Path)
    assert.NoError(t, err)
}
```

### E2E Testing

```go
var _ = Describe("Create Command", func() {
    var cli *helpers.TwiggitCLI

    BeforeEach(func() {
        cli = helpers.NewTwiggitCLI()
    })

    It("creates worktree from existing branch", func() {
        session := cli.Run("create", "feature-branch")
        Eventually(session).Should(gexec.Exit(0))

        output := string(session.Out.Contents())
        Expect(output).To(ContainSubstring("Created worktree"))
    })
}
```

### Integration Testing

```go
type WorkflowIntegrationTestSuite struct {
    suite.Suite
    testRepo *helpers.GitRepo
    workspace string
    cleanup   func()
}

func (s *WorkflowIntegrationTestSuite) SetupSuite() {
    if testing.Short() {
        s.T().Skip("Skipping integration test")
    }
    
    // Create test environment
    s.workspace, s.cleanup = helpers.TempDir(s.T(), "workspace-*")
    s.testRepo = helpers.NewGitRepo(s.T(), "test-repo-*")
}

func (s *WorkflowIntegrationTestSuite) TearDownSuite() {
    if s.testRepo != nil {
        s.testRepo.Cleanup()
    }
    if s.cleanup != nil {
        s.cleanup()
    }
}

func TestWorkflowIntegrationSuite(t *testing.T) {
    suite.Run(t, new(WorkflowIntegrationTestSuite))
}

func (s *WorkflowIntegrationTestSuite) TestFullWorkflow() {
    // Set up test repositories and worktrees
    // ... test setup using s.testRepo and s.workspace ...

    // Test the complete workflow
    err := YourWorkflowFunction(s.workspace)
    s.Assert().NoError(err)
}
```

## Migration from Old Structure

The old `internal/testutil` structure has been replaced with this simplified approach:

### Key Changes

1. **Removed Over-Engineered Components**:
   - `NewTableTestRunner()` → Standard `for _, tt := range testCases` loops within Testify suites
   - Complex suite structures → Consistent Testify suite pattern across unit/integration tests
   - Over-abstracted utilities → Focused, single-purpose helpers

2. **Simplified Imports**:
   - `github.com/amaury/twiggit/internal/testutil` → `github.com/amaury/twiggit/test/helpers`
   - `github.com/amaury/twiggit/internal/testutil/git` → `github.com/amaury/twiggit/test/helpers`
   - Added consistent Testify suite usage for unit and integration tests

3. **Standard Patterns**:
   - Custom test runners → Standard Go table-driven tests within Testify suites
   - Complex setup/teardown → Consistent Testify suite SetupTest/TearDownTest patterns
   - Mixed testing frameworks → Consistent framework usage per test type

### Benefits

- **Reduced Complexity**: ~70% reduction in testing infrastructure code
- **Improved Maintainability**: Standard patterns are easier to understand and modify
- **Better Performance**: Less overhead from complex abstractions
- **Enhanced Readability**: Tests follow familiar Go patterns

## Running Tests

### Unit Tests
```bash
go test ./...
```

### E2E Tests
```bash
go test -tags=e2e ./test/e2e/...
```

### Integration Tests
```bash
go test -tags=integration ./test/integration/...
```

### Test Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Specific Test Packages
```bash
go test ./internal/domain/...
go test ./internal/infrastructure/...
go test ./internal/services/...
```

### Using Mise Tasks
```bash
mise run test          # Run all tests (unit + integration + E2E + race)
mise run test:unit     # Run unit tests only
mise run test:integration # Run integration tests only
mise run test:e2e      # Run E2E tests only
mise run test:race     # Run tests with race detection
```

## Best Practices

For comprehensive testing best practices, framework selection, and quality standards, see [../.ai/testing.md](../.ai/testing.md).

### Infrastructure-Specific Guidelines

1. **Place Utilities Correctly**: 
   - General utilities → `test/helpers/`
   - Test data → `test/fixtures/`
   - Integration tests → `test/integration/`
   - E2E tests → `test/e2e/`

2. **Use Helpers Appropriately**:
   - `git.go`: For git repository setup and operations
   - `temp.go`: For temporary directory/file management
   - `cli.go`: For E2E CLI testing only
   - `mise.go`: For mise integration testing

3. **Keep Infrastructure Simple**:
   - Avoid over-engineering test utilities
   - Focus on single responsibility per helper
   - Prefer composition over inheritance

4. **Document Thoroughly**:
   - Include comprehensive godoc comments
   - Add example usage in comments
   - Explain cleanup responsibilities

## Contributing

When adding new test utilities, follow the infrastructure-specific guidelines in the Best Practices section above. For testing philosophy and framework decisions, refer to [../.ai/testing.md](../.ai/testing.md).