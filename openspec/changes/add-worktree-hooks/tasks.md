## 1. Domain Types

- [x] 1.1 Create `internal/domain/hook_types.go` with `HookType`, `HookConfig`, `HookDefinition`, `HookResult`, `HookFailure` types
- [x] 1.2 Add `CreateWorktreeResult` struct to `internal/domain/service_results.go`

## 2. Infrastructure Layer

- [x] 2.1 Add `HookRunner` interface to `internal/infrastructure/interfaces.go`
- [x] 2.2 Add `HookRunRequest` struct to `internal/infrastructure/interfaces.go`
- [x] 2.3 Create `internal/infrastructure/hook_runner.go` with `hookRunner` struct and `NewHookRunner` constructor
- [x] 2.4 Implement `ReadHookConfig` function to parse `.twiggit.toml` using koanf; return nil config and log warning on malformed TOML
- [x] 2.5 Implement `Run` method executing commands via `CommandExecutor` with environment variables
- [x] 2.6 Implement failure collection logic (continue on failure, collect all failures)

## 3. Application Layer

- [x] 3.1 Update `WorktreeService` interface in `internal/application/interfaces.go` to return `(*domain.CreateWorktreeResult, error)` for `CreateWorktree`

## 4. Service Layer

- [x] 4.1 Add `hookRunner` field to `worktreeService` struct in `internal/service/worktree_service.go`
- [x] 4.2 Update `NewWorktreeService` constructor to inject `HookRunner`
- [x] 4.3 Update `CreateWorktree` method to call `hookRunner.Run` after successful worktree creation
- [x] 4.4 Return `CreateWorktreeResult` with `WorktreeInfo` and `HookResult`

## 5. CMD Layer

- [x] 5.1 Update `cmd/root.go` to create `HookRunner` and inject into `WorktreeService`
- [x] 5.2 Update `cmd/create.go` to handle `CreateWorktreeResult` return type
- [x] 5.3 Add hook failure warning display logic to `cmd/create.go`

## 6. Tests

- [x] 6.1 Add unit tests for `hook_runner.go` with mock `CommandExecutor` (success, failure, malformed TOML, missing commands array, empty commands)
- [x] 6.2 Add integration test for hook execution with real `.twiggit.toml` file
- [x] 6.3 Add E2E test for `twiggit create` with hooks configured
- [x] 6.4 Add E2E test for hook failure warning display
- [x] 6.5 Add E2E test for error when `--source` branch does not exist
- [x] 6.6 Add E2E test for error when outside git context without project specification

## 7. Mocks

- [x] 7.1 Add `MockHookRunner` to `test/mocks/` for service layer testing — SKIPPED: causes import cycle; hook_runner_test uses MockCommandExecutor directly

## 8. Documentation

- [x] 8.1 Add security note to README/docs: users should review `.twiggit.toml` before trusting a repo (hooks execute arbitrary commands)
