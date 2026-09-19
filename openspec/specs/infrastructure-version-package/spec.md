# Capability: Version

## Purpose

Display the twiggit version, commit hash, and build date. Version
metadata is injected at build time via `-ldflags`.

## Requirements

### Requirement: Print version info

The system SHALL print the version string, commit hash, and build date
when the user runs `twiggit version`.

#### Scenario: Default version output

- **WHEN** user runs `twiggit version`
- **THEN** system SHALL print the version (semver, e.g. `0.5.0`)
- **AND** SHALL print the commit hash (short)
- **AND** SHALL print the build date (ISO 8601)

### Requirement: Build-time injection

The system SHALL inject `Version`, `Commit`, and `Date` variables in
`internal/version` via `-ldflags` at build time. The GoReleaser
pipeline owns the build flag wiring.

#### Scenario: ldflags targets

- **WHEN** the binary is built
- **THEN** the linker SHALL set `twiggit/internal/version.Version`
- **AND** SHALL set `twiggit/internal/version.Commit`
- **AND** SHALL set `twiggit/internal/version.Date`

### Requirement: `String()` formatter

The system SHALL provide a `String()` formatter in `internal/version`
that returns a multi-line version block. The formatter SHALL NOT
include the `twiggit ` prefix; the cmd layer prepends it.

#### Scenario: String output

- **WHEN** `version.String()` is called
- **THEN** the returned string SHALL contain the version, commit, and date
- **AND** SHALL NOT start with the literal `twiggit ` prefix
