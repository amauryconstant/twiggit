# Capability: End-to-End Test Suite

## Purpose

Ginkgo/Gomega test suite that runs against the built binary against
real git repositories. Covers all user-visible commands and the
context-aware behaviors (project, worktree, outside git).

## Requirements

### Requirement: Suite structure

The E2E suite SHALL live in `test/e2e/` and SHALL organize tests by
command: `list_test.go`, `create_test.go`, `delete_test.go`,
`prune_test.go`, `cd_test.go`, `init_test.go`, `completion_test.go`,
`version_test.go`, etc.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Test pyramid

The project's test pyramid SHALL be, in order of increasing scope and
decreasing volume:

1. **Unit** (most tests, fastest) — `internal/core/` with mocks and
   `cmd/*_test.go` covering per-command logic
2. **Integration** — `test/integration/` with real git repos
3. **E2E** (fewest tests, slowest) — `test/e2e/` against the built
   binary

The cmd layer is covered by both unit tests in `cmd/*_test.go` and
end-to-end tests in `test/e2e/`; the E2E suite verifies the full CLI
contract including cobra wiring and exit codes.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Binary build

Before E2E tests run, the suite SHALL build the twiggit binary into a
temp directory using `go build`. The built binary path SHALL be
exported to tests via an env var or shared helper.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Context isolation

Each E2E test SHALL work in its own temp directory simulating
`$HOME`, with fresh `Projects/` and `Worktrees/` subdirectories.
Tests SHALL NOT touch the user's real `$HOME`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Coverage of all commands

The E2E suite SHALL cover, at minimum: list, create, delete, prune,
cd, init, completion, version. Each command SHALL have at least one
"happy path" test and one error-path test.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Fixtures

Common git scenarios SHALL be available as fixtures:
`test/e2e/fixtures/repos/` provides a corrupted repo (`corrupted.tar.gz`),
a bare repo (`bare-main.tar.gz`), a repo with submodules
(`submodule.tar.gz`), and a detached-HEAD repo (`detached.tar.gz`). See
`testing-edge-case-fixtures`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Iostreams injection

Tests SHALL inject a fake IOStreams when testing the cmd layer in
isolation (without building the binary). The injection pattern uses
`runF` injection at the command constructor. See
`testing-helpers`.

#### Scenario: IOStreams injection via runF

- **WHEN** a test invokes `NewCmdList(testFactory, runF)` with a `runF` that reads from an `iostreams.Test()` instance
- **THEN** the command SHALL write data to the test stdout buffer (not `os.Stdout`)
- **AND** the test asserts on the buffer contents without touching the host process's stdio

### Requirement: E2E testify hygiene

Ginkgo `BeforeEach` setup SHALL use `require.NoError` / `require.NotNil` for preconditions (never `assert`). Every mock injected via `WithMocks` SHALL be verified by `mock.AssertExpectations(GinkgoT())` at suite teardown so missing expectations fail the suite rather than pass silently.

#### Scenario: Ginkgo BeforeEach uses require

- **WHEN** a Ginkgo `BeforeEach` reads `mockClient.SomeCall(...)` setup
- **THEN** the precondition checks SHALL use `Expect(err).NotTo(HaveOccurred())` (Gomega equivalent of `require`) — silent failures SHALL NOT be tolerated

#### Scenario: Mock expectations verified at teardown

- **WHEN** a Ginkgo `AfterEach` runs after a suite that injected mocks via `WithMocks`
- **THEN** `mock.AssertExpectations(GinkgoT())` SHALL be called
- **AND** missing expectations SHALL fail the suite


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
