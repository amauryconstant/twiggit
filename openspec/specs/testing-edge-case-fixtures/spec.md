# Capability: Edge-Case Fixtures

## Purpose

Pre-built git repository fixtures in `test/fixtures/` covering
unusual states (corrupted objects, bare repos, submodules, detached
HEAD) so that E2E and integration tests can exercise graceful
handling.

## Requirements

### Requirement: Corrupted repository fixture

The fixture SHALL provide a repository with intentionally corrupted
git objects so tests can verify graceful error handling.

#### Scenario: List on corrupted repo

- **WHEN** `twiggit list` runs against a repo with corrupted
  `.git/objects`
- **THEN** the system SHALL return a descriptive error
- **AND** SHALL NOT panic

#### Scenario: Context detection in corrupted repo

- **WHEN** context detection runs in a corrupted repository
- **THEN** the system SHALL identify the directory as a git-context
- **AND** SHALL handle read errors gracefully

### Requirement: Bare repository fixture

The fixture SHALL provide a bare git repository. Worktree operations
on bare repos SHALL return descriptive errors.

#### Scenario: List on bare repo

- **WHEN** `twiggit list` runs against a bare repository
- **THEN** the system SHALL return an appropriate error or empty list
- **AND** SHALL NOT panic

#### Scenario: Create worktree on bare repo

- **WHEN** user attempts `twiggit create myproject/feature` against a
  bare repository
- **THEN** the system SHALL return a descriptive error
- **AND** SHALL explain the bare-repository limitation

### Requirement: Submodule repository fixture

The fixture SHALL provide a repo with git submodules. Context
detection SHALL correctly identify the main repo and SHALL NOT
confuse a submodule for the main project.

#### Scenario: List with submodules

- **WHEN** `twiggit list` runs against a repo with submodules
- **THEN** the system SHALL list main repository worktrees
- **AND** submodules SHALL be handled appropriately

#### Scenario: Context detection with submodules

- **WHEN** context detection runs in a repo with submodules
- **THEN** the system SHALL identify the main repository
- **AND** SHALL NOT confuse a submodule path for the main project

### Requirement: Detached HEAD fixture

The fixture SHALL provide a repo in detached HEAD state. The system
SHALL still list worktrees and SHALL report the detached state.

#### Scenario: List in detached HEAD

- **WHEN** `twiggit list` runs while in detached HEAD
- **THEN** the system SHALL list worktrees correctly
- **AND** SHALL report the detached state in the worktree info

#### Scenario: Context detection in detached HEAD

- **WHEN** context detection runs in detached HEAD
- **THEN** the system SHALL identify the context correctly
- **AND** the branch name MAY be reported as `HEAD` or `detached`
