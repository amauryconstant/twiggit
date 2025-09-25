# Testing Philosophy and Patterns

## Testing Philosophy

### Pragmatic Test-Driven Development (TDD)

- **Write tests BEFORE implementation**: Tests WILL serve as safety net for refactoring, not as bureaucracy
- **Focus on integration tests that verify user workflows**: Tests WILL verify how components work together
- **Unit tests for complex algorithms, not simple getters/setters**: Tests WILL focus on business logic that matters
- **E2E tests for CLI commands**: Tests WILL ensure user-facing functionality works correctly

### Testing Hierarchy

#### Unit Tests
- **Purpose**: Test individual components in isolation
- **Scope**: Business logic, algorithms, complex calculations
- **Approach**: Good coverage with mocked dependencies, test behavior not implementation
- **Tools**: Testify for assertions and mock generation
- **WILL Provide**: Fast execution, isolated environment, repeatable results
- **When to use**: For complex business logic, validation rules, algorithms

#### Integration Tests
- **Purpose**: Test component interactions and workflows
- **Scope**: Real git repositories in temporary directories
- **Approach**: Test how services work together with real dependencies using Testify suites
- **Tools**: Testify suites with build tags to separate from unit tests, skip in short mode with `testing.Short()`
- **WILL Provide**: Slower execution than unit tests, but still isolated environment
- **When to use**: For service interactions, git operations, configuration workflows

#### E2E Tests
- **Purpose**: Test CLI commands from user perspective
- **Scope**: Complete user workflows and error scenarios
- **Approach**: Build actual binary and execute commands
- **Tools**: Ginkgo/Gomega for CLI interaction testing, gexec for process management
- **WILL Provide**: Slowest but most realistic execution, complete user experience testing
- **When to use**: For complete CLI workflows, user interaction testing, error scenarios



## Quality Standards

### Coverage Requirements

- **Overall test coverage**: SHOULD exceed 80%
- **CI Pipeline Enforcement**: Coverage threshold is automatically enforced in CI pipeline - builds WILL fail if coverage drops below 80%
- **Happy path coverage**: SHALL have 100% coverage for expected behavior and critical business logic
- **Error handling coverage**: SHOULD test error paths but 100% coverage not required for all scenarios
- **Package-specific strategies**:
  - **cmd/ package**: Tested via E2E tests only (no coverage monitoring)
  - **Infrastructure packages**: Integration tests preferred for external dependencies
  - **Utility packages**: Unit tests with appropriate mocking/temporary files
  - **Services packages**: Focus on happy paths and critical business logic

### Current Coverage Status

As of the latest implementation, the following packages meet or exceed coverage requirements:

- **Version Package**: 100.0% coverage ✅ 
- **Mise Infrastructure**: 90.4% coverage ✅
- **Services Layer**: 82.0% coverage ✅ (exceeds 80% target)
- **Domain Package**: 89.7% coverage ✅
- **Config Infrastructure**: 83.3% coverage ✅
- **Validation Infrastructure**: 95.2% coverage ✅
- **DI Package**: 100.0% coverage ✅
- **Git Infrastructure**: 57.3% coverage (below target, not in scope for current requirements)

### Coverage Monitoring Process

1. **Local Development**: Use `mise run test:unit` to generate coverage reports locally
2. **CI Pipeline**: Automatic coverage threshold enforcement during merge requests and main branch pushes
3. **Coverage Reports**: HTML and XML coverage reports are generated and available as CI artifacts
4. **Threshold Check**: Pipeline fails with clear error message if coverage drops below 80%
5. **Coverage Merging**: Unit and integration test coverage profiles are automatically merged for comprehensive reporting

### Test Quality

- **Tests SHOULD be maintainable**: Clear structure, good naming, minimal duplication
- **Tests SHOULD be reliable**: Consistent results, no flakiness
- **Tests SHOULD be fast**: Unit tests in milliseconds, integration tests in seconds
- **Tests SHOULD provide good feedback**: Clear failure messages, helpful debugging information

## Testing Commands

### Complete Testing Command Reference

- `mise run test` - Run all tests (unit + integration + E2E + race)
- `mise run test:unit` - Run unit tests only
- `mise run test:integration` - Run integration tests only
- `mise run test:e2e` - Run CLI end-to-end tests
- `mise run test:race` - Run tests with race condition detection
- `mise run test:single` - Run single test (usage: `mise run test:single TestName ./pkg/module`)

### Build Commands for Testing

