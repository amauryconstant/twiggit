# Testing Conventions

## Test Organization

### Test File Locations
- Domain tests: `internal/domain/*_test.go`
- Service tests: `internal/service/*_test.go`
- Infrastructure tests: `internal/infrastructure/*_test.go`
- Mocks: `test/mocks/*.go`

### Naming Conventions
- Test files: `<source_file>_test.go`
- Test functions: `Test<SuiteName>_<Scenario>` for suite-based tests
- Test cases: Descriptive names following "given_when_then" pattern where applicable

## Test Patterns

### Testify Suite Pattern (Required)

All unit tests SHALL use Testify suites for consistency:

```go
type MyTestSuite struct {
    suite.Suite
    // Test fixtures
    service MyService
    mockDependency *mocks.MockDependency
}

func TestMySuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}

func (s *MyTestSuite) SetupTest() {
    // Setup before each test
    s.mockDependency = mocks.NewMockDependency()
    s.service = NewMyService(s.mockDependency)
}

func (s *MyTestSuite) TearDownTest() {
    // Cleanup after each test (optional)
}

func (s *MyTestSuite) TestOperation_Success() {
    testCases := []struct {
        name     string
        input    Input
        expected Output
    }{
        {
            name:     "simple case",
            input:    Input{...},
            expected: Output{...},
        },
    }

    for _, tc := range testCases {
        s.Run(tc.name, func() {
            result := s.service.Do(tc.input)
            s.Equal(tc.expected, result)
        })
    }
}
```

### Table-Driven Tests

Use table-driven tests within suites for multiple scenarios:

```go
func (s *MyTestSuite) TestValidation() {
    tests := []struct {
        name        string
        input       string
        expectError bool
    }{
        {"valid input", "valid", false},
        {"empty input", "", true},
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            err := s.validate(tt.input)
            if tt.expectError {
                s.Error(err)
            } else {
                s.NoError(err)
            }
        })
    }
}
```

## Assertion Guidelines

### When to Use `require` (Fail Fast)
Use `require.*` for critical checks that should stop the test immediately:

```go
s.Require().NoError(err)          // Error assertions (per testifylint)
s.Require().NotNil(obj)            // Critical setup checks
s.Require().Equal(expected, actual) // When test can't continue
```

### When to Use `assert` (Continue)
Use `assert.*` for non-critical checks where you want to see all failures:

```go
s.Equal(expected, actual)  // Value comparisons
s.Contains(str, substr)     // String checks
s.True(condition)           // Boolean checks
```

**Testifylint Requirement**: All error assertions MUST use `require.Error()` or `require.NoError()`

## Mock Guidelines

### Centralized Mocks
All mocks SHALL be centralized in `test/mocks/`:

```go
// test/mocks/my_service_mock.go
package mocks

type MockMyService struct {
    mock.Mock
}

func NewMockMyService() *MockMyService {
    return &MockMyService{}
}
```

### Mock Patterns
Two patterns used across codebase:

1. **Testify mocks** (testify/mock):
   ```go
   type MockMyService struct {
       mock.Mock
   }
   func (m *MockMyService) Method() error {
       args := m.Called()
       return args.Error(0)
   }
   ```

2. **Functional mocks** (with function fields):
   ```go
   type MockMyService struct {
       MethodFunc func() error
   }
   func (m *MockMyService) Method() error {
       if m.MethodFunc != nil {
           return m.MethodFunc()
       }
       return nil
   }
   ```

### Using Mocks in Tests

```go
import "twiggit/test/mocks"

func (s *MyTestSuite) SetupTest() {
    s.mockService = mocks.NewMockMyService()
    s.mockService.On("Method").Return(nil)
}
```

## Setup/Teardown Patterns

### Suite Setup/Teardown
- `SetupTest()`: Runs before each test method
- `TearDownTest()`: Runs after each test method (optional)
- `SetupSuite()`: Runs once before all tests (rarely needed)
- `TearDownSuite()`: Runs once after all tests (rarely needed)

### Helper Functions
For complex setup, use helper functions:

```go
func (s *MyTestSuite) setupTestService() MyService {
    config := domain.DefaultConfig()
    mock := mocks.NewMockDependency()
    return NewMyService(mock, config)
}
```

## Error Testing

### Testing Error Conditions
```go
func (s *MyTestSuite) TestErrorHandling() {
    s.Run("invalid input", func() {
        err := s.service.Process("")
        s.Require().Error(err)
        s.Contains(err.Error(), "invalid input")
    })
}
```

### Error Type Checking
```go
func (s *MyTestSuite) TestSpecificErrorType() {
    err := s.service.DoSomething()
    s.Require().Error(err)

    var myErr *domain.MyError
    s.Require().ErrorAs(err, &myErr)
    s.Equal("expected field", myErr.Field())
}
```

## Quality Requirements

- All tests MUST pass `mise run test:full`
- All tests MUST pass `mise run test:race`
- All golangci-lint checks MUST pass
- Test coverage SHOULD be >80% for new code
- Tests SHOULD be fast (<100ms per test)
- Tests MUST be deterministic (no randomness)

## Build Tags

Use build tags for integration tests:

```go
//go:build integration

package service_test

func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    // Real dependencies
}
```
