## Purpose

Provides shell initialization functionality for twiggit with stdout output as the default mode, enabling eval-based activation, and file installation as an opt-in mode via flags.

## Requirements

### Requirement: Shell Type Inference

The system SHALL accept shell type as a positional argument (bash, zsh, fish) with auto-detection from $SHELL when omitted, replacing the previous --shell flag.

#### Scenario: Use explicit shell type from positional argument

- **WHEN** user runs `twiggit init bash` or `twiggit init zsh` or `twiggit init fish`
- **THEN** system SHALL use the explicitly specified shell type
- **AND** no --shell flag SHALL be required

#### Scenario: Auto-detect shell from SHELL env when positional omitted

- **WHEN** user runs `twiggit init` without positional argument
- **AND** SHELL environment variable is set to supported shell
- **THEN** system SHALL detect shell type from SHELL environment variable
- **AND** system SHALL proceed with detected shell type

#### Scenario: Return error when shell cannot be detected

- **WHEN** user runs `twiggit init` without positional argument
- **AND** SHELL environment variable is not set or contains unsupported shell
- **THEN** system SHALL return detection error
- **AND** error message SHALL suggest specifying shell type as positional argument
- **AND** system SHALL provide list of supported shells (bash, zsh, fish)

---

### Requirement: Install to Explicit Config File

The system SHALL install shell wrapper to a config file only when --install flag is provided, with optional custom config file via --config flag.

#### Scenario: Install to auto-detected config file with --install

- **WHEN** user runs `twiggit init --install` or `twiggit init bash --install`
- **AND** no --config flag is provided
- **AND** wrapper is not already installed
- **THEN** system SHALL auto-detect config file for the shell type
- **AND** system SHALL generate shell-specific wrapper
- **AND** system SHALL append wrapper to auto-detected config file
- **AND** wrapper SHALL include block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** system SHALL append Carapace completion sourcing
- **AND** completion SHALL include block delimiters (`### BEGIN/END TWIGGIT COMPLETION`)
- **AND** success message SHALL indicate installation completed
- **AND** success message SHALL include config file path

#### Scenario: Install to explicit config file with --config

- **WHEN** user runs `twiggit init bash --install --config ~/.customrc`
- **AND** --install flag is present
- **AND** --config flag specifies a path
- **AND** wrapper is not already installed
- **THEN** system SHALL generate shell-specific wrapper
- **AND** system SHALL append wrapper to specified config file
- **AND** wrapper SHALL include block delimiters
- **AND** success message SHALL indicate installation completed
- **AND** success message SHALL include specified config file path

#### Scenario: Return error when --config used without --install

- **WHEN** user runs `twiggit init bash --config ~/.bashrc` without --install
- **THEN** system SHALL return validation error
- **AND** error message SHALL indicate --config requires --install flag

#### Scenario: Return error for missing parent directory

- **WHEN** user runs `twiggit init bash --install --config /nonexistent/path/config`
- **AND** parent directory does not exist
- **THEN** system SHALL return error indicating directory not found
- **AND** error message SHALL include config file path
- **AND** installation SHALL not proceed

---

### Requirement: Force Reinstall with Block Delimiters

The system SHALL remove existing wrapper blocks before reinstalling when --force flag is provided alongside --install.

#### Scenario: Force reinstall removes old blocks

- **WHEN** user runs `twiggit init bash --install --force`
- **AND** config file contains existing wrapper and/or completion blocks
- **THEN** system SHALL detect WRAPPER block delimiters
- **AND** system SHALL detect COMPLETION block delimiters
- **AND** system SHALL remove both blocks entirely including delimiters
- **AND** system SHALL preserve all other config file content
- **AND** system SHALL append new wrapper and completion blocks
- **AND** success message SHALL indicate reinstallation completed

#### Scenario: Return error when --force used without --install

- **WHEN** user runs `twiggit init bash --force` without --install
- **THEN** system SHALL return validation error
- **AND** error message SHALL indicate --force requires --install flag

#### Scenario: Handle missing end delimiter on force

- **WHEN** user runs `twiggit init bash --install --force`
- **AND** config file contains only BEGIN delimiter without END delimiter (WRAPPER or COMPLETION block)
- **THEN** system SHALL treat as incomplete installation
- **AND** system SHALL remove partial blocks
- **AND** system SHALL append complete wrapper and completion blocks
- **AND** warning message SHALL indicate incomplete blocks were removed

#### Scenario: Handle missing begin delimiter on force

- **WHEN** user runs `twiggit init bash --install --force`
- **AND** config file contains only END delimiter without BEGIN delimiter (WRAPPER or COMPLETION block)
- **THEN** system SHALL treat as orphaned delimiter
- **AND** system SHALL remove orphaned delimiters
- **AND** system SHALL append complete wrapper and completion blocks
- **AND** warning message SHALL indicate orphaned delimiters were removed

---

### Requirement: Stdout Output Mode

The system SHALL output shell wrapper to stdout by default (without --install), enabling eval-based activation without file modification.

#### Scenario: Output wrapper to stdout with auto-detected shell

- **WHEN** user runs `twiggit init` without --install flag
- **AND** SHELL environment variable is set to supported shell
- **THEN** system SHALL detect shell type from SHELL environment variable
- **AND** system SHALL generate shell-specific wrapper
- **AND** system SHALL output wrapper to stdout
- **AND** output SHALL be eval-safe (no metadata, only wrapper block)
- **AND** output SHALL include block delimiters
- **AND** output SHALL include Carapace completion sourcing
- **AND** system SHALL NOT modify any files

#### Scenario: Output wrapper to stdout with explicit shell

- **WHEN** user runs `twiggit init zsh` without --install flag
- **THEN** system SHALL generate zsh-specific wrapper
- **AND** system SHALL output wrapper to stdout
- **AND** output SHALL be eval-safe (no metadata, only wrapper block)
- **AND** system SHALL NOT modify any files

#### Scenario: Eval activation works correctly

- **WHEN** user runs `eval "$(twiggit init bash)"`
- **THEN** twiggit shell function SHALL be defined in current shell
- **AND** Carapace completion SHALL be sourced
- **AND** `twiggit cd <branch>` SHALL change directory
- **AND** `builtin cd <path>` SHALL use shell built-in

#### Scenario: Stdout output includes completion

- **WHEN** user runs `twiggit init bash` without --install
- **THEN** output SHALL include wrapper block with delimiters
- **AND** output SHALL include completion block with delimiters
- **AND** completion block SHALL source Carapace for the specified shell
