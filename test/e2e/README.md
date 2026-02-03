# E2E Test Suite

## Organization

The E2E test suite is organized into the following directories:

```
test/e2e/
├── e2e_suite_test.go           # Suite setup (BeforeSuite/AfterSuite)
├── infrastructure_verification_test.go  # Infrastructure validation tests
├── fixtures/                    # Test fixtures and helpers
│   ├── e2e_fixtures.go         # Main fixture with worktree tracking
│   ├── context_helper.go        # Context-aware testing (project/worktree/outside)
│   └── test_scenarios.go       # Pre-built test scenarios (future)
├── helpers/                     # Testing helpers and utilities
│   ├── cli_helper.go            # CLI execution with gexec
│   ├── config_helper.go         # Config management and validation
│   ├── assertions.go            # Domain-specific assertions
│   └── test_id_generator.go     # Unique test ID generation
├── cmd/                         # Command-specific tests
│   └── help_test.go            # Help command tests
├── context/                     # Context-aware tests
│   ├── project_context_test.go  # Tests from project directory
│   ├── worktree_context_test.go # Tests from worktree directory
│   └── outside_git_test.go      # Tests outside git repository
└── workflows/                   # End-to-end workflow tests
    ├── rebase_workflow_test.go     # Rebase workflow tests
    ├── navigation_workflow_test.go  # Navigation workflow tests
    └── cleanup_workflow_test.go    # Cleanup workflow tests
```

## Key Improvements

### 1. Robust Cleanup

The `E2ETestFixture.Cleanup()` method now includes:
- **Nil checks** for all helpers to prevent panics
- **Retry logic** with exponential backoff for worktree removal
- **Force flag** (`--force`) to handle uncommitted changes
- **Idempotency** - can be called multiple times safely

### 2. Debugging Infrastructure

The fixture now provides `Inspect()` method for debugging:

```go
fixture.CreateWorktreeSetup("test")
GinkgoT().Log(fixture.Inspect())

// Output:
// === E2ETestFixture State ===
// TempDir: /tmp/...
//
// === Projects ===
//   test: /tmp/.../test [✓]
//
// === Worktrees ===
//   [0] /tmp/.../wt-feature-1 [✓]
//   [1] /tmp/.../wt-feature-2 [✓]
```

### 3. Cleanup Validation

Use `ValidateCleanup()` to ensure cleanup succeeded:

```go
fixture.Cleanup()
err := fixture.ValidateCleanup()
Expect(err).NotTo(HaveOccurred())
```

### 4. Domain-Specific Assertions

Use `NewTwiggitAssertions()` for cleaner test code:

```go
assertions := helpers.NewTwiggitAssertions()

assertions.ShouldCreateWorktree(session, "feature-1")
assertions.ShouldOutputWorktreeList(session, []string{"main", "feature-1"})
assertions.ShouldHaveWorktree(worktreePath)
```

## Writing Tests

### Basic Test Structure

```go
package e2e

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "twiggit/test/e2e/fixtures"
    "twiggit/test/e2e/helpers"
)

var _ = Describe("list command", func() {
    var fixture *fixtures.E2ETestFixture
    var cli *helpers.TwiggitCLI
    var assertions *helpers.TwiggitAssertions

    BeforeEach(func() {
        fixture = fixtures.NewE2ETestFixture()
        cli = helpers.NewTwiggitCLI()
        assertions = helpers.NewTwiggitAssertions()
    })

    AfterEach(func() {
        fixture.Cleanup()
    })

    It("lists all projects", func() {
        fixture.SetupMultiProject()
        cli = cli.WithConfigDir(fixture.Build())

        session := cli.Run("list")
        assertions.ShouldListProjects(session, []string{"test-project-1", "test-project-2"})
    })
})
```

### Context-Aware Testing

```go
It("creates worktree from project directory", func() {
    fixture.SetupSingleProject("myproject")
    cli = cli.WithConfigDir(fixture.Build())

    ctxHelper := fixtures.NewContextHelper(fixture, cli)
    session := ctxHelper.FromProjectDir("myproject", "create", "feature-1")

    assertions.ShouldCreateWorktree(session, "feature-1")
})
```

### Debugging Failed Tests

```go
AfterEach(func() {
    if CurrentSpecReport().Failed() {
        GinkgoT().Log("Test failed - current state:\n", fixture.Inspect())
    }
    fixture.Cleanup()
})
```

## Running Tests

### Run all E2E tests
```bash
mise run test:e2e
```

### Run specific test suite
```bash
cd test/e2e && ginkgo --tags=e2e --focus="list command"
```

### Run with verbose output
```bash
cd test/e2e && ginkgo --tags=e2e -v
```

## Best Practices

1. **Always call `Cleanup()` in `AfterEach`**
2. **Use `fixture.Inspect()` when debugging failures**
3. **Use domain assertions for cleaner test code**
4. **Test from all three contexts** (project, worktree, outside git)
5. **Verify cleanup with `ValidateCleanup()` in critical tests**
6. **Organize tests by command or workflow**

## Critical Infrastructure Features

- ✅ Robust cleanup with retry logic
- ✅ Debugging visibility via `Inspect()`
- ✅ Idempotent cleanup
- ✅ Context-aware testing support
- ✅ Domain-specific assertions
- ✅ Nil-safe operations
- ✅ Force flag for worktree removal
