# 01 — Architecture by Scale

This reference expands `SKILL.md`'s Tier 1→3 growth path in depth: it defines three architectural tiers for Go CLIs, each building on the previous one. The tiers are not rigid categories — they are waypoints on a continuous scale. The goal is to give you the simplest viable architecture for your current needs, with a clear upgrade path when complexity demands it.

---

## The Three Concerns: Parse, Execute, Respond

`SKILL.md` defines the two principles this reference builds on — **Functional Core,
Imperative Shell** (pure computation in `internal/core/`, all I/O at the edges) and the
**Parse → Execute → Respond** shape every command takes. This reference doesn't
re-derive them; it shows *when* each tier separates those concerns into distinct code
units. One nuance worth restating: the Respond step in a CLI owns the *entire* output
surface — stdout format, stderr messaging, and exit-code selection — which is broader
than a web responder, and that is deliberate, to keep output concerns out of the core.

```
┌─────────────────────────────────────────────────┐
│                   CLI Binary                     │
│                                                  │
│  ┌──────────┐   ┌──────────┐   ┌─────────────┐  │
│  │  Parse    │──▶│   Core   │──▶│   Respond   │  │
│  │          │   │ (Execute)│   │             │  │
│  │ flags    │   │ validate │   │ stdout fmt  │  │
│  │ args     │   │ compute  │   │ stderr msgs │  │
│  │ env vars │   │ orchestr.│   │ exit code   │  │
│  │ stdin    │   │ (pure)   │   │ (TTY-aware) │  │
│  └──────────┘   └──────────┘   └─────────────┘  │
│                                                  │
└─────────────────────────────────────────────────┘
```

These three concerns exist even in a 50-line tool — they just might all live in the same function. The tiers below are about *when and how* to separate them into distinct code units.

---

## Tier 1: Minimal CLI

**When:** Single command or 2–3 subcommands. One domain. Under ~500 LOC. The kind of tool you might otherwise write in Bash.

### Layout

```
myapp/
├── main.go          # One-liner: os.Exit(run())
├── app.go           # appEnv struct, fromArgs(), run()
├── app_test.go      # Table-driven tests against run()
├── go.mod
└── go.sum
```

Two packages: `main` (one-liner) and everything else. Don't split prematurely. As Dave Cheney argued, consider fewer, larger packages.

### The Single-Line `main`

```go
package main

import "os"

func main() {
    os.Exit(run())
}
```

`func main` cannot return errors. This pattern creates a clear pathway between errors and program termination. It avoids the `check(err)` / `must()` anti-pattern, and makes the CLI testable — `run()` is a normal function that returns an int.

### The `appEnv` Struct

All parsed configuration lives in a single struct. This mirrors Mat Ryer's HTTP `server` struct approach:

```go
type appEnv struct {
    out     io.Writer
    errOut  io.Writer
    verbose bool
    format  string
    // domain-specific fields
    target  string
    dryRun  bool
}

func (app *appEnv) fromArgs(args []string) error {
    fl := flag.NewFlagSet("myapp", flag.ContinueOnError)
    fl.SetOutput(app.errOut)
    fl.BoolVar(&app.verbose, "verbose", false, "verbose output")
    fl.StringVar(&app.format, "format", "plain", "output format (plain, json)")
    fl.StringVar(&app.target, "target", "", "target to process")
    fl.BoolVar(&app.dryRun, "dry-run", false, "show what would happen")
    return fl.Parse(args)
}

func (app *appEnv) run() error {
    // Parse: input is already parsed into appEnv fields
    // Execute: do the actual work in the core
    result, err := process(app.target, app.dryRun)
    if err != nil {
        return err
    }
    // Respond: format output
    return app.writeOutput(result)
}
```

The three concerns are present but inlined — `fromArgs` is Parse, `process` is Execute (the core), `writeOutput` is Respond. No package boundaries needed yet.

### The `run()` Entry Point

