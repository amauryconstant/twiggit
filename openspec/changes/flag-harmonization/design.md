## Context

Current state:
- `cmd/create.go` defines `--cd` as a String flag (zombie - defined but ignored)
- `cmd/delete.go` uses `--change-dir` with `-C` short form (boolean, works)
- `cmd/init.go` has `--force` without short form
- Shell wrapper templates only handle `cd` command, not `create` or `delete` with `-C` flag
- cmd/AGENTS.md documents flags that don't exist and omits implemented flags
- Documentation drift between `--help` output and AGENTS.md

Constraints:
- NO code comments allowed unless explicitly requested
- Tests written after implementation
- Table-driven test patterns with fixtures
- Must maintain backward compatibility for commands without `-C` flag
- Shell wrapper pattern matching must work across bash/zsh/fish

Assumptions:
- User expects consistent flag naming across similar functionality
- Shell wrapper is primary integration point for directory changes
- Path output to stdout is sufficient for wrapper integration (no JSON/machine-readable format)

## Goals / Non-Goals

**Goals:**
- Standardize `-C, --cd` flag across `create` and `delete` commands
- Implement missing `create -C` path output for shell wrapper
- Add `-f` short form for `--force` in `delete` and `init` commands
- Update shell wrapper templates to handle `create -C` and `delete -C` flags
- Align all documentation (AGENTS.md, Long descriptions) with implemented reality
- Maintain backward compatibility (commands work without `-C` flag)

**Non-Goals:**
- Changing spec-level behavior (only implementation refactoring)
- Adding new user-facing capabilities
- Modifying `cd` command behavior (already works)
- Changing any other command flags outside scope
- Implementing verbose logging system (separate parallel work)

## Decisions

1. **Flag naming: Standardize to `-C, --cd` for create/delete**

   **Choice:** Use `-C` as short form and `--cd` as long form
   
   **Rationale:** 
   - Consistent with `delete` command (already uses `-C`)
   - `-C` is common Unix convention for "change directory" (tar uses `-C` for `--directory`)
   - More discoverable than `--change-dir`
   - Aligns with user mental model (they already know `twiggit cd`)

   **Alternatives considered:**
   - Keep `--change-dir` everywhere: More explicit but inconsistent with `cd` command name
   - Use only `--cd` (no short form): Simpler but less discoverable
   - Use `-d` for short form: Conflicts with git's `-d` for detached HEAD

2. **Output format: Path only on separate line when `-C` flag set**

   **Choice:** Output only the absolute path to stdout, no other text

   **Rationale:**
   - Cleaner for shell wrapper parsing
   - Wrapper can check if output is non-empty and exit code is 0
   - Verbose messaging will be handled by separate parallel work
   - Consistent with existing `cd` command behavior

   **Alternatives considered:**
   - JSON output: Overkill for simple use case
   - Message + path: Requires wrapper to parse which line is the path
   - Environment variable: Adds complexity, harder to use in shell wrapper

3. **Delete navigation: Context-aware path resolution**

   **Choice:** When deleting with `-C` from worktree context, navigate to project root (resolve "main"). From project or outside git context, output nothing.

   **Rationale:**
   - Logical flow: delete worktree → return to project
   - No-op when already at project root (avoids unnecessary cd)
   - Nothing when outside git (no sensible navigation target)
   - Uses existing `NavigationService.ResolvePath` infrastructure

   **Alternatives considered:**
   - Always navigate to `currentCtx.Path`: Would stay in deleted worktree during deletion
   - Navigate to parent directory: Unpredictable, depends on directory structure
   - Navigate to home directory: Arbitrary choice, not context-aware

4. **Shell wrapper pattern: Explicit command handling with case statement**

   **Choice:** Use case/switch statements in bash/zsh/fish to explicitly check for `-C` flag in `create` and `delete` commands

   **Rationale:**
   - Predictable behavior (Option B from exploration)
   - Explicit about which commands support `-C` flag
   - Won't capture output unexpectedly from other commands
   - Pattern matches approach used for `cd` command

   **Alternatives considered:**
   - Check for `-C` in all commands: Could capture output from unintended commands
   - Subprocess all commands: Overkill, performance impact
   - Require user to opt-in to wrapper capture: Burdens user

5. **Init flag ordering: Alphabetical by long flag name**

    **Choice:** Order init command flags: `--check` → `--dry-run` → `-f, --force` → `--shell`

    **Rationale:**
    - Consistent with common CLI flag conventions (man pages list alphabetically by long name)
    - Easier to find flags in help output
    - Matches user expectation from other tools

    **Alternatives considered:**
    - Keep existing order (based on addition date): Arbitrary, hard to navigate
    - Logical grouping (related flags together): Subjective, harder to maintain
    - Order by short form (`-f`, `-c`, `-s`): Confusing when flags have no short form

6. **Flag registration: Use *Var/*VarP for command-specific flags, Get*() for global flags**

    **Choice:** Use `*Var` and `*VarP` for command-specific flags, use `Get*()` for global persistent flags

    **Rationale:**
    - Command-specific flags: Type-safe, compile-time checked, direct variable access, single declaration point
    - Global persistent flags: Runtime lookup appropriate for flags accessed globally (e.g., verbose in util.go)
    - Consistent with create/delete commands for command-specific flags
    - Consistent with root/util for global flags

    **Examples:**
    - Command-specific: `cmd.Flags().BoolVar(&check, "check", false, "...")` ✓
    - Command-specific: `cmd.Flags().BoolVarP(&force, "force", "f", false, "...")` ✓
    - Global persistent: `cmd.Flags().GetCount("verbose")` in util.go ✓

    **Alternatives considered:**
    - Use GetBool for all flags: Runtime-only type safety, string lookups, less robust
    - Use *Var for all flags: Forces passing global flags through all functions (verbose parameter everywhere)

## Risks / Trade-offs

[Risk] Shell wrapper pattern matching may not handle complex flag combinations
→ **Mitigation:** E2E tests across all three shells (bash/zsh/fish) with various flag combinations

[Risk] Delete navigation might resolve to unexpected path in edge cases
→ **Mitigation:** Comprehensive unit tests for all context types (project, worktree, outside git)

[Risk] Breaking change if user scripts parse `--change-dir` flag
→ **Mitigation:** Keep `-C` short form working, document clearly. Scripts using long form need update but short form continues to work.

[Risk] Pattern matching in shell wrapper might capture unintended command output
→ **Mitigation:** Explicit case/switch for create/delete commands only, not wildcard matching

[Trade-off] Path-only output loses user-friendly success message
→ **Acceptable:** Verbose logging will restore messages in future parallel work

[Trade-off] More complex shell wrapper (case statements) vs simple pattern
→ **Acceptable:** Complexity is localized to wrapper functions, not user-visible

## Open Questions

None
