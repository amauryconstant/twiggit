# E2E Test Fixtures

This directory contains pre-built git repository fixtures for E2E testing.

## Overview

Pre-built git repository fixtures are used for:

- **Speed**: Archive extraction is faster than `git init` + commits
- **Reproducibility**: Same git state every test run
- **Cross-platform**: Archives work on any OS
- **Version control**: Repository structure is tracked in git
- **Eliminates state pollution**: Fresh git state for each test

## Available Fixtures

| Fixture | Description | Branches | Use Cases |
|---------|-------------|-----------|------------|
| `bare-main.tar.gz` | Bare repository with main branch | `main` | Basic worktree operations |
| `single-branch.tar.gz` | Repository with main branch and 3 commits | `main` | Most create/list/delete tests |
| `multi-branch.tar.gz` | Repository with main + 2 feature branches | `main`, `feature-1`, `feature-2` | Multi-branch scenarios, switching tests |

## Fixture Structure

Each `.tar.gz` archive contains a complete git repository with:
- `.git/` directory with full git history
- Working tree files (README.md, feature files, etc.)
- All git references and objects

## Creating New Fixtures

To create a new fixture:

1. Create a setup script in `scripts/fixtures/<name>.sh`:

```bash
#!/bin/bash
set -e

REPO_DIR=$1

git init "$REPO_DIR"
cd "$REPO_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"

# Create commits, branches, etc.
echo "content" > file.txt
git add file.txt
git commit -m "Initial commit"

# Create branches if needed
git checkout -b feature-branch
echo "feature content" > feature.txt
git add feature.txt
git commit -m "Feature commit"

git checkout main
```

2. Run the fixture generation script:

```bash
./scripts/generate-repo-fixtures.sh
```

3. This will create `test/e2e/fixtures/repos/<name>.tar.gz`

## Using Fixtures in Tests

```go
var _ = Describe("create command", func() {
    var fixture *fixtures.E2ETestFixture

    BeforeEach(func() {
        fixture = fixtures.NewE2ETestFixture()
        // Uses single-branch fixture automatically
        fixture.SetupSingleProject("test-project")
    })

    It("creates worktree", func() {
        projectPath := fixture.GetProjectPath("test-project")
        // Test using the pre-built repo
    })
})
```

## Fixture States

When creating fixtures, ensure they are in a clean, reproducible state:

- **Clean working tree**: No uncommitted changes
- **Detached HEAD** (optional): For detached worktree tests
- **Known commit SHAs**: For reliable referencing
- **Minimal history**: Only necessary commits to reduce size

## Future: On-Demand Repository Creation

For scenarios requiring dynamic git states (specific commit sequences, custom configurations),
on-demand repository creation could be implemented as an alternative to pre-built archives.
This would be slower but more flexible for complex test scenarios.

## Troubleshooting

### Fixture not found

If you see "repo fixture 'X' not found", ensure:
- The fixture archive exists in `test/e2e/fixtures/repos/`
- The name matches exactly (including case)

### Archive extraction fails

If extraction fails, regenerate the fixture:
```bash
./scripts/generate-repo-fixtures.sh
```

### State pollution persists

If tests fail with "branch already exists" or similar errors:
- Ensure `TrackWorktree()` is called when creating worktrees
- Verify `Cleanup()` is running in `AfterEach()`
- Check that `GinkgoT().TempDir()` is being used

## Performance

Fixture extraction typically takes:
- `bare-main`: ~10ms
- `single-branch`: ~15ms
- `multi-branch`: ~20ms

This is significantly faster than creating repos from scratch (100-200ms per repo).
