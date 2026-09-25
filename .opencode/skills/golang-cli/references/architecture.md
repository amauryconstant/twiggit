# Architecture by Scale

Three tiers, each a waypoint on a continuous scale: take the simplest one that fits today and move up when a growth trigger fires. Every tier keeps the same Parse → Execute → Respond shape; the tiers differ only in how far those concerns are split into separate code units.

## Contents

- [Why concerns, not layers](#why-concerns-not-layers)
- [Tier 1: minimal CLI](#tier-1-minimal-cli)
- [Tier 2: standard CLI](#tier-2-standard-cli)
- [Tier 2 growth: command groups](#tier-2-growth-command-groups)
- [Tier 3: large-scale CLI](#tier-3-large-scale-cli)
- [When to extract a package](#when-to-extract-a-package)
- [When to add formal layers](#when-to-add-formal-layers)
- [Decision matrix and growth path](#decision-matrix-and-growth-path)

## Why concerns, not layers

Each boundary costs an interface, a constructor, a type mapping, and a place to look while debugging, so a CLI keeps exactly the two boundaries that pay:

- **Pure logic vs I/O** — the _Functional Core_ is testable without mocks.
- **Stable types vs external formats** — adapters return core types, so a tool's output format changes in one package.

Respond owns the whole output surface — stdout format, stderr messages, and the exit code — which keeps presentation out of the core.

## Tier 1: minimal CLI

**When:** one job or 2–3 subcommands, one domain, under ~500 LOC — the tool you might otherwise write in Bash.

```text
myapp/
├── main.go        # main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
├── app.go         # appEnv, fromArgs (Parse), run (Execute → Respond)
├── app_test.go    # table tests against run() with buffers
└── go.mod
```

[tier1_app.go](../assets/examples/tier1_app.go) shows the whole pattern:

- `run` returns the exit code (2 for flag errors, 1 for failures), so tests call it with buffers. The signal context is created inside `run`; `main` stays one line.
- `appEnv` holds parsed input and the output writers: no package-level variables, no `init()`.
- `process` is the core: typed values in, typed result out.

| Need | Pick |
| --- | --- |
| Flags only | `flag.NewFlagSet` + `appEnv` |
| Flags + env vars + config file | `peterbourgon/ff` + `appEnv` |
| 2–3 subcommands | `switch os.Args[1]` with one `flag.FlagSet` per subcommand |

**Move to Tier 2 when** you need a fourth subcommand, commands share config or dependencies, you want completions or persistent flags, or one file passes ~300 LOC.

## Tier 2: standard CLI

**When:** 5–15 commands, shared configuration, one or two domains — the typical developer tool or API client. Cobra, the Factory, and IOStreams arrive here (→ `commands.md`).

```text
myapp/
├── main.go                  # composition root: Factory, signals, exit code
├── cmd/
│   ├── root.go              # root command, global flags
│   ├── create.go            # one file per command (+ _test.go)
│   └── list.go
├── internal/
│   ├── core/                # Functional Core: types, rules, validation, errors
│   ├── config/              # config loading
│   ├── git/                 # an adapter (or api/, fs/)
│   ├── output/              # formatters, error display, prompts
│   ├── iostreams/           # terminal abstraction, styles
│   └── cmdutil/             # Factory, exit codes, Cobra helpers
└── testdata/                # golden files
```

**Layout choice.** `main.go` at the module root and commands in `cmd/` — this overrides `samber/cc-skills-golang@golang-project-layout`'s `cmd/<name>/main.go`. A CLI ships one binary, so a nested `cmd/<name>/` buys nothing, and a flat `cmd/` puts every command in one obvious place. When command packages must be unimportable by other modules, `internal/cli/` is an equivalent home; pick one and keep it.

**Move to Tier 3 when** you have 15+ commands across distinct domains, commands have very different dependencies, you need plugins, or several teams contribute.

## Tier 2 growth: command groups

Five or more commands under one noun become a sub-directory before a full Tier 3 split:

```text
cmd/
├── root.go
├── worktree/        # `myapp worktree create|delete|list|prune`
│   ├── worktree.go  # group command: Args: cmdutil.NoSubcommand, RunE: Help
│   ├── create.go
│   └── list.go
└── version.go
```

## Tier 3: large-scale CLI

**When:** 20+ commands, several domains, plugins, several teams — `gh`, `kubectl`, `terraform`. Tier 3 adds vertical slices, a lazy `IOStreams` field on the Factory, and an exit-code registry (→ `errors.md`).

```text
myapp/
├── main.go
├── cmd/
│   ├── root.go
│   ├── deploy/
│   │   ├── deploy.go          # group command
│   │   ├── create/create.go   # `myapp deploy create` (+ create_test.go)
│   │   └── rollback/rollback.go
│   └── auth/{login,logout,status}/
├── internal/
│   ├── cmdutil/               # Factory, exit registry, shared helpers; cmdtest/ for test helpers
│   ├── config/  auth/  output/  iostreams/
│   └── httpmock/              # test utilities
└── api/                       # API client (internal/api if not exported)
```

**Vertical slices**: each command package holds its options, run function, and tests; adding `myapp deploy rollback` adds one directory instead of touching `handlers/`, `services/`, and `models/`. Shared concerns live in `internal/` packages the slices import.

## When to extract a package

Extract when you have **5+ types or functions** sharing a concern, **an I/O boundary** you want to test on its own, or **a distinct external dependency** (a tool, an API, a library). Below that, keep the code beside its only caller — a package for two functions adds indirection without a boundary.

## When to add formal layers

Ports-and-adapters or clean-architecture boundaries pay off when the CLI is also a library imported by other modules, when whole subsystems swap (local ↔ cloud storage), or when the domain is complex enough to need enforced boundaries. Otherwise commands → Factory → core → formatters already separate the concerns.

## Decision matrix and growth path

| Situation | Tier | Key pattern |
| --- | --- | --- |
| Script replacement | 1 | `appEnv` + `flag` |
| Single tool needing config files | 1 | `ff` + `appEnv` |
| Developer tool, API client, 5–15 commands | 2 | Cobra + Factory + IOStreams |
| Wrapper around one external tool | 1–2 | Thin core, one adapter |
| Tool that is also a library | 2–3 | Public API in `pkg/`, formal layers |
| Platform CLI, 20+ commands, plugins, teams | 3 | Vertical slices + exit registry |

| Concern | Tier 1 | Tier 2 | Tier 3 |
| --- | --- | --- | --- |
| Layout | `main.go` + `app.go` | `main.go` + `cmd/` | `cmd/<group>/<verb>/` |
| Dependencies | `appEnv` fields | Factory (lazy) | Factory + lazy IOStreams |
| Exit codes | `run()` returns int | `ExitCodeFor` | Registry |
| Output | inline | `Formatter` + `--output` | + TTY-adaptive tables |
| Flags | `flag` / `ff` | Cobra | Cobra + hook middleware |
| Tests | `run()` table tests | `runF` + golden files | + httpmock + E2E |

Each step keeps what came before: `appEnv` fields become Factory and IOStreams fields, inline formatting becomes a `Formatter`, and Tier 2 command files move into per-command packages.
