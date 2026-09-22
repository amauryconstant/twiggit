# Tasks

## 1. Setup

- [ ] 1.1 Confirm `cli-functional-core-shell` is applied (or note as a precondition); verify by `openspec list --changes --json` includes `cli-functional-core-shell` with `state: applied`.
- [ ] 1.2 Verify `internal/core/git.go` exists as the planned target file for the role interface declarations; verify by `ls internal/core/git.go` succeeding.
- [ ] 1.3 Verify `internal/git/gogit_client.go` and `internal/git/cli_client.go` carry the methods to be split (`OpenRepository`, `ValidateRepository`, `ListBranches`, `BranchExists`, `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo`, `ListRemotes`, `CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees`, `DeleteBranch`, `IsBranchMerged`); verify by grepping each name.

## 2. Role interface declarations

- [ ] 2.1 Create `internal/core/git.go` with the `package core` declaration and a godoc comment naming the role-interface segregation contract; verify by `go build ./internal/core/` clean.
- [ ] 2.2 Declare `core.RepositoryOpener` with `OpenRepository(path) (*Repository, error)` and `ValidateRepository(path) error`; verify by `go build ./internal/core/` clean.
- [ ] 2.3 Declare `core.BranchReader` with `ListBranches(ctx context.Context, repoPath string) ([]BranchInfo, error)` and `BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)`; verify by `go build ./internal/core/` clean.
- [ ] 2.4 Declare `core.RepositoryReader` with `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo`; verify by `go build ./internal/core/` clean.
- [ ] 2.5 Declare `core.RemoteReader` with `ListRemotes(ctx context.Context, repoPath string) ([]RemoteInfo, error)`; verify by `go build ./internal/core/` clean.
- [ ] 2.6 Declare `core.WorktreeWriter` with the four worktree-lifecycle methods; godoc comment documents the 4-method skill-rule violation and the deferred follow-up; verify by `go build ./internal/core/` clean.
- [ ] 2.7 Declare `core.BranchWriter` with `DeleteBranch` and `IsBranchMerged`; verify by `go build ./internal/core/` clean.

## 3. Compile-time interface satisfaction checks

- [ ] 3.1 Add `var _ core.RepositoryOpener = (*GoGitClient)(nil)` (and the three other `GoGitClient` role checks) at the bottom of `internal/git/gogit_client.go`; verify by `go build ./internal/git/...` clean.
- [ ] 3.2 Add `var _ core.WorktreeWriter = (*CLIClient)(nil)` and `var _ core.BranchWriter = (*CLIClient)(nil)` at the bottom of `internal/git/cli_client.go`; verify by `go build ./internal/git/...` clean.
- [ ] 3.3 Run `go build ./...` end-to-end; verify by exit code 0.

## 4. Service constructor refactor

- [ ] 4.1 Update `core.NewWorktreeService` signature to take `(branchReader BranchReader, repoReader RepositoryReader, worktreeWriter WorktreeWriter, branchWriter BranchWriter, projectService ProjectService, config *Config, hookRunner HookRunner)`; update the body to consume role interfaces; verify by `go build ./...` clean.
- [ ] 4.2 Update `core.NewProjectService` signature to take `(repoOpener RepositoryOpener, repoReader RepositoryReader, remoteReader RemoteReader, repoLocator RepoLocator, config *Config)`; verify by `go build ./...` clean.
- [ ] 4.3 Update `core.NewNavigationService` signature to take `(repoLocator RepoLocator, contextResolver ContextResolver, config *Config)`; verify by `go build ./...` clean.
- [ ] 4.4 Update `core.NewContextResolver` signature to take `(repoOpener RepositoryOpener, repoReader RepositoryReader, config *Config)`; verify by `go build ./...` clean.
- [ ] 4.5 Update all `cmd/*.go` call sites that construct the four services; verify by `go build ./cmd/...` clean.
- [ ] 4.6 Update `main.go` to invoke the new constructor signatures; verify by `go build ./...` clean.

## 5. Factory wiring

