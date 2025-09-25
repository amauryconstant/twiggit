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
- **Approach**: Test how services work together with real dependencies
- **Tools**: Build tags to separate from unit tests, skip in short mode with `testing.Short()`
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
- **Critical path coverage**: SHALL have 100% coverage for core business logic
- **Error handling coverage**: SHALL test all error paths and edge cases

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

## Concrete Test Examples

For comprehensive Go code patterns and testing structure, see code-style-guide.md. This section provides concrete examples aligned with the testing philosophy.

### Unit Test Example (Testify)
```go
func TestProjectValidator_ValidateName(t *testing.T) {
    validator := NewProjectValidator()
    
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "my-project", false},
        {"empty name", "", true},
        {"invalid chars", "my@project", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateName(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Test Example (Ginkgo/Gomega)
```go
var _ = Describe("WorktreeCreator", func() {
    var tempDir string
    var creator *WorktreeCreator
    
    BeforeEach(func() {
        var err error
        tempDir, err = os.MkdirTemp("", "twiggit-test")
        Expect(err).ToNot(HaveOccurred())
        creator = NewWorktreeCreator()
    })
    
    AfterEach(func() {
        os.RemoveAll(tempDir)
    })
    
    It("should create worktree from existing branch", func() {
        // Setup real git repository
        repoPath := filepath.Join(tempDir, "repo")
        // ... git setup code
        
        worktreePath := filepath.Join(tempDir, "worktree")
        err := creator.CreateWorktree(repoPath, worktreePath, "main")
        
        Expect(err).ToNot(HaveOccurred())
        Expect(worktreePath).To(BeADirectory())
    })
})
```

## Testing Anti-Patterns

### Common Testing Mistakes

1. **Testing implementation details**: Tests SHOULD NOT test implementation details; test behavior instead
2. **Over-mocking**: Tests SHOULD NOT over-mock; mock only what you need to control and use real objects when possible
3. **Ignoring error paths**: Tests SHOULD NOT ignore error paths; test both success and failure scenarios
4. **Flaky tests**: Tests SHOULD NOT be flaky; ensure tests are deterministic and reliable
5. **Slow unit tests**: Unit tests SHOULD NOT be slow; move slow operations to integration tests

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

This testing philosophy provides a comprehensive framework for building high-quality, maintainable tests that provide real value. By focusing on pragmatic TDD, clear testing hierarchy, and consistent patterns, we ensure that tests serve as a safety net for refactoring while providing confidence in the system's correctness. The separation between unit, integration, and E2E tests allows us to test at the right level of abstraction for each scenario, balancing speed with realism.