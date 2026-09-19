# 14 — Library Reference

Consolidated recommendation tables. **Bold** entries are the recommended default for new projects. Alternatives are listed with the context where they'd be preferred.

`samber/cc-skills-golang@golang-popular-libraries` is the ecosystem source of truth for library selection; this table is the CLI-focused subset, tuned to this skill's opinions (koanf over viper, errgroup over any second concurrency lib).

---

## CLI Frameworks

| Library | Philosophy | Recommendation |
|---------|-----------|----------------|
| **`spf13/cobra`** | Full-featured, hierarchical commands | **Default for multi-command CLIs** (Tier 2+) |
| **`peterbourgon/ff`** | Flag-first config, wraps `flag.FlagSet` | **Default for single-purpose tools** (Tier 1) |
| `urfave/cli` | Simpler, less opinionated | Alternative if Cobra feels heavy |
| `alecthomas/kong` | Struct-tag-driven | Minimal boilerplate, built-in config |
| Standard `flag` | Minimal, idiomatic | Viable for the simplest cases |

---

## Configuration

| Library | Use Case | Notes |
|---------|----------|-------|
| **`knadh/koanf`** | **General config (flags + env + files)** | Modular, preserves key casing, explicit merge |
| `spf13/viper` | Cobra integration out of the box | Key-casing issues, binary bloat |
| `peterbourgon/ff` | Flag-first config | Lightweight, Tier 1 |
| `kelseyhightower/envconfig` | Env-var-only | Minimal, struct tags |
| `caarlos0/env` | Env var parsing into structs | Zero-dependency |
| `adrg/xdg` | XDG path resolution | Cross-platform |

---

## Logging

| Library | CLI Fit | Notes |
|---------|---------|-------|
| **`log/slog`** (stdlib) | **Recommended default** | Zero deps, structured, `LevelVar` for dynamic switching |
| `charmbracelet/log` | Good for Charm ecosystem | Pretty terminal output, Lip Gloss integration |
| `rs/zerolog` | Overkill for most CLIs | Zero-allocation JSON. Only if logging is high-volume. |
| `uber-go/zap` | Overkill for CLIs | Designed for servers |
| `sirupsen/logrus` | Legacy | Prefer slog for new projects |

---

## Interactive Prompts

| Library | Status | Notes |
|---------|--------|-------|
| **`charmbracelet/huh`** | **Recommended** | Forms + prompts. Standalone or Bubble Tea. Generics. Accessible mode. |
| `AlecAivazis/survey` | **Archived** | Author recommends Bubble Tea |
| `manifoldco/promptui` | Low activity | Polished templates, spinners |
| `c-bata/go-prompt` | Active | Tab completion, dynamic suggestions, REPL |
| `charmbracelet/gum` | Active | Interactive prompts as standalone binaries for shell scripts |

---

## TUI Frameworks

| Library | Architecture | Notes |
|---------|-------------|-------|
| **`charmbracelet/bubbletea`** | Elm Architecture | **Default for TUI apps.** 18k+ apps built with it. |
| `charmbracelet/bubbles` | Pre-built Bubble Tea components | Text input, lists, spinners, viewports, tables |
| `rivo/tview` | Widget-based | Closer to traditional GUI toolkit. Less composable. |
| `jroimartin/gocui` | Older TUI framework | Simpler but less maintained |

---

## Output Formatting

| Library | Purpose | Notes |
|---------|---------|-------|
| **`charmbracelet/lipgloss`** | Terminal styling | **CSS-like API.** 9,187+ importers. |
| `charmbracelet/glamour` | Markdown rendering | Terminal markdown |
| `muesli/termenv` | Terminal color detection | Lower-level than Lip Gloss |
| `fatih/color` | Simple color output | Widely used, minimal |
| **`olekukonko/tablewriter`** | Tables | ASCII/Unicode/Markdown/HTML. Use v1.1.x. |
| `jedib0t/go-pretty` | Tables, lists, progress | All-in-one formatting |
| `cli/go-gh/pkg/tableprinter` | Tables (gh-style) | Column-formatted to TTY, TSV to pipes. Auto-fits. |
| `text/tabwriter` (stdlib) | Basic tab-aligned text | Zero dependencies |

---

## Progress and Spinners

