# Spec Delta: Domain Typed Errors

## MODIFIED Requirements

### Requirement: ValidationError builder and getter names

`ValidationError` SHALL support immutable builder methods:

- `WithSuggestions([]string) *ValidationError`
- `WithContext(string) *ValidationError`

Getters SHALL be: `Field()`, `Value()`, `Message()`, `Request()`,
`Suggestions()`, and `Detail()` (the prior `Context()` getter is
renamed to `Detail()` to avoid collision with the `domain.Context`
type at call sites that pass both).

#### Scenario: Detail() returns the contextual explanation

- **WHEN** a caller invokes `e.Detail()` on a `*ValidationError`
- **THEN** it SHALL return the same string that the prior
  `e.Context()` getter returned
- **AND** it SHALL NOT collide with the `domain.Context` type when
  the caller writes `e.Detail()` next to a `*domain.Context`
  argument in the same scope

#### Scenario: Context() getter no longer exists

- **WHEN** the domain package is compiled
- **THEN** `(*ValidationError).Context()` SHALL NOT be defined
- **AND** any caller that referenced `e.Context()` SHALL fail to
  compile until migrated to `e.Detail()`
