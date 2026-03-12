## 1. Domain Layer Cleanup

- [x] 1.1 Remove `DryRun` field from `SetupShellRequest` in `internal/domain/shell_requests.go`
- [x] 1.2 Remove `DryRun` field from `SetupShellResult` in `internal/domain/shell_results.go`

## 2. Service Layer Updates

- [x] 2.1 Remove `DryRun` branch from `SetupShell()` in `internal/service/shell_service.go`
- [x] 2.2 Update `SetupShell()` to no longer check/handle `req.DryRun`

## 3. Command Layer Rewrite

- [x] 3.1 Update `NewInitCmd()` to use `[shell]` positional argument instead of `[config-file]`
- [x] 3.2 Add `-i, --install` flag for file installation mode
- [x] 3.3 Add `-c, --config` flag for custom config file path
- [x] 3.4 Remove `--dry-run` flag
- [x] 3.5 Remove `--shell` flag
- [x] 3.6 Remove `--check` flag and `runInitCheck()` function
- [x] 3.7 Implement flag validation: `--config` requires `--install`
- [x] 3.8 Implement flag validation: `--force` requires `--install`
- [x] 3.9 Implement stdout mode: call `GenerateWrapper()` and print to stdout (default behavior)
- [x] 3.10 Implement install mode: call `SetupShell()` when `--install` flag present
- [x] 3.11 Update `displayInitResults()` to only handle install mode (stdout mode has no metadata)
- [x] 3.12 Update command `Use` string to `init [shell]`
- [x] 3.13 Update command `Short` and `Long` descriptions for new behavior
- [x] 3.14 Update `Args` validation to accept 0-1 positional args

## 4. Update Unit Tests

- [x] 4.1 Update `cmd/init_test.go` for new flag structure
- [x] 4.2 Remove tests for `--dry-run`, `--shell`, `--check` flags
- [x] 4.3 Add tests for `-i, --install` flag
- [x] 4.4 Add tests for `-c, --config` flag validation
- [x] 4.5 Add tests for `--force` requiring `--install`
- [x] 4.6 Add tests for stdout output mode (default)
- [x] 4.7 Add tests for positional `[shell]` argument
- [x] 4.8 Update `internal/service/shell_service_test.go` to remove DryRun test cases

## 5. Update E2E Tests

- [x] 5.1 Update `test/e2e/init_test.go` for new command behavior
- [x] 5.2 Add E2E test: `twiggit init` outputs to stdout
- [x] 5.3 Add E2E test: `twiggit init bash` outputs bash wrapper to stdout
- [x] 5.4 Add E2E test: `twiggit init --install` writes to auto-detected config
- [x] 5.5 Add E2E test: `twiggit init bash --install --config <path>` writes to custom config
- [x] 5.6 Add E2E test: `twiggit init --config <path>` errors without `--install`
- [x] 5.7 Add E2E test: `twiggit init --force` errors without `--install`
- [x] 5.8 Remove E2E tests for removed flags (`--dry-run`, `--shell`, `--check`)

## 6. Update Shell Completion

- [x] 6.1 Update shell completion if needed for new positional argument

## 7. Verification

- [x] 7.1 Run `mise run check` to ensure all linting and tests pass
- [x] 7.2 Verify `eval "$(twiggit init bash)"` works correctly
- [x] 7.3 Verify `twiggit init --install` writes to config file
- [x] 7.4 Verify error messages guide users when flags are misused
