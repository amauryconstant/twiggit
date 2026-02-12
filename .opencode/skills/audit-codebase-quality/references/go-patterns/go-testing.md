# Go Testing Patterns

Go-specific patterns for testing frameworks, organization, and mock usage in codebase auditing.

## Frameworks

- **Testify**: `github.com/stretchr/testify` - most common
  - `assert` package for assertions
  - `require` package for fatal assertions
  - `suite` package for test suites
- **Table-driven tests**: Subtests with `t.Run()` for different scenarios
- **Mocking**: Functional mocks (function fields) or `mock.Mock` from testify

## Test Organization

- **Test alongside source**: `file_test.go` next to `file.go`
- **Package tests**: `_test.go` suffix
- **Build tags**: `//go:build integration` for integration tests
- **Short mode**: `if testing.Short() { t.Skip() }`

## Common Patterns

```go
// Table-driven test
func TestCreateWorktree(t *testing.T) {
    testCases := []struct {
        name string
        input *domain.CreateWorktreeRequest
        wantErr bool
    }{
        {
            name: "valid request",
            input: &domain.CreateWorktreeRequest{...},
            wantErr: false,
        },
        {
            name: "invalid project name",
            input: &domain.CreateWorktreeRequest{ProjectName: ""},
            wantErr: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := service.CreateWorktree(context.Background(), tc.input)
            if tc.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

// Mock setup
func setupTestWorktreeService() application.WorktreeService {
    gitClient := mocks.NewMockGitClient()
    projectService := mocks.NewMockProjectService()
    config := domain.DefaultConfig()
    return NewWorktreeService(gitClient, projectService, config)
}
```

## Anti-Patterns

- **Inline mocks**: Mocks defined in test files instead of centralized `test/mocks/`
- **Test logic in tests**: Complex setup/teardown that should be helpers
- **Brittle assertions**: Checking exact error messages instead of error types

## Audit-Specific Patterns

### For Test Pattern Audits

- Verify all test files follow `<source>_test.go` naming
- Check for inline mocks that should be in `test/mocks/`
- Identify duplicate mock implementations
- Find files without corresponding tests (if logic exists)

### For Mock Centralization Audits

- Check for inline mock definitions in test files
- Identify duplicate mock implementations for same interface
- Verify centralized mocks exist in `test/mocks/`
- Check for mock naming inconsistencies
