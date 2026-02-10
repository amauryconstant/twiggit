## ADDED Requirements

### Requirement: Shell Integration Setup

The system SHALL install shell wrapper functions for bash, zsh, and fish to enable seamless directory navigation between worktrees and projects via the `twiggit cd` command.

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

### Requirement: Shell Wrapper Behavior

The installed shell wrapper SHALL intercept `twiggit cd` calls, change directories automatically, provide escape hatch for built-in cd, and pass through all other commands.

#### Scenario: Intercept twiggit cd commands

- **WHEN** shell wrapper is installed and user runs `twiggit cd <target>`
- **THEN** wrapper SHALL intercept command
- **AND** wrapper SHALL execute `twiggit cd` to resolve path
- **AND** wrapper SHALL change to shell's working directory to resolved path
- **AND** directory change SHALL be silent (no extra output)

#### Scenario: Provide escape hatch with builtin cd

- **WHEN** user runs `builtin cd <path>` with wrapper installed
- **THEN** wrapper SHALL pass through to shell's built-in cd command
- **AND** wrapper SHALL NOT intercept or modify command
- **AND** user SHALL have full control of built-in cd functionality

#### Scenario: Pass through non-cd commands

- **WHEN** user runs any command that is not `twiggit cd`
- **THEN** wrapper SHALL NOT intercept command
- **AND** command SHALL execute normally without wrapper involvement
- **AND** wrapper SHALL have no side effects on other commands

#### Scenario: Output path for wrapper consumption

- **WHEN** `twiggit cd` command executes successfully
- **THEN** system SHALL output absolute path to stdout
- **AND** wrapper SHALL capture this output
- **AND** wrapper SHALL use captured path to change directory

#### Scenario: Handle cd command errors

- **WHEN** `twiggit cd` command fails to resolve path
- **THEN** system SHALL output error message to stderr
- **AND** non-zero exit code SHALL be returned
- **AND** wrapper SHALL NOT change directory
- **AND** error message SHALL be displayed to user

---

### Requirement: Shell Installation Validation

The system SHALL validate whether shell integration is properly installed for a given shell type.

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

#### Scenario: Detect configuration file location with fallback

- **WHEN** system validates or installs shell wrapper for a shell type
- **THEN** system SHALL detect appropriate configuration file
- **AND** bash SHALL try `.bashrc`, `.bash_profile`, `.profile` in order
- **AND** zsh SHALL try `.zshrc`, `.zprofile`, `.profile` in order
- **AND** fish SHALL try `.config/fish/config.fish`, `config.fish`, `.fishrc` in order
- **AND** first existing config file SHALL be used
- **AND** detection SHALL handle both standard and user-specific locations

#### Scenario: Generate wrapper with timestamp

- **WHEN** system generates a shell wrapper
- **THEN** system SHALL include a timestamp in the wrapper comment
- **AND** timestamp format SHALL be "2006-01-02 15:04:05"
- **AND** timestamp SHALL be generated at wrapper creation time
- **AND** wrapper SHALL indicate when it was generated

#### Scenario: Compose wrapper with template replacements

- **WHEN** system generates a wrapper from a template
- **THEN** system SHALL replace template placeholders with values
- **AND** `{{SHELL_TYPE}}` SHALL be replaced with actual shell type
- **AND** `{{TIMESTAMP}}` SHALL be replaced with generated timestamp
- **AND** replacement SHALL be deterministic and side-effect free
- **AND** all occurrences of each placeholder SHALL be replaced
