# Capability: Init Command

## Purpose

Generate or install the shell wrapper that intercepts `twiggit cd` to
enable navigation between worktrees and projects. Two modes:
**stdout** (default, eval-safe) for `eval "$(twiggit init)"` activation,
and **install** (`--install`) for persistent writes to the user's shell
config file. Service-layer generation and validation contracts are owned
by `application-shell-service`.

## Requirements

### Requirement: Positional shell argument

The system SHALL accept the shell type as a positional argument
(`bash`, `zsh`, `fish`) and SHALL auto-detect from the `$SHELL`
environment variable when the positional is omitted. The previous
`--shell=<type>` flag form is no longer supported.

#### Scenario: Explicit positional shell

- **WHEN** user runs `twiggit init bash` (or `zsh`, `fish`)
- **THEN** system SHALL use the explicitly specified shell type
- **AND** no `--shell` flag SHALL be required

#### Scenario: Auto-detect from `$SHELL`

- **WHEN** user runs `twiggit init` with no positional argument
- **AND** `$SHELL` is set to a supported shell path
- **THEN** system SHALL detect the shell type from `$SHELL`

#### Scenario: Detection failure

- **WHEN** user runs `twiggit init` with no positional argument
- **AND** `$SHELL` is unset or points to an unsupported shell
- **THEN** system SHALL return a detection error
- **AND** error message SHALL list the supported shells (`bash`, `zsh`, `fish`)
- **AND** error message SHALL suggest specifying the shell positionally

#### Scenario: Unsupported shell type

- **WHEN** user runs `twiggit init powershell` or any unsupported type
- **THEN** system SHALL return a `ValidationError` on field `shellType`
- **AND** message SHALL list supported shells

### Requirement: Stdout mode (default)

The system SHALL default to writing the wrapper to stdout with no file
modification, eval-safe and containing both the wrapper block and the
Carapace completion block.

#### Scenario: Eval activation

- **WHEN** user runs `eval "$(twiggit init)"`
- **THEN** `twiggit` shell function SHALL be defined in the current shell
- **AND** Carapace completion for the detected shell SHALL be sourced
- **AND** `twiggit cd <branch>` SHALL change directory
- **AND** `builtin cd <path>` SHALL fall back to the shell built-in

#### Scenario: Explicit shell, stdout mode