- [ ] 5.1 Add per-role lazy fields (`RepoOpener`, `BranchReader`, `RepoReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) to `cmdutil.Factory`; each field type is `func() (core.Role, error)`; verify by `go build ./internal/cmdutil/` clean.
- [ ] 5.2 Add `sync.Once` cache inside the Factory so the underlying `*GoGitClient` / `*CLIClient` is built exactly once; subsequent calls return the same instance typed as the requested role; verify by reading the Factory implementation.
- [ ] 5.3 Remove the composite `GitClient` and `CLIClient` lazy fields from `cmdutil.Factory`; update `main.go` to assign the per-role fields; verify by `go build ./...` clean.
- [ ] 5.4 Verify each command consumes only the factory fields it actually invokes; verify by grepping `cmd/*.go` for `f\.BranchReader\(` etc. and confirming no command invokes a role it does not consume.

## 6. Mock rewrites

- [ ] 6.1 Delete `MockGitClientBundle` struct from `test/mocks/git_service_mock.go`; verify by `grep MockGitClientBundle test/` returning no matches.
- [ ] 6.2 Create `test/mocks/repository_opener_mock.go` with `MockRepositoryOpener` (2 methods) using testify/mock; verify by `go test ./test/mocks/...` clean.
- [ ] 6.3 Create `test/mocks/branch_reader_mock.go` with `MockBranchReader` (2 methods); verify by `go test ./test/mocks/...` clean.
- [ ] 6.4 Create `test/mocks/repository_reader_mock.go` with `MockRepositoryReader` (3 methods); verify by `go test ./test/mocks/...` clean.
- [ ] 6.5 Create `test/mocks/remote_reader_mock.go` with `MockRemoteReader` (1 method); verify by `go test ./test/mocks/...` clean.
- [ ] 6.6 Create `test/mocks/worktree_writer_mock.go` with `MockWorktreeWriter` (4 methods); verify by `go test ./test/mocks/...` clean.
- [ ] 6.7 Create `test/mocks/branch_writer_mock.go` with `MockBranchWriter` (2 methods); verify by `go test ./test/mocks/...` clean.

## 7. Integration test updates

- [ ] 7.1 Update each test in `test/integration/` that constructed `MockGitClientBundle` to construct only the role mocks matching its command's consumed roles; verify by `go test ./test/integration/...` passes.
- [ ] 7.2 Update each test in `test/integration/` to pass the per-role mocks to the new constructor signatures; verify by `go test ./test/integration/...` passes.
- [ ] 7.3 Verify no remaining references to `MockGitClientBundle`, `MockGoGitClient`, or `MockCLIClient` in `test/integration/`; verify by `grep -rn 'MockGitClientBundle\|MockGoGitClient\|MockCLIClient' test/integration/` returning no matches.

## 8. Concurrent test updates

- [ ] 8.1 Update each test in `test/concurrent/` that constructed `MockGitClientBundle` to construct only the role mocks matching its command's consumed roles; verify by `go test ./test/concurrent/...` passes.
- [ ] 8.2 Verify no remaining references to `MockGitClientBundle`, `MockGoGitClient`, or `MockCLIClient` in `test/concurrent/`; verify by `grep -rn 'MockGitClientBundle\|MockGoGitClient\|MockCLIClient' test/concurrent/` returning no matches.

## 9. E2E fixture updates

- [ ] 9.1 Update each fixture in `test/e2e/fixtures/` that constructed `MockGitClientBundle` to construct only the role mocks matching its command's consumed roles; verify by `go test ./test/e2e/fixtures/...` passes.
- [ ] 9.2 Verify no remaining references to `MockGitClientBundle`, `MockGoGitClient`, or `MockCLIClient` in `test/e2e/fixtures/`; verify by `grep -rn 'MockGitClientBundle\|MockGoGitClient\|MockCLIClient' test/e2e/fixtures/` returning no matches.

## 10. Spec authoring

- [ ] 10.1 Create `openspec/changes/interface-segregation/specs/core-git/spec.md` with the role interface method sets, consumer-side placement, factory wiring contract, and test mock structure; verify by reading the file.
- [ ] 10.2 Create `openspec/changes/interface-segregation/specs/git-client/spec.md` updating the "Routing table" requirement to reference the role interfaces by name; verify by reading the file.

## 11. Final lint pass

- [ ] 11.1 Run `golangci-lint run`; verify by zero new findings from this change.
- [ ] 11.2 Run `gofmt -l` and `goimports -l` over the touched files; verify by exit code 0.

## 12. Build + test verification

- [ ] 12.1 Run `go build ./...`; verify by exit code 0.
- [ ] 12.2 Run `go test ./...`; verify by exit code 0.

## 13. OpenSpec verification

- [ ] 13.1 Run `openspec validate interface-segregation --json`; verify by `valid: true, issues: []` (INFO entries about `git-client` target spec not yet in source-of-truth tree are acceptable per the locked decision).
- [ ] 13.2 Run `openspec status --change interface-segregation --json`; verify by `isPlanningComplete: true`.
