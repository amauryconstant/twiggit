## 1. Simplify Error Messages

- [x] 1.1 Modify `ServiceError.Error()` in `internal/domain/service_errors.go` to return user-friendly messages without internal operation names
- [x] 1.2 Modify `WorktreeServiceError.Error()` to hide internal operation names while keeping worktree context
- [x] 1.3 Modify `ProjectServiceError.Error()` to hide internal operation names while keeping project context
- [x] 1.4 Modify `NavigationServiceError.Error()` to hide internal operation names while keeping navigation context
- [x] 1.5 Ensure `ValidationError` messages remain unchanged (already user-friendly)

## 2. Granular Exit Codes

- [x] 2.1 Add new exit code constants to `cmd/error_handler.go`: `ExitCodeConfig`, `ExitCodeGit`, `ExitCodeValidation`, `ExitCodeNotFound`
- [x] 2.2 Update `GetExitCodeForError()` to map error types to new exit codes
- [x] 2.3 Add `ErrorCategoryNotFound` to the `ErrorCategory` enum
- [x] 2.4 Update `CategorizeError()` to detect not-found errors using existing `IsNotFound()` methods
- [x] 2.5 Update `cmd/error_formatter.go` hints to match simplified error messages

## 3. Panic Recovery

- [x] 3.1 Add panic recovery with defer/recover in `main.go`
- [x] 3.2 Display "Internal error: <panic value>" to stderr on panic
- [x] 3.3 Check `TWIGGIT_DEBUG` environment variable and show stack trace when set
- [x] 3.4 Exit with code 1 on recovered panic

## 4. Testing

- [x] 4.1 Add E2E tests verifying exit codes for different error categories
- [x] 4.2 Add E2E tests verifying user-friendly error message format
- [x] 4.3 Add E2E tests verifying panic recovery with and without `TWIGGIT_DEBUG`
- [x] 4.4 Run full test suite (`mise run test:full`) and ensure all tests pass
- [x] 4.5 Run linting (`mise run lint:fix`) and fix any issues
