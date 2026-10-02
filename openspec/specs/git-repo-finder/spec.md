# Capability: Git Repo Finder

## Purpose

Defines the I/O adapter that enumerates git repositories under a directory tree. Used by `list --all` to enumerate projects under `Config.ProjectsDirectory`. Lives in `internal/git/repo_finder.go`. Every filesystem walk SHALL execute under a `context.Context` deadline. Every opened directory handle SHALL be `Close()`d via `defer` immediately.

## Requirements

### Requirement: RepoFinder interface

`git.RepoFinder` SHALL be a consumer-side interface declared where consumed. The interface SHALL have one method: `FindGitRepositories(ctx context.Context, root string, opts ...git.RepoFinderOption) ([]git.RepoFound, error)`. Production code uses the concrete `git.NewRepoFinder()`.

#### Scenario: FindGitRepositories enumerates under root

- **WHEN** a caller invokes `FindGitRepositories(ctx, "$HOME/Projects")`
- **THEN** the result SHALL be `[]git.RepoFound` with one entry per top-level project directory containing a `.git` directory
- **AND** the walk SHALL honor `ctx` cancellation

### Requirement: RepoFound and RepoValidator

`git.RepoFound` SHALL carry `Name string` (basename of the project directory), `Path string` (absolute path), `DefaultBranch string`, and `WorktreeCount int`. `git.RepoValidator` SHALL expose `IsValid(path string) bool` and `IsBare(path string) bool` to filter enumeration candidates.

#### Scenario: RepoFound carries DefaultBranch

- **WHEN** `FindGitRepositories` enumerates a project whose main branch is `develop`
- **THEN** `RepoFound.DefaultBranch == "develop"` (read from `HEAD` symbolic ref)

### Requirement: NewRepoFinder returns the production concrete

`git.NewRepoFinder() *git.repoFinder` SHALL return the production finder. The concrete SHALL use `filepath.WalkDir` for the directory traversal (not `filepath.Walk` — `WalkDir` returns `fs.DirEntry` and avoids per-op stat syscalls).

#### Scenario: NewRepoFinder returns non-nil

- **WHEN** a caller invokes `f := git.NewRepoFinder()`
- **THEN** `f` SHALL be non-nil and SHALL satisfy the `git.RepoFinder` interface

### Requirement: RepoFinderOption functional options

`git.WithMaxDepth(n int) RepoFinderOption` SHALL limit traversal depth; `git.WithExcludePaths(paths ...string) RepoFinderOption` SHALL skip matching paths during walk. The options pattern follows the `golang-design-patterns` functional-options rule; per the `golang-naming` `no-stuttering` rule, the option type SHALL NOT repeat the package name in the method form.

#### Scenario: WithMaxDepth limits walk depth

- **WHEN** a caller invokes `FindGitRepositories(ctx, root, git.WithMaxDepth(2))`
- **THEN** the walk SHALL descend at most 2 levels below `root`
- **AND** deeper git repositories SHALL NOT appear in the result

### Requirement: Validation failures return typed errors

When `IsValid` fails for a candidate (corrupted `.git` directory), the finder SHALL skip the entry rather than failing the entire enumeration. Any per-entry error surfaced (e.g., via an out-parameter or debug log) SHALL be lowercase without trailing punctuation.

#### Scenario: Corrupted repo is skipped

- **WHEN** the walk encounters a directory containing a `.git` directory that cannot be read
- **THEN** the finder SHALL skip that entry
- **AND** the enumeration SHALL continue with the remaining directories
- **AND** the returned slice SHALL NOT contain the corrupted entry