```go
func run() int {
    app := &appEnv{
        out:    os.Stdout,
        errOut: os.Stderr,
    }
    if err := app.fromArgs(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        return 2
    }
    if err := app.run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        return 1
    }
    return 0
}
```

This aligns with Peter Bourgon's guidelines: no package-level variables, no `func init`.

### When to Graduate

Move to Tier 2 when any of these become true:

- You need more than 3 subcommands
- Multiple commands share configuration or dependencies
- You want shell completions, persistent flags, or help generation
- The single file exceeds ~300 LOC and you're losing track of what's where

### Framework Choice at Tier 1

| Need | Recommendation |
|------|---------------|
| Single command, flags only | `flag.NewFlagSet` + `appEnv` |
| Single command, flags + env vars + config file | `ff` (Peter Bourgon) |
| 2–3 subcommands | Hand-rolled `switch` on `os.Args[1]`, one `flag.FlagSet` per subcommand |

---

## Tier 2: Standard CLI

**When:** 5–15 commands, shared configuration, one or two domains. The typical developer tool, API client, or workflow automation CLI.

This tier introduces Cobra (or your framework of choice), explicit package separation, and the Parse/Execute/Respond structure as distinct code units. The Factory pattern is introduced here to provide lazy dependency injection.

> **Command placement.** Following `SKILL.md`, `main.go` sits at the module root as
> the composition root and command files live in `cmd/` (this skill's documented
> override of `golang-project-layout` — see `SKILL.md` §Project Structure). If you
> need the command packages to be non-importable by other modules, an `internal/cli/`
> package is an equivalent home for them; pick one and use it consistently.

### Layout

```
myapp/
├── main.go                      # Composition root: builds Factory, runs root cmd
├── cmd/                         # Command definitions (Parse + Execute)
│   ├── root.go                  # Root command, global flags, IOStreams setup
│   ├── deploy.go                # deploy command
│   ├── deploy_test.go
│   ├── status.go                # status command
│   └── status_test.go
├── internal/
│   ├── core/                    # Functional Core (no I/O)
│   │   ├── deploy.go            # Deploy logic — no I/O, no CLI concerns
│   │   ├── deploy_test.go
│   │   ├── status.go
│   │   └── status_test.go
│   ├── output/                  # Output formatting (Respond step)
│   │   ├── formatter.go         # Formatter interface + implementations
│   │   ├── json.go
│   │   ├── table.go
│   │   └── plain.go
│   ├── iostreams/               # I/O abstraction, TTY detection
│   │   └── iostreams.go
│   ├── cmdutil/                 # Factory, exit codes
│   │   ├── factory.go
│   │   └── exit.go
│   └── config/                  # Configuration loading, validation
│       └── config.go
├── testdata/                    # Golden files
├── go.mod
└── go.sum
```

### IOStreams and the Factory

Tier 2 is where you *introduce* the two threading mechanisms `SKILL.md` defines
canonically: the **IOStreams** struct (stdout/stderr, TTY detection, styling, with the
`System()` and `Test()` constructors) and the **Factory** that holds it alongside
lazily-initialized dependencies. Don't redefine them here — construct the Factory once
in `main` and pass it to each `NewCmdXxx`. What is specific to Tier 2 is the *wiring*:
the root command sets IOStreams state from persistent flags, and `main` builds the
Factory's lazy fields.

### Wiring: Root Command Setup

```go
package cmd

func NewRootCommand(f *cmdutil.Factory) *cobra.Command {
    var verbose bool
    
    root := &cobra.Command{
        Use:           "myapp",
        SilenceErrors: true,
        SilenceUsage:  true,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            f.IOStreams.Verbose = verbose
            return nil
        },
    }
    
    root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
    
    root.AddCommand(
        NewCmdDeploy(f, nil),
        NewCmdStatus(f, nil),
    )
    
    return root
}
```

In `main.go`, construct the factory and pass it in:

