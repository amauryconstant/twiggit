---
name: golang-cli-architecture
description: "Golang CLI architecture — pragmatic, proportional design for Go command-line tools, derived from how gh, kubectl, and docker are built. Use when scaffolding a new Go CLI, adding or restructuring commands, deciding where logic lives, designing error handling or output, or reviewing an existing CLI's structure. Also triggers for Go CLI project layout, package dependency rules, and choosing between Cobra, ff, or stdlib flag."
user-invocable: true
license: MIT
compatibility: "Designed for Claude Code or similar AI coding agents, and for projects using Golang."
metadata:
  author: amaury
  version: "2.0.0"
  openclaw:
    emoji: "⌨️"
    homepage: https://github.com/amaurybrisou/go-cli-skill
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
---

**Persona:** You are a Go CLI architect. You build tools that feel native to the Unix shell — composable, scriptable, and predictable under automation. You choose proportional complexity: the simplest structure that solves the problem without ceremony.

**Modes:**

- **Build** — scaffolding a new CLI from scratch: apply the Tier 1→3 growth path, wire the Factory, define IOStreams and root command, follow the three-concern layout.
- **Extend** — adding commands, flags, or completions to an existing CLI: read the current command tree first, extend without breaking existing patterns.
- **Review** — auditing an existing CLI for correctness: check the Decision Triggers table, verify stdout/stderr discipline, error-type mapping, exit code contract, and testability.

> **Ecosystem positioning.** This skill is a standalone plugin that complements `samber/cc-skills-golang`; every `samber/cc-skills-golang@…` cross-reference below assumes that plugin is installed alongside this one. When both are present, this skill is the authority for CLI *structure* and **takes precedence** over these skills for CLI work:
> - `samber/cc-skills-golang@golang-cli` — deeper architectural model (Functional Core / Imperative Shell, Tier 1→3 growth path, Factory pattern, IOStreams abstraction); prefers `koanf`/TOML over `viper`/YAML and `lipgloss`/`huh` over `fatih/color`
> - `samber/cc-skills-golang@golang-dependency-injection` — Factory pattern with lazy function fields replaces DI containers for CLI-scale applications
> - `samber/cc-skills-golang@golang-project-layout` — CLI-specific Tier 1/2/3 layouts apply instead of the generic layout guide
> - `samber/cc-skills-golang@golang-concurrency` (partially) — sequential-first; reach for `errgroup` (with `SetLimit`) only when I/O latency is measurably user-perceptible
> - `samber/cc-skills-golang@golang-testing` (partially) — CLI testing pyramid (Core → Command via `runF` → Integration → E2E) takes precedence
>
> For general Go craft it **defers** to the ecosystem — load these alongside it:
> - `samber/cc-skills-golang@golang-spf13-cobra` — Cobra command-tree API (hook chain, args validators, completion directives)
> - `samber/cc-skills-golang@golang-spf13-viper` — if the project uses `viper` instead of the recommended `koanf`
> - `samber/cc-skills-golang@golang-error-handling` — `%w` wrapping, `errors.Is/As/AsType`, the single-handling rule
> - `samber/cc-skills-golang@golang-samber-oops` — structured production errors when the lightweight local error types aren't enough
> - `samber/cc-skills-golang@golang-samber-slog` — `slog` handler ecosystem for debug logging
> - `samber/cc-skills-golang@golang-stretchr-testify` — `assert`/`require` API used in command tests
> - `samber/cc-skills-golang@golang-popular-libraries` — source of truth for library selection ([14-libraries.md](./references/14-libraries.md) mirrors it)

## Design Philosophy

Core beliefs behind this guide:

1. Start with structure, not libraries — the three-concern layout (Parse / Execute / Respond) applies even to 200-line tools.
2. Vertical slices over horizontal layers — a command is a self-contained slice from input to output, not a traversal through N architectural layers.
3. Honor the Unix contract — stdout is for data, stderr is for humans, exit codes signal success/failure.
4. Pure logic is testable without mocks — if a function touches I/O, it does not belong in the core.
5. Grow into complexity — adopt Tier 2/3 patterns only when the pain of staying at Tier 1 is real, not predicted.
6. Interfaces where consumed — define an interface next to the code that calls it, not next to the implementation.
7. Side effects are imperative — CLIs are synchronous request→response invocations. Keep orchestration explicit; don't hide it behind event buses or async dispatch.

# Pragmatic Architecture for Go CLIs

An architecture for Go command-line tools derived from how successful CLIs
(gh, kubectl, docker) are actually built — not from web-service patterns
adapted to the terminal. Prioritizes discoverability, testability, and
proportional complexity: the architecture should be as simple as the
problem allows and no simpler.

## Foundational Model: Functional Core, Imperative Shell

The architecture is organized around one structural principle:

**Functional Core** — pure computation with no I/O. Types with behavior,
validation, business rules, decision logic. Depends on nothing except
stdlib and curated zero-dependency libraries. Tested with values in,
values out — no mocks.

**Imperative Shell** — everything that does I/O. External tool calls, API
requests, filesystem access, terminal output, user prompts. Orchestrates
calls to the core and to external systems. Tested with integration tests
or by injecting test doubles for I/O boundaries.

This is Gary Bernhardt's "Functional Core, Imperative Shell" pattern,
applied to CLIs. The core has many logic paths but no dependencies. The
shell has many dependencies but few logic paths. You test the core
exhaustively with unit tests and the shell sparingly with integration tests.

