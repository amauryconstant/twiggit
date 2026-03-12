## Context

Current error handling has three issues:

1. **Verbose error messages**: `ServiceError.Error()` exposes internal details like `WorktreeService.CreateWorktree` that confuse users and provide no actionable guidance

2. **Generic exit codes**: Only 0 (success), 1 (error), 2 (usage) exist, making reliable scripting impossible when users need to distinguish between config errors, git errors, validation errors, etc.

3. **No panic recovery**: Unexpected panics crash the CLI with Go stack traces, providing poor user experience

Current error flow:
```
service layer → wraps in ServiceError with internal names
cmd layer → formats via ErrorFormatter
main.go → no panic recovery
```

Constraints:
- Domain layer (`internal/domain/`) MUST NOT depend on external packages
- Error types already implement `Unwrap()` for error chain support
- `ValidationError` messages are already user-friendly and SHOULD remain unchanged

## Goals / Non-Goals

**Goals:**
- Simplify `ServiceError.Error()` to return user-friendly messages without internal operation names
- Add granular exit codes (3-6) for different error categories
- Add panic recovery in main.go with optional stack traces via `TWIGGIT_DEBUG`
- Keep internal details available for debugging when `TWIGGIT_DEBUG` is set

**Non-Goals:**
- Changing `ValidationError` format (already user-friendly)
- Changing infrastructure error types (`GitRepositoryError`, `ConfigError`, etc.)
- Adding new error types
- Creating structured logging or error telemetry

## Decisions

### Decision 1: Simplify ServiceError messages at source (domain layer)

**Choice:** Modify `ServiceError.Error()` to return simplified user-friendly messages.

**Rationale:** 
- Domain errors should be self-describing
- ErrorFormatter already has type-specific logic; simplifying at source is cleaner than adding more formatter complexity
- Internal operation names (`WorktreeService.CreateWorktree`) provide no value to users

**Alternatives considered:**
- Keep domain errors verbose, simplify only in ErrorFormatter: Would require formatter to parse and strip operation names, which is fragile
- Add `UserMessage()` method to error types: Adds API surface without benefit since `Error()` is the standard interface

**Pattern:**
```go
// Before:
// "WorktreeService.CreateWorktree failed: could not create worktree"

// After:
// "could not create worktree for 'myproject'"
```

### Decision 2: Add exit codes 3-6 for specific error categories

**Choice:** Add four new exit codes mapped to error categories.

**Rationale:**
- Enables reliable shell scripting where users can distinguish error types
- Follows conventions where exit code 2 is typically reserved for usage errors
- Categories align with existing `ErrorCategory` enum in `error_handler.go`

**Exit codes:**
| Code | Constant | Category |
|------|----------|----------|
| 0 | ExitCodeSuccess | Success |
| 1 | ExitCodeError | General/unclassified error |
| 2 | ExitCodeUsage | Command usage error (Cobra) |
| 3 | ExitCodeConfig | Configuration error |
| 4 | ExitCodeGit | Git operation error |
| 5 | ExitCodeValidation | Input validation error |
| 6 | ExitCodeNotFound | Resource not found |

**Alternatives considered:**
- Use exit codes 64-78 (BSD sysexits): Overkill for this tool, less intuitive for shell scripts
- Single additional exit code: Insufficient for distinguishing error types

### Decision 3: Panic recovery with debug-aware output

**Choice:** Add defer/recover in main.go that checks `TWIGGIT_DEBUG` for stack traces.

**Rationale:**
- Provides graceful handling of unexpected panics
- Stack traces only shown when explicitly requested
- Simple implementation with clear debug mode opt-in

**Pattern:**
```go
defer func() {
    if r := recover(); r != nil {
        fmt.Fprintf(os.Stderr, "Internal error: %v\n", r)
        if os.Getenv("TWIGGIT_DEBUG") != "" {
            debug.PrintStack()
        }
        os.Exit(1)
    }
}()
```

**Alternatives considered:**
- Always show stack traces: Too verbose for users
- Write panics to a log file: Adds complexity for minimal value
- Custom panic handler type: Overkill for this scope

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Simplified messages may lose debugging context | Keep full details available when `TWIGGIT_DEBUG` is set; error chain still accessible via `Unwrap()` |
| Exit code changes may break existing scripts | Document clearly in CHANGELOG; existing scripts using `!= 0` checks still work |
| Panic recovery may hide real bugs | Always exit with code 1 for panics; stack traces available in debug mode |
| Mapping errors to exit codes may be ambiguous | Use first-match-wins ordering in `GetExitCodeForError()` with clear category precedence |
