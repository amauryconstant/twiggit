## MODIFIED Requirements

### Requirement: Shell Integration Setup

The system SHALL install shell wrapper functions for bash, zsh, and fish to enable seamless directory navigation between worktrees and projects via `twiggit cd` command.

#### Scenario: Install shell wrapper with inferred shell type

- **WHEN** user runs `twiggit init ~/.bashrc`
- **AND** bash is a supported shell
- **AND** wrapper is not already installed
- **THEN** system SHALL infer shell type as bash from config file name
- **AND** system SHALL generate bash-specific wrapper function
- **AND** wrapper SHALL be added to specified configuration file (~/.bashrc)
- **AND** wrapper SHALL include block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** wrapper SHALL intercept `twiggit cd` calls and change directories
- **AND** wrapper SHALL provide `builtin cd` escape hatch
- **AND** success message SHALL indicate shell wrapper installed
- **AND** success message SHALL include config file path
- **AND** user SHALL be instructed to restart shell or source config file

#### Scenario: Install shell wrapper with explicit shell override

- **WHEN** user runs `twiggit init /custom/config --shell=zsh`
- **AND** zsh is a supported shell
- **AND** wrapper is not already installed
- **THEN** system SHALL use explicit shell type (zsh) instead of inference
- **AND** system SHALL generate zsh-specific wrapper function
- **AND** wrapper SHALL be added to specified configuration file (/custom/config)
- **AND** wrapper SHALL include block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** wrapper SHALL intercept `twiggit cd` calls and change directories
- **AND** wrapper SHALL provide `builtin cd` escape hatch
- **AND** success message SHALL indicate shell wrapper installed
- **AND** success message SHALL include config file path
- **AND** user SHALL be instructed to restart shell or source config file

#### Scenario: Skip installation if wrapper already exists

- **WHEN** user runs `twiggit init <config-file>`
- **AND** wrapper is already installed in specified config file
- **AND** --force flag is not provided
- **THEN** system SHALL skip installation
- **AND** message SHALL indicate "Shell wrapper already installed"
- **AND** message SHALL include config file path
- **AND** system SHALL suggest using --force to reinstall
- **AND** no changes SHALL be made to shell configuration

#### Scenario: Force reinstall shell wrapper

- **WHEN** user runs `twiggit init <config-file> --force`
- **AND** wrapper is already installed in specified config file
- **THEN** system SHALL remove existing wrapper block using delimiters
- **AND** system SHALL regenerate and reinstall wrapper
- **AND** wrapper SHALL include block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** success message SHALL indicate shell wrapper installed (not skipped)
- **AND** success message SHALL include config file path
- **AND** user SHALL be instructed to restart shell or source config file

#### Scenario: Dry run shell wrapper installation

- **WHEN** user runs `twiggit init <config-file> --dry-run`
- **THEN** system SHALL generate wrapper content without installing
- **AND** generated wrapper SHALL include block delimiters
- **AND** generated wrapper SHALL be displayed to stdout
- **AND** message SHALL indicate "Would install wrapper for <shell>"
- **AND** message SHALL include config file path
- **AND** no changes SHALL be made to shell configuration
- **AND** wrapper SHALL not be written to any file

#### Scenario: Return error when shell type inference fails

- **WHEN** user runs `twiggit init /custom/config.txt`
- **AND** shell type cannot be inferred from config file name
- **AND** --shell flag is not provided
- **THEN** system SHALL return error indicating cannot infer shell type
- **AND** error SHALL include config file path
- **AND** error SHALL suggest using --shell to specify shell type
- **AND** installation SHALL not proceed

---

### Requirement: Shell Installation Validation

The system SHALL validate whether shell integration is properly installed for a given shell type using block delimiters.

#### Scenario: Validate shell wrapper installation

- **WHEN** user runs `twiggit init --check ~/.bashrc`
- **AND** wrapper is present in ~/.bashrc file
- **THEN** validation SHALL succeed
- **AND** result SHALL indicate "Shell wrapper is installed"
- **AND** configuration file path (~/.bashrc) SHALL be included in result
- **AND** validation SHALL check for block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)

#### Scenario: Return not installed when wrapper missing

- **WHEN** user runs `twiggit init --check ~/.bashrc`
- **AND** wrapper is not present in ~/.bashrc file
- **THEN** validation SHALL fail
- **AND** result SHALL indicate "Shell wrapper not installed"
- **AND** configuration file path (~/.bashrc) SHALL be included in result
- **AND** validation SHALL check for block delimiters (`### BEGIN/END TWIGGIT WRAPPER`)
