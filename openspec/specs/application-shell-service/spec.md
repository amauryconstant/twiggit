# Capability: Shell Service

## Purpose

Service-layer orchestration for shell wrapper generation, installation,
and validation. The cmd layer (`cli-init`) delegates to this service.

## Requirements

### Requirement: Service contract surface

The `ShellService` interface SHALL expose:

- `SetupShell(ctx, *SetupShellRequest) (*SetupShellResult, error)`
- `ValidateInstallation(ctx, *ValidateInstallationRequest) (*ValidateInstallationResult, error)`
- `GenerateWrapper(ctx, *GenerateWrapperRequest) (*GenerateWrapperResult, error)`

Implementation lives in `internal/service/`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Generate wrapper content

`GenerateWrapper` SHALL produce the eval-safe shell wrapper content
for the requested shell type (without writing any file).

#### Scenario: Generate bash wrapper

- **WHEN** `GenerateWrapper(ctx, &GenerateWrapperRequest{ShellType: "bash"})` is called
- **THEN** system SHALL return a string containing the wrapper block
  with `### BEGIN/END TWIGGIT WRAPPER` delimiters
- **AND** the Carapace completion sourcing block with
  `### BEGIN/END TWIGGIT COMPLETION` delimiters

#### Scenario: Generate zsh wrapper

- **WHEN** `GenerateWrapper(ctx, &GenerateWrapperRequest{ShellType: "zsh"})` is called
- **THEN** system SHALL return a zsh-flavoured wrapper

#### Scenario: Generate fish wrapper

- **WHEN** `GenerateWrapper(ctx, &GenerateWrapperRequest{ShellType: "fish"})` is called
- **THEN** system SHALL return a fish-flavoured wrapper

#### Scenario: Unsupported shell type

- **WHEN** an unsupported shell type is requested
- **THEN** system SHALL return a `domain.ValidationError` on `shellType`
- **AND** message SHALL list supported shells (`bash`, `zsh`, `fish`)

### Requirement: Setup shell (install)

`SetupShell` SHALL write the wrapper to the shell config file, detect
the appropriate config file if none is given, and respect
`ForceOverwrite`. The infrastructure layer (`ShellInfrastructure`)
owns the actual file write; this service orchestrates the full flow.

#### Scenario: Fresh install

- **WHEN** `SetupShell` is called and the wrapper is not yet installed
- **THEN** system SHALL detect the config file via
  `ShellInfrastructure.DetectConfigFile`
- **AND** SHALL generate and install the wrapper
- **AND** SHALL return `SetupShellResult.Installed = true`

#### Scenario: Already installed

- **WHEN** `SetupShell` is called and the wrapper is already installed
- **AND** `ForceOverwrite = false`
- **THEN** system SHALL skip installation
- **AND** SHALL return `SetupShellResult.Skipped = true`

#### Scenario: Force reinstall

- **WHEN** `SetupShell` is called with `ForceOverwrite = true`
- **THEN** system SHALL remove the existing wrapper and completion blocks
- **AND** SHALL install fresh blocks

### Requirement: Validate installation

`ValidateInstallation` SHALL check whether the wrapper is installed in
the user's shell config file, returning an actionable result.

#### Scenario: Installed

- **WHEN** the wrapper block is present in the config file
- **THEN** `ValidateInstallationResult.Installed` SHALL be `true`
- **AND** `ValidateInstallationResult.ConfigFile` SHALL be set

#### Scenario: Not installed

- **WHEN** the wrapper block is missing
- **THEN** `ValidateInstallationResult.Installed` SHALL be `false`

### Requirement: Auto-detect config file

The service SHALL defer config-file detection to
`ShellInfrastructure.DetectConfigFile`. See
`infrastructure-shell-detect` for the precedence rules.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Service construction

The `ShellService` constructor SHALL accept `ShellInfrastructure`,
`*domain.Config`, and any other infrastructure dependencies via
constructor injection. No globals.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
