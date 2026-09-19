# 05 — Error Handling and Exit Codes

---

This reference owns the CLI-specific concerns: the error→exit-code mapping and
TTY-aware formatting. The general Go error craft it builds on — `%w` wrapping,
`errors.Is`/`errors.As`/`errors.AsType`, sentinel vs typed errors, `errors.Join`,
`samber/oops` — is owned by `samber/cc-skills-golang@golang-error-handling`; load
it alongside.

## Principles

1. **`main.go` is the only place that calls `os.Exit`.** Every other layer returns `error` values. The entry point translates errors into exit codes.
2. **Single-handling rule:** an error is either **logged or returned, never both** (from `golang-error-handling`). Returning up to `main.go` where it is formatted and logged once prevents duplicate lines in log aggregators.
3. **Errors are abstraction layers.** Each layer wraps errors with its own context using `%w`. The CLI layer decides how to present them.
4. **User-facing errors answer three questions:** What happened? Why? What to do next.

---

## Error Wrapping with Context

```go
return fmt.Errorf("reading config %s: %w", path, err)
```

Always wrap with `%w` to preserve the error chain for `errors.Is` and `errors.As` inspection upstream.

---

## The UserError Pattern

`UserError` is an optional convenience type for ad-hoc rich errors raised at the
shell. It is one case in the boundary formatter, which `SKILL.md` calls
`FormatError` and which dispatches on every core type (`ValidationError`,
`NotFoundError`, `OperationError`, `ExternalError`) via `errors.As`. The
structured core types carry their own actionable text (`Suggestions`), so prefer
them; reach for `UserError` only when you need a one-off Message/Reason/Hint at a
call site that has no dedicated type.

For errors that should display rich context to the user:

```go
type UserError struct {
    Message string  // What happened
    Reason  string  // Why
    Hint    string  // What to do next
    Err     error   // Underlying error (for debug logging)
}

func (e *UserError) Error() string { return e.Message }
func (e *UserError) Unwrap() error { return e.Err }
```

The UserError branch of `FormatError` (the canonical dispatcher lives in `SKILL.md`):

```go
func formatUserError(w io.Writer, err error) {
    var ue *UserError
    if errors.As(err, &ue) {
        fmt.Fprintf(w, "Error: %s\n", ue.Message)
        if ue.Reason != "" {
            fmt.Fprintf(w, "  Reason: %s\n", ue.Reason)
        }
        if ue.Hint != "" {
            fmt.Fprintf(w, "  Hint: %s\n", ue.Hint)
        }
    } else {
        fmt.Fprintf(w, "Error: %v\n", err)
    }
}
```

Output example:

```
Error: could not connect to API server
  Reason: connection refused at https://api.example.com:443
  Hint: Check your network connection, or set MYAPP_API_URL to use a different server
```

---

## Core Error Types

Define semantic error types in the functional **core**. These carry meaning but
NOT exit codes or formatting — those are CLI-layer concerns. `SKILL.md` shows
`ValidationError` as the representative shape; this reference is the canonical home for
the full hierarchy:

```go
// internal/core/errors.go

// ValidationError — a domain value failed a rule (bad branch name, protected
// branch). This is a runtime failure → exit 1, NOT a CLI usage error.
type ValidationError struct {
    Field       string
    Value       string
    Message     string
    Suggestions []string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("invalid %s %q: %s", e.Field, e.Value, e.Message)
}

type NotFoundError struct {
    Entity string
    Name   string
}
func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s %q not found", e.Entity, e.Name)
}

type OperationError struct {
    Op          string
    Message     string
    Cause       error
    Suggestions []string
}
func (e *OperationError) Error() string { return e.Message }
func (e *OperationError) Unwrap() error { return e.Cause }

// UsageError — the invocation itself was wrong (unknown flag, wrong arg count).
// This is the ONLY core error type that maps to exit code 2. Cobra's own arg
// errors should be wrapped in this.
type UsageError struct {
    Message string
}
func (e *UsageError) Error() string { return e.Message }
```

Errors that originate from external I/O are defined near their origin (e.g.
`internal/git/errors.go`), not in the core — see `SKILL.md`.

---

## The Exit Code Registry

The registry pattern maps errors to exit codes declaratively. It is the scale-up form of
`SKILL.md`'s flat `ExitCodeFor` — the same `UsageError`→2 decision, expressed as a table
so new mappings are one line. This keeps exit-code concerns out of domain errors and makes
the mapping explicit. `SKILL.md` §Exit Code Mapping owns the rationale for the three-code
contract (0/1/2) and the sysexits escape hatch; this reference is only the mechanics:

```go
type ExitCode int

const (
    ExitOK    ExitCode = 0
    ExitError ExitCode = 1
    ExitUsage ExitCode = 2  // matches Bash convention for usage/argument errors
)

type exitMapping struct {
    match func(error) bool
    code  ExitCode
}

// Only UsageError maps to exit 2 ("invalid flags or arguments", per
// samber/cc-skills-golang@golang-cli). Domain failures like ValidationError
// and NotFoundError are runtime errors → exit 1.
var defaultMappings = []exitMapping{
    {match: func(err error) bool { var e *UsageError; return errors.As(err, &e) }, code: ExitUsage},
}

func ExitCodeFor(err error) ExitCode {
    if err == nil {
        return ExitOK
    }
    for _, m := range defaultMappings {
        if m.match(err) {
            return m.code
        }
    }
    return ExitError
}
```

`main.go` is the only place that maps errors to exit codes — it formats via the
canonical `FormatError` and applies `ExitCodeFor`:

```go
func main() {
    f := cmdutil.NewFactory() // builds IOStreams + lazy deps (see SKILL.md §The Factory)
    root := cmd.NewRootCommand(f)
    if err := root.Execute(); err != nil {
        output.FormatError(f.IOStreams, err)
        os.Exit(int(cmdutil.ExitCodeFor(err)))
    }
}
```

---

## Cobra Integration

Set `SilenceErrors: true` and `SilenceUsage: true` on the root command. This prevents Cobra from printing errors and usage text automatically — you handle formatting in `Execute()`.

```go
root := &cobra.Command{
    Use:           "myapp",
    SilenceErrors: true,
    SilenceUsage:  true,
}
```

Commands return errors from `RunE`. The error propagates up to `Execute()`, where it's formatted and mapped to an exit code.

---

## Formatting Guidelines

From clig.dev:

- Lowercase error messages (Go convention)
- No trailing punctuation
- Include the filename/resource that caused the error
- Attribute external errors: `API returned 403: insufficient permissions`
- Color: red for the `Error:` prefix, normal for the message body
- If `--debug` is enabled, print the full wrapped error chain after the user-friendly message

---

## Anti-Pattern: `check(err)` / `must()`

These collapse all error handling into a single response (abort via panic or `log.Fatal`). Fine for throwaway scripts, breaks down as CLIs grow. Every call site loses the ability to recover, retry, or provide context-specific feedback.

---

**See also:** `samber/cc-skills-golang@golang-error-handling` — deeper coverage of `%w` wrapping, `errors.Join`, sentinel errors, and production error patterns with structured logging
