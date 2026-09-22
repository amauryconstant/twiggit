# Spec Delta

## Purpose

Defines the value-object pattern and validation pipeline for the `internal/core/` package, making the rule that every domain type is always-valid post-construction. The pipeline's `ValidateAll` uses `errors.Join` to preserve the full error chain; `output.FormatError` walks each joined error independently.

## ADDED Requirements

### Requirement: Value objects expose always-valid construction

Every value object (`core.NewBranchName`, `core.NewProjectName`, `core.NewWorktreePath`) SHALL return `(T, error)` from its constructor. The returned value SHALL be safe to use without further validation; if the input fails the rules, the constructor SHALL return a `*core.ValidationError`.

#### Scenario: Valid branch name constructs successfully
- **WHEN** the caller invokes `b, err := core.NewBranchName("feat/foo")`
- **THEN** `err` is nil and `b` is a non-zero `core.BranchName`

#### Scenario: Invalid branch name returns ValidationError
- **WHEN** the caller invokes `b, err := core.NewBranchName("feat..bad")`
- **THEN** `err` is a `*core.ValidationError` carrying `Field="BranchName"`, `Value="feat..bad"`, and a `Message` naming the rule violated

### Requirement: Pipeline[T] composes Validator[T] with Validate (fail-fast) or ValidateAll (collect-all)

The `core.Pipeline[T]` type SHALL compose zero or more `core.Validator[T]` functions. `Pipeline.Validate(t)` SHALL run each validator in order and return the first error (fail-fast); `Pipeline.ValidateAll(t)` SHALL run all validators and return an error composed via `errors.Join` of every failure. `FormatError` SHALL walk the joined chain via `errors.As` and render each failure as a separate validation hint.

#### Scenario: Validate stops at first failure
- **WHEN** a pipeline of two validators runs `pipeline.Validate(input)` and the first fails
- **THEN** the second validator SHALL NOT run and the returned error SHALL be the first validator's error

#### Scenario: ValidateAll collects every failure via errors.Join
- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and validators 1 and 3 fail
- **THEN** `errors.Is(err, validator1Err)` returns `true` and `errors.Is(err, validator3Err)` returns `true` (both preserved in the chain)
- **AND** `errors.Is(err, validator2Err)` returns `false` (the passing validator is not in the chain)

### Requirement: Core types are immutable post-construction

Once constructed, a value object SHALL expose read-only fields (or accessor methods). Mutation methods SHALL NOT exist on core types.

#### Scenario: No mutation methods on BranchName
- **WHEN** the package user searches `core.BranchName` for setter or mutation methods
- **THEN** no method SHALL match `Set`, `Mutate`, `Assign`, `Replace`, or `Update` patterns