- **WHEN** user runs `twiggit init zsh`
- **THEN** system SHALL output the zsh wrapper to stdout
- **AND** output SHALL include the wrapper block (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** output SHALL include the completion block (`### BEGIN/END TWIGGIT COMPLETION`)
- **AND** system SHALL NOT modify any files

#### Scenario: Stdout output is eval-safe

- **WHEN** system writes the wrapper to stdout
- **THEN** output SHALL contain only the wrapper and completion blocks
- **AND** SHALL NOT include any human-readable metadata lines

### Requirement: Install mode

The system SHALL write the wrapper to a shell config file when
`--install` (`-i`) is provided, optionally targeting a custom file
via `--config` (`-c`). `--config` and `--force` SHALL be rejected
unless `--install` is also present.

#### Scenario: Install to auto-detected config file

- **WHEN** user runs `twiggit init --install` (or `twiggit init bash --install`)
- **AND** `--config` is NOT provided
- **AND** the wrapper is not already installed
- **THEN** system SHALL auto-detect the config file for the shell
- **AND** system SHALL append the wrapper block and completion block
- **AND** success message SHALL include the config file path

#### Scenario: Install to explicit config file

- **WHEN** user runs `twiggit init bash --install --config ~/.customrc`
- **AND** the parent directory of `~/.customrc` exists
- **THEN** system SHALL write to `~/.customrc`
- **AND** success message SHALL include `~/.customrc` as the config file

#### Scenario: `--config` without `--install` rejected

- **WHEN** user runs `twiggit init bash --config ~/.bashrc` without `--install`
- **THEN** system SHALL return validation error: "--config requires --install"

#### Scenario: `--force` without `--install` rejected

- **WHEN** user runs `twiggit init bash --force` without `--install`
- **THEN** system SHALL return validation error: "--force requires --install"

#### Scenario: Missing parent directory

- **WHEN** user runs `twiggit init bash --install --config /nonexistent/path/cfg`
- **AND** `/nonexistent/path/` does not exist
- **THEN** system SHALL return an error naming the missing directory
- **AND** SHALL NOT write any file

### Requirement: Reinstall and force

The system SHALL detect existing wrapper and completion blocks via their
`### BEGIN/END TWIGGIT WRAPPER` and `### BEGIN/END TWIGGIT COMPLETION`
delimiters. Without `--force` the system SHALL refuse to reinstall; with
`--force` it SHALL remove the old blocks and rewrite fresh ones, tolerating
partial or orphaned delimiters.

#### Scenario: Already installed, no force

- **WHEN** user runs `twiggit init bash --install`
- **AND** the wrapper block is already present
- **AND** `--force` is NOT set
- **THEN** system SHALL skip installation
- **AND** SHALL emit "Shell wrapper already installed for bash"
- **AND** SHALL emit "Use --force to reinstall"

#### Scenario: Force reinstall clean

- **WHEN** user runs `twiggit init bash --install --force`
- **AND** the config file contains a complete WRAPPER and COMPLETION block
- **THEN** system SHALL remove both blocks (including delimiters)
- **AND** preserve all other content
- **AND** append fresh wrapper and completion blocks

#### Scenario: Force reinstall with partial delimiters

- **WHEN** user runs `twiggit init bash --install --force`
- **AND** the config file contains only a `BEGIN` delimiter (no matching `END`)
- **THEN** system SHALL treat it as an incomplete block
- **AND** SHALL remove the partial block
- **AND** SHALL append complete blocks
- **AND** SHALL emit a warning noting incomplete blocks were removed

#### Scenario: Force reinstall with orphaned END delimiter

- **WHEN** user runs `twiggit init bash --install --force`
- **AND** the config file contains a stray `END` delimiter with no matching `BEGIN`
- **THEN** system SHALL remove the orphaned delimiter
- **AND** SHALL append complete blocks
- **AND** SHALL emit a warning noting orphaned delimiters were removed

### Requirement: Positional completion

The system SHALL provide Carapace-driven completion for the positional
`[shell]` argument offering `bash`, `zsh`, `fish`.

#### Scenario: Tab completion of shell

- **WHEN** user types `twiggit init <TAB>`
- **THEN** the completion menu SHALL offer `bash`, `zsh`, `fish`

### Requirement: Activation instructions

The system SHALL print, after successful installation, the activate-now
and activate-later instructions referencing the actual config file
written.

#### Scenario: Print activate instructions

- **WHEN** install completes and the config file exists
- **THEN** system SHALL print the steps: "1. Restart your shell, or 2. Run:
  source <config-file>"
- **AND** SHALL print usage examples: `twiggit cd <branch>` and
  `builtin cd <path>`

### Requirement: Wrapper runtime behavior

The installed wrapper SHALL intercept `twiggit cd`, navigate to the
absolute path printed by the command, pass `builtin cd` through to the
shell, and pass through all other commands. The wrapper SHALL NOT
modify any other commands.

#### Scenario: Intercept `twiggit cd`

- **WHEN** shell wrapper is installed
- **AND** user runs `twiggit cd <target>`
- **THEN** wrapper SHALL execute `twiggit cd` to resolve the target
- **AND** SHALL change the shell's working directory to the resolved
  absolute path
- **AND** cd SHALL be silent (no extra wrapper output)

#### Scenario: `builtin cd` escape hatch

- **WHEN** user runs `builtin cd <path>`
- **THEN** wrapper SHALL pass through to the shell built-in
- **AND** SHALL NOT intercept the call

#### Scenario: Non-cd commands

- **WHEN** user runs any command other than `twiggit cd`
- **THEN** wrapper SHALL pass through without modification

### Requirement: Service contract

The cmd layer SHALL delegate wrapper generation, installation, and
validation to `application-shell-service`. The service contract is owned
by `application-shell-service`. This spec is silent on service-layer
internals.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
