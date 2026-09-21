# Spec Delta

## MODIFIED Requirements

### Requirement: Precondition failures use t.Fatal

Helper functions in `test/helpers/` that establish test
preconditions (creating a fixture repo, configuring git, etc.)
SHALL signal a precondition failure via `t.Fatal` (or
`t.Fatalf`) rather than `panic`. Helpers that do not have
access to a `*testing.T` (e.g., pure constructors) SHALL
return an error and the caller SHALL propagate via
`t.Fatal`.

#### Scenario: Repo creation precondition uses t.Fatal
- **WHEN** `test/helpers/repo.go` is called with a project
  name that cannot be created (e.g., empty string)
- **THEN** the helper calls `t.Fatal("project name cannot
  be empty")` instead of `panic(...)`; the test fails with
  a normal test failure rather than a panic stack trace

## ADDED Requirements

### Requirement: t.Context migration

Test functions in `cmd/`, `internal/service/`,
`internal/infrastructure/`, and `test/integration/` SHALL
replace `context.Background()` literals with `t.Context()`
(Go 1.24+) so that test cancellation propagates correctly
and the test process exits promptly when the test ends.

#### Scenario: Test uses t.Context
- **WHEN** a service-layer test creates a context to pass to
  a service method
- **THEN** the context comes from `t.Context()` and is
  automatically cancelled when the test ends

#### Scenario: Mock matcher uses Anything for context
- **WHEN** a mock expectation is registered with a context
  argument and the production code now passes `t.Context()`
- **THEN** the mock uses `mock.Anything` or
  `mock.MatchedBy(func(ctx context.Context) bool { return
  ctx != nil })` instead of `context.Background()` as the
  exact-match argument
