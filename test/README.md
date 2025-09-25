# Testing Infrastructure

This directory contains the simplified testing infrastructure for the twiggit project.

## Overview

The testing infrastructure has been streamlined to reduce complexity and improve maintainability:

- **Simplified Structure**: Reduced from ~1000 lines of over-engineered infrastructure to ~310 lines of focused utilities
- **Standard Patterns**: Uses standard Go table-driven test patterns instead of custom runners
- **Clear Organization**: Separated concerns between helpers, fixtures, and integration tests

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

#### Assertion Helpers (`assertion_helpers.go`)
- `AssertDirExists()`: Directory existence assertions
- `AssertFileExists()`: File existence assertions
- `AssertFileContains()`: File content assertions

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

### Standard Table-Driven Tests

Instead of using custom test runners, use standard Go patterns:

```go
func TestSomething(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "simple case",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        {
            name:     "error case",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

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
func TestCreateCommand(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test")
    }

    // Create CLI wrapper
    cli := helpers.NewTwiggitCLI()
    
    // Run CLI command
    session := cli.Run("create", "feature-branch")
    defer session.Terminate()
    
    // Verify results
    Expect(session.ExitCode()).To(Equal(0))
    Expect(session.Out.Contents()).To(ContainSubstring("Created worktree"))
}

### Integration Testing

```go
func TestFullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Create test environment
    workspace, cleanup := helpers.TempDir(t, "workspace-*")
    defer cleanup()

    // Set up test repositories and worktrees
    // ... test setup ...

    // Test the complete workflow
    err := YourWorkflowFunction(workspace)
    assert.NoError(t, err)
}
```

## Migration from Old Structure

The old `internal/testutil` structure has been replaced with this simplified approach:

### Key Changes

1. **Removed Over-Engineered Components**:
   - `NewTableTestRunner()` → Standard `for _, tt := range testCases` loops
   - Complex suite structures → Simple testify suites or standard tests
   - Over-abstracted utilities → Focused, single-purpose helpers

2. **Simplified Imports**:
   - `github.com/amaury/twiggit/internal/testutil` → `github.com/amaury/twiggit/test/helpers`
   - `github.com/amaury/twiggit/internal/testutil/git` → `github.com/amaury/twiggit/test/helpers`

3. **Standard Patterns**:
   - Custom test runners → Standard Go table-driven tests
   - Complex setup/teardown → Simple defer cleanup patterns

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

1. **Use Table-Driven Tests**: For testing multiple scenarios with the same logic
2. **Cleanup Resources**: Always use `defer` for cleanup operations
3. **Test Names**: Use descriptive test names that explain the scenario
4. **Test Isolation**: Each test should be independent and not rely on execution order
5. **Mock External Dependencies**: Use mocks for external services and filesystem operations
6. **Integration Tests**: Mark integration tests with build tags and skip in short mode

## Contributing

When adding new test utilities:

1. **Place in Correct Package**: 
   - General utilities → `test/helpers/`
   - Test data → `test/fixtures/`
   - Integration tests → `test/integration/`

2. **Follow Existing Patterns**:
   - Use standard Go patterns
   - Include comprehensive godoc comments
   - Add example usage in comments

3. **Keep It Simple**:
   - Avoid over-engineering
   - Focus on single responsibility
   - Prefer composition over inheritance