# Spec Delta

## ADDED Requirements

### Requirement: Shell-detection errors surface via sentinel

`GenerateWrapper`, `ValidateInstallation`, and any other
`ShellService` method that internally calls
`ShellInfrastructure.DetectConfigFile` SHALL return an error
satisfying `errors.Is(err, domain.ErrShellInferenceFailed)` when
the shell type cannot be detected. The cmd layer's typed-error
dispatch SHALL classify the error as a configuration failure
without falling back to string matching.

#### Scenario: Undetectable shell type returns typed error
- **WHEN** `GenerateWrapper` is called with a shell type that
  cannot be resolved from `$SHELL` or the user's environment
- **THEN** the returned error satisfies
  `errors.Is(err, domain.ErrShellInferenceFailed)` and the
  cmd layer formats it with the shell-detection hint
