# Capability: Cd Worktree

## Purpose

Resolve a project/branch identifier to an absolute filesystem path and
emit it to stdout, for the shell wrapper to `cd` into. Has no other
output and no side effects.

## Requirements

### Requirement: Resolve identifier to path

The system SHALL print the absolute path of the resolved worktree to
stdout on success, and nothing else.

#### Scenario: Resolve worktree by branch

- **WHEN** user runs `twiggit cd feature` from inside a project
- **AND** a worktree for `feature` exists
- **THEN** system SHALL print the worktree's absolute path to stdout
- **AND** SHALL exit with status 0

#### Scenario: Resolve by project/branch

- **WHEN** user runs `twiggit cd myproject/feature` from any context
- **AND** the worktree exists
- **THEN** system SHALL print its absolute path to stdout

### Requirement: Navigation on `twiggit cd` failure

The system SHALL NOT print a path to stdout when resolution fails.
Errors SHALL go to stderr with an actionable hint.

#### Scenario: Unknown target

- **WHEN** user runs `twiggit cd missing` and no worktree matches
- **THEN** system SHALL print an error to stderr
- **AND** SHALL exit with non-zero status
- **AND** SHALL NOT print any path to stdout

### Requirement: Wrapper invocation

The system SHALL assume the shell wrapper is installed when `twiggit cd`
is invoked; the wrapper intercepts the path on stdout and `cd`s to it.
The wrapper runtime behavior is owned by `cli-init`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Resolution via context resolver

The system SHALL delegate identifier resolution to `(*git.contextResolver).ResolveIdentifier(currentCtx, target)` constructed via `git.NewContextResolver(cfg, gitClient, gitClient)`. The resolution contract itself (detection priority, identifier parsing, suggestion generation) is owned by `git-context-resolver`; the cmd-layer wiring is owned here. Every error returned by the resolution path SHALL be lowercase with no trailing punctuation. The resolved `string` SHALL be returned as a defensive copy so the caller cannot mutate the resolver's internal buffer.

#### Scenario: Resolution delegates to git.NewContextResolver

- **WHEN** `twiggit cd <target>` runs
- **THEN** the cmd layer SHALL compose `git.NewContextDetector(cfg)` and `git.NewContextResolver(cfg, gitClient, gitClient)`
- **AND** SHALL invoke `(*contextResolver).ResolveIdentifier(ctx, target)` rather than any removed `application.*` API


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
