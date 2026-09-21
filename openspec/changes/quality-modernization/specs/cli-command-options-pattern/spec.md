# Spec Delta

## ADDED Requirements

### Requirement: Every leaf command sets SilenceUsage and SilenceErrors

Every leaf cobra command SHALL set `SilenceUsage = true` and
`SilenceErrors = true` so that a `RunE` failure prints only the
error message and exit code, not the full command usage block.
Persistent flags SHALL NOT be affected; this requirement applies
to leaf commands only.

#### Scenario: create fails and prints error without usage
- **WHEN** `twiggit create` fails (e.g., the requested branch
  name is invalid) and `RunE` returns the validation error
- **THEN** the user sees the error message and exit code 5
  (validation), with no usage block printed

#### Scenario: delete fails silently usage but still prints error
- **WHEN** `twiggit delete feature/x` fails because the
  worktree has uncommitted changes
- **THEN** the user sees the error message and exit code 5;
  the usage block for `delete` is not printed

#### Scenario: help still works despite SilenceUsage
- **WHEN** the user runs `twiggit delete --help` or
  `twiggit help delete`
- **THEN** the help text is printed (help is unaffected by
  SilenceUsage)
