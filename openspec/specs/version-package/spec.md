# Purpose

Version package for build-time information injection.

## Requirements

### Requirement: Version variables in dedicated package

The version variables SHALL be defined in the `internal/version` package with the following characteristics:
- `Version` variable of type string with default value "dev"
- `Commit` variable of type string with default value "" (empty)
- `Date` variable of type string with default value "" (empty)
- All variables SHALL be accessible from external packages

#### Scenario: Default values in development build
- **WHEN** no ldflags are injected during build
- **THEN** Version SHALL equal "dev"
- **AND** Commit SHALL equal ""
- **AND** Date SHALL equal ""

#### Scenario: Variables accessible from cmd package
- **WHEN** cmd/version.go imports internal/version
- **THEN** version.Version, version.Commit, version.Date SHALL be accessible

### Requirement: Formatted version string output

The `internal/version` package SHALL provide a `String()` function that returns a formatted version string. The String() function SHALL NOT include the "twiggit " prefix - the command layer prepends it.

#### Scenario: Complete version information
- **WHEN** Version is "1.0.0", Commit is "abc123def456", Date is "2024-01-15"
- **THEN** String() SHALL return "1.0.0 (abc123def456) 2024-01-15"

#### Scenario: Short commit (≤7 chars)
- **WHEN** Commit has 7 or fewer characters
- **THEN** String() SHALL use commit as-is with no truncation

#### Scenario: Missing commit information
- **WHEN** Commit is empty string
- **THEN** String() SHALL return "<version> () " with empty parens and trailing space

#### Scenario: Missing date information
- **WHEN** Date is empty string but Commit is present
- **THEN** String() SHALL return "<version> (<commit>) " with trailing space

### Requirement: GoReleaser ldflags configuration

The GoReleaser configuration SHALL inject build-time values using the new package path.

#### Scenario: Ldflags target correct package
- **WHEN** GoReleaser builds the binary
- **THEN** ldflags SHALL target `twiggit/internal/version.Version`
- **AND** ldflags SHALL target `twiggit/internal/version.Commit`
- **AND** ldflags SHALL target `twiggit/internal/version.Date`
