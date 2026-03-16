## Context

Config loading via koanf does not expand environment variables. TOML values like `$HOME/Worktrees` are stored as literal strings, causing `filepath.IsAbs()` validation to fail. The expansion must occur between unmarshaling and validation.

Current flow:
```
Load defaults → Load TOML → Unmarshal → Validate → Return
```

New flow:
```
Load defaults → Load TOML → Unmarshal → Expand env vars → Validate → Return
```

## Goals / Non-Goals

**Goals:**
- Expand `$VAR`, `${VAR}`, and `~` in config path fields
- Apply expansion to `ProjectsDirectory`, `WorktreesDirectory`, `Shell.Wrapper.BackupDir`
- Maintain backward compatibility with literal paths

**Non-Goals:**
- Shell command substitution (e.g., `$(whoami)`)
- Recursive environment variable expansion
- Expansion of non-path config fields

## Decisions

### 1. Expansion Location: Post-Load in ConfigManager

**Choice:** Add `normalizeConfigPaths()` function called in `Load()` after unmarshal, before validation.

**Rationale:** 
- Centralized: All path expansion in one place
- Clean separation: Expansion is infrastructure concern, not domain
- Early validation: Expanded paths validated immediately

**Alternatives:**
- In `Validate()`: Mixes concerns, domain layer would need env access
- Custom koanf parser: More complex, harder to test
- In `DefaultConfig()`: Only helps defaults, not user config

### 2. Expansion Implementation

**Choice:** Pure functions using `os.ExpandEnv()` + manual `~` handling.

**Rationale:**
- `os.ExpandEnv()` handles `$VAR` and `${VAR}` natively
- `~` requires manual expansion via `os.UserHomeDir()`
- Pure functions are easily testable

**Pattern:**
```go
func expandConfigPath(path string) string {
    if strings.HasPrefix(path, "~") {
        home, _ := os.UserHomeDir()
        return filepath.Join(home, strings.TrimPrefix(path, "~"))
    }
    return os.ExpandEnv(path)
}
```

### 3. Fields to Expand

**Choice:** Expand only path fields that commonly use `$HOME`:
- `ProjectsDirectory`
- `WorktreesDirectory`
- `Shell.Wrapper.BackupDir`

**Rationale:** These are the only fields where users expect shell-style expansion.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| `os.UserHomeDir()` fails | Fallback to `$HOME` env, then `/tmp` |
| Empty env var leaves `$VAR` literal | Document behavior; user must set vars |
| Breaking change for users with literal `$` in paths | Unlikely edge case; `$$` escapes if needed |
