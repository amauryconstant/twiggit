## Context

Current state: Debug output is scattered throughout the codebase with ad-hoc `fmt.Fprintf(os.Stderr, "DEBUG: ...")` calls in both the command layer (`cmd/create.go`) and service layer (`internal/services/worktree_service.go`). The "DEBUG:" prefix is developer-focused, and users have no control over verbosity.

Constraints:
- No new dependencies allowed
- Must follow existing DDD architecture (cmd/, services/, domain/, infrastructure/)
- No breaking changes to CLI behavior or external interfaces
- cmd/ package tested via E2E only
- Code conventions: NO comments in code

## Goals / Non-Goals

**Goals:**
- Provide user-controlled verbosity levels with `--verbose` flag (`-v`) that can be repeated for higher verbosity
- Implement two distinct verbosity levels (level 1: high-level flow, level 2: detailed parameters with indentation)
- Centralize verbose output handling in command layer only
- Remove all developer-focused debug output from service layer
- Maintain clean normal output (no verbose messages by default)

**Non-Goals:**
- Structured logging (JSON, timestamps, component tags)
- Color-coded output
- Configuration file support for verbosity settings
- Logging to files
- Debug mode for internal development (use standard Go debugging tools instead)

## Decisions

### 1. Verbose Flag Implementation

**Decision:** Use Cobra's `PersistentFlags().CountP("verbose", "v", ...)` to allow multiple `-v` flags

**Rationale:**
- CountP returns integer count of flag occurrences (0, 1, 2, etc.)
- Persistent flag available to all subcommands without duplication
- Standard Unix convention (git, docker, rsync)
- No new dependencies - uses existing Cobra framework

**Alternatives considered:**
- Separate flags (`--verbose`, `--very-verbose`): More verbose, less flexible
- Enum/string values (`--verbosity=low|high`): Harder to increment, not Unix-conventional
- Environment variables: Overkill, adds configuration complexity

### 2. Log Helper Function Location

**Decision:** Create `cmd/util.go` with `logv()` helper function

**Rationale:**
- Centralized implementation ensures consistency across all commands
- Easy to test and modify format in one place
- Separates verbose output logic from command business logic
- cmd/util.go is appropriate location for command-level utilities

**Alternatives considered:**
- Inline `if` checks in each command: Duplicates code, harder to maintain
- Helper in domain layer: Violates DDD, domain should not know about CLI concerns
- Separate package (cmd/logger): Overkill for a single function

### 3. Verbose Output Placement

**Decision:** Verbose output SHALL only appear in command layer (`cmd/*.go`), never in service layer

**Rationale:**
- Respects DDD separation - services handle business logic, not UI/output
- Makes services more reusable (could be called by non-CLI interfaces)
- Command layer has access to Cobra command object for flag checking
- Consistent with existing architecture pattern

**Alternatives considered:**
- Allow verbose in services: Violates separation of concerns
- Pass logger to services: Adds unnecessary complexity for this use case

### 4. Output Format

**Decision:** Plain text, no color, indentation for level 2, no prefix

**Rationale:**
- Matches Unix tool conventions (git, docker, go build)
- Pipe-friendly and scriptable
- Indentation provides visual hierarchy without noise
- No "DEBUG:" prefix - messages describe what's happening to users

**Alternatives considered:**
- With "[VERBOSE]" prefix: Unnecessary noise
- With colored output: Harder to parse, not pipe-friendly, requires colorama/terminal detection
- Structured format: Overkill, users want simple readable output

### 5. logv() Function Signature

**Decision:** `func logv(cmd *cobra.Command, level int, format string, args ...interface{})`

**Rationale:**
- Pass Cobra command to access verbose flag count via `cmd.Flags().GetCount("verbose")`
- Level parameter enables granular control (1 for high-level, 2 for detailed)
- Format + args follows fmt.Printf convention, familiar to Go developers
- Simple and extensible

**Alternatives considered:**
- Global verbose variable: Requires initialization, harder to test
- Closure approach: More complex, unnecessary abstraction
- Separate functions for each level (logv1, logv2): Duplicates code

## Risks / Trade-offs

**Risk: Verbose output may expose implementation details to users**

→ Mitigation: Use user-focused language, describe operations (what) not internals (how). Level 1 messages describe high-level actions, level 2 provides parameters without exposing service-layer internals.

**Risk: Removing debug output from services may impact developer debugging**

→ Mitigation: Use standard Go debugging tools (debuggers, delve) and add verbose output at command layer. Service layer remains pure and testable.

**Risk: Inconsistent verbose output across commands**

→ Mitigation: Centralized `logv()` function ensures consistent formatting. Provide guidance in cmd/AGENTS.md on what level to use for different types of output.

**Trade-off: Two-level verbosity is simpler than unlimited levels but less flexible**

→ Justification: Simplicity is prioritized for this use case. Two levels cover 99% of user needs (high-level flow vs. detailed parameters). Can always add level 3+ later if needed.

**Trade-off: Verbose output goes to stderr, not stdout**

→ Justification: Matches standard Unix convention (verbose/debug output to stderr, normal output to stdout). Allows piping normal output while seeing verbose messages. Command success/failure messages continue to stdout.

## Migration Plan

1. Add persistent verbose flag to `cmd/root.go`
2. Create `cmd/util.go` with `logv()` function
3. Update `cmd/create.go` to use `logv()` and migrate relevant service-layer debug
4. Update other commands (delete, list, cd, setup-shell) with `logv()` as appropriate
5. Remove all `fmt.Fprintf(os.Stderr, "DEBUG: ...")` from `internal/services/worktree_service.go`
6. Test verbose output levels manually and via E2E tests
7. Update cmd/AGENTS.md with verbose output guidelines

**Rollback strategy:** Git revert if issues arise. No data changes or external state modifications.

## Open Questions

None. Design is straightforward with clear decisions based on Unix conventions and existing architecture.