```go
func main() {
    f := &cmdutil.Factory{
        IOStreams: iostreams.System(),
    }
    f.Config = func() (*config.AppConfig, error) {
        return config.Load()
    }
    f.Client = func() (*api.Client, error) {
        cfg, err := f.Config()
        if err != nil { return nil, err }
        return api.NewClient(cfg.Token), nil
    }
    
    root := cmd.NewRootCommand(f)
    if err := root.Execute(); err != nil {
        output.FormatError(f.IOStreams, err)
        os.Exit(int(cmdutil.ExitCodeFor(err)))
    }
}
```

In tests, inject `bytes.Buffer` instances via `iostreams.Test()`.

### Parse / Execute / Respond in Practice

Each command is one instance of the command pattern from `SKILL.md` §The Command Pattern
— an Options struct, a `NewCmdXxx(f, runF)` constructor that wires flags and `RunE`, and a
`runXxx(opts)` that runs Parse→Execute→Respond. At Tier 2 the concrete shape is:

```go
// cmd/deploy.go
func NewCmdDeploy(f *cmdutil.Factory, runF func(*DeployOptions) error) *cobra.Command {
    opts := &DeployOptions{}
    cmd := &cobra.Command{
        Use:  "deploy [target]",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.IO, opts.Client, opts.Target = f.IOStreams, f.Client, args[0]
            if runF != nil {
                return runF(opts) // test injection point
            }
            return runDeploy(opts) // validate → core.Deploy → format
        },
    }
    cmd.Flags().StringVar(&opts.Tag, "tag", "", "version tag to deploy")
    return cmd
}
```

The Execute step calls into the pure core, which at Tier 2 lives in `internal/core/`
and imports no I/O:

```go
// internal/core/deploy.go — the Core

// Deploy contains pure business logic. No io.Writer, no slog.Logger,
// no cobra.Command. Takes typed inputs, returns typed outputs.
func Deploy(target, tag string, dryRun bool) (*DeployResult, error) {
    if target == "" {
        return nil, &ValidationError{Field: "target", Message: "target is required"}
    }
    if tag == "" {
        tag = "latest"
    }
    return &DeployResult{
        Target:    target,
        Tag:       tag,
        DryRun:    dryRun,
        Timestamp: time.Now(),
    }, nil
}
```

### The Thin-Core Case

`SKILL.md` covers the Functional Core / Imperative Shell split. The Tier-2 nuance worth
adding: when your CLI primarily orchestrates external tools (git wrappers, API clients),
the functional core may be very thin — validate input → call external tool → parse
output. That's fine. Not every CLI has a rich domain; don't fabricate abstraction to
make the core look bigger. The Parse/Execute/Respond structure still applies — the core
is just small.

### When to Graduate

Move to Tier 3 when:

- You have 15+ commands across multiple distinct domains
- Commands have substantially different dependency needs (some talk to APIs, some to databases, some to the filesystem)
- You need a plugin system
- The IOStreams + direct constructor injection becomes unwieldy (too many parameters threading through)
- Multiple teams contribute to the same CLI

---

## Tier 3: Large-Scale CLI

**When:** 20+ commands, multiple domains, plugin system, diverse output formats, multiple teams. Think `gh`, `kubectl`, `terraform`.

This tier introduces vertical slice organization, a factory for lazy dependency injection, and explicit exit code management.

### Layout