**Why this and not DDD / Clean Architecture / Hexagonal**: Those patterns
were designed for long-lived, stateful, multi-user systems with complex
domain model ambiguity. CLIs are short-lived, single-user, stateless
process invocations. The hard problems in CLIs are I/O orchestration,
error presentation, and cross-platform behavior — not domain model
ambiguity. Importing DDD's tactical toolkit (aggregates, specifications,
domain services, bounded contexts) adds ceremony without proportional
benefit for most CLIs.

## Three Concerns, Not Five Layers

Every CLI command does three things:

1. **Parse** — collect input from flags, arguments, environment, config
2. **Execute** — validate, decide, call external systems, produce a result
3. **Respond** — format the result for the terminal (or pipe, or JSON)

These three concerns map to concrete code, not to abstract layers:

```text
main.go          → wires dependencies, runs root command
cmd/             → Parse + Execute: Cobra commands, flag definitions, I/O orchestration
internal/core/   → Functional Core: types, validation, rules (no I/O)
internal/output/ → Respond: formatting, error display, styles
internal/<dep>/  → I/O adapters: git, api, config, filesystem
```

Boundaries come from package structure and interfaces placed where they are
consumed — not from layers mandated upfront (`application/`, `presenter/`, or a
`service/` vs `infrastructure/` split).

### Why Not More Layers?

Each layer boundary has a cost: an interface definition, a constructor,
a mapping between types, a place to look when debugging. These costs are
justified when the boundary enables independent change or independent
testing. In a CLI:

- You will never swap Cobra for another framework without rewriting `cmd/`.
- You will never serve the same logic over HTTP and CLI simultaneously.
- You will never have two implementations of "call git" running in prod.

The boundaries that matter are: pure logic vs I/O (testability) and
stable types vs external formats (change isolation). Two boundaries, not
five.

## The Command Pattern (How gh Does It)

Each CLI command follows a self-contained pattern inspired by the GitHub
CLI. A command is a package (or a file, for small CLIs) containing:

1. An **Options struct** — holds all inputs: parsed flags, resolved
   dependencies (as functions for lazy init), I/O streams
2. A **constructor** `NewCmdFoo(f *Factory, runF ...) *cobra.Command` —
   builds the Cobra command, defines flags, wires RunE
3. A **run function** `runFoo(opts *FooOptions) error` — the actual work:
   validate, execute, respond

```go
// cmd/create.go

type CreateOptions struct {
    IO       *iostreams.IOStreams
    Config   func() (*config.AppConfig, error)  // lazy
    Client   func() (*git.Client, error)        // lazy
    Ctx      context.Context
    
    // Flags
    Name     string
    Source   string
    Force    bool
    Verbose  bool
    Format   string // --output: json | table | plain
}

func NewCmdCreate(f *Factory, runF func(*CreateOptions) error) *cobra.Command {
    opts := &CreateOptions{}
    
    cmd := &cobra.Command{
        Use:   "create <name>",
        Short: "Create a new worktree",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.IO = f.IOStreams
            opts.Config = f.Config
            opts.Client = f.Client
            opts.Ctx = cmd.Context()
            opts.Name = args[0]
            
            if runF != nil {
                return runF(opts) // test injection point
            }
            return runCreate(opts)
        },
    }
    
    cmd.Flags().StringVarP(&opts.Source, "source", "s", "main", "Source branch")
    cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation")
    cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Show verbose output")
    cmd.Flags().StringVarP(&opts.Format, "output", "o", "", "Output format: json, table, plain")
    return cmd
}

func runCreate(opts *CreateOptions) error {
    // 1. Validate (functional core)
    name, err := core.NewBranchName(opts.Name)
    if err != nil {
        return err
    }
    
    // 2. Execute (imperative shell)
    client, err := opts.Client()
    if err != nil {
        return err
    }
    result, err := client.CreateWorktree(opts.Ctx, name, opts.Source)
    if err != nil {
        return err
    }
    
    // 3. Respond (output)
    if opts.IO.IsStdoutTTY() {
        fmt.Fprintf(opts.IO.Out, "Created %s at %s\n",
            opts.IO.Styles().Success(result.Name),
            opts.IO.Styles().Dim(result.Path))
    } else {
        fmt.Fprintln(opts.IO.Out, result.Path)
    }
    return nil
}
```

**Key properties of this pattern:**

- **Self-contained**: one file contains parse + execute + respond for one
  command. You read one file to understand one command.
- **Testable without mocks**: the `runF` parameter lets tests inject a
  function that receives the fully-parsed options, bypassing Cobra.
- **Lazy dependencies**: `opts.Client` is a `func()` — it's only called
  if the command actually needs it. Commands that don't use git don't
  pay for git client initialization.
- **Output co-located with logic**: the command knows what it produced
  and how to format it. No distant presenter package with a method per
  result type.

### When Output Gets Complex

If a command has complex output (multi-format table/JSON/plain), extract
the formatting into a local helper or a shared output function:

```go
// For structured output, select a Formatter by the --output value:
if opts.Format != "" {
    return output.NewFormatter(opts.Format).Write(opts.IO.Out, items)
}
// Otherwise, human-readable default
output.PrintItems(opts.IO, items)
```

The `output` package provides shared formatting utilities (table printer,
JSON formatter, styles). Individual commands compose these — they don't
delegate to a Presenter.

