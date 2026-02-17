## 1. Domain Types

- [ ] 1.1 Create `internal/domain/hook_types.go` with `HookType`, `HookConfig`, `HookDefinition`, `HookResult`, `HookFailure` types
- [ ] 1.2 Add `CreateWorktreeResult` struct to `internal/domain/service_results.go`

## 2. Infrastructure Layer

- [ ] 2.1 Add `HookRunner` interface to `internal/infrastructure/interfaces.go`
- [ ] 2.2 Add `HookRunRequest` struct to `internal/infrastructure/interfaces.go`
- [ ] 2.3 Create `internal/infrastructure/hook_runner.go` with `hookRunner` struct and `NewHookRunner` constructor
- [ ] 2.4 Implement `ReadHookConfig` function to parse `.twiggit.toml` using koanf; return nil config and log warning on malformed TOML
- [ ] 2.5 Implement `Run` method executing commands via `CommandExecutor` with environment variables
- [ ] 2.6 Implement failure collection logic (continue on failure, collect all failures)

## 3. Application Layer

- [ ] 3.1 Update `WorktreeService` interface in `internal/application/interfaces.go` to return `(*domain.CreateWorktreeResult, error)` for `CreateWorktree`

## 4. Service Layer

- [ ] 4.1 Add `hookRunner` field to `worktreeService` struct in `internal/service/worktree_service.go`
- [ ] 4.2 Update `NewWorktreeService` constructor to inject `HookRunner`
- [ ] 4.3 Update `CreateWorktree` method to call `hookRunner.Run` after successful worktree creation
- [ ] 4.4 Return `CreateWorktreeResult` with `WorktreeInfo` and `HookResult`

## 5. CMD Layer

- [ ] 5.1 Update `cmd/root.go` to create `HookRunner` and inject into `WorktreeService`
- [ ] 5.2 Update `cmd/create.go` to handle `CreateWorktreeResult` return type
- [ ] 5.3 Add hook failure warning display logic to `cmd/create.go`

## 6. Tests

- [ ] 6.1 Add unit tests for `hook_runner.go` with mock `CommandExecutor` (success, failure, malformed TOML, missing commands array, empty commands)
- [ ] 6.2 Add integration test for hook execution with real `.twiggit.toml` file
- [ ] 6.3 Add E2E test for `twiggit create` with hooks configured
- [ ] 6.4 Add E2E test for hook failure warning display

## 7. Mocks

- [ ] 7.1 Add `MockHookRunner` to `test/mocks/` for service layer testing

## 8. Documentation

- [ ] 8.1 Add security note to README/docs: users should review `.twiggit.toml` before trusting a repo (hooks execute arbitrary commands)
