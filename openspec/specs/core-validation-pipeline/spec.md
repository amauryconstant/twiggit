# Capability: Validation Pipeline

## Purpose

Defines the `core.ValidationPipeline[T]` type and `core.ValidationFunc[T]` function type used to compose zero or more checks against a candidate value of type `T`. The pipeline is the canonical composition mechanism for `core.ValidateBranchName`, `core.ValidateProjectName`, and `core.ValidateShellType` (each implemented as a pre-built pipeline of one or more validators). Lives in `internal/core/validation.go`.

## Requirements

### Requirement: ValidationFunc[T] signature

`core.ValidationFunc[T]` SHALL be declared as `func(T) error`. A pipeline SHALL compose zero or more `ValidationFunc[T]` instances in declaration order. The constructor SHALL use the `New` prefix: `core.NewValidationPipeline[T any](validators ...core.ValidationFunc[T]) *core.ValidationPipeline[T]`.

#### Scenario: Empty pipeline is valid

- **WHEN** `core.NewValidationPipeline[string]()` is invoked with no validators
- **THEN** `pipeline.Validate(input)` SHALL return `nil` for any input (vacuously true)
- **AND** `pipeline.ValidateAll(input)` SHALL return `nil`

### Requirement: Validate runs fail-fast

`(*core.ValidationPipeline[T]).Run(t T) error` SHALL execute each validator in declaration order and SHALL return the FIRST non-nil error without running subsequent validators.

#### Scenario: First failure short-circuits

- **WHEN** a pipeline of `[v1, v2, v3]` runs `Run(input)` and `v1(input)` returns `err1`
- **THEN** `v2` SHALL NOT run
- **AND** `v3` SHALL NOT run
- **AND** the returned error SHALL be `err1`

### Requirement: ValidateAll collects every failure

`(*core.ValidationPipeline[T]).RunAll(t T) error` SHALL execute every validator regardless of intermediate failures and SHALL return the `errors.Join`-composed error preserving every failure. Each failure is preserved independently in the chain so `errors.Is` and `errors.AsType` can inspect them.

#### Scenario: All failures preserved

- **WHEN** a pipeline of `[v1, v2, v3]` runs `RunAll(input)` and `v1`, `v3` fail with `err1`, `err3`
- **THEN** `errors.Is(err, err1) == true` AND `errors.Is(err, err3) == true`
- **AND** `errors.Is(err, err2) == false` (`v2` did not fail)

#### Scenario: Empty failures returns nil

- **WHEN** all validators return `nil`
- **THEN** `RunAll(input)` SHALL return `nil` (the `errors.Join` of zero errors is `nil`)

### Requirement: Pre-built pipelines for canonical types

The following pre-built pipelines SHALL be exported from `internal/core/validation.go`:

| Pipeline | Constructor | Validation rules |
|---|---|---|
| Branch name | `core.ValidateBranchName(name string) core.Result[bool]` | Non-empty; ≤ 128 chars; no `..`, no `~`, `^`, `:`, `?`, `*`, `[`, no leading `-`, no trailing `/`, no whitespace |
| Project name | `core.ValidateProjectName(name string) core.Result[bool]` | Non-empty; matches `^[A-Za-z0-9._-]+$`; ≤ 64 chars; no leading `.` |
| Shell type | `core.ValidateShellType(s string) core.Result[bool]` | One of `bash`, `zsh`, `fish`; case-sensitive |

Each pre-built pipeline is a single `core.ValidationFunc[string]` for simplicity (no internal composition); future expansion to multiple rules will wrap in a `core.NewValidationPipeline[string](...)` chain.

#### Scenario: Empty branch name rejected

- **WHEN** `core.ValidateBranchName("")` runs
- **THEN** the result SHALL carry a `*core.ValidationError` with `Field == "BranchName"` and a message naming the empty-input rule

#### Scenario: Invalid shell type rejected

- **WHEN** `core.ValidateShellType("powershell")` runs
- **THEN** the result SHALL carry a `*core.ValidationError` with `Field == "ShellType"` and a message listing the supported set

### Requirement: Result wrapping

The pre-built pipelines return `core.Result[bool]` (owned by `core-types`). `Result[bool]` carries either `Value(true)` for valid input or `Err(*core.ValidationError)` for invalid input. Callers SHALL inspect the result via `r.IsSuccess()` / `r.IsError()` and `r.Err()`.

#### Scenario: Result accessors

- **WHEN** `r := core.ValidateBranchName("feat/foo")` returns success
- **THEN** `r.IsSuccess() == true` and `r.Value() == true`
- **AND** `r.Err() == nil`