- `mise run build:cli` - Build CLI binary
- `mise run build:e2e` - Build CLI binary for E2E tests
- `mise run build:clean` - Clean build artifacts

## Testing Framework Strategy

### Framework Selection Rationale

- **Unit Tests**: SHALL use Testify suite pattern for consistency and better setup/teardown management
- **Integration Tests**: SHALL use Testify suites with build tags for structured test organization and consistent assertions
- **E2E Tests**: SHALL use Ginkgo/Gomega for CLI interaction testing and user workflow descriptions

### Mock Strategy Guidelines

- **Service Tests**: SHALL use centralized mocks from `test/mocks/` package to promote reuse and consistency
- **Domain Tests**: SHOULD use inline mocks within test files to keep them self-contained and avoid over-engineering
- **Infrastructure Tests**: MAY use either approach based on complexity and reuse needs

### Test Structure Standards

- **Unit Tests**: SHALL follow table-driven test patterns within Testify suites
- **Integration Tests**: SHALL use real git repositories in temporary directories with proper cleanup
- **E2E Tests**: SHALL build actual binaries and test complete user workflows

## Concrete Test Examples

For comprehensive Go code patterns and testing structure, see code-style-guide.md. This section provides concrete examples aligned with the testing philosophy.

### Unit Test Example (Testify Suite)
```go
type ProjectTestSuite struct {
    suite.Suite
    Project   *Project
    Workspace *Workspace
}

func (s *ProjectTestSuite) SetupTest() {
    var err error
    s.Project, err = NewProject("test-project", "/repo/path")
    s.Require().NoError(err)
    
    s.Workspace, err = NewWorkspace("/test/workspace")
    s.Require().NoError(err)
}

func TestProjectSuite(t *testing.T) {
    suite.Run(t, new(ProjectTestSuite))
}

func (s *ProjectTestSuite) TestProject_NewProject() {
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
    
    for _, tt := range testCases {
        s.Run(tt.name, func() {
            project, err := NewProject(tt.projectName, tt.gitRepo)
            
            if tt.expectError {
                s.Assert().Error(err)
                s.Assert().Contains(err.Error(), tt.errorMessage)
            } else {
                s.Assert().NoError(err)
                s.Assert().Equal(tt.projectName, project.Name())
                s.Assert().Equal(tt.gitRepo, project.GitRepoPath())
            }
        })
    }
}
```

### Integration Test Example (Testify Suite)
```go
//go:build integration
// +build integration

package integration

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type WorktreeIntegrationTestSuite struct {
    suite.Suite
    testRepo        *IntegrationTestRepo
    worktreeCreator *services.WorktreeCreator
}

func (s *WorktreeIntegrationTestSuite) SetupSuite() {
    // Skip if not in integration test mode
    if testing.Short() {
        s.T().Skip("Skipping integration test")
    }
    
    // Create test git repository
    s.testRepo = NewTestGitRepo(s.T())
    
    // Initialize services with real dependencies
    s.worktreeCreator = services.NewWorktreeCreator(/* real dependencies */)
}

func (s *WorktreeIntegrationTestSuite) TearDownSuite() {
    if s.testRepo != nil {
        s.testRepo.Cleanup()
    }
}

func TestWorktreeIntegrationSuite(t *testing.T) {
    suite.Run(t, new(WorktreeIntegrationTestSuite))
}

func (s *WorktreeIntegrationTestSuite) TestCreateWorktreeFromExistingBranch() {
    worktreePath := filepath.Join(filepath.Dir(s.testRepo.RepoDir()), "feature-worktree")
    
    err := s.worktreeCreator.Create(context.Background(), s.testRepo.RepoDir(), "feature-1", worktreePath)
    s.Assert().NoError(err)
    
    // Verify worktree was created
    _, err = os.Stat(worktreePath)
    s.Assert().NoError(err, "Worktree directory should exist")
}
```

## Testing Anti-Patterns

### Common Testing Mistakes

1. **Testing implementation details**: Tests SHOULD NOT test implementation details; test behavior instead
2. **Over-mocking**: Tests SHOULD NOT over-mock; mock only what you need to control and use real objects when possible
3. **Ignoring error paths**: Tests SHOULD NOT ignore error paths; test both success and failure scenarios
4. **Flaky tests**: Tests SHOULD NOT be flaky; ensure tests are deterministic and reliable
5. **Slow unit tests**: Unit tests SHOULD NOT be slow; move slow operations to integration tests
6. **Coverage-driven testing**: Tests SHOULD NOT focus on coverage percentage; focus on meaningful tests for happy paths and critical business logic
7. **Wrong test type for package**: Tests SHOULD NOT use unit tests for CLI commands or integration tests for simple utilities; match test type to package responsibility

