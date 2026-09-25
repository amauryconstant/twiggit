# Commands and Wiring

How a Tier 2+ CLI is wired: one pattern per command, a Factory for shared dependencies, IOStreams for the terminal, a root command, and an entry point that owns the exit code. The Cobra API itself (hooks, validators, completion directives) → See `samber/cc-skills-golang@golang-spf13-cobra`.

## Contents

- [The command pattern](#the-command-pattern)
- [The Factory](#the-factory)
- [IOStreams](#iostreams)
- [Root command](#root-command)
- [Middleware via Cobra hooks](#middleware-via-cobra-hooks)
- [Entry point](#entry-point)

## The command pattern

A command is a file (Tier 2) or a package (Tier 3) holding three things — [create.go](../assets/examples/create.go):

1. **Options struct** — I/O, lazy dependencies, context, and parsed flags and arguments. Tests build it directly.
2. **Constructor** `NewCmdX(f *cmdutil.Factory, runF func(*XOptions) error) *cobra.Command` — defines flags and a `RunE` that copies dependencies from the Factory into the options, then calls `runF` when set, else `runX`.
3. **Run function** `runX(opts *XOptions) error` — Parse → Execute → Respond: validate through the core, call the adapter, write to `opts.IO`.

Properties this buys:

- **One file per command** — parse, execute, and respond for a command read top to bottom in one place.
- **`runF` test seam** — a test receives the fully parsed options without running the command ([create_test.go](../assets/examples/create_test.go)).
- **Lazy dependencies** — `opts.Client` is a `func`, called only on the code path that needs it.
- **Consumer-side interfaces** — the options hold `func() (worktreeCreator, error)`, an interface declared in the command file, so tests pass a fake and the adapter package declares no interface.
- **Output beside the logic** — the command knows what it produced and formats it with shared helpers from `internal/output`; there is no presenter layer.

A query command resolves its `--output` formatter before doing any work, so a bad value fails fast as a usage error ([list.go](../assets/examples/list.go)).

## The Factory

[factory.go](../assets/examples/factory.go) — created once in `main.go`, passed to every constructor:

```go
type Factory struct {
    IOStreams *iostreams.IOStreams
    Config    func() (*config.AppConfig, error)
    Client    func() (*git.Client, error)
}
```

- **Function fields are lazy**: a dependency is built only when a command calls it inside `RunE`, so `myapp version` never reads config and import cycles never form.
- **`sync.OnceValues` caches each field** — `Client` calls `Config`, and the config file is still read once per invocation.
- **The Factory is a composition-time tool, not a service locator**: only constructors in `cmd/` receive it; `RunE` extracts concrete values and the core never imports `cmdutil`. That keeps `golang-dependency-injection`'s "never pass the container" rule intact.
- **Manual wiring scales** — `gh` runs 30+ commands on this pattern with no DI framework.
- **Split when it bloats**: past ~10 fields with conditional setup, move rarely used dependencies into the commands that need them.

At Tier 3, `IOStreams` becomes a function field too, so trivial commands pay for nothing.

## IOStreams

[iostreams.go](../assets/examples/iostreams.go) and [styles.go](../assets/examples/styles.go):

| Member | Purpose |
| --- | --- |
| `In`, `Out`, `ErrOut` | stdin, data output, everything else |
| `IsStdoutTTY()` | Human output vs bare data |
| `IsInteractive()` | stdout **and** stdin are terminals — the gate for every prompt |
| `Styles()`, `ErrStyles()` | lipgloss styles for `Out` and `ErrOut`, each keyed on its own stream's TTY; identity functions when color is off (`NO_COLOR`, pipes) |
| `Verbose`, `Verbosef()` | `-v` progress lines on stderr |
| `Logger` | `slog` logger, discarding unless `--debug` / `MYAPP_DEBUG` |
| `System()` / `Test()` | Real streams / buffers with a non-TTY, colorless setup |
| `SetTTY(stdout, stdin)` | Test helper for the interactive path |

TTY state is captured once in `System()`, so every check goes through the struct and tests control it.

## Root command

[root.go](../assets/examples/root.go):

- `SilenceUsage: true` and `SilenceErrors: true` — `main.go` prints each error once, without the usage dump.
- Persistent `-v/--verbose` and `--debug` flags; `PersistentPreRunE` sets `IOStreams.Verbose` and `IOStreams.Logger`.
- **Usage-error wiring** ([args.go](../assets/examples/args.go)) so misuse exits 2:
  - `root.SetFlagErrorFunc(cmdutil.FlagError)` — unknown flags and bad values; subcommands inherit it.
  - `Args: cmdutil.UsageArgs(cobra.ExactArgs(1))` on every command with positional arguments.
  - `Args: cmdutil.NoSubcommand` plus `RunE: cmd.Help` on the root and on group commands — Cobra's own unknown-command error is untyped, so this replaces it (suggestions kept).
- Commands come from `NewCmdX(f, nil)` calls in `NewRootCommand`, never from `init()` — each test builds a fresh tree.

Cobra's required-flag and flag-group errors stay untyped. Check required input in `runX` instead and return `&core.UsageError{}`, which also leaves room for a prompt fallback (→ `prompts-and-tui.md`).

## Middleware via Cobra hooks

Cross-cutting concerns (auth, timing) use the hook chain `PersistentPreRunE → PreRunE → RunE → PostRunE → PersistentPostRunE`. A child's `PersistentPreRunE` **replaces** the parent's: Cobra runs only the nearest one. A group hook therefore calls its ancestor's hook first, or the root sets `cobra.EnableTraverseRunHooks = true` (global) to run every level:

```go
// runParentPreRun runs the nearest ancestor hook above owner, the command
// that defines the calling hook. Walk from owner, not from c (the leaf):
// starting at c finds owner's own hook and recurses.
func runParentPreRun(owner, c *cobra.Command, args []string) error {
    for p := owner.Parent(); p != nil; p = p.Parent() {
        if p.PersistentPreRunE != nil {
            return p.PersistentPreRunE(c, args)
        }
    }
    return nil
}

admin.PersistentPreRunE = func(c *cobra.Command, args []string) error {
    if err := runParentPreRun(admin, c, args); err != nil { // root: logging
        return err
    }
    return checkAdminToken(c, args)
}
```

To stack several middlewares on one command, `addPreRun` chains them. It composes only hooks set on that same command and does not fix parent shadowing:

```go
func addPreRun(cmd *cobra.Command, mw func(*cobra.Command, []string) error) {
    next := cmd.PersistentPreRunE
    cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
        if err := mw(c, args); err != nil {
            return err
        }
        if next != nil {
            return next(c, args)
        }
        return nil
    }
}
```

An auth hook skips commands annotated as authless and returns an error with the next step when the token is missing:

```go
func requireAuth(loadToken func() (string, error)) func(*cobra.Command, []string) error {
    return func(c *cobra.Command, _ []string) error {
        if c.Annotations["auth"] == "skip" { // set on version, completion, auth login
            return nil
        }
        if tok, err := loadToken(); err != nil || tok == "" {
            return &core.OperationError{Op: "auth", Message: "not logged in", Cause: err,
                Suggestions: []string{"run 'myapp auth login'"}}
        }
        return nil
    }
}
```

Wire it with `addPreRun(root, requireAuth(...))`; commands that need no token set `Annotations: map[string]string{"auth": "skip"}`.

All three, exported from `cmdutil`: [hooks.go](../assets/examples/hooks.go).

## Entry point

[main.go](../assets/examples/main.go) is the same shape at every tier: `func main() { os.Exit(run()) }`. `run()`:

1. Defers panic recovery — prints `fatal: …`, the stack only under `MYAPP_DEBUG`, returns 1.
2. Creates `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` and runs `root.ExecuteContext(ctx)`.
3. Returns 130 when the context was canceled by a signal; otherwise calls `output.FormatError` once and returns `cmdutil.ExitCodeFor(err)`.

`os.Exit` skips deferred calls, which is why every defer lives in `run()` and `main` only exits.
