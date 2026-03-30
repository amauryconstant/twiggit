## 1. Error Handling - Domain Layer

- [ ] 1.1 Add IsNotFound() to ProjectServiceError in internal/domain/service_errors.go
- [ ] 1.2 Add IsNotFound() to NavigationServiceError in internal/domain/service_errors.go
- [ ] 1.3 Add IsNotFound() to ResolutionError in internal/domain/service_errors.go

## 2. Error Handling - Cmd Layer

- [ ] 2.1 Fix cmd/create.go:65 to return ValidationError directly (remove branchValidation wrapper)
- [ ] 2.2 Fix cmd/create.go:77-78 to return error directly from parseProjectBranch
- [ ] 2.3 Fix cmd/create.go:91-92 to return ValidationError directly for invalid source branch
- [ ] 2.4 Fix cmd/delete.go:108 to return NavigationServiceError with NotFound
- [ ] 2.5 Fix cmd/cd.go:90-93 to return ValidationError directly

## 3. Type Safety

- [ ] 3.1 Add nil check for req.Context in internal/service/worktree_service.go:326-329
- [ ] 3.2 Add validation for empty ResolvedPath in cmd/delete.go:75-84

## 4. Concurrency

- [ ] 4.1 Add mutex for result modifications in PruneMergedWorktrees (internal/service/worktree_service.go:409-488)
- [ ] 4.2 Add synchronization to pruneProjectWorktrees (internal/service/worktree_service.go:497-529)

## 5. Code Deduplication

- [ ] 5.1 Extract shell auto-detection to shared helper in internal/service/shell_service.go
- [ ] 5.2 Extract navigation target resolution to cmd/util.go
- [ ] 5.3 Extract path validation to shared function in internal/infrastructure/context_resolver.go

## 6. CLI Improvements

- [ ] 6.1 Use config.Validation.DefaultSourceBranch in cmd/create.go when --source not specified
- [ ] 6.2 Add short flag -m to --merged-only in cmd/delete.go
- [ ] 6.3 Add short flag to --delete-branches in cmd/prune.go
- [ ] 6.4 Add preview display before confirmation in cmd/prune.go:72-81

## 7. Hardcoded Values

- [ ] 7.1 Use config.Git.CLITimeout in internal/infrastructure/cli_client.go:76
- [ ] 7.2 Add hook timeout to config and use in internal/infrastructure/hook_runner.go:29

## 8. Testing

- [ ] 8.1 Add unit tests for cmd/error_formatter.go
- [ ] 8.2 Add unit tests for cmd/util.go
- [ ] 8.3 Add e2e test for prune --delete-branches in test/e2e/
- [ ] 8.4 Add integration test for shell wrapper block

## 9. Verification

- [ ] 9.1 Run mise run lint:fix && mise run check
- [ ] 9.2 Run mise run test
- [ ] 9.3 Run mise run test:e2e
- [ ] 9.4 Run go test -race ./... to verify concurrency fixes
- [ ] 9.5 Verify go build ./... compiles successfully