### Test Organization Anti-Patterns

1. **Giant test files**: Tests SHOULD NOT use giant test files; split large test files into focused test suites
2. **Inconsistent naming**: Tests SHOULD NOT use inconsistent naming; use consistent naming conventions across all tests
3. **Missing cleanup**: Tests SHOULD NOT have missing cleanup; always clean up resources, even when tests fail
4. **Hardcoded paths**: Tests SHOULD NOT use hardcoded paths; use temporary directories and relative paths
5. **Test interdependence**: Tests SHOULD NOT be interdependent; each test should be independent and isolated

### Testing Philosophy Anti-Patterns

1. **Testing for coverage percentage**: Testing SHOULD NOT focus on coverage percentage; focus on meaningful tests instead
2. **Writing tests after implementation**: Tests SHOULD NOT be written after implementation; tests should drive design, not just verify
3. **Skipping integration tests**: Testing SHOULD NOT skip integration tests; don't rely only on unit tests; test real interactions
4. **Ignoring E2E tests**: Testing SHOULD NOT ignore E2E tests; don't skip user-facing tests; they catch real-world issues
5. **Treating tests as second-class code**: Tests SHOULD NOT be treated as second-class code; tests should be as clean as production code
6. **One-size-fits-all testing**: Testing SHOULD NOT use the same approach for all packages; match test strategy to package responsibility and dependencies
7. **Inconsistent framework usage**: Tests SHALL NOT mix testing frameworks within the same test type; use the designated framework for each test type

### Optional Testing Restrictions
- **Parallel Test Execution**: Tests MAY NOT be executed in parallel if they interfere with each other
- **Test Timeout Configuration**: Individual tests MAY NOT have custom timeout configurations
- **Test Data Management**: External test data MAY NOT be used; all test data SHOULD be generated within tests
- **Test Environment Variables**: Tests MAY NOT rely on specific environment variable configurations

## Shell Integration Testing

### Testing Philosophy
Shell integration tests WILL validate that the context-aware navigation system works correctly across all supported shells and contexts.

### Test Requirements

#### Core Functionality Testing
- Tests SHALL validate context detection for all three context types (project, worktree, outside)
- Tests SHALL verify identifier resolution from all contexts
- Tests SHALL confirm special case handling for `main` branch navigation
- Tests SHALL validate cross-project navigation scenarios

#### Shell Compatibility Testing
- Tests SHALL verify bash shell integration with proper `builtin cd` usage
- Tests SHALL validate zsh shell integration with completion support
- Tests SHALL confirm fish shell integration with completion support
- Tests SHALL ensure escape hatch functionality works in all supported shells

#### Error Handling Testing
- Tests SHALL validate context-aware error messages
- Tests SHALL confirm proper error handling for invalid targets
- Tests SHALL verify that navigation fails gracefully when targets don't exist

### Test Data Requirements
Test environments SHALL include:
- Multiple project repositories with varied worktree structures
- Cross-project navigation scenarios with different project layouts
- Edge cases for context detection and resolution
- Invalid target scenarios for error handling validation

### Quality Standards
- Test coverage SHOULD exceed 80% for all navigation functionality
- Integration tests SHOULD cover real-world usage scenarios
- Tests SHOULD validate both successful and failed navigation attempts
- Performance SHOULD be acceptable for environments with many projects

### Testing Best Practices
- Tests SHOULD use realistic project and worktree structures
- Test data SHOULD represent common developer workflows
- Error scenarios SHOULD include helpful user guidance validation
- Tests SHOULD NOT rely on specific file system layouts beyond the configured paths

## Summary

This testing philosophy provides a comprehensive framework for building high-quality, maintainable tests that provide real value. By focusing on pragmatic TDD, clear testing hierarchy, and consistent patterns, we ensure that tests serve as a safety net for refactoring while providing confidence in the system's correctness. 

The separation between unit, integration, and E2E tests allows us to test at the right level of abstraction for each scenario, balancing speed with realism. Our framework strategy ensures consistency across the codebase while leveraging the strengths of each testing approach: Testify suites for structured unit and integration testing, and Ginkgo/Gomega for expressive E2E workflow testing.

This documentation SHALL be kept current with implementation to ensure that our testing practices remain aligned with our architectural vision and quality standards.