| Library | Notes |
|---------|-------|
| **`schollz/progressbar`** | Thread-safe. Auto-converts to spinner for unknown length. Reader/Writer wrappers. |
| `vbauerster/mpb` | Multiple concurrent progress bars. Good for parallel downloads. |
| `cheggaaa/pb` | Older, stable. Simple API. |
| `briandowns/spinner` | 90+ configurable spinner types. Pure spinner. |
| `charmbracelet/bubbles` spinner | Spinner as Bubble Tea component |

---

## Concurrency

| Library | Use Case | Notes |
|---------|----------|-------|
| **`golang.org/x/sync/errgroup`** | **Bounded concurrent I/O** | `WithContext` cancels siblings on first error; `SetLimit(n)` is a built-in worker pool. One primitive covers simple concurrency and bounded fan-out — no second library. |
| `golang.org/x/sync/singleflight` | Deduplicate concurrent calls | Cache-stampede prevention |
| stdlib `sync` | Mutex, WaitGroup, Once, atomics | See `samber/cc-skills-golang@golang-concurrency` |

Do **not** add `sourcegraph/conc` — `errgroup.SetLimit` covers the CLI fan-out case the ecosystem targets.

## Testing

| Library | Purpose | Notes |
|---------|---------|-------|
| **`stretchr/testify`** | Assertions | Practically universal |
| **`google/go-cmp`** | Deep equality comparison | Better than `reflect.DeepEqual` |
| `gotest.tools/v3/golden` | Golden file tests | With `-update` flag |
| `sebdah/goldie` | Golden files + templates | Dynamic values in golden files |
| `jarcoal/httpmock` | HTTP mocking | Request/response mocking |
| `spf13/afero` | In-memory filesystem | Testing file operations |
| `testcontainers-go` | Real Docker containers | Database/service integration tests |
| **`matryer/moq`** | Mock generation | Function-field mocks, no reflection. Use for interfaces with 5+ methods; hand-write smaller doubles. |
| **`rogpeppe/go-internal/testscript`** | CLI E2E | Declarative `.txtar` script tests driving the real binary |

---

## Shell Completion

| Library | Notes |
|---------|-------|
| Cobra built-in | bash/zsh/fish/powershell. Sufficient for most CLIs. |
| **`carapace-sh/carapace`** | **Recommended for advanced needs.** Multi-part values, caching, 8+ shells, bridge system. |
| `posener/complete` | Standalone, used by HashiCorp tools |

---

## Storage

| Library | Use Case | Notes |
|---------|----------|-------|
| **`etcd-io/bbolt`** | Embedded key/value | Pure Go, ACID, crash-safe. Good default for local state. |
| **`modernc.org/sqlite`** | Embedded relational DB | Pure Go, no CGO. Better for cross-compilation. |
| `mattn/go-sqlite3` | SQLite with CGO | More mature but requires CGO |
| `zalando/go-keyring` | OS keyring access | Cross-platform credentials |
| `gofrs/flock` | File locking | Cross-platform advisory locks |

---

## Dependency Injection

| Tool | Approach | CLI Fit |
|------|----------|---------|
| **Manual wiring** | Constructor injection | **Default. Almost always sufficient.** |
| `google/wire` | Compile-time code generation | Large apps with complex dependency graphs |
| `uber-go/dig` | Runtime reflection | More flexible, less idiomatic |
| `uber-go/fx` | Application framework on Dig | Lifecycle management. Probably overkill for CLIs. |

---

## Distribution

| Tool | Purpose |
|------|---------|
| **`goreleaser`** | Build, package, publish Go binaries | **Standard for Go CLI distribution** |
| `ko` | Go container images without Docker |
| `cosign` | Binary/container signing |
| `creativeprojects/go-selfupdate` | GitHub-based self-update |

---

## Plugin Systems

| Library | Model | Notes |
|---------|-------|-------|
| PATH discovery (custom) | Git-style `<prefix>-<name>` | Zero coupling, any language |
| **`hashicorp/go-plugin`** | gRPC/RPC | **Default for typed plugin interfaces** |
| `gopher-lua` | Embedded Lua | User scripting |
| `google/cel-go` | Expression evaluation | Policy/filter rules |

---

**See also:** `samber/cc-skills-golang@golang-samber-lo` — full API reference for the `lo` collection utilities used in the core · `samber/cc-skills-golang@golang-design-patterns` — idiomatic Go patterns for the libraries listed above · `samber/cc-skills-golang@golang-popular-libraries` — broader library recommendations across all Go project types
