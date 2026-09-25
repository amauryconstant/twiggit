# The Functional Core

`internal/core/` holds the pure computation: value objects, types with behavior, rule functions, validation, and semantic errors. It fits CLIs because they are short-lived, single-user processes whose hard problems are I/O orchestration, error presentation, and platform differences — so plain types and functions carry the domain, tested with values in and values out.

## Contents

- [What belongs in the core](#what-belongs-in-the-core)
- [Value objects](#value-objects)
- [Types with behavior and rule functions](#types-with-behavior-and-rule-functions)
- [Validation pipeline](#validation-pipeline)
- [The thin core](#the-thin-core)
- [Dependency policy](#dependency-policy)

## What belongs in the core

| Belongs | Lives in the shell |
| --- | --- |
| Value objects (`BranchName`) | Flag parsing (`cmd/`) |
| Types with behavior (`Project.ActiveWorktrees`) | Config loading and serialization (`internal/config`) |
| Rule functions (`CanDelete`) | git, HTTP, filesystem calls (adapters) |
| Validation pipelines | Terminal output, prompts (`internal/output`, `cmd/`) |
| Semantic error types ([core_errors.go](../assets/examples/core_errors.go)) | Exit codes, error formatting, concurrency, `context.Context` |

Examples: [core_types.go](../assets/examples/core_types.go), [validation.go](../assets/examples/validation.go).

## Value objects

A value object validates on construction, so every holder has a valid value:

```go
type BranchName struct{ value string }

func NewBranchName(raw string) (BranchName, error) {
    if err := branchRules.Validate(raw); err != nil {
        return BranchName{}, err
    }
    return BranchName{value: raw}, nil
}
```

- **Create one** when the same value is validated or normalized in more than one place — the rule then lives once.
- **Keep plain types** for values whose Go type already states every invariant (`time.Duration`, a count, a pass-through string).

## Types with behavior and rule functions

Business rules are methods on the types they concern (`Project.ActiveWorktrees`). A rule spanning several values is a plain function that receives exactly the values it needs:

```go
func CanDelete(p *Project, wt Worktree, protected []string) error
```

It takes the protected-branch list, not the config struct, so the core never depends on config loading. The command reads config and passes `cfg.Protected`.

## Validation pipeline

`Pipeline[T]` composes validators with two modes:

| Mode | Use for | Why |
| --- | --- | --- |
| `Validate` — stop at first failure | User input, value objects | One clear error per mistake |
| `ValidateAll` — `errors.Join` every failure | Config files, multi-field input | The user fixes everything in one pass |

`output.FormatError` prints the whole joined message, so every failure reaches the user.

## The thin core

A CLI that mostly orchestrates external tools (git wrappers, API clients) has a thin core: validate input, call the tool, parse its output. That is a correct result; the core holds what the domain actually has, and Parse → Execute → Respond still shapes each command.

## Dependency policy

The core imports the standard library (minus I/O packages: `os`, `net`, `os/exec`, `io/fs`) and `github.com/samber/lo`. Enforce it with the depguard rule in [.golangci.yml](../assets/examples/.golangci.yml). Verify with `golangci-lint run` (no `depguard` findings) and `go list -f '{{join .Imports "\n"}}' ./internal/core/... | sort -u` (only stdlib and `github.com/samber/lo`).

`lo` earns its place with zero dependencies and concrete helpers (`Filter`, `Contains`, `Find`) that replace loop boilerplate; prefer `slices`/`maps`/`iter` where the stdlib has the equivalent. Optional values are `*T` and fallible operations return `(T, error)` — the idiomatic Go forms — so the core carries no monad library.

→ See `samber/cc-skills-golang@golang-samber-lo` for the `lo` API.
