## ADDED Requirements

### Requirement: Error formatting uses explicit strategy pattern

The error formatter SHALL use an explicit strategy pattern with `errors.As()` matching for type detection, not reflection.

#### Scenario: Formatter registers matcher-formatter pair
- **WHEN** a matcher function and formatter function are registered
- **THEN** the formatter stores the pair for later matching

#### Scenario: Formatter iterates matchers in registration order
- **WHEN** formatting an error
- **THEN** the formatter iterates through registered matchers in order
- **AND** uses the first matching formatter

### Requirement: Matcher functions use errors.As()

Matcher functions SHALL use `errors.As()` for type detection.

#### Scenario: Validation error matcher
- **WHEN** checking if an error is a ValidationError
- **THEN** `errors.As(err, &ValidationError{})` returns true

#### Scenario: Worktree error matcher
- **WHEN** checking if an error is a WorktreeServiceError
- **THEN** `errors.As(err, &WorktreeServiceError{})` returns true

#### Scenario: Project error matcher
- **WHEN** checking if an error is a ProjectServiceError
- **THEN** `errors.As(err, &ProjectServiceError{})` returns true

#### Scenario: Service error matcher
- **WHEN** checking if an error is a ServiceError
- **THEN** `errors.As(err, &ServiceError{})` returns true

### Requirement: Formatter SHALL NOT use reflection

The error formatter SHALL NOT use the `reflect` package for type detection.

#### Scenario: No reflect usage in Format method
- **WHEN** formatting any error
- **THEN** no reflection operations are performed

### Requirement: Formatter functions accept error directly

Formatter functions SHALL accept an `error` parameter directly, not via receiver.

#### Scenario: Formatter function signature
- **WHEN** defining a formatter function
- **THEN** the signature is `func(error) string`

## UNCHANGED Requirements

### Requirement: Quiet mode suppresses hints

The formatter SHALL suppress hint output when quiet mode is enabled.

#### Scenario: Quiet mode enabled
- **WHEN** formatting an error with quiet mode true
- **THEN** the hint portion is omitted from output

#### Scenario: Quiet mode disabled
- **WHEN** formatting an error with quiet mode false
- **THEN** the hint portion is included in output
