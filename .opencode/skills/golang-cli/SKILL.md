---
name: golang-cli
description: "Golang CLI architecture and conventions — command tree layout, flags and argument validation, config layering (koanf or viper), stdout/stderr discipline and --output formats, error-to-exit-code mapping, signal handling, version embedding, shell completion, interactive prompts, and CLI testing. Use when building, extending, or reviewing a Go command-line tool, deciding where command logic lives, or when the codebase imports spf13/cobra, spf13/viper, urfave/cli, peterbourgon/ff, or knadh/koanf. For Cobra API details → See `samber/cc-skills-golang@golang-spf13-cobra` skill; for Viper specifics → See `samber/cc-skills-golang@golang-spf13-viper` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: amaury
  version: "2.1.0"
  openclaw:
    emoji: "⌨️"
    homepage: https://github.com/amaurybrisou/go-cli-skill
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go CLI architect. You build tools that feel native to the Unix shell, and you keep the structure _proportional_: the simplest layout the tool needs today, grown when a trigger fires.

**Modes:**

- **Build** — pick the tier from [Tiers](#tiers) (ask when unclear) and build only its parts. **Done when** `go build ./... && go vet ./... && go test ./...` pass, the core import rule holds, only `main.go` calls `os.Exit`, and every interactive command is _scriptable_ with both paths tested.
- **Extend** — read the command tree and Factory, match the tier, check [Decision Triggers](#decision-triggers); refactor an unpatterned command into [The Command Pattern](#the-command-pattern) before testing it. **Done when** the new command has a `runF` test, plus a non-TTY test if it prompts.
- **Review** — **Done when** every [Common Mistakes](#common-mistakes) row is checked and each finding cites `file:line`; for a large CLI, fan out one sub-agent per area (streams, errors, core, prompts, tests).

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-cli` skill takes precedence.

# Go CLI Architecture

Every command does _Parse_ → _Execute_ → _Respond_. The _Functional Core_ (types, validation, rules) does no I/O and gets exhaustive table tests; the _shell_ (commands, adapters, output) does the I/O and gets a few integration tests.

## Principles

1. **The _Functional Core_ imports only stdlib and `samber/lo`** — all I/O lives in the _shell_, so core tests are values in, values out.
2. **stdout carries data; stderr carries errors, progress, prompts, and logs, all through `IOStreams`** — `myapp list -o json | jq .` works, and tests swap in non-TTY buffers.
3. **Commands return errors; `main.go` formats each once — the message, then the `Suggestions` of typed errors in the chain — and exits** — defers run and nothing prints twice.
4. **Exit 0 success, 1 failure, 2 `UsageError` (bad flag, argument, missing input), 130 interrupted** — [exit.go](./assets/examples/exit.go). When automation dispatches on codes, add sysexits 64–78 for runtime categories only; usage stays 2.
5. **Every interactive command is _scriptable_**: a flag bypasses each prompt, prompts need an interactive TTY, and missing input without one is a `UsageError` — so CI never hangs.
6. **The Factory stays in `cmd/`** — `RunE` pulls dependencies and passes plain values to the core; interfaces live where they are consumed.
7. **Context flows from `main` into every I/O call**; the core takes none.

## References

- [architecture](./references/architecture.md) — Tier, layout, package extraction
- [commands](./references/commands.md) — Commands, Factory, IOStreams, root, entry point, hooks
- [functional-core](./references/functional-core.md) — Value objects, rules, validation, core imports
- [errors](./references/errors.md) — Error types, exit codes, formatting, adapters
- [output](./references/output.md) — `--output`, TTY, stdin, progress, color
- [config](./references/config.md) — koanf/viper/ff, precedence, env, files
- [command-ux](./references/command-ux.md) — Naming, help, destructive ops, deprecation
- [prompts-and-tui](./references/prompts-and-tui.md) — Prompts, wizards, Bubble Tea
- [concurrency](./references/concurrency.md) — Signals, shutdown, `errgroup`
- [testing](./references/testing.md) — Test pyramid, helpers, golden files, E2E
- [logging](./references/logging.md) — `--debug` vs `--verbose`, `slog`
- [completion](./references/completion.md) — Shell completion
- [versioning](./references/versioning.md) — Version embedding, releases
- [state](./references/state.md) — XDG paths, state, locks, tokens
- [plugins](./references/plugins.md) — Plugin models
- [libraries](./references/libraries.md) — Library choice

Compiling examples: [assets/examples/](./assets/examples/) (each header names its project path).

## Tiers

Start at the lowest tier that fits; move up when a [Decision Trigger](#decision-triggers) fires.

| Signal | Tier | Layout | Key pattern |
| --- | --- | --- | --- |
| One job, ≤3 subcommands, <500 LOC | 1 | `main.go` + `app.go` | `appEnv` struct + stdlib `flag` or `ff` — [tier1_app.go](./assets/examples/tier1_app.go) |
| 5–15 commands, shared config | 2 | `main.go` + `cmd/` + `internal/{core,config,output,iostreams,cmdutil,<adapter>}` | Cobra + Factory + IOStreams |
| 20+ commands, several domains or teams, plugins | 3 | `cmd/<group>/<verb>/` packages | Tier 2 + vertical slices + exit-code registry |

## The Command Pattern

Each command is an Options struct, a constructor `NewCmdX(f, runF)`, and a `runX(opts)` that does Parse → Execute → Respond:

```go
func NewCmdCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
    opts := &CreateOptions{}
    cmd := &cobra.Command{
        Use:  "create <name>",
        Args: cmdutil.UsageArgs(cobra.ExactArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.IO, opts.Ctx, opts.Name = f.IOStreams, cmd.Context(), args[0]
            if runF != nil {
                return runF(opts) // test seam: receives parsed options
            }
            return runCreate(opts) // validate in core → call adapter → write to opts.IO
        },
    }
    cmd.Flags().StringVarP(&opts.Source, "source", "s", "main", "branch to start from")
    return cmd
}
```

Every Cobra command uses this shape, even a one-flag one; tiers change the layout, not the command. Full version: [create.go](./assets/examples/create.go), tests: [create_test.go](./assets/examples/create_test.go), query with `--output`: [list.go](./assets/examples/list.go).

## Package Dependency Rules

```text
main.go ──► cmd/ ──► internal/output ──► internal/iostreams
                 ├─► internal/<adapter> (git, api, fs) ──► internal/core
                 ├─► internal/config ──► internal/core
                 └─► internal/core  (imports nothing outside stdlib + lo)
```

Adapters return core types; `output` formats core errors. Enforce the core rule with depguard — [functional-core](./references/functional-core.md#dependency-policy).

## Decision Triggers

| When you see… | Do… |
| --- | --- |
| 3+ output formats for the same data | Shared `Formatter` + `--output` flag |
| 3+ commands sharing one I/O adapter | Extract the adapter into its own package |
| Validation duplicated across commands | Extract a value object into the core |
| 5+ commands in a group | Sub-directory under `cmd/` |
| Independent I/O: >500ms sequential, or N>10 items | `errgroup`; `SetLimit(n)` for N items |
| Config errors reported one at a time | Collect-all validation (`ValidateAll`) |
| Test setup duplicated across 3+ files | `internal/cmdutil/cmdtest`: test streams, Factory, shared fakes |
| A business rule spans 2+ types plus config | A rule function in the core |
| Factory with 10+ fields and conditional init | Move rarely-used dependencies into their commands |
| Commands need dynamic completion | Completion funcs in `cmd/` calling adapters |
| A destructive operation touches many items | Plan → confirm → execute |

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| `fmt.Println` / `os.Stdout` in a command, or diagnostics on stdout | Untestable, corrupts pipes. Data to `opts.IO.Out`, the rest to `opts.IO.ErrOut` |
| `os.Exit` or `log.Fatal` inside `RunE` | Defers never run. Return the error; `main.go` exits |
| Logging an error and returning it | It prints twice. Return it; diagnostics go to `IOStreams.Logger` (`slog`, `--debug`), never `log` |
| Adapter returns raw tool output (`[]byte`, `string`) | Callers break when the format changes. Parse into core types in the adapter |
| Root without `SilenceUsage` / `SilenceErrors` | Usage dumps and duplicate messages. Set both |
| Flag, argument, or unknown-command errors exit 1 | Misuse looks like failure. Convert to `UsageError` ending `Run '<cmd> --help'` — [args.go](./assets/examples/args.go) |
| `ValidationError` mapped to exit 2 | A rejected value is a failure (1); 2 means a malformed invocation |
| `os`, `net`, `exec`, or the Factory imported in `internal/core` | The core needs mocks. Pass values in; enforce with depguard |
| Prompt without a TTY gate or a flag bypass | Hangs in CI. Make it _scriptable_ |
| Completion func that prompts, prints non-candidates, or offers files for values | Breaks Tab. Print only candidates, return `ShellCompDirectiveNoFileComp`; install via `eval "$(myapp completion zsh)"` |
| Query command without `--output` | Scripts parse changing text. Offer `json`, `table`, `plain` |
| Config file required | First run fails. A missing file means defaults |
| `-v` and `--debug` merged into one flag | Users get noise, developers lose structure. Verbose is prose, debug is `slog` |
| Logic inline in `RunE`, tested only through `cmd.Execute()` | Untestable without Cobra. Add `runF` and `runX(opts)`; test both — [create_test.go](./assets/examples/create_test.go) |
| Commands as package globals registered in `init()` | Tests share state. Build the tree in `NewRootCommand(f)` |
| Context not passed to I/O, or Ctrl-C exits 1 | Work outlives the interrupt. `signal.NotifyContext`, exit 130 — [main.go](./assets/examples/main.go) |

## Cross-References

- → See `samber/cc-skills-golang@golang-spf13-cobra`, `@golang-spf13-viper`, `@golang-error-handling`, `@golang-testing`, `@golang-stretchr-testify`, `@golang-concurrency`, and `@golang-context` for the general APIs and idioms.
- For CLI work this skill overrides `@golang-project-layout` (root `main.go`, `cmd/`), `@golang-dependency-injection` (lazy Factory), and `@golang-testing` (the CLI pyramid).