```
myapp/
├── main.go                          # Composition root
├── cmd/                             # Command tree (vertical slices)
│   ├── root.go                      # Root command, error→exit code mapping
│   ├── deploy/
│   │   ├── deploy.go                # Parent command
│   │   ├── create/
│   │   │   ├── create.go            # `myapp deploy create`
│   │   │   └── create_test.go
│   │   ├── status/
│   │   │   ├── status.go            # `myapp deploy status`
│   │   │   └── status_test.go
│   │   └── rollback/
│   │       ├── rollback.go
│   │       └── rollback_test.go
│   ├── repo/
│   │   ├── repo.go
│   │   ├── clone/
│   │   ├── create/
│   │   └── list/
│   └── auth/
│       ├── auth.go
│       ├── login/
│       ├── logout/
│       └── status/
├── internal/
│   ├── cmdutil/                     # Factory (lazy init), shared cmd utilities
│   │   ├── factory.go
│   │   ├── exit.go                  # error→exit code registry
│   │   └── errors.go                # shared error types
│   ├── config/                      # Config loading, validation, XDG paths
│   │   └── config.go
│   ├── auth/                        # Auth token management
│   │   └── auth.go
│   ├── output/                      # Output formatting
│   │   ├── formatter.go
│   │   ├── table.go
│   │   └── json.go
│   ├── iostreams/                   # IOStreams + TTY detection
│   │   └── iostreams.go
│   └── httpmock/                    # Test utilities
│       └── httpmock.go
├── api/                             # API client (or internal/api if not exported)
│   └── client.go
├── testdata/
└── go.mod
```

### Vertical Slices

Commands are organized by what they do, not by architectural layer. Each command package contains everything it needs: its options struct, run function, and tests. Shared concerns (API client, output formatting, IOStreams) live in separate packages that command slices import.

This is how `gh` and `kubectl` are organized. The benefit over horizontal layers: when you add a new command, you add a new directory. You don't scatter changes across `handlers/`, `services/`, `models/`, `repositories/`.

```
Adding `myapp deploy rollback`:

Vertical slice (one directory):     Horizontal layers (four files):
  cmd/deploy/rollback/               internal/handlers/rollback.go
    rollback.go                        internal/services/rollback.go
    rollback_test.go                   internal/models/rollback.go
                                       internal/handlers/rollback_test.go
```

### The Factory Pattern at Scale

`SKILL.md` §The Factory defines the Factory, its lazy function fields, and the
`sync.Once` caching. At Tier 3 that same pattern gains two wrinkles.

**1. IOStreams becomes lazy too.** With 20+ commands, even IOStreams is exposed as a
function field, so a command like `myapp version` pays for nothing it doesn't use:

```go
// internal/cmdutil/factory.go
type Factory struct {
    IOStreams  func() *iostreams.IOStreams
    HttpClient func() (*http.Client, error)
    Config     func() (config.Config, error)
    AuthToken  func() (string, error)
    Logger     func() *slog.Logger
}
```

**2. Vertical-slice consumption.** Commands live in `cmd/<group>/<verb>/` packages
and pull only what they need from the factory inside `RunE`:

```go
func NewCmdDeployCreate(f *cmdutil.Factory, runF func(*DeployCreateOptions) error) *cobra.Command {
    opts := &DeployCreateOptions{}
    return &cobra.Command{
        Use: "create",
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.IO = f.IOStreams()
            client, err := f.HttpClient()
            if err != nil {
                return err
            }
            opts.Client = client
            if runF != nil {
                return runF(opts) // test injection point
            }
            return runDeployCreate(opts)
        },
    }
}
```

This is the pattern `gh` uses — the natural evolution of the Tier-1 `appEnv` struct: one
dependency holder, now with lazy initialization and per-command isolation.

### Exit Code Registry

At scale, the flat `ExitCodeFor` from `SKILL.md` becomes a declarative registry — a
slice of `(match func(error) bool, code int)` mappings the root command walks to convert
any `RunE` error into an exit code. That registry, its `UsageError`→2 mapping, and the
sysexits guidance are owned by [05-errors.md](./05-errors.md) §The Exit Code Registry.
The Tier-3 point is only *where* it runs: `main.go` formats the error and applies the
registry in one place, keeping exit-code concerns out of domain error types.

### Middleware via Cobra Hooks

Cross-cutting concerns (auth checks, timing, debug headers) use Cobra's hook chain:

```
PersistentPreRunE (root) → PreRunE (command) → RunE → PostRunE → PersistentPostRunE
```

