

## Purpose
Pre-built git repos for fast, reproducible E2E tests

## Available Fixtures

| Fixture | Branches | Use Cases |
|---------|----------|-----------|
| bare-main.tar.gz | main | Basic operations |
| single-branch.tar.gz | main | Create/list/delete tests |
| multi-branch.tar.gz | main, feature-1, feature-2 | Multi-branch scenarios |

## Creating Fixtures

Write script in `scripts/fixtures/<name>.sh`. Run: `./scripts/generate-repo-fixtures.sh`. Archives created at: `test/e2e/fixtures/repos/`

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
`./scripts/generate-repo-fixtures.sh`

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
