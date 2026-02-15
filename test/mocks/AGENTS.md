# Test Mocks

Test doubles for unit/integration tests using testify/mock.

## Mocks

| Mock | Implements | File |
|------|------------|------|
| `MockWorktreeService` | `application.WorktreeService` | `cmd_mocks.go` |
| `MockProjectService` | `application.ProjectService` | `cmd_mocks.go` |
| `MockContextService` | `application.ContextService` | `cmd_mocks.go` |
| `MockNavigationService` | `application.NavigationService` | `cmd_mocks.go` |
| `MockShellService` | `application.ShellService` | `cmd_mocks.go` |
| `MockGitClient` | `infrastructure.GitClient` | `git_service_mock.go` |
| `MockShellInfrastructure` | `infrastructure.ShellInfrastructure` | `shell_infrastructure_mock.go` |
| `MockContextDetector` | `domain.ContextDetector` | `mock_context_detector.go` |
| `MockContextResolver` | `domain.ContextResolver` | `mock_context_resolver.go` |

## Usage Pattern

```go
func (s *MyTestSuite) SetupTest() {
    s.mock = mocks.NewMockWorktreeService()
    s.mock.On("CreateWorktree", mock.Anything, mock.Anything).
        Return(&domain.WorktreeInfo{Path: "/tmp/wt", Branch: "feature"}, nil)
}

func (s *MyTestSuite) TestCreate() {
    result, err := s.service.CreateWorktree(ctx, req)
    s.Require().NoError(err)
    s.mock.AssertCalled(s.T(), "CreateWorktree", ctx, req)
}
```

## Adding New Mock

1. Create `<name>_mock.go` in `test/mocks/`
2. Embed `mock.Mock` in struct
3. Implement interface methods with `m.Called()` + type assertions
4. Add `NewMock<Name>()` constructor
