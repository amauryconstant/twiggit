## ADDED Requirements

### Requirement: Shell Type Inference

The system SHALL infer shell type from configuration file path using pattern matching to reduce command verbosity.

#### Scenario: Infer bash from standard config file

- **WHEN** config file path is `.bashrc`, `.bash_profile`, or `.profile`
- **THEN** system SHALL infer shell type as bash
- **AND** inference SHALL be based on filename pattern matching

#### Scenario: Infer bash from custom config file

- **WHEN** config file path is `custom.bash`, `my-bash-config`, or ends with `.bash`
- **THEN** system SHALL infer shell type as bash
- **AND** inference SHALL be based on filename pattern matching

#### Scenario: Infer zsh from standard config file

- **WHEN** config file path is `.zshrc` or `.zprofile`
- **THEN** system SHALL infer shell type as zsh
- **AND** inference SHALL be based on filename pattern matching

#### Scenario: Infer zsh from custom config file

- **WHEN** config file path ends with `.zsh`
- **THEN** system SHALL infer shell type as zsh
- **AND** inference SHALL be based on filename pattern matching

#### Scenario: Infer fish from config file

- **WHEN** config file path contains "fish" in name
- **AND** path includes `config.fish` or `.fishrc`
- **THEN** system SHALL infer shell type as fish
- **AND** inference SHALL be based on filename pattern matching

#### Scenario: Return error when shell type cannot be inferred

- **WHEN** config file path does not match any shell type pattern
- **THEN** system SHALL return inference error
- **AND** error message SHALL indicate cannot infer shell type
- **AND** error message SHALL include config file path
- **AND** error message SHALL suggest using --shell flag to specify shell type
- **AND** system SHALL provide list of supported shells (bash, zsh, fish)

---

### Requirement: Install to Explicit Config File

The system SHALL install shell wrapper to explicitly specified configuration file path.

#### Scenario: Install to existing config file

- **WHEN** user runs `twiggit init /custom/path/config` with existing config file
- **AND** shell type is inferred or specified
- **AND** wrapper is not already installed
- **THEN** system SHALL generate shell-specific wrapper
- **AND** system SHALL append wrapper to specified config file
- **AND** wrapper SHALL include block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** system SHALL not modify existing file content
- **AND** success message SHALL indicate installation completed
- **AND** success message SHALL include config file path

#### Scenario: Install to missing config file

- **WHEN** user runs `twiggit init /custom/path/config` with missing config file
- **AND** parent directory exists and is writable
- **THEN** system SHALL create empty config file with permissions 0644
- **AND** system SHALL append wrapper to new config file
- **AND** wrapper SHALL include block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** success message SHALL indicate installation completed
- **AND** success message SHALL indicate file was created
- **AND** success message SHALL include config file path

#### Scenario: Return error for unwritable config directory

- **WHEN** user runs `twiggit init /custom/path/config`
- **AND** config file does not exist
- **AND** parent directory is not writable
- **THEN** system SHALL return error indicating config file not writable
- **AND** error message SHALL include config file path
- **AND** installation SHALL not proceed

#### Scenario: Return error for missing parent directory

- **WHEN** user runs `twiggit init /custom/path/config`
- **AND** parent directory does not exist
- **THEN** system SHALL return error indicating directory not found
- **AND** error message SHALL include config file path
- **AND** installation SHALL not proceed

---

### Requirement: Force Reinstall with Block Delimiters

The system SHALL remove existing wrapper blocks before reinstalling when --force flag is provided.

#### Scenario: Remove old wrapper block before reinstall

- **WHEN** user runs `twiggit init <config-file> --force`
- **AND** config file contains existing wrapper block
- **THEN** system SHALL detect block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** system SHALL remove entire wrapper block including delimiters
- **AND** system SHALL preserve all other config file content
- **AND** system SHALL append new wrapper block with delimiters
- **AND** system SHALL not create duplicate wrapper blocks
- **AND** success message SHALL indicate reinstallation completed

#### Scenario: Handle missing end delimiter on force

- **WHEN** user runs `twiggit init <config-file> --force`
- **AND** config file contains only BEGIN delimiter without END delimiter
- **THEN** system SHALL treat as incomplete wrapper installation
- **AND** system SHALL remove partial wrapper block
- **AND** system SHALL append complete wrapper block
- **AND** warning message SHALL indicate incomplete wrapper was removed

#### Scenario: Handle missing begin delimiter on force

- **WHEN** user runs `twiggit init <config-file> --force`
- **AND** config file contains only END delimiter without BEGIN delimiter
- **THEN** system SHALL treat as orphaned delimiter
- **AND** system SHALL remove orphaned delimiter
- **AND** system SHALL append complete wrapper block
- **AND** warning message SHALL indicate orphaned delimiter was removed
