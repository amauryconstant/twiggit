# Capability: Git Shell Detect

## Purpose

Documents the Tier 2 split of shell-related logic. Pure derivation helpers (`DetectShellFromEnv`, `InferShellTypeFromPath`, `core.ShellType`, `core.IsValidShellType`) live in `internal/core/shell.go`; filesystem probing of config files lives in `internal/git/shell_detect.go` (I/O adapter); the `ComposeWrapper` renderer lives in `internal/output/wrapper.go`.

## Requirements

### Requirement: Pure shell derivation lives in internal/core

`core.ShellType` SHALL be a string type defined in `internal/core/shell.go`. The system SHALL recognize exactly three values: `bash`, `zsh`, `fish`. The type SHALL declare an explicit `Unknown` sentinel at `iota` position 0 per the `enum-unknown-zero` rule. Other non-Unknown values SHALL be rejected by `core.IsValidShellType`. The derivation helpers `core.DetectShellFromEnv()` (reads `$SHELL`) and `core.InferShellTypeFromPath(path)` (maps config-file basename to shell type) SHALL live in `internal/core/shell.go` and SHALL NOT touch the filesystem.

#### Scenario: Shell-detect file path is internal/core

- **WHEN** the pure-derivation shell-detect source file is located
- **THEN** it SHALL be `internal/core/shell.go`

#### Scenario: ShellType is core.ShellType

- **WHEN** the package user imports the shell type
- **THEN** the import path SHALL be `twiggit/internal/core` and the type SHALL be `core.ShellType`

#### Scenario: core does not import os for Stat

- **WHEN** the `internal/core/shell_detect.go` import list is read
- **THEN** it SHALL NOT contain `"os"` for `os.Stat` calls (filesystem I/O lives in the adapter package)

### Requirement: Filesystem probing lives in internal/git

Config-file probing (`os.Stat`) SHALL live in `internal/git/shell_detect.go`. For a given shell, the system SHALL probe the following config files in order, using the first one that exists:

| Shell | Files (preference order) |
|---|---|
| bash | `.bashrc`, `.bash_profile`, `.profile` |
| zsh | `.zshrc`, `.zprofile`, `.profile` |
| fish | `.config/fish/config.fish`, `config.fish`, `.fishrc` |

Stat failures SHALL be wrapped as `*core.OperationError` with `Op = "shell.probe"`.

#### Scenario: Probe returns the first existing file

- **WHEN** `shell = "bash"` and `$HOME/.bashrc` exists
- **THEN** the probe returns `$HOME/.bashrc` without checking subsequent files

#### Scenario: Probe failure wraps as *core.OperationError

- **WHEN** every file in the preference list errors on stat
- **THEN** `errors.AsType[*core.OperationError](err)` SHALL return a non-nil result with `Op == "shell.probe"`

### Requirement: ComposeWrapper lives in internal/output

`output.ComposeWrapper(template string, shellType core.ShellType) (string, error)` SHALL live in `internal/output/wrapper.go` (rendering concern). It SHALL replace `{{SHELL_TYPE}}` and `{{TIMESTAMP}}` placeholders in the template. The `{{TIMESTAMP}}` SHALL be formatted as `2006-01-02 15:04:05` at the time of invocation.

#### Scenario: ComposeWrapper file path is internal/output

- **WHEN** the wrapper source file is located
- **THEN** it SHALL be `internal/output/wrapper.go`

#### Scenario: ShellType parameter is core.ShellType

- **WHEN** `ComposeWrapper` is called
- **THEN** the `shellType` parameter SHALL be of type `core.ShellType`

#### Scenario: Timestamp placeholder resolves at invocation time

- **WHEN** `ComposeWrapper("at {{TIMESTAMP}}", "bash")` runs at `2026-09-22 14:30:00`
- **THEN** the result contains `"at 2026-09-22 14:30:00"`
