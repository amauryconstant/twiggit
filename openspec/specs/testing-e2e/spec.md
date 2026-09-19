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

1. **Unit** (most tests, fastest) — `internal/service/` and
   `internal/domain/` with mocks
2. **Integration** — `test/integration/` with real git repos
3. **E2E** (fewest tests, slowest) — `test/e2e/` against the built
   binary

The cmd layer is tested exclusively via E2E; see
`infrastructure-release` for the rationale.



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
`test/fixtures/` provides a corrupted repo, a bare repo, a repo with
submodules, and a detached-HEAD repo. See `testing-edge-case-fixtures`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Iostreams injection

Tests SHALL inject a fake IOStreams when testing the cmd layer in
isolation (without building the binary). The injection pattern uses
`runF` injection at the command constructor. See
`testing-helpers`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
