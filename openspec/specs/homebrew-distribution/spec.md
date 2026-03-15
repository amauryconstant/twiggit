# Homebrew Distribution

## Purpose

 TBD: Define purpose for Homebrew distribution capability

## Requirements

### Requirement: GoReleaser generates Homebrew cask
The `.goreleaser.yml` configuration SHALL include a `homebrew_casks` section that generates a Homebrew cask for twiggit.

#### Scenario: Homebrew cask configuration
- **WHEN** GoReleaser runs during release
- **THEN** a Homebrew cask SHALL be generated with correct binary name, description, and homepage

#### Scenario: Tap repository publishing
- **WHEN** a release is created
- **THEN** the cask SHALL be pushed to `amoconst/homebrew-tap` in the `Casks` directory

### Requirement: Cask bypasses macOS quarantine for unsigned binary
The Homebrew cask SHALL include a post-install hook to remove the quarantine attribute from the unsigned binary.

#### Scenario: Quarantine removal on macOS
- **WHEN** the cask is installed on macOS
- **THEN** the com.apple.quarantine attribute SHALL be removed from the twiggit binary

### Requirement: Cask installs binary correctly
The cask SHALL install the twiggit binary to the Homebrew bin directory.

#### Scenario: Binary installation
- **WHEN** user runs `brew install amoconst/tap/twiggit`
- **THEN** the `twiggit` binary SHALL be available in PATH on macOS
