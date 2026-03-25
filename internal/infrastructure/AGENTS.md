## Infrastructure Layer

Layer: External integrations (git, config, CLI execution)

**Interfaces:** Defined in `application/` - implementations here satisfy those contracts.

**Compile-time checks:** Each implementation includes `var _ Interface = (*Implementation)(nil)`.

## Error Wrapping

**Rule:** Return domain error types, not plain `fmt.Errorf`.

| Operation | Pattern |
|-----------|---------|
| Repository | `domain.NewGitRepositoryError(path, message, cause)` |
| Worktree | `domain.NewGitWorktreeError(worktreePath, branchName, message, cause)` |
| Context | `domain.NewContextDetectionError(path, message, cause)` |
| Config | `domain.NewConfigError(path, message, cause)` |

**Exception:** String-based checks OK for parsing external CLI output.

## Git Client Routing

| Operation | GoGitClient | CLIClient | Rationale |
|-----------|:-----------:|:---------:|-----------|
| Open repo, List branches, Branch exists | ✅ | ❌ | Portable, deterministic |
| Get status, Validate repo, Get info | ✅ | ❌ | Portable, deterministic |
| List remotes, Get commit info | ✅ | ❌ | Portable, deterministic |
| Create/Delete/List worktree, Prune | ❌ | ✅ | go-git lacks support |
| Is branch merged, Delete branch | ❌ | ✅ | go-git limitations |

## GoGitClient Implementation

```go
// Compile-time interface check
var _ application.GoGitClient = (*GoGitClientImpl)(nil)

type GoGitClientImpl struct { ... }
```

**Cache:** LRU cache (default 25 repos) prevents memory leak.
```go
NewGoGitClient(cacheEnabled...)           // default size 25
NewGoGitClientWithSize(cacheSize, ...)    // custom size
```

## CLIClient Implementation

```go
// Compile-time interface check
var _ application.CLIClient = (*CLIClientImpl)(nil)

type CLIClientImpl struct { ... }
```

## Path Utilities

| Function | Purpose |
|----------|---------|
| `IsPathUnder(base, target)` | Check target under base, resolves symlinks |
| `ExtractProjectFromWorktreePath(path, worktreesDir)` | Get project name from `{worktreesDir}/{project}/{branch}/...` |
| `NormalizePath(path)` | Absolute path, symlinks resolved |

## Context Detection

**Priority:** Worktree folder → Project folder (`.git/` found) → Outside git

**Worktree pattern:** `$HOME/Worktrees/<project>/<branch>/` with valid `.git` file

## Context Resolution

| Context | `<branch>` | `<project>` | `<project>/<branch>` |
|---------|------------|-------------|----------------------|
| Project | Current project worktree | Different project main | Cross-project worktree |
| Worktree | Different worktree, same project | Different project main | Cross-project worktree |
| Outside | - | Project main | Cross-project worktree |

## ShellInfrastructure Implementation

```go
// Compile-time interface check
var _ application.ShellInfrastructure = (*ShellInfrastructureImpl)(nil)

type ShellInfrastructureImpl struct { ... }
```

| Shell | Config Files (preference order) |
|-------|--------------------------------|
| Bash | `.bashrc`, `.bash_profile`, `.profile` |
| Zsh | `.zshrc`, `.zprofile`, `.profile` |
| Fish | `.config/fish/config.fish`, `config.fish`, `.fishrc` |

## Configuration

**Location:** `$HOME/.config/twiggit/config.toml` (XDG)
**Priority:** defaults → config file → env vars (`TWIGGIT_*`) → flags

**Path expansion:** `$VAR`, `${VAR}`, and `~` expanded in path fields:
- `ProjectsDirectory`, `WorktreesDirectory`, `Shell.Wrapper.BackupDir`
- Example: `worktrees_directory = "$HOME/Worktrees"` → `/home/user/Worktrees`

**Completion timeout:**
```toml
[completion]
timeout = "500ms"  # Optional, default 500ms
```

Slow git operations gracefully degrade to empty suggestions.

## HookRunner Implementation

```go
// Compile-time interface check
var _ application.HookRunner = (*HookRunnerImpl)(nil)

type HookRunnerImpl struct { ... }
```

**Configuration:** `.twiggit.toml` at repository root
```toml
[hooks.post-create]
commands = [
    "mise trust",
    "npm install",
]
```

**Env vars set during execution:**

| Variable | Description |
|----------|-------------|
| `TWIGGIT_WORKTREE_PATH` | Path to new worktree (also cwd) |
| `TWIGGIT_PROJECT_NAME` | Project identifier |
| `TWIGGIT_BRANCH_NAME` | New branch name |
| `TWIGGIT_SOURCE_BRANCH` | Branch created from |
| `TWIGGIT_MAIN_REPO_PATH` | Main repository location |

**Failure handling:** All commands run even if previous fail; failures collected and returned.
