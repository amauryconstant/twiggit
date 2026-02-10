## Context

**Current State:**
The `setup-shell` command uses a flag-based API where users specify shell type via `--shell=bash|zsh|fish` and the system auto-detects the appropriate config file location (`~/.bashrc`, `~/.zshrc`, etc.). Shell wrapper validation uses partial markers (checking for `twiggit() {` and `# Twiggit <shell> wrapper`) which makes proper wrapper replacement difficult.

**Constraints:**

- Must maintain existing shell wrapper behavior (cd interception, builtin cd escape hatch, pass-through)
- Must support bash, zsh, and fish shells
- Must preserve --dry-run and --force functionality as optional flags
- Tests must achieve >80% coverage
- No code comments unless explicitly requested

**Problem:**
The current API couples command to shell types (requiring --shell flag), limits custom config file locations (must auto-detect), and uses partial validation markers that prevent clean wrapper replacement. The command name `setup-shell` is less intuitive than `init`.

## Goals / Non-Goals

**Goals:**

- Rename command from `setup-shell` to `init` for clearer intent
- Change API to positional argument (`init <config-file>`) for explicit target specification
- Infer shell type from config file path (`.bash*` → bash, `.zsh*` → zsh, contains "fish" → fish)
- Add explicit `--shell` flag as override for inference failures
- Add block delimiters (`### BEGIN/END TWIGGIT WRAPPER`) for proper wrapper removal/replacement
- Update install.sh to orchestrate detection, prompting, and `init` call
- Maintain backward-compatible wrapper behavior

**Non-Goals:**

- Automatic backward compatibility for old wrapper format (manual cleanup by user)
- Shell type detection from environment (install.sh handles this)
- Multi-shell wrapper installation (single config file per init call)
- Configuration file backup (beyond basic create-if-missing)

## Decisions

### Decision 1: Shell Type Inference from File Path

**Choice:** Infer shell type from config file name using pattern matching rather than explicit --shell flag requirement.

**Rationale:**

- Reduces command verbosity: `twiggit init ~/.bashrc` vs `twiggit setup-shell --shell=bash`
- Enables custom config file locations without manual shell specification
- Follows convention-over-configuration principle (filename implies shell type)
- Maintains escape hatch for non-standard names via --shell override

**Alternatives Considered:**

- Keep --shell flag as required: More explicit but less ergonomic
- Detect from $SHELL environment: Adds complexity to CLI, install.sh already handles this
- Require explicit shell always: Prevents custom config locations

### Decision 2: Block Delimiters for Wrapper Management

**Choice:** Add explicit `### BEGIN TWIGGIT WRAPPER` and `### END TWIGGIT WRAPPER` delimiters around wrapper content.

**Rationale:**

- Enables reliable wrapper removal before re-installation with --force
- Prevents duplicate/conflicting wrappers in config files
- Makes validation deterministic (check for both markers)
- Standard practice for shell script snippet management

**Alternatives Considered:**

- Keep partial markers: Simpler but unreliable for replacement
- Use unique UUID markers: Over-engineering for this use case
- Multiple wrapper support: Out of scope, increases complexity

### Decision 3: Install Shell Service Orchestrates Detection

**Choice:** Update install.sh to detect shell and config file, prompt user for confirmation, then call `twiggit init`.

**Rationale:**

- Separates concerns: CLI does one thing (install to specified file), script orchestrates
- install.sh already handles OS detection and binary installation
- Script can provide better interactive experience (confirm before overwrite)
- CLI remains testable without shell environment

**Alternatives Considered:**

- Auto-detect in CLI: Adds complexity to command, harder to test
- Remove install.sh wrapper installation: Loses interactive guidance
- Detect in both places: Duplicated logic, inconsistency risk

### Decision 4: Create Empty Config File If Missing

**Choice:** When config file doesn't exist, create it as empty file before appending wrapper.

**Rationale:**

- Shell configs are sourced (executed), so empty file is valid
- Simpler than adding shebang (not needed for sourced files)
- Avoids assumptions about file format or existing content
- Fails gracefully if directory is not writable

**Alternatives Considered:**

- Return error requiring manual file creation: Poor UX
- Add shebang (#!/bin/bash): Wrong for sourced files
- Add comment header: Unnecessary, adds complexity

### Decision 5: Rename Command to "init"

**Choice:** Rename `setup-shell` to `init` for clarity and brevity.

**Rationale:**

- "init" is a standard Unix convention (git init, npm init, etc.)
- "setup-shell" is verbose and shell-specific (future might support other integrations)
- Shorter command: `twiggit init ~/.bashrc` vs `twiggit setup-shell --shell=bash`

**Alternatives Considered:**

- Keep "setup-shell": More explicit but less idiomatic
- Use "shell-init": Still verbose, redundant with file path
- Use "install-shell": Longer than init, same meaning

## Risks / Trade-offs

### Risk 1: Shell Type Inference Fails for Custom Paths

[Risk] Users with non-standard config file names (e.g., `config.txt`) cannot infer shell type, requiring --shell flag.

**Mitigation:**

- Provide clear error message: "cannot infer shell type from path: config.txt (use --shell to specify)"
- Document --shell flag usage in help text
- Auto-detect in install.sh handles standard paths, reducing manual specification needs

### Risk 2: Old Wrapper Format Causes Conflicts

[Risk] Existing installations with old wrapper format (without block delimiters) will fail validation and require manual cleanup.

**Mitigation:**

- Accept as intentional breakage (proposal states backward compatibility is user's responsibility)
- Validation error will be clear: "wrapper already installed" with suggestion to use --force
- Documentation can include migration note if needed

### Risk 3: Install Script Failure Doesn't Block Installation

[Risk] If `twiggit init` fails during install.sh, script continues and binary may be installed without shell wrapper.

**Mitigation:**

- Show warning but don't fail (install.sh is convenience, not essential)
- Display remediation steps: "You can run manually: twiggit init <config-file>"
- Binary installation is primary goal, wrapper is optional feature

### Risk 4: Force Flag Overwrites Without Backup

[Risk] `--force` flag removes old wrapper block without backup, potentially losing user modifications.

**Mitigation:**

- Wrapper blocks are machine-generated, not user-editable content
- Standard practice is to remove old before install (atomic replacement)
- Users who manually modified wrapper should back up themselves

### Trade-off: API Simplicity vs Explicitness

**Trade-off:** Positional API is simpler but less explicit than flag-based API about shell type.

**Acceptance:**

- Inference covers 95%+ of use cases (standard config files)
- --shell override provides explicitness for edge cases
- Follows Unix philosophy: "Do one thing well" (install to file, don't detect shell)
