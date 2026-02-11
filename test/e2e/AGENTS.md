## E2E Test Structure
Purpose: Test CLI commands from user perspective

**Framework:** Ginkgo/Gomega + gexec for binary execution
**Build tag:** `//go:build e2e`

## Ginkgo Pattern

```go
//go:build e2e

var _ = ginkgo.Describe("create command", func() {
    var fixture *fixtures.E2ETestFixture

    ginkgo.BeforeEach(func() {
        fixture = fixtures.NewE2ETestFixture()
        fixture.SetupSingleProject("test-project")
    })

    ginkgo.It("creates worktree from default branch", func() {
        session := fixtures.ExecuteCLI(fixture, "create", "feature-1")
        gomega.Eventually(session).Should(gexec.Exit(0))
    })

    ginkgo.AfterEach(func() { fixture.Cleanup() })
})
```

## CLI Execution

**Helper:** `fixtures.ExecuteCLI(fixture, args...)`

Executes built CLI binary with:
- Binary path from fixture
- Arguments passed to command
- Output/exit code capture
- Timeout support

**Example:**
```go
session := fixtures.ExecuteCLI(fixture, "list", "--all")
gomega.Eventually(session).Should(gexec.Exit(0))
output := string(session.Out.Contents())
```

## Test Fixtures

**Fixture Helper:** `fixtures.NewE2ETestFixture()`
See `test/e2e/fixtures/AGENTS.md` for available fixtures and usage.

## E2E Test Patterns

```go
ginkgo.It("lists worktrees successfully", func() {
    session := fixtures.ExecuteCLI(fixture, "list")
    gomega.Eventually(session).Should(gexec.Exit(0))
})

ginkgo.It("fails with invalid project", func() {
    session := fixtures.ExecuteCLI(fixture, "create", "invalid-project", "feature")
    gomega.Eventually(session).Should(gexec.HaveExitCode(1))
})
```

## Running E2E Tests
```bash
mise run test:e2e      # All E2E tests
mise run build:e2e     # Build CLI for E2E testing
```

## Testing Verbose Output

The `--verbose` flag (`-v`, `-vv`) outputs to stderr. Use `ShouldVerboseOutput()` and `ShouldNotHaveVerboseOutput()` helpers.

**Example:**
```go
// Test no verbose output by default
session := cli.Run("create", "feature-1")
cli.ShouldSucceed(session)
cli.ShouldNotHaveVerboseOutput(session)

// Test level 1 verbose output
session := cli.Run("create", "feature-1", "-v")
cli.ShouldSucceed(session)
cli.ShouldVerboseOutput(session, "Creating worktree")

// Test level 2 verbose output
session := cli.Run("create", "feature-1", "-vv")
cli.ShouldSucceed(session)
cli.ShouldVerboseOutput(session, "Creating worktree")
cli.ShouldVerboseOutput(session, "  from branch: main")
```

**Helper Methods:**
- `ShouldVerboseOutput(session, expected)` - Asserts stderr contains expected verbose text
- `ShouldNotHaveVerboseOutput(session)` - Asserts no verbose messages in stderr