## The Factory (Dependency Wiring)

A Factory struct provides lazy access to shared dependencies. It is
created once in `main.go` and passed to every command constructor.

```go
// internal/cmdutil/factory.go

type Factory struct {
    IOStreams  *iostreams.IOStreams
    Config    func() (*config.AppConfig, error)
    Client    func() (*git.Client, error)
    Logger    func() *slog.Logger
    
    // Eagerly initialized
    AppVersion string
    Executable string
}
```

**Lazy initialization via function fields** is the default DI mechanism.
The function is called at most once per command execution (the command
calls it in RunE, not in the constructor). This avoids:
- Constructing expensive clients for commands that don't need them
- Import cycles between packages
- DI containers and reflection

The Factory is initialized in `main.go`:

```go
func main() {
    f := &cmdutil.Factory{
        IOStreams:   iostreams.System(),
        AppVersion: version.Version,
        Executable: "myapp",
    }
    
    // Lazy: only constructed when a command calls f.Config()
    f.Config = func() (*config.AppConfig, error) {
        return config.Load()
    }
    
    f.Client = func() (*git.Client, error) {
        appCfg, err := f.Config()
        if err != nil { return nil, err }
        return git.NewClient(appCfg.Timeout), nil
    }
    
    f.Logger = func() *slog.Logger {
        level := slog.LevelWarn
        if os.Getenv("MYAPP_DEBUG") != "" {
            level = slog.LevelDebug
        }
        return slog.New(slog.NewTextHandler(os.Stderr, 
            &slog.HandlerOptions{Level: level}))
    }
    
    root := cmd.NewRootCommand(f)
    if err := root.Execute(); err != nil {
        output.FormatError(f.IOStreams, err)
        os.Exit(int(cmdutil.ExitCodeFor(err)))
    }
}
```

**No DI containers.** Manual wiring scales for CLI-sized applications.
Even `gh` with 30+ commands uses this pattern without a framework.

**The Factory is not a service locator.** `golang-dependency-injection`'s rule
"never pass the container as a dependency" still holds: the Factory is passed
only to command *constructors* in `cmd/` — the composition layer — never into
the functional core. A command pulls concrete dependencies from it inside
`RunE` and hands plain values to `core`. The core never imports the Factory.

→ See also: `samber/cc-skills-golang@golang-dependency-injection` (manual
injection, lazy init, interfaces-where-consumed) and
`samber/cc-skills-golang@golang-design-patterns`

### Cached Lazy Initialization

Without caching, calling `cfg.Config()` twice in one command invocation
loads the config file twice. For expensive operations (disk reads, network
calls), add `sync.Once` inside the factory function:

```go
var (
    cachedConfig    *config.AppConfig
    cachedConfigErr error
    configOnce      sync.Once
)

cfg.Config = func() (*config.AppConfig, error) {
    configOnce.Do(func() {
        cachedConfig, cachedConfigErr = config.Load()
    })
    return cachedConfig, cachedConfigErr
}
```

This is necessary when multiple dependencies chain through `cfg.Config()`:
`cfg.Client` calls `cfg.Config()` and so does a hypothetical `cfg.Validator`.
Without caching, each call independently re-reads the config file.

## The Functional Core (`internal/core/`)

The core package contains all pure computation:

- **Value objects** with always-valid construction (`BranchName`, `ProjectPath`)
- **Types with behavior** — methods that encode business rules
- **Validation pipelines** — composable, fail-fast or collect-all
- **Business rule functions** — pure functions over domain types
- **Error types** — semantically meaningful, not exit-code-aware

### Value Objects

```go
// internal/core/branch_name.go

type BranchName struct {
    value string
}

func NewBranchName(raw string) (BranchName, error) {
    if err := validateBranchName(raw); err != nil {
        return BranchName{}, err
    }
    return BranchName{value: raw}, nil
}

func (b BranchName) String() string { return b.value }
```

**When to use value objects**: when you validate or normalize the same
value in multiple places. A `BranchName` that enforces git naming rules
prevents duplicate validation in every command that takes a branch name.

**When NOT to use value objects**: for simple pass-through strings, ints,
durations that have no invariants beyond their Go type. Don't wrap
`time.Duration` in a value object.

### Types with Behavior

```go
// internal/core/project.go

type Project struct {
    name     string
    path     string
    worktrees []Worktree
}

func (p *Project) ActiveWorktrees() []Worktree {
    return lo.Filter(p.worktrees, func(w Worktree, _ int) bool {
        return !w.IsArchived()
    })
}

func (p *Project) CanPrune() bool {
    return len(p.ActiveWorktrees()) > 1
}

func (p *Project) FindWorktree(name string) *Worktree {
    wt, found := lo.Find(p.worktrees, func(w Worktree) bool {
        return w.Name() == name
    })
    if !found { return nil }
    return &wt
}
```

These methods replace what the previous architecture called "aggregate
behavior" and "domain services." They're just methods on types that
encode business rules. No DDD terminology needed.

### Error Types

Errors are defined in the core package with semantic meaning. They carry
enough information for both programmatic handling and human-readable
formatting. They do NOT carry exit codes — that mapping happens at the
CLI boundary. `ValidationError` is the representative shape:

```go
// internal/core/errors.go

type ValidationError struct {
    Field       string
    Value       string
    Message     string
    Suggestions []string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("invalid %s %q: %s", e.Field, e.Value, e.Message)
}
```

The full core hierarchy — `NotFoundError`, `OperationError`, and `UsageError` (the one
type that maps to exit code 2) — lives in [05-errors.md](./references/05-errors.md).

Errors that originate from external I/O are defined near their origin,
not in the core:

```go
// internal/git/errors.go

type ExternalError struct {
    Tool      string
    Operation string
    Message   string
    Cause     error
    Kind      ErrorKind // NotFound, Timeout, Permission
}
```

### Business Rules as Functions

When a rule spans multiple types or needs config, write a standalone
function in the core:

```go
// internal/core/rules.go

func CanDelete(project *Project, worktree *Worktree, cfg *Config) error {
    if len(project.ActiveWorktrees()) <= 1 {
        return &OperationError{
            Op:      "delete",
            Message: "cannot delete the last active worktree",
        }
    }
    if cfg.IsProtected(worktree.Branch()) {
        return &ValidationError{
            Field:   "branch",
            Value:   worktree.Branch(),
            Message: "branch is protected",
        }
    }
    return nil
}
```

### Validation Pipeline

A generic, composable validation mechanism:

```go
// internal/core/validation.go

type Validator[T any] func(T) error

type Pipeline[T any] struct {
    validators []Validator[T]
}

func NewPipeline[T any](vs ...Validator[T]) *Pipeline[T] {
    return &Pipeline[T]{validators: vs}
}

// Validate stops on first error (value objects, user input)
func (p *Pipeline[T]) Validate(value T) error {
    for _, v := range p.validators {
        if err := v(value); err != nil {
            return err
        }
    }
    return nil
}

// ValidateAll collects all errors (config, complex DTOs)
func (p *Pipeline[T]) ValidateAll(value T) error {
    var errs []error
    for _, v := range p.validators {
        if err := v(value); err != nil {
            errs = append(errs, err)
        }
    }
    return errors.Join(errs...)
}
```

### Core Dependency Policy

The core package may import:
- Go standard library (excluding I/O packages: no `os`, `net`, `exec`)
- `samber/lo` — collection utilities (zero deps, stdlib only)

No other dependencies. No I/O. No filesystem. No network. No exec.
Enforce with `depguard` in CI.

**Why `samber/lo` but not `samber/mo`**: `lo` provides concrete utility
(`Filter`, `Map`, `Contains`, `Find`) that replaces verbose loops. It's
widely adopted (45k+ stars), zero-dependency, and the overhead is ~4%
vs hand-written loops. `mo` (Option, Result monads) is non-idiomatic Go
— use `*T` for optional values and `(T, error)` for fallible operations.
The stdlib is catching up via `iter.Seq` and the `slices`/`maps`
packages; prefer stdlib where equivalent, use `lo` for what stdlib
doesn't cover.

→ See also: `samber/cc-skills-golang@golang-samber-lo`

## Error Handling

