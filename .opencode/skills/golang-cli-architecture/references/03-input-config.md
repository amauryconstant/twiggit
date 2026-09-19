# 03 — Input and Configuration

---

## The Precedence Hierarchy

The standard config precedence for CLI tools, from highest to lowest priority:

1. **Flags** — explicit user intent for this invocation
2. **Environment variables** — session or deployment context
3. **Config file** — persistent user preferences
4. **Defaults** — sensible zero-configuration behavior

This hierarchy enables 12-factor compliance and supports diverse contexts (local dev, containers, CI).

---

## Framework Options

### Cobra + Viper (Incumbent)

Designed to work together. Viper handles the full precedence chain out of the box.

**Known problems:**
- Force-lowercases all keys internally, breaking JSON/YAML/TOML specs
- Hard-codes all parsers in the core — pulls heavy dependencies even if you only read JSON
- Binary bloat: a Viper "hello world" reading JSON produces a binary ~3x larger than koanf equivalent
- `Get()` returns references to internal slices/maps — external mutations can corrupt config state

**When it's still fine:** Small-to-medium CLIs where binary size doesn't matter and you want Cobra integration out of the box. If a project commits to viper, defer to `samber/cc-skills-golang@golang-spf13-viper` for `BindPFlag`, `SetEnvKeyReplacer`, `mapstructure` tags, and test isolation — this skill recommends koanf but does not re-teach viper.

### Koanf (Recommended Default)

Modular architecture — core is tiny, providers and parsers are separate modules.

**Key advantages over Viper:**
- Preserves key casing
- Explicit merge semantics — you control order by calling `k.Load()` multiple times
- Providers: file, env, posflag, confmap, rawbytes, structs, S3, Consul, Vault, etc.
- Parsers: JSON, YAML, TOML, HCL — each installed separately
- File watching via provider-level `Watch()`

**Caveat:** Not goroutine-safe during `Load()`. Needs a mutex if concurrent reads happen during load.

### ff (Lightweight Alternative)

`ff` wraps `flag.FlagSet` with env var and config file support. No subcommands, no dependency tree. Best for Tier 1 CLIs:

```go
fs := flag.NewFlagSet("myapp", flag.ContinueOnError)
listen := fs.String("listen", ":8080", "listen address")
debug := fs.Bool("debug", false, "debug mode")

ff.Parse(fs, os.Args[1:],
    ff.WithEnvVarPrefix("MYAPP"),
    ff.WithConfigFileFlag("config"),
    ff.WithConfigFileParser(ff.PlainParser),
)
```

### Other Options

For env-var-only config (`kelseyhightower/envconfig`, `caarlos0/env`) or
struct-tag parsers with built-in config (`kong`), see
[14-libraries.md](./14-libraries.md#configuration).

---

## Config File Format

| Format | Best For |
|--------|----------|
| YAML | Human-edited config. Familiar to most developers. |
| TOML | Simpler spec, less ambiguity than YAML. Good for tools. |
| JSON | Machine-generated config. Not ideal for human editing (no comments). |

Support env var overrides regardless of file format.

---

## XDG Compliance

Config files SHOULD live under `$XDG_CONFIG_HOME/<app>/` (default
`~/.config/<app>/`), resolved with `adrg/xdg` for cross-platform paths.
[10-state.md](./10-state.md#what-clis-persist) owns the full XDG directory table
(config, data, cache, runtime) and what each holds.

---

## Config Schema Validation

The primary validation mechanism is the core `Pipeline[T]` / `ValidateAll()`
(defined in `SKILL.md`) — run it in collect-all mode after the precedence merge
so users see every problem at once. It keeps validation rules in the functional
core, testable with plain values.

For simple field-level constraints, `go-playground/validator` struct tags are a
lighter alternative:

```go
type AppConfig struct {
    ProjectsDir string        `toml:"projects_dir" validate:"required,dir"`
    Timeout     time.Duration `toml:"timeout" validate:"gte=1s,lte=300s"`
    Format      string        `toml:"format" validate:"oneof=json table plain"`
}
```

Either way, validation is a separate concern from config loading — validate after the full precedence merge is complete.

---

## Environment Variable Conventions

- Prefix all env vars with your app name: `MYAPP_API_KEY`, `MYAPP_DEBUG`
- Use underscores for word separation
- Document supported env vars in `--help` output and README
- Support `MYAPP_DEBUG=1` as a universal debug toggle (see [09-logging.md](./09-logging.md))

