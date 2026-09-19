# Capability: Shell Auto-Detection

## Purpose

Detect the user's shell from `$SHELL` and from inferred shell config
file names, and validate shell types against the supported set.

## Requirements

### Requirement: `ShellType` and supported set

`domain.ShellType` SHALL be a string type. The system SHALL recognize
exactly three values: `bash`, `zsh`, `fish`. Other values SHALL be
rejected by `domain.IsValidShellType`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `DetectShellFromEnv`

`DetectShellFromEnv()` SHALL read `$SHELL` and return the matching
`ShellType` by substring match: a path containing `bash` maps to
`bash`, `zsh` to `zsh`, `fish` to `fish`. If `$SHELL` is unset, the
function SHALL return an error with code `SHELL_DETECTION_FAILED`.

#### Scenario: Detect bash

- **WHEN** `$SHELL = /bin/bash`
- **THEN** system SHALL return `ShellType("bash")`

#### Scenario: Detect zsh

- **WHEN** `$SHELL = /usr/local/bin/zsh`
- **THEN** system SHALL return `ShellType("zsh")`

#### Scenario: Detect fish

- **WHEN** `$SHELL = /usr/local/bin/fish`
- **THEN** system SHALL return `ShellType("fish")`

#### Scenario: Unsupported shell

- **WHEN** `$SHELL = /bin/sh`
- **THEN** system SHALL return an error with `SHELL_DETECTION_FAILED`
- **AND** message SHALL include the detected shell path
- **AND** message SHALL list supported shells

#### Scenario: Unset `$SHELL`

- **WHEN** `$SHELL` is empty or unset
- **THEN** system SHALL return an error indicating `SHELL` is not set

### Requirement: `InferShellTypeFromPath`

`InferShellTypeFromPath(path)` SHALL map the config file basename to
a shell type:

| Basename | Shell |
|---|---|
| `.bashrc`, `.bash_profile`, `.profile`, `bash_profile`, `bashrc` | `bash` |
| `.zshrc`, `.zprofile`, `zshrc` | `zsh` |
| `.config/fish/config.fish`, `config.fish`, `.fishrc` | `fish` |

#### Scenario: Infer from bash config

- **WHEN** `path = ~/.bashrc`
- **THEN** system SHALL return `ShellType("bash")`

#### Scenario: Unrecognized path

- **WHEN** `path = ~/.config/starship.toml`
- **THEN** system SHALL return `ShellType("")`
- **AND** caller SHALL fall back to env detection

### Requirement: Config-file preference order

For a given shell, the system SHALL probe the following config files
in order, using the first one that exists:

| Shell | Files (preference order) |
|---|---|
| bash | `.bashrc`, `.bash_profile`, `.profile` |
| zsh | `.zshrc`, `.zprofile`, `.profile` |
| fish | `.config/fish/config.fish`, `config.fish`, `.fishrc` |

#### Scenario: First existing config wins

- **WHEN** `.bashrc` exists in the user's home
- **THEN** system SHALL return `~/.bashrc` for bash
- **AND** SHALL NOT probe `.bash_profile`

### Requirement: Compose wrapper with placeholders

`ShellInfrastructure.ComposeWrapper(template, shellType)` SHALL
replace `{{SHELL_TYPE}}` and `{{TIMESTAMP}}` placeholders in the
template. Replacements SHALL be deterministic and side-effect free;
the `{{TIMESTAMP}}` SHALL be formatted as `2006-01-02 15:04:05` at
the time of `ComposeWrapper` invocation.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
