# Spec Delta

## ADDED Requirements

### Requirement: Build-once TestMain

The E2E test suite SHALL define a single `TestMain` that
compiles the `twiggit` binary once into a shared temporary
directory. The binary path SHALL be exported to E2E tests via
an environment variable. Each E2E test SHALL NOT rebuild the
binary; the build happens exactly once per test process
invocation.

#### Scenario: TestMain builds binary once
- **WHEN** `go test -tags=e2e ./test/e2e/...` runs
- **THEN** `TestMain` invokes `go build` exactly once; the
  suite-level setup time is reduced from O(tests) to O(1)

#### Scenario: Test references binary via env var
- **WHEN** an E2E test needs to invoke the binary
- **THEN** it reads the binary path from the env var set by
  `TestMain`; no per-test binary compilation occurs

### Requirement: Shuffle-on test execution

The unit and integration test tasks (`mise run test`,
`mise run test:race`, `mise run test:e2e`) SHALL include the
`-shuffle=on` flag so that test order is randomized. Tests
SHALL be order-independent; any test that depends on
shared package-level state SHALL fail under shuffle.

#### Scenario: Order-dependent test fails under shuffle
- **WHEN** a test mutates package-level state and a later
  test depends on the initial value
- **THEN** running with `-shuffle=on` exposes the dependency
  and the test fails until the dependency is removed

### Requirement: Goleak test-exit verification

The test suite SHALL install `go.uber.org/goleak.VerifyTestMain`
as the final step of `TestMain`. Any goroutine that is still
running when the test process exits SHALL be reported as a
leak, and the process SHALL exit non-zero.

#### Scenario: Goroutine leak surfaces on test exit
- **WHEN** a test starts a goroutine that does not terminate
  by the time the test ends
- **THEN** `goleak.VerifyTestMain` reports the leak with the
  goroutine's stack trace and the test process exits non-zero

### Requirement: Mock cleanup fixture

Every test that uses a testify mock via `.On(...)` SHALL
register a `t.Cleanup(mock.AssertExpectations)` so that any
expected but unfulfilled call is reported at the end of the
test. The fixture SHALL be applied via a helper function
(e.g., `mocks.NewMockX(t)`) rather than per-call boilerplate.

#### Scenario: Unfulfilled expectation is reported
- **WHEN** a test registers `.On("Method", ...).Return(...)`
  but the production code does not call `Method`
- **THEN** `t.Cleanup(mock.AssertExpectations)` fails the
  test with a message naming the unfulfilled expectation

#### Scenario: Helper wraps mock construction
- **WHEN** a test calls `mocks.NewMockX(t)`
- **THEN** the helper constructs the mock, registers the
  `AssertExpectations` cleanup, and returns the mock for
  the test to configure

### Requirement: Testing-short guards for unguarded integration files

The integration test files that lack a `testing.Short()` skip
guard SHALL add one at the top of every test (or at the file
level via `testing.Short()` check in `TestMain`). The
following files SHALL be guarded:
`cli_commands_test.go`, `completion_test.go`,
`config_test.go`, `context_detection_test.go`,
`hook_runner_test.go`, `path_utilities_test.go`.

#### Scenario: Short mode skips integration tests
- **WHEN** `go test -short ./test/integration/...` runs
- **THEN** the formerly-unguarded integration tests exit
  immediately via `t.Skip(...)`; the suite completes in
  seconds rather than minutes
