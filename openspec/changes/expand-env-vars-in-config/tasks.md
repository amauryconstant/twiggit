## 1. Core Implementation

- [x] 1.1 Add `expandConfigPath()` pure function in `config_manager.go`
- [x] 1.2 Add `normalizeConfigPaths()` function to expand all path fields
- [x] 1.3 Update `Load()` to call `normalizeConfigPaths()` after unmarshal

## 2. Testing

- [x] 2.1 Add unit tests for `expandConfigPath()` covering `$VAR`, `${VAR}`, `~`, absolute paths
- [x] 2.2 Add unit tests for `normalizeConfigPaths()` covering all three path fields
- [x] 2.3 Add integration test verifying config with `$HOME` paths loads successfully
- [x] 2.4 Run `mise run check` to verify all tests pass and lint is clean
