# Homebrew Distribution

## Purpose

 TBD: Define purpose for Homebrew distribution capability

## Requirements

### Requirement: GoReleaser generates Homebrew formula
The `.goreleaser.yml` configuration SHALL include a `brews` section that generates a Homebrew formula for twiggit.

#### Scenario: Brews section configuration
- **WHEN** GoReleaser runs during release
- **THEN** a Homebrew formula SHALL be generated with correct binary name, description, and homepage

#### Scenario: Tap repository publishing
- **WHEN** a release is created
- **THEN** the formula SHALL be pushed to `amoconst/homebrew-tap` in the `Formula` directory

### Requirement: Formula includes test command
The generated Homebrew formula SHALL include a test command that verifies the binary works.

#### Scenario: Test command execution
- **WHEN** Homebrew runs `brew test twiggit`
- **THEN** the command `twiggit version` SHALL execute successfully

### Requirement: Formula installs binary correctly
The formula SHALL install the twiggit binary to the Homebrew bin directory.

#### Scenario: Binary installation
- **WHEN** user runs `brew install amoconst/tap/twiggit`
- **THEN** the `twiggit` binary SHALL be available in PATH
