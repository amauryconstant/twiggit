# Spec Delta: Domain Context Types

## MODIFIED Requirements

### Requirement: PathType enum values

The system SHALL define `PathType` with constants
`PathTypeUnknown`, `PathTypeProject`, `PathTypeWorktree`,
`PathTypeInvalid`. The `iota` zero value SHALL be `PathTypeUnknown`
so that an uninitialized `PathType` variable is detectable as
invalid. `String()` SHALL return `"unknown"`, `"project"`,
`"worktree"`, `"invalid"` respectively.

#### Scenario: Zero-value PathType is Unknown

- **WHEN** a `PathType` variable is declared without explicit
  initialization
- **THEN** `var p PathType` SHALL equal `PathTypeUnknown`
- **AND** `p.String()` SHALL return `"unknown"`

#### Scenario: Existing constants retain their strings

- **WHEN** `PathTypeProject`, `PathTypeWorktree`, `PathTypeInvalid`
  are rendered via `String()`
- **THEN** the result SHALL equal `"project"`, `"worktree"`,
  `"invalid"` respectively

#### Scenario: Numeric shift documented for callers

- **WHEN** downstream code compares `PathType` against an integer
  literal
- **THEN** the comparison SHALL use the named constant
  (`domain.PathTypeProject`, etc.) rather than the underlying integer
  value, because the integer values shift with the addition of
  `PathTypeUnknown = 0`