```go
// Auth check as middleware
root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    // Skip auth for commands that don't need it
    if cmdutil.IsAuthless(cmd) {
        return nil
    }
    token, err := f.AuthToken()
    if err != nil {
        return &AuthError{Err: err, Hint: "run 'myapp auth login' to authenticate"}
    }
    // Store token in context for downstream use
    cmd.SetContext(withAuthToken(cmd.Context(), token))
    return nil
}
```

For composing multiple middleware functions, wrap the original hook:

```go
func addMiddleware(cmd *cobra.Command, mw func(*cobra.Command, []string) error) {
    original := cmd.PersistentPreRunE
    cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        if err := mw(cmd, args); err != nil {
            return err
        }
        if original != nil {
            return original(cmd, args)
        }
        return nil
    }
}
```

---

## Dependency Injection Strategy by Tier

| Tier | Strategy | Mechanism |
|------|----------|-----------|
| Tier 1 | Field injection on `appEnv` | Struct fields set at construction |
| Tier 2 | Factory with lazy initialization | Pass `*Factory` to each `NewCmdXxx`; IOStreams is a field of Factory |
| Tier 3 | Factory + vertical slices | Same Factory pattern, commands in `cmd/` vertical packages |

At all tiers, the principle is the same: **accept interfaces, return structs**. Inject `io.Writer` instead of using `os.Stdout`. Inject `*http.Client` (or an interface wrapper) for HTTP calls. Use `fs.FS` for filesystem abstraction.

DI frameworks (Wire, Dig, Fx) are available but rarely justified for CLIs. Manual wiring is almost always sufficient and easier to debug.

---

## Architectural Patterns: When to Reach for More

The Parse/Execute/Respond foundation handles most CLIs. Here's when to consider
additional structure. (Functional Core / Imperative Shell is not "more" — it's the
baseline, canonical in `SKILL.md` and assumed from Tier 2 up.)

### Vertical Slices (Default at Tier 3)

Organize by command, not by layer. Each command package is self-contained. Shared utilities live in `internal/cmdutil/` or `internal/output/`.

### Horizontal Layers (Optional, use with caution)

Full hexagonal / clean architecture adds ports-and-adapters boundaries. Justified when:

- Your CLI is also a library (the domain is imported by other Go modules)
- You need to swap entire subsystems (e.g., local storage ↔ cloud storage)
- The domain is genuinely complex and benefits from formal boundary enforcement

For most CLIs, this is over-engineering. The implicit layers of Cobra commands → factory → core functions → output formatters provide sufficient separation.

---

## Decision Matrix

| Situation | Tier | Key Pattern |
|-----------|------|-------------|
| Script replacement, single purpose | 1 | `appEnv` + `flag.NewFlagSet` |
| Developer tool, API client, 5-15 commands | 2 | Cobra + Factory + Parse/Execute/Respond split |
| Platform CLI, 20+ commands, plugins, teams | 3 | Cobra + Factory + vertical slices + exit registry |
| Single-purpose tool needing config file support | 1 | `ff` + `appEnv` |
| Tool that's also a library | 2-3 | Horizontal layers with `pkg/` for public API |
| CLI wrapping a single external tool | 1-2 | Thin domain, Facade pattern is honest |

---

## The Growth Path

```
Tier 1                    Tier 2                    Tier 3
─────────────────────────────────────────────────────────────
main.go + app.go    →    main.go + cmd/         →   cmd/<group>/ vertical slices
appEnv struct        →    Factory (lazy init)   →   Factory + vertical slices
switch on error      →    error types           →   Exit code registry
inline formatting    →    Formatter interface    →   TTY-adaptive + format flag
flag.NewFlagSet      →    Cobra                 →   Cobra + middleware hooks
direct testing       →    golden files          →   golden files + httpmock + E2E
```

Each transition preserves what came before. The `appEnv` struct's fields become IOStreams fields. The inline formatting becomes a Formatter interface. The Cobra commands from Tier 2 move into per-command packages at Tier 3. Nothing is thrown away — it's refined.
