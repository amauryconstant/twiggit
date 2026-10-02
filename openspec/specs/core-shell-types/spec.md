# Capability: Shell Types

## Purpose

Defines the request and result value objects for shell-wrapper composition, installation, and validation. The cmd layer (`cli-init`) constructs these and passes them to `output.ComposeWrapper`, `output.InstallWrapper`, and `output.ValidateInstallation`. Types live in `internal/core/shell_requests.go` and `internal/core/shell_results.go`.

## Requirements

### Requirement: GenerateWrapperRequest and Result

`core.GenerateWrapperRequest` SHALL carry `ShellType core.ShellType` (one of `bash`, `zsh`, `fish`), `Timestamp time.Time` (formatted as `2006-01-02 15:04:05` at invocation time), and `IncludeCompletion bool` (default `true` — embed Carapace completion). `core.GenerateWrapperResult` SHALL carry `Wrapper string` (the rendered template with `{{SHELL_TYPE}}` and `{{TIMESTAMP}}` placeholders resolved), `Completion string` (the Carapace completion block), and `Combined string` (wrapper + completion ready for `eval`).

#### Scenario: Generate bash wrapper

- **WHEN** `GenerateWrapperRequest{ShellType: core.ShellType("bash"), IncludeCompletion: true}` is dispatched
- **THEN** `result.Wrapper` SHALL contain the bash wrapper block (`### BEGIN/END TWIGGIT WRAPPER`)
- **AND** `result.Combined` SHALL be `result.Wrapper + "\n" + result.Completion`

### Requirement: SetupShellRequest and Result

`core.SetupShellRequest` SHALL carry `ShellType core.ShellType`, `ConfigFile string` (absolute path to the user's shell config; empty means auto-detect), and `Force bool` (rewrite existing blocks). `core.SetupShellResult` SHALL carry `ConfigFile string`, `AlreadyInstalled bool`, `Installed bool`, and `OldBlocksRemoved int`.

#### Scenario: Already installed

- **WHEN** `SetupShellRequest{ShellType: "bash", ConfigFile: "", Force: false}` runs against a file containing the wrapper block
- **THEN** `result.AlreadyInstalled == true`, `Installed == false`, `OldBlocksRemoved == 0`

### Requirement: ValidateInstallationRequest and Result

`core.ValidateInstallationRequest` SHALL carry `ShellType core.ShellType` and `ConfigFile string` (empty means auto-detect). `core.ValidateInstallationResult` SHALL carry `Installed bool`, `HasWrapperBlock bool`, `HasCompletionBlock bool`, `Missing []string` (the list of missing block names), and `Corruption []string` (the list of orphaned or partial block fragments found).

#### Scenario: Orphaned END delimiter detected

- **WHEN** the config file contains a `### END TWIGGIT WRAPPER` line with no matching `### BEGIN TWIGGIT WRAPPER`
- **THEN** `ValidateInstallationResult.Corruption` SHALL contain one entry naming the orphaned END delimiter

### Requirement: RequestWithShellType constructors

`core.NewGenerateWrapperRequest(shellType)`, `core.NewSetupShellRequest(shellType, configFile, force)`, and `core.NewValidateInstallationRequest(shellType, configFile)` SHALL be the canonical constructors. Each constructor SHALL validate the `ShellType` via `core.IsValidShellType` and SHALL return a `*core.ValidationError` on invalid input per the `err-prefix-suffix` rule.

#### Scenario: Invalid shell type rejected at construction

- **WHEN** `core.NewGenerateWrapperRequest(core.ShellType("powershell"))` is invoked
- **THEN** the function SHALL return a `*core.ValidationError` with `Field == "ShellType"`