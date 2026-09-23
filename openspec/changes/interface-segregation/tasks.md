# Tasks

## 1. Setup

- [ ] 1.1 Confirm `cli-functional-core-shell` is archived (it is the
  precondition that collapsed the service layer and deleted `test/mocks/`);
  verify by `openspec show cli-functional-core-shell 2>&1` (or
  `ls openspec/changes/archive/2026-09-23-cli-functional-core-shell/`)
  existing.
- [ ] 1.2 Verify `internal/git/reader.go` carries the read-side methods to
  be declared in role interfaces (`ValidateRepository`, `ListBranches`,
  `BranchExists`, `GetRepositoryStatus`, `GetRepositoryInfo`,
  `GetCommitInfo`, `ListRemotes`); verify by grepping each name in
  `internal/git/reader.go`.
- [ ] 1.3 Verify `internal/git/writer.go` carries the write-side methods
  (`CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees`,
  `DeleteBranch`, `IsBranchMerged`); verify by grepping each name in
  `internal/git/writer.go`.
- [ ] 1.4 Verify `internal/core/git.go` does **not** yet exist as the
  target file for the role interface declarations; verify by
  `ls internal/core/git.go` reporting no such file or directory.

## 2. Role interface declarations

- [ ] 2.1 Create `internal/core/git.go` with `package core` and a godoc
  preamble naming the role-interface segregation contract and citing
  the `golang-structs-interfaces` 1-3 method rule; verify by `go build
  ./internal/core/` clean.
- [ ] 2.2 Declare `core.RepositoryOpener` with exactly one method:
  `ValidateRepository(path string) error`; godoc documents why
  `OpenRepository` is excluded (no command consumer; return-type
  `*go-git.Repository` violates `core-isolation` depguard); verify by
  `go build ./internal/core/` clean.
- [ ] 2.3 Declare `core.BranchReader` with `ListBranches(ctx
  context.Context, repoPath string) ([]core.BranchInfo, error)` and
  `BranchExists(ctx context.Context, repoPath, branchName string) (bool,
  error)`; verify by `go build ./internal/core/` clean.
- [ ] 2.4 Declare `core.RepositoryReader` with `GetRepositoryStatus`,
  `GetRepositoryInfo`, and `GetCommitInfo` (3 methods, upper-bound of the
  1-3 rule); verify by `go build ./internal/core/` clean.
- [ ] 2.5 Declare `core.RemoteReader` with `ListRemotes(ctx
  context.Context, repoPath string) ([]core.RemoteInfo, error)`; verify
  by `go build ./internal/core/` clean.
- [ ] 2.6 Declare `core.WorktreeWriter` with `CreateWorktree`,
  `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees` (4 methods);
  godoc documents the 1-3 rule violation and the deferred split; verify
  by `go build ./internal/core/` clean.
- [ ] 2.7 Declare `core.BranchWriter` with `DeleteBranch` and
  `IsBranchMerged`; verify by `go build ./internal/core/` clean.

## 3. Compile-time role satisfaction

- [ ] 3.1 Append a `var (`/`)` block to the bottom of
  `internal/git/client.go` with six `var _ core.Role = (*git.Client)(nil)`
  lines covering `RepositoryOpener`, `BranchReader`, `RepositoryReader`,
  `RemoteReader`, `WorktreeWriter`, `BranchWriter`; verify by `go build
  ./internal/git/...` clean.
- [ ] 3.2 Add the `_ = (*git.Client)(nil)` placement at the bottom of
  `internal/git/client.go` (right after the existing `var _ =
  lru.New[...]` line) so the new role assertions sit with the other
  compile-time guards; verify by `go build ./internal/git/...` clean.
- [ ] 3.3 Run `go build ./...` end-to-end; verify by exit code 0.

## 4. Drift metatest

