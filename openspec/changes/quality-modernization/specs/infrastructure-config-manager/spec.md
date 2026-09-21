# Spec Delta

## ADDED Requirements

### Requirement: Path confinement for config writes

When writing the user config file under `$HOME/.config/twiggit/`
or `$XDG_CONFIG_HOME/twiggit/`, the system SHALL confine writes
to that directory via the Go 1.24+ `os.Root` type. Any write
attempt that resolves to a path outside the configured root
SHALL return a `domain.ConfigError` (exit code 3) naming the
rejected path.

#### Scenario: Symlink traversal attempt is rejected
- **WHEN** the user has placed a symlink at
  `$HOME/.config/twiggit/config.toml` that points outside
  `$HOME/.config/twiggit/`
- **THEN** the write is rejected and the user sees a config
  error naming the symlink target

### Requirement: ProtectedBranches slice is defensively cloned

When the configuration loader populates the
`Config.Validation.ProtectedBranches` slice, the loader SHALL
clone the slice via `slices.Clone` so that downstream service
mutations (via `append` or index assignment) cannot leak back to
the config. Services SHALL iterate over the cloned slice only.

#### Scenario: In-place append does not mutate the config
- **WHEN** a service iterates over `ProtectedBranches` and
  appends a branch name to a per-iteration working slice
- **THEN** the config's `ProtectedBranches` is unchanged
  after the loop
