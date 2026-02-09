## Integration Test Structure
Purpose: Test component interactions with real dependencies

## Testify Suite Pattern

```go
//go:build integration

type WorktreeIntegrationTestSuite struct {
    suite.Suite
    testRepo        *helpers.IntegrationTestRepo
    worktreeService *services.WorktreeService
}

func (s *WorktreeIntegrationTestSuite) SetupSuite() {
    if testing.Short() { s.T().Skip() }
    s.testRepo = helpers.NewIntegrationTestRepo(s.T())
    s.worktreeService = services.NewWorktreeService(/* deps */)
}

func (s *WorktreeIntegrationTestSuite) TearDownSuite() {
    if s.testRepo != nil { s.testRepo.Cleanup() }
}
```

## Real Git Repos

**Test Helper:** `helpers.NewIntegrationTestRepo(t)`

Creates temporary git repository in `t.TempDir()` with auto-cleanup. Configured with test user: `test@twiggit.dev` / `Test User`.

```go
repo := helpers.NewIntegrationTestRepo(s.T())
repo.CreateBranch("feature-1")
repo.CreateWorktree("feature-1")
```

## Test Organization

### Unit Tests
- **Framework:** Testify with table-driven tests
- **Focus:** Component logic in isolation
- **Mocking:** Mock external dependencies

### Integration Tests
- **Framework:** Testify suites with `//go:build integration`
- **Focus:** Component interactions
- **Setup:** Real git repos in temp dirs
- **Skip in short mode:** `if testing.Short() { t.Skip() }`

### E2E Tests
- **Framework:** Ginkgo/Gomega + gexec
- **Focus:** CLI user workflows
- **Setup:** Build actual binary, execute commands

## Testing Commands
```bash
mise run test              # All tests
mise run test:unit         # Unit only
mise run test:integration  # Integration only
mise run test:e2e          # E2E only
```
