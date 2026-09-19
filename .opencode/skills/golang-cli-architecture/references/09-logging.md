# 09 — Logging

---

## Two Different Intentions

These are often conflated but serve distinct purposes:

**Verbose (`-v` / `--verbose`)** — User-facing operational detail. "Tell me more about what you're doing." Written to stderr in human-readable prose. Examples: which files are being processed, which API endpoints are being called, timing information. The user expects to read this.

**Debug (`--debug` / `MYAPP_DEBUG=1`)** — Developer-facing diagnostic detail. "Help me understand what's going wrong internally." Structured logging (key=value or JSON), directed at developers or support engineers. Examples: full HTTP request/response bodies, internal state transitions, config resolution traces. The user redirects this to a file or pastes it into a bug report.

---

## Implementation

### Verbose: Simple Boolean Gate

```go
// Canonical definition in SKILL.md §Output and I/O Streams
func (s *IOStreams) Verbosef(format string, args ...any) {
    if !s.Verbose {
        return
    }
    fmt.Fprintln(s.ErrOut, s.styles.Dim(fmt.Sprintf(format, args...)))
}
```

Verbose is prose, not structured. It goes to stderr so it doesn't pollute pipeable stdout.

### Debug: slog with Dynamic Levels

`log/slog` (stdlib, Go 1.21+) is the recommended default:

```go
func newLogger(debug bool) *slog.Logger {
    level := slog.LevelWarn
    if debug {
        level = slog.LevelDebug
    }
    return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: level,
    }))
}
```

For runtime level switching, use `slog.LevelVar`:

```go
var programLevel = new(slog.LevelVar) // default: INFO

func setupLogging(debug bool) {
    if debug {
        programLevel.Set(slog.LevelDebug)
    }
    handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: programLevel,
    })
    slog.SetDefault(slog.New(handler))
}
```

---

## The slog Level Spectrum

| Level | Value | CLI Usage |
|-------|-------|-----------|
| DEBUG | -4 | Internal diagnostics, config resolution, HTTP traces |
| INFO | 0 | Operational messages (default when verbose) |
| WARN | 4 | Recoverable issues, deprecation notices |
| ERROR | 8 | Failures that affect output |

Custom levels are defined between these. A useful CLI-specific level is TRACE (`slog.Level(-8)`) for extremely granular output.

---

## Handler Choice

| Context | Handler | Why |
|---------|---------|-----|
| Human at terminal | `slog.NewTextHandler` | key=value pairs, readable |
| Piped to file / CI | `slog.NewJSONHandler` | Machine-parseable |
| Pretty development output | `charmbracelet/log` | Colored, styled with Lip Gloss |

Auto-detect: use `TextHandler` when stderr is a TTY, `JSONHandler` when it's not. Or provide `--log-format=text|json`.

---

## Contextual Logging

Add persistent context for multi-phase operations:

```go
logger := slog.Default().With("command", cmd.Name())
logger.Debug("starting execution", "args", args)

for _, item := range items {
    itemLogger := logger.With("item", item.ID)
    itemLogger.Debug("processing")
}
```

---

## Environment Variable Override

Support `MYAPP_DEBUG=1` as an env var override for `--debug`. This is valuable for debugging completion scripts, init sequences, and other contexts where flags aren't practical:

```go
debug = debug || os.Getenv("MYAPP_DEBUG") != ""
```

---

## Integration with IOStreams

The full pattern: flags configure IOStreams, which carries both verbose state and the logger:

```go
// Root command PersistentPreRunE
rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    streams.Verbose = verbose
    streams.Logger = newLogger(debug || os.Getenv("MYAPP_DEBUG") != "")
    return nil
}
```

Commands call `streams.Verbosef()` for user-facing detail and `streams.Logger.Debug()` for developer diagnostics.

---

## Library Comparison

Recommended default: `log/slog` (stdlib). See
[14-libraries.md](./14-libraries.md#logging) for the full comparison (`charmbracelet/log`,
`rs/zerolog`, `uber-go/zap`, `sirupsen/logrus`).

---

**See also:** `samber/cc-skills-golang@golang-observability` — Prometheus metrics, OpenTelemetry tracing, Pyroscope profiling, and Grafana dashboards for production services · `samber/cc-skills-golang@golang-samber-slog` — `slog-multi` handler pipelines, sampling, formatters, and backend routing (Datadog, Loki, Sentry)
