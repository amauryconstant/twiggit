# Spec Delta

## ADDED Requirements

### Requirement: Caller context propagates through resolver

The context resolver SHALL thread the caller's `ctx` (received
as the first argument to `ResolveIdentifier`,
`GetResolutionSuggestions`, and related methods) through every
git client call it makes (`ListWorktrees`, `GetRepositoryStatus`,
`ListBranches`, etc.). The resolver SHALL NOT call
`context.Background()` to substitute a fresh context.

#### Scenario: Caller context cancellation aborts resolver
- **WHEN** a caller passes a context that is already cancelled
  and invokes `ResolveIdentifier(ctx, "feature")`
- **THEN** the resolver returns promptly with the
  context-cancelled error, without performing any git I/O

#### Scenario: Caller context with deadline is respected
- **WHEN** a caller passes a context with a 5-second deadline
  and invokes `GetResolutionSuggestions(ctx, partial)`
- **THEN** the resolver returns within 5 seconds plus a small
  grace period, even if the git client is slow

### Requirement: Project summary replaces local ProjectRef

The resolver SHALL return `*domain.ProjectSummary` from
`discoverProjects` and `GetResolutionSuggestions` instead of
defining a local `ProjectRef` struct. Callers SHALL consume the
domain type directly without re-wrapping. The local
`ProjectRef` type SHALL be removed.

#### Scenario: ProjectSuggestion carries domain.ProjectSummary
- **WHEN** the resolver returns a project suggestion
- **THEN** the suggestion's `Project` field is a
  `*domain.ProjectSummary` and the caller can pass it to
  other domain-typed APIs without conversion

### Requirement: Chdir-based tests use t.Chdir

The integration test suite SHALL use `t.Chdir` (Go 1.24+) for
tests that need to change the working directory. The
`os.Chdir` + `defer os.Chdir` antipattern SHALL NOT appear in
new tests.

#### Scenario: Integration test uses t.Chdir
- **WHEN** an integration test needs to operate from a
  temporary directory
- **THEN** it calls `t.Chdir(tempDir)` and the test framework
  restores the previous working directory on test end

#### Scenario: Shuffle-safe working directory changes
- **WHEN** the test suite runs with `-shuffle=on`
- **THEN** tests that change working directory do not affect
  subsequent tests because `t.Chdir` restoration is per-test
