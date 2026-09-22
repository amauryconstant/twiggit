# Spec Delta

## MODIFIED Requirements

### Requirement: Shell detect lives in internal/core

The shell-detect logic (`DetectShellFromEnv`, `InferShellTypeFromPath`, shell-config-file probing) SHALL live in `internal/core/shell_detect.go`. It is pure logic with no I/O dependency beyond `os.Getenv` and `os.Stat`. The `core.ShellType` type SHALL replace the legacy `domain.ShellType`.

#### Scenario: Shell-detect file path is internal/core
- **WHEN** the shell-detect source file is located
- **THEN** it SHALL be `internal/core/shell_detect.go`

#### Scenario: ShellType is core.ShellType
- **WHEN** the package user imports the shell type
- **THEN** the import path SHALL be `twiggit/internal/core` and the type SHALL be `core.ShellType`

### Requirement: ComposeWrapper lives in internal/output

`ShellInfrastructure.ComposeWrapper(template, shellType)` SHALL live in `internal/output/wrapper.go` (rendering concern). It SHALL replace `{{SHELL_TYPE}}` and `{{TIMESTAMP}}` placeholders in the template. The `{{TIMESTAMP}}` SHALL be formatted as `2006-01-02 15:04:05` at the time of invocation.

#### Scenario: ComposeWrapper file path is internal/output
- **WHEN** the wrapper source file is located
- **THEN** it SHALL be `internal/output/wrapper.go`

#### Scenario: ShellType parameter is core.ShellType
- **WHEN** `ComposeWrapper` is called
- **THEN** the `shellType` parameter SHALL be of type `core.ShellType`
