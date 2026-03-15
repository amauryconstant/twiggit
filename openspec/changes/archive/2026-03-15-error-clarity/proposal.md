## Why

Current error messages expose internal implementation details like service names (`WorktreeService.CreateWorktree`) that confuse users and provide no actionable guidance. Exit codes are generic (0 for success, 1 for error, 2 for usage), making reliable scripting impossible. Users see messages like:

```
Error: failed to create worktree: worktree service operation 'CreateWorktree' failed for myproject: ...
```

Instead of actionable guidance like:

```
Error: could not create worktree for 'myproject'
Hint: Check that the project exists and you have permission
```

This is Stream 2 of pre-release improvements to polish the user experience before v1.0.

## What Changes

- **Simplify error messages**: Service errors will return user-friendly messages, hiding internal operation names while keeping details available for debugging via `TWIGGIT_DEBUG` environment variable
- **Add granular exit codes**: New exit codes for different error categories (config=3, git=4, validation=5, not-found=6) enable reliable scripting
- **Add panic recovery**: Unexpected panics will be caught and converted to user-friendly messages with stack traces only in debug mode

## Capabilities

### New Capabilities

- `error-clarity`: User-facing error formatting with actionable messages, granular exit codes for scripting reliability, and panic recovery for robustness

### Modified Capabilities

(None - this introduces new capability without changing existing spec requirements)

## Impact

- `internal/domain/service_errors.go` - Simplify `Error()` methods for user-facing output
- `cmd/error_formatter.go` - Enhanced formatting with hints for all error types
- `cmd/error_handler.go` - New exit code constants and mapping logic
- `main.go` - Add panic recovery with defer/recover pattern
- E2E tests - Verify new exit codes and error message format
