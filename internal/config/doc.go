// Package config owns twiggit configuration loading: koanf wiring, TOML
// file resolution under XDG_CONFIG_HOME, environment variable expansion,
// and the merge order (defaults -> file -> env).
//
// DefaultConfig returns the hardcoded fallback. Load() reads from disk and
// overlays env on top. Errors are surfaced as *core.OperationError so the
// CLI boundary can dispatch them through output.FormatError uniformly.
package config
