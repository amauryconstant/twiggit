## 1. Core Implementation

- [ ] 1.1 Add `expandConfigPath()` pure function in `config_manager.go`
- [ ] 1.2 Add `normalizeConfigPaths()` function to expand all path fields
- [ ] 1.3 Update `Load()` to call `normalizeConfigPaths()` after unmarshal

## 2. Testing

- [ ] 2.1 Add unit tests for `expandConfigPath()` covering `$VAR`, `${VAR}`, `~`, absolute paths
- [ ] 2.2 Add unit tests for `normalizeConfigPaths()` covering all three path fields
- [ ] 2.3 Add integration test verifying config with `$HOME` paths loads successfully
- [ ] 2.4 Run `mise run check` to verify all tests pass and lint is clean
