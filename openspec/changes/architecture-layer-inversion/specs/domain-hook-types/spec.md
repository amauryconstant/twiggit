# Spec Delta: Domain Hook Types

## MODIFIED Requirements

### Requirement: HookResult field names

The system SHALL provide a `HookResult` struct with the following
shape:

| Field | Type | Purpose |
|---|---|---|
| `HookType` | `HookType` | Type of hook that ran |
| `HasExecuted` | `bool` | Were any commands configured for this hook? |
| `IsSuccessful` | `bool` | Did every executed command succeed? |
| `Failures` | `[]HookFailure` | Empty when `IsSuccessful` is true |

The field names `HasExecuted` and `IsSuccessful` SHALL be used in
place of the prior `Executed` and `Success` so the boolean fields
read as predicates (`Is`-/`Has`-prefixed booleans per project
naming rules).

#### Scenario: HasExecuted replaces Executed

- **WHEN** a caller inspects whether hooks ran for a worktree
- **THEN** it SHALL read `result.HasExecuted`
- **AND** the prior `result.Executed` field SHALL NOT exist

#### Scenario: IsSuccessful replaces Success

- **WHEN** a caller inspects whether all commands succeeded
- **THEN** it SHALL read `result.IsSuccessful`
- **AND** the prior `result.Success` field SHALL NOT exist

#### Scenario: Empty Failures signals success

- **WHEN** `result.IsSuccessful` is true
- **THEN** `result.Failures` SHALL be empty (length 0)
