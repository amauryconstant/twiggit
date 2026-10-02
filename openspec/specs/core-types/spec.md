# Capability: Types

## Purpose

Defines the value-object pattern and validation pipeline for the `internal/core/` package, making the rule that every domain type is always-valid post-construction. The pipeline's `ValidateAll` uses `errors.Join` to preserve the full error chain; `output.FormatError` walks each joined error independently.

## Requirements

### Requirement: Value objects expose always-valid construction

Every value object (`core.NewBranchName`, `core.NewProjectName`, `core.NewWorktreePath`) SHALL return `(T, error)` from its constructor. The returned value SHALL be safe to use without further validation; if the input fails the rules, the constructor SHALL return a `*core.ValidationError`. The validation pipeline exposes `core.ValidateBranchName(name) Result[bool]`, `core.ValidateProjectName(name) Result[bool]`, and `core.ValidateShellType(s) Result[bool]` for callers that prefer pipeline composition over constructor invocation.

#### Scenario: Valid branch name constructs successfully

- **WHEN** the caller invokes `r := core.ValidateBranchName("feat/foo")`
- **THEN** `r.IsSuccess()` returns `true` and `r.Value()` is `true`

#### Scenario: Invalid branch name returns ValidationError

- **WHEN** the caller invokes `r := core.ValidateBranchName("feat..bad")`
- **THEN** `r.IsError()` returns `true` and the wrapped error is a `*core.ValidationError` carrying `Field="BranchName"`, `Value="feat..bad"`, and a `Message` naming the rule violated

### Requirement: Pipeline[T] composes Validator[T] with Validate (fail-fast) or ValidateAll (collect-all)

The `core.Pipeline[T]` type SHALL compose zero or more `core.Validator[T]` functions. `Pipeline.Validate(t)` SHALL run each validator in order and return the first error (fail-fast); `Pipeline.ValidateAll(t)` SHALL run all validators and MUST use `errors.Join(...)` to compose every failure into one error that preserves the full chain. `FormatError` SHALL walk the joined chain via `errors.As` and render each failure as a separate validation hint. `ValidateAll` SHALL NOT aggregate failures into a single `*core.ValidationError` with a `Suggestions` slice; each failure SHALL remain an independent error in the join so `errors.Is` and `errors.As` can inspect them individually.

#### Scenario: Validate stops at first failure

- **WHEN** a pipeline of two validators runs `pipeline.Validate(input)` and the first fails
- **THEN** the second validator SHALL NOT run and the returned error SHALL be the first validator's error

#### Scenario: ValidateAll collects every failure via errors.Join

- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and validators 1 and 3 fail
- **THEN** `errors.Is(err, validator1Err)` returns `true` and `errors.Is(err, validator3Err)` returns `true` (both preserved in the chain)
- **AND** `errors.Is(err, validator2Err)` returns `false` (the passing validator is not in the chain)

#### Scenario: ValidateAll wraps each failure independently

- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and validators 1 and 3 fail
- **THEN** the returned error SHALL be the `errors.Join` composition and SHALL NOT be a single aggregated `*core.ValidationError`
- **AND** `errors.AsType[*core.ValidationError](err)` SHALL return a non-nil result twice (once per independent failure), confirming each failure is its own error in the chain

#### Scenario: ValidateAll with no failures returns nil

- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and all validators pass
- **THEN** the returned error SHALL be `nil` (an empty `errors.Join` result is normalized to nil)

#### Scenario: ValidateAll with one failure preserves errors.Is

- **WHEN** a pipeline of one validator runs `pipeline.ValidateAll(input)` and the validator returns `validatorErr`
- **THEN** `errors.Is(err, validatorErr)` returns `true` via the join chain

### Requirement: Core types are immutable post-construction

Once constructed, a value object SHALL expose read-only fields (or accessor methods). Mutation methods SHALL NOT exist on core types.

#### Scenario: No mutation methods on BranchName

- **WHEN** the package user searches `core.BranchName` for setter or mutation methods
- **THEN** no method SHALL match `Set`, `Mutate`, `Assign`, `Replace`, or `Update` patterns

### Requirement: Result[T] wraps a value or an error

The `core.Result[T]` type SHALL carry either a `Value T` and a `Success bool`, or a non-nil `Err error`. Constructors `core.NewResult(value T)` and `core.NewErrResult(err error)` SHALL produce success and failure variants respectively. `Result.IsError()` and `Result.IsSuccess()` SHALL dispatch on the `Success` flag without unwrapping. `Result.Value()` SHALL return the stored value when `IsSuccess()` is true; `Result.Err()` SHALL return the stored error when `IsError()` is true.

#### Scenario: Success result exposes Value

- **WHEN** the caller invokes `r := core.NewResult("feat/foo")`
- **THEN** `r.IsSuccess()` returns `true` and `r.Value()` returns `"feat/foo"`

#### Scenario: Error result exposes Err

- **WHEN** the caller invokes `r := core.NewErrResult(core.NewValidationError(...))`
- **THEN** `r.IsError()` returns `true` and `errors.Is(r.Err(), &core.ValidationError{})` matches

### Requirement: Value objects use New prefix per branch

Exported constructors for value objects SHALL use the `New` prefix (`core.NewBranchName`, `core.NewProjectName`, `core.NewWorktreePath`). Sentinels SHALL use the `Err` prefix; error types SHALL use the `Error` suffix. Enum-typed value objects (`ShellType`, `ContextType`, `PathType`) SHALL place an explicit `Unknown`/`Invalid` variant at `iota` position 0.

#### Scenario: Constructors use New prefix

- **WHEN** the package user enumerates exported constructors in `internal/core/`
- **THEN** every constructor SHALL match `^New[A-Z][A-Za-z0-9]*$`

#### Scenario: Enum types declare zero-value sentinel

- **WHEN** the package user reads `core.ShellType`, `core.ContextType`, or `core.PathType`
- **THEN** each type SHALL declare an `Unknown` or `Invalid` constant at `iota` position 0