- [ ] 4.1 Append the `_DriftCheck` sentinel struct to
  `internal/git/client_test.go`:

  ```go
  type _DriftCheck struct {
      _ core.RepositoryOpener
      _ core.BranchReader
      _ core.RepositoryReader
      _ core.RemoteReader
      _ core.WorktreeWriter
      _ core.BranchWriter
  }

  var _ = func() any { var c _DriftCheck; return &c }((*git.Client)(nil))
  ```

  verify by `go build ./internal/git/...` and `go test ./internal/git/...`
  clean.
- [ ] 4.2 Rename or remove any of the 14 role methods on `*reader` or
  `*cliClient` temporarily to confirm the sentinel fails the build with a
  type-mismatch error referencing the affected role; then revert the
  rename. (This is a one-shot confidence check; the build clean
  confirmation replaces the assertion.)

## 5. Factory wiring (additive)

- [ ] 5.1 Add six per-role lazy function fields to `cmdutil.Factory`:
  `RepoOpener func() (core.RepositoryOpener, error)`, `BranchReader`,
  `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`;
  verify by `go build ./internal/cmdutil/` clean.
- [ ] 5.2 Initialize the six per-role fields in `NewFactory()` so each
  field's body routes through `f.GitClient()` and returns the cached
  `*git.Client` typed as the requested role; the `sync.Once` per
  concrete is the existing one on `f.GitClient`; verify by reading the
  Factory implementation.
- [ ] 5.3 Update `Factory.Init()` to touch the six per-role fields and
  collect any errors via `errors.Join`; verify by `go build ./internal/cmdutil/`
  clean.
- [ ] 5.4 Confirm the composite `f.GitClient()` field is unchanged
  (`func() (*git.Client, error)`, still wired to `git.NewClient()`);
  existing `cmd/*.go` call sites need no modification; verify by
  `go build ./cmd/...` clean.

## 6. Build + test + lint verification

- [ ] 6.1 Run `go build ./...`; verify by exit code 0.
- [ ] 6.2 Run `go test ./...`; verify by exit code 0.
- [ ] 6.3 Run `golangci-lint run`; verify by zero new findings from this
  change.
- [ ] 6.4 Run `gofmt -l` and `goimports -l` over the touched files
  (`internal/core/git.go`, `internal/git/client.go`,
  `internal/git/client_test.go`, `internal/cmdutil/factory.go`); verify
  by exit code 0.
- [ ] 6.5 Run `go vet ./internal/git/...` and `go vet ./internal/core/...`;
  verify by exit code 0.

## 7. OpenSpec verification

- [ ] 7.1 Run `openspec validate interface-segregation --strict
  --json`; verify by `valid: true, issues: []`.
- [ ] 7.2 Run `openspec status --change interface-segregation --json`;
  verify by `isPlanningComplete: true`.

---

## Tasks dropped at revise time (with rationale)

The following tasks from the original `tasks.md` were dropped because the
underlying infrastructure was deleted by `cli-functional-core-shell` or
was never built. Tracking them under this change would have produced
no-op or incorrect edits.

| Original task group | Reason dropped |
|---|---|
| §4.1–4.6 Service constructor refactor (`core.NewWorktreeService`, `NewProjectService`, `NewNavigationService`, `NewContextResolver`) | No service constructors exist; the service layer was collapsed in `cli-functional-core-shell`. |
| §6.1–6.7 Mock rewrites (`test/mocks/repository_opener_mock.go`, etc.) | `test/mocks/` directory deleted; tests use real `*git.Client` + Factory `runF` seam. |
| §7.1–7.3, §8.1–8.2, §9.1–9.2 Integration, concurrent, e2e fixture updates | No mocks to update; integration tests already use `git.NewClient()`. |
| §10.1, §10.2 Spec authoring for `core-git` and `git-client` delta | The original spec drafts assumed a deleted architecture (referenced `internal/core/git.go`, `internal/git/gogit_client.go`, `MockGitClientBundle`). They were re-authored atomically with this `tasks.md` revision rather than marked done against stale content. |

If a future change reintroduces a service layer or a mock surface, those
tasks become live again — the role interfaces declared here will be the
constructor argument shapes and the mock method sets.
