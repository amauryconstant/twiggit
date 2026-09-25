# Input and Configuration

## Contents

- [Precedence](#precedence)
- [koanf loader](#koanf-loader)
- [Choosing a library](#choosing-a-library)
- [File format and location](#file-format-and-location)
- [Validation](#validation)
- [Environment variables](#environment-variables)

## Precedence

Highest wins:

1. **Flags** — explicit intent for this invocation
2. **Environment variables** — session or deployment context
3. **Config file** — persistent preferences
4. **Defaults** — zero-configuration behavior

## koanf loader

[config.go](../assets/examples/config.go) loads defaults → file → `MYAPP_*` env with koanf and returns an `AppConfig`:

- Config is its own package (`internal/config`) — it knows file formats and koanf, which are shell concerns. The core receives only the values it needs (`cfg.Protected`, not `*AppConfig`).
- A missing file falls back to defaults; any other read or parse error names the path.
- `MYAPP_DEFAULT_BRANCH` maps to `default_branch` by trimming the prefix and lowercasing.
- **Flags win at the command**: the command binds flags into its Options, and uses config only for values the user left unset (`cmd.Flags().Changed("timeout")`). When many flags mirror config keys, load them into koanf last with the `posflag` provider instead.
- The Factory wraps `config.Load` in `sync.OnceValues`, so the file is read once per invocation.

koanf is not goroutine-safe during `Load`; load once at startup, as the Factory does.

## Choosing a library

| Library | Pick when |
| --- | --- |
| **`knadh/koanf`** (default) | Flags + env + files; modular providers and parsers, preserved key casing, explicit merge order |
| `spf13/viper` | The project already uses it — then → See `samber/cc-skills-golang@golang-spf13-viper` for `BindPFlag`, `SetEnvKeyReplacer`, and test isolation |
| `peterbourgon/ff` | Tier 1: wraps `flag.FlagSet` with env vars and a config file |
| `caarlos0/env` | Env-var-only config into structs |

koanf over viper for new projects: viper lowercases every key (breaking case-sensitive formats), pulls every parser into the binary, and `Get` returns internal maps that callers can mutate.

Tier 1 with `ff`:

```go
fs := flag.NewFlagSet("myapp", flag.ContinueOnError)
listen := fs.String("listen", ":8080", "listen address")
err := ff.Parse(fs, os.Args[1:],
    ff.WithEnvVarPrefix("MYAPP"),
    ff.WithConfigFileFlag("config"),
    ff.WithConfigFileParser(ff.PlainParser),
)
```

## File format and location

| Format             | Best for                                            |
| ------------------ | --------------------------------------------------- |
| **TOML** (default) | Tool config: small spec, comments, little ambiguity |
| YAML               | Teams already standardized on it                    |
| JSON               | Machine-generated config (no comments)              |

The file lives at `$XDG_CONFIG_HOME/<app>/config.toml` (default `~/.config/<app>/`), resolved with `adrg/xdg`. The other XDG directories (data, cache, runtime) → `state.md`.

## Validation

Validate after the full merge, in collect-all mode, so users fix every problem in one pass. [config.go](../assets/examples/config.go) runs a `core.Pipeline` with `ValidateAll` and wraps the result with the file path.

For simple field constraints, `go-playground/validator` struct tags (`validate:"required,oneof=json table plain"`) are a lighter alternative.

## Environment variables

- Prefix every variable with the app name: `MYAPP_API_URL`, `MYAPP_DEBUG`.
- Separate words with underscores.
- List supported variables in `--help` and the README.
- Support `MYAPP_DEBUG=1` as a debug toggle (→ `logging.md`).
