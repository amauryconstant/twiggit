# Spec Delta

## Purpose

Exposes the bash/zsh/fish shell wrapper templates as pure domain
functions so test fixtures and service-layer code can read or
seed wrapper content without constructing a real
`ShellInfrastructure` (which writes to the user's home directory).

## ADDED Requirements

### Requirement: Pure domain wrapper function

The system SHALL provide `domain.ShellWrapper(shellType)` that
returns the eval-safe wrapper template string for the requested
shell type. The function SHALL NOT perform any I/O (no filesystem
access, no environment inspection) and SHALL return the same byte
sequence on every call.

#### Scenario: Bash wrapper content is deterministic
- **WHEN** `domain.ShellWrapper("bash")` is called twice in the
  same process
- **THEN** both calls return byte-identical strings

#### Scenario: Unsupported shell type returns an error
- **WHEN** `domain.ShellWrapper("tcsh")` is called for a shell
  type that has no registered template
- **THEN** the function returns an empty string and a non-nil
  error satisfying `errors.Is(err, domain.ErrInvalidShellType)`

### Requirement: Service-layer tests use the domain function

The service-layer tests SHALL call `domain.ShellWrapper(shellType)`
to seed mock return values rather than constructing a real
`ShellInfrastructure` to capture wrapper content. The mock setup
pattern SHALL be: call `domain.ShellWrapper(shellType)` once per
shell type under test, store the result, and use it as the
mock's return value.

#### Scenario: Service test seeds bash wrapper via domain function
- **WHEN** a `ShellService` test sets up a mock to return
  `GenerateWrapper` output for shell type "bash"
- **THEN** the test imports only `internal/domain` (not
  `internal/infrastructure`) and the mock return value equals
  `domain.ShellWrapper("bash")`

#### Scenario: Service test seeds zsh and fish wrappers
- **WHEN** a table-driven test exercises three shell types
  ("bash", "zsh", "fish")
- **THEN** each row seeds its mock return value via
  `domain.ShellWrapper(shellType)` with no filesystem dependency

### Requirement: Wrapper content is byte-identical to prior infrastructure output

The wrapper template content returned by `domain.ShellWrapper`
SHALL be byte-for-byte identical to the content previously
returned by the infrastructure layer's wrapper generator. The
move from infrastructure to domain is a pure relocation of
behavior, not a content change.

#### Scenario: Golden test compares wrapper content across relocation
- **WHEN** the test suite compares the current wrapper content
  against a stored golden file
- **THEN** the comparison passes with no diff (byte-identical
  content), proving the relocation did not alter user-visible
  output
