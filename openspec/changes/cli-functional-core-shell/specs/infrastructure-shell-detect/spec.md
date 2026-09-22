# Spec Delta

## MODIFIED Requirements

### Requirement: `ShellType` and supported set

`core.ShellType` SHALL be a string type. The system SHALL recognize exactly three values: `bash`, `zsh`, `fish`. Other values SHALL be rejected by `core.IsValidShellType`. The previous `domain.ShellType` and `domain.IsValidShellType` are removed; the package qualifier is `core` for every call site. `core.ShellType` lives in `internal/core/shell_detect.go` (pure derivation; no `os` import for `os.Stat`).

#### Scenario: IsValidShellType accepts the three values
- **WHEN** the value is `"bash"`, `"zsh"`, or `"fish"`
- **THEN** `core.IsValidShellType(s)` returns `true`

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `DetectShellFromEnv`

`core.DetectShellFromEnv()` SHALL read `$SHELL` and return the matching `core.ShellType` by substring match: a path containing `bash` maps to `bash`, `zsh` to `zsh`, `fish` to `fish`. If `$SHELL` is unset, the function SHALL return a `*core.OperationError` with `Op = "shell.detect"`. The previous `SHELL_DETECTION_FAILED` error code is removed in favor of the `errors.As` walk to `*core.OperationError`.

#### Scenario: Detect bash
- **WHEN** `$SHELL = /bin/bash`
- **THEN** system SHALL return `core.ShellType("bash")`

#### Scenario: Detect zsh
- **WHEN** `$SHELL = /usr/local/bin/zsh`
- **THEN** system SHALL return `core.ShellType("zsh")`

#### Scenario: Detect fish
- **WHEN** `$SHELL = /usr/local/bin/fish`
- **THEN** system SHALL return `core.ShellType("fish")`

#### Scenario: Unsupported shell
- **WHEN** `$SHELL = /bin/sh`
- **THEN** system SHALL return an error implementing `errors.As(err, &*core.OperationError{})` with `Op = "shell.detect"`
- **AND** message SHALL include the detected shell path
- **AND** message SHALL list supported shells

#### Scenario: Unset `$SHELL`
- **WHEN** `$SHELL` is empty or unset
- **THEN** system SHALL return a `*core.OperationError` indicating `SHELL` is not set

### Requirement: `InferShellTypeFromPath`

`core.InferShellTypeFromPath(path)` SHALL map the config file basename to a shell type:

| Basename | Shell |
|---|---|
| `.bashrc`, `.bash_profile`, `.profile`, `bash_profile`, `bashrc` | `bash` |
| `.zshrc`, `.zprofile`, `zshrc` | `zsh` |
| `.config/fish/config.fish`, `config.fish`, `.fishrc` | `fish` |

The function lives in `internal/core/shell_detect.go` (pure derivation; no filesystem access).

#### Scenario: Infer from bash config
- **WHEN** `path = ~/.bashrc`
- **THEN** system SHALL return `core.ShellType("bash")`

#### Scenario: Unrecognized path
- **WHEN** `path = ~/.config/starship.toml`
- **THEN** system SHALL return `core.ShellType("")`
- **AND** caller SHALL fall back to env detection

### Requirement: Config-file preference order

For a given shell, the system SHALL probe the following config files in order, using the first one that exists:

| Shell | Files (preference order) |
|---|---|
| bash | `.bashrc`, `.bash_profile`, `.profile` |
| zsh | `.zshrc`, `.zprofile`, `.profile` |
| fish | `.config/fish/config.fish`, `config.fish`, `.fishrc` |

The probing logic lives in `internal/git/shell_detect.go` (an I/O adapter) because it uses `os.Stat`. Stat failures SHALL be wrapped as `*core.OperationError` with `Op = "shell.probe"`.

#### Scenario: First existing config wins
- **WHEN** `.bashrc` exists in the user's home
- **THEN** system SHALL return `~/.bashrc` for bash
- **AND** SHALL NOT probe `.bash_profile`

#### Scenario: Probe failure wraps as *core.OperationError
- **WHEN** every file in the preference list errors on stat
- **THEN** the returned error SHALL implement `errors.As(err, &*core.OperationError{})` with `Op = "shell.probe"`

### Requirement: Compose wrapper with placeholders

`output.ComposeWrapper(template string, shellType core.ShellType) (string, error)` SHALL live in `internal/output/wrapper.go` (rendering concern). It SHALL replace `{{SHELL_TYPE}}` and `{{TIMESTAMP}}` placeholders in the template. Replacements SHALL be deterministic and side-effect free; the `{{TIMESTAMP}}` SHALL be formatted as `2006-01-02 15:04:05` at the time of `ComposeWrapper` invocation. The previous `ShellInfrastructure.ComposeWrapper` is removed.

#### Scenario: ComposeWrapper file path is internal/output
- **WHEN** the wrapper source file is located
- **THEN** it SHALL be `internal/output/wrapper.go`

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
