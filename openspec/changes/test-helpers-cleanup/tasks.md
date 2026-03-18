## 1. Verify t.Helper() in test helper constructors

> Note: WorktreeTestHelper excluded from scope - constructor does not accept *testing.T parameter
>
- [x] 1.1 Verify t.Helper() call is present in NewGitTestHelper constructor in test/helpers/git.go (already present)
- [ ] 1.2 Verify t.Helper() call is present in helper constructor in test/helpers/shell.go (already present, confirm)
- [x] 1.3 Verify t.Helper() call is present in NewRepoTestHelper constructor in test/helpers/repo.go (already present)

## 2. Add t.Cleanup() patterns to constructors

> Note: Task 2.2 must be completed before other tasks if additional cleanup is needed
>
- [ ] 2.1 Verify GitTestHelper uses t.TempDir() for automatic cleanup in test/helpers/git.go (cleanup already handled by testing package)
- [ ] 2.2 Register t.Cleanup() in RepoTestHelper constructor in test/helpers/repo.go to call Cleanup() method

## 3. Documentation and verification

- [ ] 3.1 Update test/helpers/AGENTS.md with cleanup documentation
- [ ] 3.2 Run mise run test:full to verify no regressions
