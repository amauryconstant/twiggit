## Purpose

Pre-built git repos for fast, reproducible E2E tests

## Available Fixtures

| Fixture              | Branches                   | Use Cases                            |
| -------------------- | -------------------------- | ------------------------------------ |
| bare-main.tar.gz     | main                       | Basic operations, bare repo handling |
| single-branch.tar.gz | main                       | Create/list/delete tests             |
| multi-branch.tar.gz  | main, feature-1, feature-2 | Multi-branch scenarios               |
| corrupted.tar.gz     | main                       | Corrupted repository error handling  |
| detached.tar.gz      | main (detached HEAD)       | Detached HEAD state handling         |
| submodule.tar.gz     | main + submodule           | Repository with submodules           |

## Edge Case Fixtures

| Fixture          | State                   | Use Case                        |
| ---------------- | ----------------------- | ------------------------------- |
| corrupted.tar.gz | Invalid `.git/objects`  | Test graceful error handling    |
| detached.tar.gz  | HEAD detached at commit | Test detached HEAD detection    |
| submodule.tar.gz | Contains git submodule  | Test submodule-aware operations |

## Creating Fixtures

Write script in `.mise/tasks/fixtures/<name>` (with `#MISE hide=true` header). Run: `mise run fixtures:generate`. Archives created at: `test/e2e/fixtures/repos/`

## Usage Pattern

```go
var _ = ginkgo.Describe("create command", func() {
    var fixture *fixtures.E2ETestFixture

    ginkgo.BeforeEach(func() {
        fixture = fixtures.NewE2ETestFixture()
        fixture.SetupSingleProject("test-project")
    })

    ginkgo.It("creates worktree", func() {
        session := fixtures.ExecuteCLI(fixture, "create", "feature-1")
        gomega.Eventually(session).Should(gexec.Exit(0))
    })
})
```

## Regeneration

If fixture missing or extraction fails:
`mise run fixtures:generate`

## Fixture States

Requirements:

- Clean working tree (no uncommitted changes)
- Known commit SHAs for deterministic referencing
- Minimal history (only necessary commits)

## Performance

Fixture extraction times:

- bare-main: ~10ms
- single-branch: ~15ms
- multi-branch: ~20ms

Much faster than creating repos from scratch (100-200ms).
