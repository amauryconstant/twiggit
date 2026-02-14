## Integration Test Structure
Purpose: Test component interactions with real dependencies

## Testify Suite Pattern (Required)

All unit/integration tests SHALL use Testify suites:

```go
//go:build integration

type MyTestSuite struct {
    suite.Suite
    service MyService
    mockDep *mocks.MockDependency
}

func TestMySuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}

func (s *MyTestSuite) SetupTest() {
    s.mockDep = mocks.NewMockDependency()
    s.service = NewMyService(s.mockDep)
}

func (s *MyTestSuite) TestOperation_Success() {
    tests := []struct {
        name     string
        input    Input
        expected Output
    }{
        {"simple case", Input{...}, Output{...}},
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            result := s.service.Do(tt.input)
            s.Equal(tt.expected, result)
        })
    }
}
```

## Assertion Guidelines

| Use `require.*` (fail fast) | Use `assert.*` (continue) |
|-----------------------------|---------------------------|
| `s.Require().NoError(err)` | `s.Equal(exp, act)` |
| `s.Require().NotNil(obj)` | `s.Contains(str, substr)` |
| `s.Require().Equal(exp, act)` | `s.True(cond)` |

**Testifylint:** Error assertions MUST use `require.Error()` or `require.NoError()`

## Mock Patterns

All mocks centralized in `test/mocks/`:

```go
// test/mocks/my_service_mock.go
type MockMyService struct {
    mock.Mock
}

func (m *MockMyService) Method() error {
    args := m.Called()
    return args.Error(0)
}
```

**Usage:**
```go
func (s *MyTestSuite) SetupTest() {
    s.mock = mocks.NewMockMyService()
    s.mock.On("Method").Return(nil)
}
```

## Setup/Teardown

| Method | When |
|--------|------|
| `SetupTest()` | Before each test method |
| `TearDownTest()` | After each test method (optional) |
| `SetupSuite()` | Once before all tests (rare) |
| `TearDownSuite()` | Once after all tests (rare) |

## Error Testing

```go
func (s *MyTestSuite) TestSpecificErrorType() {
    err := s.service.DoSomething()
    s.Require().Error(err)

    var myErr *domain.MyError
    s.Require().ErrorAs(err, &myErr)
    s.Equal("expected field", myErr.Field())
}
```

## Real Git Repos

**Helper:** `helpers.NewIntegrationTestRepo(t)` - temp repo with auto-cleanup

```go
repo := helpers.NewIntegrationTestRepo(s.T())
repo.CreateBranch("feature-1")
repo.CreateWorktree("feature-1")
```

Test user: `test@twiggit.dev` / `Test User`

## Test Types

| Type | Framework | Focus | Mocking |
|------|-----------|-------|---------|
| Unit | Testify + table-driven | Component isolation | Mock external deps |
| Integration | Testify + `//go:build integration` | Component interactions | Real git repos |
| E2E | Ginkgo/Gomega + gexec | CLI workflows | Built binary |

**Skip in short mode:** `if testing.Short() { t.Skip() }`

## Commands

```bash
mise run test              # All tests
mise run test:unit         # Unit only
mise run test:integration  # Integration only
mise run test:e2e          # E2E only
```
