# Errors and Exit Codes

The CLI-specific half of error handling: semantic error types, the exit-code contract, formatting at the boundary, and turning Cobra's misuse errors into usage errors. General Go error craft (`%w`, `errors.Is`/`As`/`AsType`, sentinels, `errors.Join`, `samber/oops`) → See `samber/cc-skills-golang@golang-error-handling`.

## Contents

- [Principles](#principles)
- [Error types](#error-types)
- [The exit-code contract](#the-exit-code-contract)
- [Usage errors from Cobra](#usage-errors-from-cobra)
- [Formatting at the boundary](#formatting-at-the-boundary)
- [The exit-code registry (Tier 3)](#the-exit-code-registry-tier-3)
- [Message guidelines](#message-guidelines)

## Principles

1. **`main.go` is the only place that exits** — every layer returns `error`; `run()` formats once and picks the code ([main.go](../assets/examples/main.go)).
2. **Each error is logged or returned, never both** — returning to one boundary keeps every message single.
3. **Each layer wraps with its own context** — `fmt.Errorf("reading config %s: %w", path, err)` — and the boundary prints the full chain.
4. **A user-facing error answers what happened, why, and what to do next** — `OperationError{Message, Cause, Suggestions}` carries all three.
5. **Abort helpers (`check(err)`, `must()`, `log.Fatal`) belong only in throwaway scripts** — they remove every caller's chance to add context, retry, or recover.

## Error types

[core_errors.go](../assets/examples/core_errors.go) — they carry meaning, never exit codes or formatting:

| Type | Meaning | Exit |
| --- | --- | --- |
| `ValidationError{Field, Value, Message, Suggestions}` | A domain value broke a rule | 1 |
| `NotFoundError{Entity, Name}` | A named thing does not exist | 1 |
| `OperationError{Op, Message, Cause, Suggestions}` | An operation failed; what, why, next | 1 |
| `UsageError{Message, Cause}` | The invocation was malformed | 2 |

Errors born in I/O are defined beside their origin, not in the core. An adapter:

- takes `ctx` and runs tools with `exec.CommandContext`, so Ctrl-C stops them;
- parses tool output into core types and never returns raw stdout;
- wraps with the operation and target (`fmt.Errorf("cloning %s: %w", url, err)`) and returns — no `log` call; diagnostic detail goes to the debug `slog` logger.

```go
// internal/git/errors.go
type ExternalError struct {
    Tool, Operation, Message string
    Kind                     ErrorKind // NotFound, Timeout, Permission
    Cause                    error
}
```

## The exit-code contract

| Code | When | Produced by |
| --- | --- | --- |
| 0 | Success | `run()` |
| 1 | Any runtime failure | `ExitCodeFor` default |
| 2 | `UsageError`: bad flag, bad argument, unknown command, missing input without a TTY | `ExitCodeFor` |
| 130 | Interrupted (128 + SIGINT) | `run()` when the signal context is canceled |

Most consumers check `$? -ne 0` and read stderr, so the error _type_ carries the meaning and the _code_ stays a blunt signal. Custom codes such as 3–6 follow no convention, and scripts rarely test for them. When automation genuinely dispatches on codes, add codes from the `sysexits.h` range 64–78 for runtime categories (69 service unavailable, 75 temporary failure, 78 invalid config), which avoids Bash's 2/126/127 and the 128+N signal codes. A usage error still exits 2, not 64, so the 0/1/2/130 contract holds.

A missing config file gets no code at all: the loader falls back to defaults. Only an unreadable or invalid file fails.

## Usage errors from Cobra

Cobra reports misuse as untyped errors; [args.go](../assets/examples/args.go) converts them:

| Misuse | Hook |
| --- | --- |
| Unknown flag, bad flag value | `root.SetFlagErrorFunc(cmdutil.FlagError)` (inherited by subcommands) |
| Wrong argument count | `Args: cmdutil.UsageArgs(cobra.ExactArgs(1))` |
| Unknown subcommand | `Args: cmdutil.NoSubcommand` + `RunE: cmd.Help` on root and group commands |
| Missing required input | Check in `runX`, return `&core.UsageError{}` |

Set `SilenceErrors` and `SilenceUsage` on the root so Cobra prints neither; `main.go` prints once.

## Formatting at the boundary

[format_error.go](../assets/examples/format_error.go) prints `error: <full message>` to stderr, then the `Suggestions` of any `ValidationError` or `OperationError` in the chain:

```text
error: invalid config ~/.config/myapp/config.toml:
invalid timeout "0s": must be positive
invalid default_branch "": must not be empty
```

Printing the full message keeps every wrap layer and every joined error; typed errors only add next steps. Colors come from `IOStreams.ErrStyles()`, keyed on stderr's TTY state, so CI logs and redirected stderr get plain text.

## The exit-code registry (Tier 3)

At scale, `ExitCodeFor` becomes a table so each new mapping is one line:

```go
var exitMappings = []struct {
    match func(error) bool
    code  ExitCode
}{
    {func(err error) bool { var e *core.UsageError; return errors.As(err, &e) }, ExitUsage},
}
```

`ExitCodeFor` walks the table and falls back to `ExitError`. `run()` in `main.go` still applies it in one place.

## Message guidelines

- Name the file or resource involved: `reading config ~/.config/myapp/config.toml: permission denied`.
- Attribute external failures: `API returned 403: insufficient permissions`.
- Put the fix in `Suggestions`, one action per line.
- Under `--debug`, log extra diagnostic context through `IOStreams.Logger`.
