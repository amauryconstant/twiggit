# Capability: Version Package

## Purpose

Defines the build-time-injected version metadata for the binary. Lives in `internal/version/version.go`. The `cmd/version` subcommand reads `version.String()` and prepends the `twiggit ` prefix; GoReleaser injects the actual `Version`, `Commit`, and `Date` values via `-X` ldflags at the goreleaser pipeline.

## Requirements

### Requirement: Variables inject build metadata

The `internal/version` package SHALL export three string variables: `Version` (default `"dev"`), `Commit` (default `""`), and `Date` (default `""`). GoReleaser and local builds inject the actual values via `-ldflags '-X twiggit/internal/version.Version=<v> -X twiggit/internal/version.Commit=<sha> -X twiggit/internal/version.Date=<iso>'`.

#### Scenario: Dev build defaults

- **WHEN** the binary is built without `-ldflags` injection (a development build)
- **THEN** `Version == "dev"`, `Commit == ""`, `Date == ""`

#### Scenario: Release build injected values

- **WHEN** the binary is built with `-ldflags '-X twiggit/internal/version.Version=1.2.3 -X twiggit/internal/version.Commit=abc1234 -X twiggit/internal/version.Date=2026-09-22T14:30:00Z'`
- **THEN** `Version == "1.2.3"`, `Commit == "abc1234"`, `Date == "2026-09-22T14:30:00Z"`

### Requirement: String() formatter

`version.String() string` SHALL return a formatted version line:
- Full info: `<version> (<commit>) <date>`
- No commit: `<version> () `
- No date: `<version> (<commit>) `

The `twiggit ` prefix is added by the cmd layer, NOT by `String()`.

#### Scenario: Complete info format

- **WHEN** `Version="1.2.3"`, `Commit="abc1234"`, `Date="2026-09-22"`
- **THEN** `version.String()` returns `"1.2.3 (abc1234) 2026-09-22"`

#### Scenario: Dev build format

- **WHEN** `Version="dev"`, `Commit=""`, `Date=""`
- **THEN** `version.String()` returns `"dev () "`

#### Scenario: No trailing prefix

- **WHEN** `version.String()` is called
- **THEN** the result SHALL NOT start with the literal `twiggit ` prefix
- **AND** the cmd layer prepends `twiggit ` when emitting the final user-facing line

### Requirement: Version command output

`twiggit version` SHALL print `twiggit <version.String()>` to stdout and exit 0. The output SHALL go to stdout (not stderr) so scripts can capture it via `twiggit version | awk '{print $2}'`.

#### Scenario: Version command stdout

- **WHEN** user runs `twiggit version`
- **THEN** stdout SHALL contain `twiggit <version.String()>` followed by a newline
- **AND** exit code SHALL be 0

### Requirement: Variables are package-level for ldflags

`Version`, `Commit`, and `Date` SHALL be package-level `var` declarations (not `const`, not unexported fields on a struct). The Go linker only injects string values into package-level `var` declarations of string type; struct fields or const declarations would not receive the injection.

#### Scenario: Variables are package-level

- **WHEN** `internal/version/version.go` is read
- **THEN** the three variables SHALL be declared with `var` (not `const`)
- **AND** they SHALL NOT be hidden behind a struct or unexported accessor