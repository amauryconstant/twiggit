# Logging: Verbose vs Debug

## Contents

- [Two channels](#two-channels)
- [Wiring](#wiring)
- [slog levels and handlers](#slog-levels-and-handlers)
- [Contextual logging](#contextual-logging)

## Two channels

| Channel | Flag | Audience | Form | Destination |
| --- | --- | --- | --- | --- |
| Verbose | `-v` / `--verbose` | The user: "tell me what you are doing" | Prose lines | stderr |
| Debug | `--debug` / `MYAPP_DEBUG=1` | A developer or bug report | Structured `slog` records | stderr |

Verbose shows which files, endpoints, and timings are involved. Debug shows internals: request and response bodies, state transitions, config resolution. Keep them separate so users get readable progress and developers get structured data.

## Wiring

Both live on `IOStreams` ([iostreams.go](../assets/examples/iostreams.go)); the root's `PersistentPreRunE` sets them from flags ([root.go](../assets/examples/root.go)):

```go
f.IOStreams.Verbose = verbose
if debug || os.Getenv("MYAPP_DEBUG") != "" {
    f.IOStreams.Logger = slog.New(slog.NewTextHandler(f.IOStreams.ErrOut,
        &slog.HandlerOptions{Level: slog.LevelDebug}))
}
```

Commands call `opts.IO.Verbosef("cloning %s", src)` and `opts.IO.Logger.Debug("resolved config", "path", p)`. Without `--debug` the logger discards everything, so debug calls cost nothing to leave in place.

`MYAPP_DEBUG` matters where flags cannot reach: completion scripts, init sequences, and wrappers.

## slog levels and handlers

| Level      | CLI use                                              |
| ---------- | ---------------------------------------------------- |
| DEBUG (-4) | Internal diagnostics, HTTP traces, config resolution |
| INFO (0)   | Operational messages                                 |
| WARN (4)   | Recoverable issues, deprecation notices              |
| ERROR (8)  | Failures that affect the result                      |

A TRACE level is `slog.Level(-8)`. For level changes at runtime, pass a `*slog.LevelVar` as `HandlerOptions.Level`.

| Context                       | Handler                           |
| ----------------------------- | --------------------------------- |
| stderr is a terminal          | `slog.NewTextHandler` (key=value) |
| stderr is redirected or in CI | `slog.NewJSONHandler`             |
| Styled development output     | `charmbracelet/log`               |

Choose by stderr TTY state, or offer `--log-format=text|json`.

## Contextual logging

```go
logger := opts.IO.Logger.With("command", cmd.Name())
for _, item := range items {
    logger.Debug("processing", "item", item.ID)
}
```

→ See `samber/cc-skills-golang@golang-samber-slog` for handler pipelines and routing, and `samber/cc-skills-golang@golang-observability` for metrics and tracing in long-running services.
