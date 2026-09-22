# Spec Delta

## Purpose

Defines the role-interface segregation contract for git operations. Establishes 1-3 method consumer-side interfaces that replace the 8-method `GoGitClient` and 6-method `CLIClient` composites, so each service receives only the roles it consumes.

## ADDED Requirements

### Requirement: RepositoryOpener exposes OpenRepository and ValidateRepository

The system SHALL expose a `RepositoryOpener` role interface with exactly two methods: `OpenRepository(path)` and `ValidateRepository(path)`. The role SHALL be defined consumer-side and SHALL NOT include any context-aware method.

#### Scenario: RepositoryOpener is the only role with both OpenRepository and ValidateRepository
- **WHEN** a caller requires both `OpenRepository` and `ValidateRepository`
- **THEN** the role SHALL be `RepositoryOpener` and SHALL NOT include any other method

#### Scenario: Concrete GoGitClient satisfies RepositoryOpener
- **WHEN** the concrete git client is type-asserted against `RepositoryOpener`
- **THEN** the assertion SHALL succeed at compile time

### Requirement: BranchReader exposes ListBranches and BranchExists

The system SHALL expose a `BranchReader` role interface with exactly two context-aware methods: `ListBranches` and `BranchExists`. The role SHALL be defined consumer-side.

#### Scenario: BranchReader is the only role with both ListBranches and BranchExists
- **WHEN** a caller requires both `ListBranches` and `BranchExists`
- **THEN** the role SHALL be `BranchReader` and SHALL NOT include any other method

#### Scenario: Concrete GoGitClient satisfies BranchReader
- **WHEN** the concrete git client is type-asserted against `BranchReader`
- **THEN** the assertion SHALL succeed at compile time

### Requirement: RepositoryReader exposes GetRepositoryStatus, GetRepositoryInfo, and GetCommitInfo

The system SHALL expose a `RepositoryReader` role interface with exactly three context-aware methods: `GetRepositoryStatus`, `GetRepositoryInfo`, and `GetCommitInfo`. The role SHALL be defined consumer-side.

#### Scenario: RepositoryReader is the only role with all three Get methods
- **WHEN** a caller requires `GetRepositoryStatus`, `GetRepositoryInfo`, and `GetCommitInfo`
- **THEN** the role SHALL be `RepositoryReader` and SHALL NOT include branch listing, remote listing, or repository opening

#### Scenario: Concrete GoGitClient satisfies RepositoryReader
- **WHEN** the concrete git client is type-asserted against `RepositoryReader`
- **THEN** the assertion SHALL succeed at compile time

### Requirement: RemoteReader exposes ListRemotes

The system SHALL expose a `RemoteReader` role interface with exactly one context-aware method: `ListRemotes`. The role SHALL be defined consumer-side.

#### Scenario: RemoteReader is the single-method role for ListRemotes
- **WHEN** a caller requires `ListRemotes` only
- **THEN** the role SHALL be `RemoteReader` and SHALL NOT include any other method

#### Scenario: Concrete GoGitClient satisfies RemoteReader
- **WHEN** the concrete git client is type-asserted against `RemoteReader`
- **THEN** the assertion SHALL succeed at compile time

### Requirement: WorktreeWriter exposes the four worktree-lifecycle methods

The system SHALL expose a `WorktreeWriter` role interface with exactly four context-aware methods: `CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, and `PruneWorktrees`. The role SHALL be defined consumer-side.

> **Note:** `WorktreeWriter` sits at 4 methods, exceeding the 1-3 method rule recommended by the interface-segregation skill. The methods form a tightly coupled worktree-lifecycle set; every code path that creates a worktree also lists, deletes, or prunes them. Further splitting is deferred as a follow-up.

#### Scenario: WorktreeWriter covers the worktree lifecycle
- **WHEN** a caller requires any combination of `CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, or `PruneWorktrees`
- **THEN** the role SHALL be `WorktreeWriter` and SHALL NOT include branch operations

#### Scenario: Concrete CLIClient satisfies WorktreeWriter
- **WHEN** the concrete CLI client is type-asserted against `WorktreeWriter`
- **THEN** the assertion SHALL succeed at compile time

### Requirement: BranchWriter exposes DeleteBranch and IsBranchMerged

The system SHALL expose a `BranchWriter` role interface with exactly two context-aware methods: `DeleteBranch` and `IsBranchMerged`. The role SHALL be defined consumer-side.

#### Scenario: BranchWriter is the only role with both DeleteBranch and IsBranchMerged
- **WHEN** a caller requires both `DeleteBranch` and `IsBranchMerged`
- **THEN** the role SHALL be `BranchWriter` and SHALL NOT include any other method

#### Scenario: Concrete CLIClient satisfies BranchWriter
- **WHEN** the concrete CLI client is type-asserted against `BranchWriter`
- **THEN** the assertion SHALL succeed at compile time

### Requirement: Role interfaces are defined consumer-side

All six role interfaces SHALL be declared in `internal/core/git.go`. No role interface SHALL be declared in `internal/git/`, `internal/cmdutil/`, or any command file. The roles SHALL be importable from `twiggit/internal/core` without importing any I/O package.

#### Scenario: Role interfaces importable from internal/core only
- **WHEN** a service in `internal/core/` requires a git role
- **THEN** the import path SHALL be `twiggit/internal/core` and the role name SHALL be one of the six declared in `git.go`

#### Scenario: No role declared in internal/git
- **WHEN** the package `internal/git/` is searched for interface declarations
- **THEN** no role interface SHALL be declared there

### Requirement: Service constructors take only the role interfaces they consume

Each service constructor SHALL accept the role interfaces it consumes as separate positional arguments. No service constructor SHALL take a composite umbrella interface that re-aggregates the role interfaces.

#### Scenario: WorktreeService takes the four role interfaces it consumes
- **WHEN** `core.NewWorktreeService` is invoked
- **THEN** the parameter list SHALL include `BranchReader`, `RepositoryReader`, `WorktreeWriter`, and `BranchWriter` and SHALL NOT include `RepositoryOpener` or `RemoteReader`

#### Scenario: ProjectService takes the three role interfaces it consumes
- **WHEN** `core.NewProjectService` is invoked
- **THEN** the parameter list SHALL include `RepositoryOpener`, `RepositoryReader`, and `RemoteReader` and SHALL NOT include `WorktreeWriter` or `BranchWriter`

### Requirement: Concrete clients satisfy their roles via compile-time checks

The concrete read-side client SHALL satisfy all four read-side roles (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`); the concrete write-side client SHALL satisfy both write-side roles (`WorktreeWriter`, `BranchWriter`). Each satisfaction SHALL be asserted at compile time via a `var _ core.Role = (*Concrete)(nil)` declaration placed in the concrete client's source file.

#### Scenario: Compile-time check is present in the concrete client source
- **WHEN** the concrete client's source file is read
- **THEN** it SHALL contain a `var _ core.<Role> = (*<Concrete>)(nil)` declaration for every role it satisfies

#### Scenario: Method-set drift fails the build
- **WHEN** a role's method is renamed or removed on the concrete client
- **THEN** `go build ./...` SHALL fail with a "cannot use" type-mismatch error referencing the affected role
