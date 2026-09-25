# Library Reference

**Bold** entries are the defaults for new projects; alternatives list the context where they fit better. This is the CLI subset of `samber/cc-skills-golang@golang-popular-libraries`, tuned to this skill's choices.

## Contents

- [Defaults at a glance](#defaults-at-a-glance)
- [CLI frameworks](#cli-frameworks) · [Configuration](#configuration) · [Logging](#logging) · [Interactive prompts](#interactive-prompts) · [TUI frameworks](#tui-frameworks)
- [Output formatting](#output-formatting) · [Progress and spinners](#progress-and-spinners) · [Concurrency](#concurrency) · [Testing](#testing)
- [Shell completion](#shell-completion) · [Storage](#storage) · [Dependency injection](#dependency-injection) · [Distribution](#distribution) · [Plugin systems](#plugin-systems)

## Defaults at a glance

`cobra`, `koanf` (TOML), `samber/lo`, `lipgloss`, `huh`, `log/slog`, `errgroup`, `testify`, `matryer/moq`, `testscript`, `golangci-lint`, `goreleaser`.

| Instead of | Use | Why |
| --- | --- | --- |
| `spf13/viper` (new projects) | `knadh/koanf` | Preserved key casing, parsers installed separately, explicit merge order |
| `samber/mo` | `*T` and `(T, error)` | The idiomatic Go forms for optional and fallible values |
| `samber/do`, `wire`, `fx` | Manual wiring with the Factory | CLI-scale graphs stay readable without a container |
| `testify/mock` | `moq` or hand-written doubles | State-based fakes survive refactors that break expectation lists |
| `sourcegraph/conc` | `errgroup` + `SetLimit` | One primitive covers parallel calls and bounded fan-out |

## CLI frameworks

| Library | Philosophy | Recommendation |
| --- | --- | --- |
| **`spf13/cobra`** | Full-featured, hierarchical commands | **Default for multi-command CLIs** (Tier 2+) |
| **`peterbourgon/ff`** | Flag-first config, wraps `flag.FlagSet` | **Default for single-purpose tools** (Tier 1) |
| `urfave/cli` | Simpler, less opinionated | Alternative if Cobra feels heavy |
| `alecthomas/kong` | Struct-tag-driven | Minimal boilerplate, built-in config |
| Standard `flag` | Minimal, idiomatic | Viable for the simplest cases |

## Configuration

| Library | Use Case | Notes |
| --- | --- | --- |
| **`knadh/koanf`** | **General config (flags + env + files)** | Modular, preserves key casing, explicit merge |
| `spf13/viper` | Cobra integration out of the box | Key-casing issues, binary bloat |
| `peterbourgon/ff` | Flag-first config | Lightweight, Tier 1 |
| `kelseyhightower/envconfig` | Env-var-only | Minimal, struct tags |
| `caarlos0/env` | Env var parsing into structs | Zero-dependency |
| `adrg/xdg` | XDG path resolution | Cross-platform |

## Logging

| Library | CLI Fit | Notes |
| --- | --- | --- |
| **`log/slog`** (stdlib) | **Recommended default** | Zero deps, structured, `LevelVar` for dynamic switching |
| `charmbracelet/log` | Good for Charm ecosystem | Pretty terminal output, Lip Gloss integration |
| `rs/zerolog` | Overkill for most CLIs | Zero-allocation JSON. Only if logging is high-volume. |
| `uber-go/zap` | Overkill for CLIs | Designed for servers |
| `sirupsen/logrus` | Legacy | Prefer slog for new projects |

## Interactive prompts

| Library | Status | Notes |
| --- | --- | --- |
| **`charmbracelet/huh`** | **Recommended** | Forms + prompts. Standalone or Bubble Tea. Generics. Accessible mode. |
| `AlecAivazis/survey` | **Archived** | Author recommends Bubble Tea |
| `manifoldco/promptui` | Low activity | Polished templates, spinners |
| `c-bata/go-prompt` | Active | Tab completion, dynamic suggestions, REPL |
| `charmbracelet/gum` | Active | Interactive prompts as standalone binaries for shell scripts |

## TUI frameworks

| Library | Architecture | Notes |
| --- | --- | --- |
| **`charmbracelet/bubbletea`** | Elm Architecture | **Default for TUI apps.** |
| `charmbracelet/bubbles` | Pre-built Bubble Tea components | Text input, lists, spinners, viewports, tables |
| `rivo/tview` | Widget-based | Closer to traditional GUI toolkit. Less composable. |
| `jroimartin/gocui` | Older TUI framework | Simpler but less maintained |

## Output formatting

| Library | Purpose | Notes |
| --- | --- | --- |
| **`charmbracelet/lipgloss`** | Terminal styling | **CSS-like API.** |
| `charmbracelet/glamour` | Markdown rendering | Terminal markdown |
| `muesli/termenv` | Terminal color detection | Lower-level than Lip Gloss |
| `fatih/color` | Simple color output | Widely used, minimal |
| **`olekukonko/tablewriter`** | Tables | ASCII/Unicode/Markdown/HTML. Use v1.1.x. |
| `jedib0t/go-pretty` | Tables, lists, progress | All-in-one formatting |
| `cli/go-gh/pkg/tableprinter` | Tables (gh-style) | Column-formatted to TTY, TSV to pipes. Auto-fits. |
| `text/tabwriter` (stdlib) | Basic tab-aligned text | Zero dependencies |

## Progress and spinners

| Library | Notes |
| --- | --- |
| **`schollz/progressbar`** | Thread-safe. Auto-converts to spinner for unknown length. Reader/Writer wrappers. |
| `vbauerster/mpb` | Multiple concurrent progress bars. Good for parallel downloads. |
| `cheggaaa/pb` | Older, stable. Simple API. |
| `briandowns/spinner` | 90+ configurable spinner types. Pure spinner. |
| `charmbracelet/bubbles` spinner | Spinner as Bubble Tea component |

## Concurrency

| Library | Use Case | Notes |
| --- | --- | --- |
| **`golang.org/x/sync/errgroup`** | **Bounded concurrent I/O** | `WithContext` cancels siblings on first error; `SetLimit(n)` is a built-in worker pool. One primitive covers simple concurrency and bounded fan-out — no second library. |
| `golang.org/x/sync/singleflight` | Deduplicate concurrent calls | Cache-stampede prevention |
| stdlib `sync` | Mutex, WaitGroup, Once, atomics | See `samber/cc-skills-golang@golang-concurrency` |

## Testing

| Library | Purpose | Notes |
| --- | --- | --- |
| **`stretchr/testify`** | Assertions | Practically universal |
| **`google/go-cmp`** | Deep equality comparison | Better than `reflect.DeepEqual` |
| `gotest.tools/v3/golden` | Golden file tests | With `-update` flag |
| `sebdah/goldie` | Golden files + templates | Dynamic values in golden files |
| `jarcoal/httpmock` | HTTP mocking | Request/response mocking |
| `spf13/afero` | In-memory filesystem | Testing file operations |
| `testcontainers-go` | Real Docker containers | Database/service integration tests |
| **`matryer/moq`** | Mock generation | Function-field mocks, no reflection. Use for interfaces with 5+ methods; hand-write smaller doubles. |
| **`rogpeppe/go-internal/testscript`** | CLI E2E | Declarative `.txtar` script tests driving the real binary |

## Shell completion

| Library | Notes |
| --- | --- |
| Cobra built-in | bash/zsh/fish/powershell. Sufficient for most CLIs. |
| **`carapace-sh/carapace`** | **Recommended for advanced needs.** Multi-part values, caching, 8+ shells, bridge system. |
| `posener/complete` | Standalone, used by HashiCorp tools |

## Storage

| Library | Use Case | Notes |
| --- | --- | --- |
| **`etcd-io/bbolt`** | Embedded key/value | Pure Go, ACID, crash-safe. Good default for local state. |
| **`modernc.org/sqlite`** | Embedded relational DB | Pure Go, no CGO. Better for cross-compilation. |
| `mattn/go-sqlite3` | SQLite with CGO | More mature but requires CGO |
| `zalando/go-keyring` | OS keyring access | Cross-platform credentials |
| `gofrs/flock` | File locking | Cross-platform advisory locks |

## Dependency injection

| Tool | Approach | CLI Fit |
| --- | --- | --- |
| **Manual wiring** | Constructor injection | **Default. Almost always sufficient.** |
| `google/wire` | Compile-time code generation | Large apps with complex dependency graphs |
| `uber-go/dig` | Runtime reflection | More flexible, less idiomatic |
| `uber-go/fx` | Application framework on Dig | Lifecycle management. Probably overkill for CLIs. |

## Distribution

| Tool | Purpose |
| --- | --- |
| **`goreleaser`** | Build, package, and publish Go binaries — the standard for Go CLI distribution |
| `ko` | Go container images without Docker |
| `cosign` | Binary/container signing |
| `creativeprojects/go-selfupdate` | GitHub-based self-update |

## Plugin systems

| Library | Model | Notes |
| --- | --- | --- |
| PATH discovery (custom) | Git-style `<prefix>-<name>` | Zero coupling, any language |
| **`hashicorp/go-plugin`** | gRPC/RPC | **Default for typed plugin interfaces** |
| `gopher-lua` | Embedded Lua | User scripting |
| `google/cel-go` | Expression evaluation | Policy/filter rules |

→ See `samber/cc-skills-golang@golang-samber-lo` for the `lo` API.