Error types (`ValidationError`, `NotFoundError`, `OperationError`, `UsageError`)
are defined in `internal/core/` — see [The Functional Core](#the-functional-core-internalcore-)
above. The CLI boundary handles mapping them to exit codes and formatted output.

This skill owns the CLI-specific concerns (error→exit-code mapping, TTY-aware
formatting). For general Go error craft — `%w` wrapping, `errors.Is`/`errors.As`
(`errors.AsType[T]` on Go 1.26+), sentinel vs typed errors, and `samber/oops`
for richer production errors — defer to `samber/cc-skills-golang@golang-error-handling`.
Two of its rules are load-bearing here:

- **Single-handling rule** — an error is either logged **or** returned, never
  both. Commands return errors up to `main.go`; only the boundary formats and
  logs them once. Logging inside `runFoo` *and* returning duplicates the output.
- **`main.go` is the only place that calls `os.Exit`.** Every layer returns
  `error`; the entry point maps it to an exit code.

See [05-errors.md](./references/05-errors.md) for the full type hierarchy and
formatting patterns.

### Exit Code Mapping

Exit codes are a CLI concern. They live in `cmd/` or a `cmdutil` package:

```go
// internal/cmdutil/exit.go

type ExitCode int

const (
    ExitOK    ExitCode = 0
    ExitError ExitCode = 1
    ExitUsage ExitCode = 2  // matches Bash convention
)

func ExitCodeFor(err error) ExitCode {
    if err == nil {
        return ExitOK
    }
    // Usage errors get code 2 (Bash convention)
    var usageErr *core.UsageError
    if errors.As(err, &usageErr) {
        return ExitUsage
    }
    // Everything else is 1
    return ExitError
}
```

**Why only three exit codes (0, 1, 2)**: Most CLI consumers check
`$? -ne 0` and read stderr. Custom codes 3-6 are non-standard, collide
with Bash/sysexits conventions, and in practice nobody writes
`if [ $? -eq 5 ]`. The error *types* carry semantic meaning for
formatting; the exit *code* is a blunt success/failure signal.

If you genuinely need granular exit codes (e.g., your CLI is consumed by
automation that dispatches on exit code), adopt the `sysexits.h` ranges
(64-78) to avoid collisions with Bash (2, 126, 127) and signals (128+N).

### Error Formatting

Error formatting lives in the `output` package. It uses `errors.As` to
dispatch on error type:

```go
// internal/output/errors.go

func FormatError(io *iostreams.IOStreams, err error) {
    var valErr *core.ValidationError
    var notFound *core.NotFoundError
    var extErr *git.ExternalError
    
    switch {
    case errors.As(err, &valErr):
        formatValidation(io, valErr)
    case errors.As(err, &notFound):
        formatNotFound(io, notFound)
    case errors.As(err, &extErr):
        formatExternal(io, extErr)
    default:
        fmt.Fprintf(io.ErrOut, "%s %s\n", io.Styles().Error("Error:"), err)
    }
}
```

Suggestions are attached to error types (not derived from a kind→tool
matrix). The error creator knows what's actionable.

## Output and I/O Streams

### The IOStreams Abstraction

All terminal I/O goes through an `IOStreams` struct that provides
stdout, stderr, TTY detection, and styling:

```go
// internal/iostreams/iostreams.go

type IOStreams struct {
    In     io.ReadCloser
    Out    io.Writer
    ErrOut io.Writer
    
    // Set from PersistentPreRunE via -v / --verbose flag
    Verbose bool
    // Set from Factory's lazy Logger; nil disables debug logging
    Logger  *slog.Logger
    
    colorEnabled bool
    isTTY        bool
    styles       *Styles
}

func System() *IOStreams {
    stdout := os.Stdout
    isTTY := term.IsTerminal(int(stdout.Fd()))
    colorEnabled := isTTY && os.Getenv("NO_COLOR") == ""
    
    return &IOStreams{
        In:           os.Stdin,
        Out:          stdout,
        ErrOut:       os.Stderr,
        isTTY:        isTTY,
        colorEnabled: colorEnabled,
        styles:       NewStyles(colorEnabled),
    }
}

func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
    in := &bytes.Buffer{}
    out := &bytes.Buffer{}
    errOut := &bytes.Buffer{}
    return &IOStreams{
        In: io.NopCloser(in), Out: out, ErrOut: errOut,
        styles: NewStyles(false),
    }, in, out, errOut
}

func (s *IOStreams) IsStdoutTTY() bool  { return s.isTTY }
func (s *IOStreams) Styles() *Styles    { return s.styles }
func (s *IOStreams) IsInteractive() bool {
    return s.isTTY && term.IsTerminal(int(os.Stdin.Fd()))
}

// Verbosef writes a dim-styled line to stderr when --verbose is set.
func (s *IOStreams) Verbosef(format string, args ...any) {
    if !s.Verbose {
        return
    }
    fmt.Fprintln(s.ErrOut, s.styles.Dim(fmt.Sprintf(format, args...)))
}
```

### Stdout vs Stderr Discipline

- **Stdout**: data output only. What gets piped. Results, JSON, tables.
- **Stderr**: everything else. Errors, progress, verbose output, prompts.

This is critical for `myapp list | jq '.'` — progress messages must not
contaminate the JSON stream.

### Styles

Use `lipgloss` centrally. Define styles once, reference everywhere:

```go
// internal/iostreams/styles.go

type Styles struct {
    Error   func(string) string
    Success func(string) string
    Warning func(string) string
    Dim     func(string) string
    Bold    func(string) string
}

func NewStyles(color bool) *Styles {
    if !color {
        identity := func(s string) string { return s }
        return &Styles{identity, identity, identity, identity, identity}
    }
    return &Styles{
        Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render,
        Success: lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render,
        Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render,
        Dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render,
        Bold:    lipgloss.NewStyle().Bold(true).Render,
    }
}
```

### Multi-Format Output (--output flag)

Query commands support `--output json|table|plain` (the ecosystem-standard
flag; see `samber/cc-skills-golang@golang-cli`). Use a shared `Formatter`
interface selected by the flag value, rather than per-command format switches:

```go
// internal/output/formatter.go

type Formatter interface {
    Write(w io.Writer, data any) error
}

// NewFormatter maps the --output value to a Formatter. Empty string means
// "no structured format requested" — the command falls through to its
// human-readable default.
func NewFormatter(format string) Formatter {
    switch format {
    case "json":
        return &JSONFormatter{}
    case "table":
        return &TableFormatter{}
    default: // "plain"
        return &PlainFormatter{}
    }
}

type JSONFormatter struct{}

func (e *JSONFormatter) Write(w io.Writer, data any) error {
    return json.NewEncoder(w).Encode(data)
}
```

Commands that support structured output check `if opts.Format != ""` before
falling through to human-readable default formatting. The `Formatter` takes an
`io.Writer` (not `IOStreams`) so it stays decoupled from terminal concerns —
pass `opts.IO.Out`. See [04-output.md](./references/04-output.md) for the full
formatter set and JSON-Lines vs JSON-array guidance.

### Verbose Output

Two distinct channels for observability:

| Channel   | Purpose              | Audience  | Destination | Control                    |
|-----------|----------------------|-----------|-------------|----------------------------|
| Debug log | Internal diagnostics | Developer | Stderr      | `--debug` / `MYAPP_DEBUG`  |
| Verbose   | Progress disclosure  | User      | Stderr      | `-v` / `--verbose`         |

```go
// Verbose output goes through IOStreams.Verbosef (defined on the struct):
opts.IO.Verbosef("cloning %s…", opts.Source)
```

Debug logging uses `slog` via `opts.IO.Logger`; verbose output uses
`opts.IO.Verbosef`. Both are co-located on the IOStreams struct — not
separate services.

## Configuration

### Loading Pipeline

Config uses `koanf` with TOML format, following a standard merge order
(last wins):

1. **Defaults** — hardcoded in `DefaultConfig()`
2. **Config file** — `$XDG_CONFIG_HOME/myapp/config.toml`
3. **Environment variables** — `MYAPP_*` prefix
4. **Flags** — CLI flags override everything (handled by Cobra)

```go
// internal/config/config.go

type AppConfig struct {
    ProjectsDir   string        `toml:"projects_dir" koanf:"projects_dir"`
    DefaultBranch string        `toml:"default_branch" koanf:"default_branch"`
    Timeout       time.Duration `toml:"timeout" koanf:"timeout"`
    Protected     []string      `toml:"protected" koanf:"protected"`
}

func (c *AppConfig) IsProtected(name string) bool {
    return lo.Contains(c.Protected, name)
}

func DefaultConfig() *AppConfig {
    return &AppConfig{
        ProjectsDir:   filepath.Join(homeDir(), "Projects"),
        DefaultBranch: "main",
        Timeout:       30 * time.Second,
        Protected:     []string{"main", "master"},
    }
}
```

**Config lives in its own package** (`internal/config/`), not in the
core. Config has serialization tags and depends on `koanf` — that's an
I/O concern. The core can accept config values as function parameters
or receive a minimal interface if it needs behavioral access:

```go
// internal/core/rules.go — accepts what it needs, not the whole config
func CanDelete(project *Project, wt *Worktree, protected []string) error {
    if lo.Contains(protected, wt.Branch()) { ... }
}
```

Missing config file falls back to defaults — this is not an error.
Config validation uses `ValidateAll()` (collect-all mode) so users see
all problems at once.

## Interactive Prompts

### Architecture: Command Decides, Prompt Co-located

The command checks whether prompting is needed. The prompt executes
in the same command function. There is no separate "interactive layer."

```go
func runDelete(opts *DeleteOptions) error {
    // ... resolve project, find worktree ...
    
    if !opts.Force && opts.IO.IsInteractive() {
        confirmed, err := promptConfirm(opts.IO,
            fmt.Sprintf("Delete worktree %s?", name))
        if err != nil { return err }
        if !confirmed { return nil }
    }
    
    // ... proceed with deletion ...
}
```

### The `scriptable` Rule

**Every interactive command is _scriptable_:** flags bypass every prompt, and
prompting is gated on an interactive TTY. `--force` or explicit arguments must
complete the command with no human interaction; in a pipe or CI (non-TTY), a
missing required input is an error, never a silent block. Test both paths. This is
the canonical definition — references invoke _scriptable_ rather than restating it.

### Prompt Utilities

Shared prompt helpers live in `internal/output/prompt.go` using `huh`:

```go
func promptConfirm(io *iostreams.IOStreams, message string) (bool, error) {
    var confirmed bool
    form := huh.NewForm(
        huh.NewGroup(huh.NewConfirm().Title(message).Value(&confirmed)),
    ).WithOutput(io.ErrOut) // prompts go to stderr
    return confirmed, form.Run()
}
```

## Concurrency

Most CLI commands are sequential. Add concurrency only when sequential
I/O latency is user-perceptible (>500ms for multiple independent calls).

**Where concurrency decisions live**: in the command's run function or
in adapter packages — not in the core.

**Default primitive**: `errgroup` from `golang.org/x/sync`. Scoped
goroutine group with shared context and first-error cancellation.

```go
func runStatus(opts *StatusOptions) error {
    g, ctx := errgroup.WithContext(opts.Ctx)
    
    var info *core.RepoInfo
    var items []core.Worktree
    
    g.Go(func() error {
        var err error
        info, err = client.GetRepoInfo(ctx, path)
        return err
    })
    g.Go(func() error {
        var err error
        items, err = client.ListWorktrees(ctx, path)
        return err
    })
    
    if err := g.Wait(); err != nil {
        return err
    }
    
    project := core.NewProject(info, items)
    // ... format output ...
}
```

**The core never uses concurrency.** It's pure computation. Concurrency
is an I/O orchestration concern — it lives in the shell.

For bounded fan-out over N items, call `errgroup.SetLimit(n)` before the `g.Go` loop —
it is the built-in worker pool and replaces hand-rolled pools. One primitive (`errgroup`)
thus covers both the simple concurrent case above and bounded fan-out; no second
concurrency library. See [06-concurrency.md](./references/06-concurrency.md) for the
`SetLimit` fan-out example plus worker-pool, pipeline, and semaphore variants (and
`samber/cc-skills-golang@golang-concurrency` for the general patterns).

## Project Structure

This overrides `samber/cc-skills-golang@golang-project-layout`'s rule that `main`
lives in `cmd/<name>/main.go` and that `cmd/` holds only `main.go`. For a CLI,
`main.go` is the module-root composition root and `cmd/` holds the command
definitions (the `cobra` convention). Rationale: a CLI ships one binary, so a
nested `cmd/<name>/` buys nothing, and keeping the command tree under `cmd/` puts
every command in one obvious place.

The two layouts below are concrete waypoints on the _Tier 1→3_ ladder;
[01-architecture.md](./references/01-architecture.md) owns the full ladder (Tier 1
minimal → Tier 3 large-scale) with per-tier growth triggers.

### Tier 2 — Standard CLI (5–15 commands, single concern)

```text
myapp/
  main.go                    # Composition root
  cmd/
    root.go                  # Root command, global flags
    create.go                # One file per command
    delete.go
    list.go
  internal/
    core/                    # Pure logic: types, validation, rules
      types.go
      errors.go
      validation.go
    config/                  # Config loading
      config.go
    git/                     # Git operations (or api/, fs/, etc.)
      client.go
      errors.go
    output/                  # Shared formatting, error display
      errors.go
      table.go
      styles.go
    iostreams/               # I/O abstraction, TTY detection
      iostreams.go
      styles.go
    cmdutil/                 # Factory, exit codes, shared cmd helpers
      factory.go
      exit.go
```

### Tier 2→3 — Command groups (15+ commands, multiple concerns)

```text
myapp/
  main.go
  cmd/
    root.go
    worktree/                # Command group as sub-directory
      create.go
      delete.go
      list.go
      prune.go
    config/
      init.go
      validate.go
    version.go
  internal/
    core/
      types.go
      errors.go
      validation.go
      rules.go
    config/
      config.go
    git/
      client.go
      reader.go
      writer.go
      errors.go
    output/
      errors.go
      exporter.go
      table.go
      prompt.go
    iostreams/
      iostreams.go
      styles.go
    cmdutil/
      factory.go
      exit.go
      json_flags.go
  tests/
    e2e/
    integration/
    fixtures/
```

### Package Dependency Rules

```text
cmd/ ───────────► internal/output/
  │                    │
  │                    ▼
  ├───────────► internal/iostreams/
  │
  ├───────────► internal/core/ ◄────── (no outbound deps)
  │                    ▲
  ├───────────► internal/git/ ─────► internal/core/
  │
  └───────────► internal/config/
```

- `cmd/` imports everything (it's the composition point)
- `internal/core/` imports nothing (stdlib + `lo` only)
- `internal/git/` imports `internal/core/` (returns core types)
- `internal/output/` imports `internal/core/` (formats core errors)
- `internal/config/` is independent (or imports `internal/core/` for
  shared types)

**Enforcing the rules** with `depguard`:

```yaml
linters-settings:
  depguard:
    rules:
      core-isolation:
        files:
          - "**/core/**"
        allow:
          - $gostd
          - github.com/samber/lo
        deny:
          - pkg: "github.com/*"
            desc: "core allows only stdlib and lo"
```

→ See also: `samber/cc-skills-golang@golang-lint`

### When to Extract a Package

Extract when you have:
- **5+ types or functions** sharing a clear concern
- **An I/O boundary** you want to test independently
- **A distinct external dependency** (a specific tool, API, or library)

Do NOT extract for:
- Fewer than 3 related things
- Matching a diagram before you have the code
- "It might grow later"

## Testing Strategy

### Testing Pyramid

```text
 ╱ E2E (few)          — full binary, testscript or gexec
╱  Integration (some)  — real I/O, component boundaries
╱   Command (many)      — runF injection, captured output
╱    Core (most)         — table-driven, no mocks, pure values
```

### Core Tests: Table-Driven, No Mocks

The core has no I/O — tests are pure functions with values in, values
out:

```go
func TestNewBranchName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "feature/login", false},
        {"empty", "", true},
        {"reserved", "HEAD", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := core.NewBranchName(tt.input)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Command Tests: runF Injection

Test the run function directly by passing the parsed options:

```go
func TestRunCreate(t *testing.T) {
    ios, _, stdout, _ := iostreams.Test()
    
    f := &cmdutil.Factory{IOStreams: ios}
    
    var gotOpts *CreateOptions
    cmd := NewCmdCreate(f, func(opts *CreateOptions) error {
        gotOpts = opts
        return nil
    })
    
    cmd.SetArgs([]string{"feature/test", "--source", "develop"})
    require.NoError(t, cmd.Execute())
    
    assert.Equal(t, "feature/test", gotOpts.Name)
    assert.Equal(t, "develop", gotOpts.Source)
}
```

To exercise the full `runCreate` (not just parsing), pass an Options struct with
test-double dependencies (a fake `Client` func) and assert on the returned error with
`require.ErrorAs`.

### Integration and E2E Tiers

The upper two tiers are owned by [07-testing.md](./references/07-testing.md): integration
tests of I/O adapters against real tools (`//go:build integration`, `t.TempDir()`), and
E2E tests that build and drive the real binary (`//go:build e2e`, asserting
stdout/stderr/exit code). The pyramid above is the CLI-specific spine; 07 has the
per-tier examples.

### Mock Strategy

**Use `moq` for generated mocks** when interfaces have 5+ methods.
`moq` generates structs with function fields — state-based, not
expectation-based. No runtime reflection.

**Use hand-written test doubles** for small interfaces (<5 methods) or
when you need stateful behavior.

**Never mock the core.** It has no I/O — test it directly with values.

The four tiers above are the CLI-specific shape. The underlying Go testing
idioms are owned by `samber/cc-skills-golang@golang-testing`: named subtests in
every table, `//go:build integration` tags on the Integration tier, `t.Parallel()`
for independent cases, `goleak` when a command spawns goroutines, and testify
(`samber/cc-skills-golang@golang-stretchr-testify`) as assertion helpers. Golden
files are a *technique* within the Command and E2E tiers (capture stdout/stderr,
diff against `testdata/*.golden`, refresh with `-update`), not a separate tier.
See [07-testing.md](./references/07-testing.md) for the full pyramid.

## Context and Signal Handling

```go
// main.go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    root := cmd.NewRootCommand(cfg)
    root.SetContext(ctx)
    // ...
}
```

All I/O operations receive `ctx` and respect cancellation. The core
never takes a context parameter (it's pure computation, it doesn't
block on I/O). On a signal, cancel in-flight work, flush, and exit **130**
for SIGINT (the `128+N` convention from `samber/cc-skills-golang@golang-cli`);
see [06-concurrency.md](./references/06-concurrency.md) for the shutdown flow.

## Panic Recovery

Top-level recovery in `main.go`:

```go
func recoverPanic() {
    if r := recover(); r != nil {
        fmt.Fprintf(os.Stderr, "Fatal: %v\n", r)
        if os.Getenv("MYAPP_DEBUG") != "" {
            fmt.Fprintf(os.Stderr, "\n%s\n", debug.Stack())
        }
        os.Exit(1)
    }
}
```

## Decision Triggers

Instead of "stages" that couple unrelated decisions, use independent
triggers. Each decision is made when its specific condition is met:

| When you see...                                | Consider...                                    |
|------------------------------------------------|------------------------------------------------|
| 3+ output formats for the same data            | Shared `Formatter` interface + `--output` flag |
| 3+ commands sharing the same I/O adapter       | Extract adapter to its own package             |
| Validation logic duplicated across commands     | Extract value object to core                   |
| 5+ commands in a group                         | Sub-directory under `cmd/`                     |
| Sequential I/O >500ms with independent calls   | `errgroup` for parallel execution              |
| N>10 items processed with independent I/O each | `errgroup` with `SetLimit(n)` for bounded fan-out |
| Config validation showing one error at a time  | Switch to `ValidateAll()` collect mode         |
| Test setup duplicated across 3+ test files     | Extract test helpers to `tests/helpers/`        |
| A business rule spans 2+ types + config        | Extract function to `core/rules.go`            |
| Factory has 10+ fields with conditional init   | Evaluate if some should be per-command          |
| Commands need dynamic shell completions        | Completion functions in cmd, calling adapters   |
| Destructive operation affects multiple items   | Plan/confirm/execute pattern                   |

## Common Libraries

[14-libraries.md](./references/14-libraries.md) is the source of truth for library
selection — comparison matrices per concern (CLI framework, config, logging,
prompts, TUI, output, concurrency, testing, completion, storage, distribution).
The CLI-specific defaults: `cobra`, `koanf` (TOML), `samber/lo`, `lipgloss`, `huh`,
`log/slog`, `errgroup`, `testify`, `matryer/moq`, `testscript`, `golangci-lint`,
`goreleaser`.

**Explicitly not recommended:**
- `samber/mo` — non-idiomatic Go; use `*T` and `(T, error)`
- `samber/do` — unnecessary for CLI-scale DI; manual wiring scales
- `testify/mock` — expectation-based; use `moq` or hand-written doubles
- `viper` — heavier than needed; `koanf` is sufficient

## Reference Documents

Load a reference when the user's task focuses on that area. Do not pre-load all references — fetch on demand.

| Reference | Load when... |
|-----------|-------------|
| [01-architecture.md](./references/01-architecture.md) | Choosing tier (Tier 1/2/3), designing project layout, or explaining the Parse/Execute/Respond foundation in depth |
| [02-command-ux.md](./references/02-command-ux.md) | Designing command names, help text, destructive-operation UX, or color/formatting rules |
| [03-input-config.md](./references/03-input-config.md) | Configuration loading pipeline, koanf setup, env var precedence, XDG config paths |
| [04-output.md](./references/04-output.md) | Multi-format output (`--output json\|table\|plain`), table rendering, structured output design |
| [05-errors.md](./references/05-errors.md) | Error type hierarchy, error formatting with suggestions, wrapping strategies |
| [06-concurrency.md](./references/06-concurrency.md) | `errgroup` vs `conc`, fan-out patterns, context cancellation under concurrency |
| [07-testing.md](./references/07-testing.md) | Testing pyramid detail, `runF` injection, integration test setup, E2E with `testscript` |
| [08-completion.md](./references/08-completion.md) | Shell completion (bash/zsh/fish/PowerShell), dynamic completion functions, Carapace |
| [09-logging.md](./references/09-logging.md) | `slog` setup, debug vs verbose channels, log levels, structured logging patterns |
| [10-state.md](./references/10-state.md) | Local state persistence (XDG dirs, state files, locking) |
| [11-plugins.md](./references/11-plugins.md) | Plugin architecture, extension points, subprocess-based plugin design |
| [12-tui.md](./references/12-tui.md) | Bubbletea TUI, when to use TUI vs plain output, model/view/update pattern |
| [13-versioning.md](./references/13-versioning.md) | Version embedding (`ldflags`), version command, goreleaser setup |
| [14-libraries.md](./references/14-libraries.md) | Full library comparison table, alternatives, and selection criteria |

## Related Skills

The full cross-reference map — which ecosystem skill owns what, and where this
skill takes precedence — is the **Ecosystem positioning** note at the top of this
file. One skill it doesn't name there:

- `samber/cc-skills-golang@golang-context` — `context.Context` propagation, deadlines, and cancellation across API boundaries
