## Context

The `init` command currently:
- Writes to a shell config file by default
- Requires `--dry-run` to preview output (but includes metadata lines)
- Has complex flag interactions: `[config-file]`, `--shell`, `--dry-run`, `--check`, `--force`

The service layer already has a pure `GenerateWrapper()` method that returns the wrapper string without file operations. The `SetupShell()` method orchestrates detection, generation, and file installation.

Current call flow:
```
cmd/init.go → SetupShell() → InstallWrapper() → os.WriteFile()
                  ↓
            GenerateWrapper() (pure, already exists)
```

## Goals / Non-Goals

**Goals:**
- Make `eval "$(twiggit init)"` the happy path (stdout by default)
- Simplify flag structure to `[shell]` + mode flags
- Preserve file installation capability via `--install`
- Minimize service layer changes (leverage existing `GenerateWrapper`)

**Non-Goals:**
- Changing the wrapper content itself
- Adding new shell types
- Modifying completion integration
- Removing `ValidateInstallation` from service layer (keep for backward compat)

## Decisions

### D1: Positional Argument Semantics

**Choice:** `[shell]` positional argument (bash|zsh|fish), auto-detect from `$SHELL` if omitted.

**Rationale:** Shell type is the primary discriminator. Config file path is secondary and only relevant with `--install`. This matches user mental model: "init zsh" vs "init bash".

**Alternatives considered:**
- Keep `--shell` flag: More verbose, inconsistent with typical CLI patterns
- Two positional args `[shell] [config]`: Ambiguous parsing (is "bash" a shell or a filename?)

### D2: Output Mode Selection

**Choice:** Default to stdout. Use `-i, --install` flag to enable file installation mode.

**Rationale:** Eval-based activation is the cleaner workflow. File installation is the "persistent" opt-in. This inverts the current default but aligns with modern shell tool conventions (direnv, starship).

**Alternatives considered:**
- `--stdout` flag: Makes eval awkward (`twiggit init --stdout`)
- New `shell-init` command: Fragments the init concept, more surface area

### D3: Config File Specification

**Choice:** `-c, --config` flag, only valid with `--install`. Error if used without `--install`.

**Rationale:** Config file is only relevant for installation. Making it a flag avoids positional ambiguity and makes the requirement explicit.

**Alternatives considered:**
- Positional when `--install` present: Parsing complexity, unclear precedence
- Auto-detect only: Removes ability to customize (existing feature regression)

### D4: Flag Removals

**Choice:** Remove `--dry-run`, `--shell`, `--check`. Keep `--force` (only with `--install`).

**Rationale:**
- `--dry-run`: Redundant when stdout is default
- `--shell`: Replaced by positional `[shell]`
- `--check`: No persistent state to check with eval model; users can `type twiggit` in shell

**Alternatives considered:**
- Keep `--check` for file-based users: Adds complexity for edge case; `--install` without `-f` already indicates if installed

### D5: Service Layer Changes

**Choice:** Minimal changes. Remove `DryRun` field from request/result types. Remove `DryRun` branch from `SetupShell()`. Keep `ValidateInstallation` for now.

**Rationale:** The cmd layer handles routing to `GenerateWrapper()` (stdout) or `SetupShell()` (--install). Service layer doesn't need to know about stdout mode.

**Alternatives considered:**
- Remove `ValidateInstallation` entirely: Could break external consumers; defer to later cleanup
- Add `StdoutMode` to service: Unnecessary; cmd layer handles this

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking change for existing users | Clear migration guide in release notes; `twiggit init --install` replaces old `twiggit init` |
| Users confused by new behavior | Help text shows both modes prominently; error messages guide to `--install` |
| `--check` removal frustrates users | Document `type twiggit` alternative in shell |
| Flag validation complexity | Validate flag combinations in cmd layer before calling service |

## Migration Plan

1. **Release Notes**: Document breaking change with migration examples
2. **Deprecation Period**: None (clean break - old flags error with helpful message)
3. **Help Text Update**: Examples show both stdout and --install modes

**Migration Examples:**
```
OLD                              NEW
twiggit init                     twiggit init --install
twiggit init ~/.bashrc           twiggit init bash --install --config ~/.bashrc
twiggit init --shell=zsh         twiggit init zsh --install
twiggit init --dry-run           twiggit init
twiggit init --check             (removed - use: type twiggit)
```

## Open Questions

None - design is complete.
