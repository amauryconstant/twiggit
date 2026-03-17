## 1. Add t.Helper() to test helper constructors

> Note: WorktreeTestHelper excluded from scope - constructor does not accept *testing.T parameter
>
- [ ] 1.1 Add t.Helper() call to NewGitTestHelper constructor in test/helpers/git.go (already present)
- [ ] 1.2 Add t.Helper() call to helper constructor in test/helpers/shell.go
- [ ] 1.3 Add t.Helper() call to NewRepoTestHelper constructor in test/helpers/repo.go (already present)

## 2. Add t.Cleanup() patterns to constructors

> Note: Task 2.1 must be completed before 2.2 because RepoTestHelper.SetupTestRepo uses GitTestHelper
>
- [ ] 2.1 Add t.Cleanup() pattern to GitTestHelper constructor in test/helpers/git.go
- [ ] 2.2 Add t.Cleanup() pattern to RepoTestHelper constructor in test/helpers/repo.go (calls Cleanup method on test completion)

## 3. Documentation and verification

- [ ] 3.1 Update test/helpers/AGENTS.md with cleanup documentation
- [ ] 3.2 Run mise run test:full to verify no regressions
