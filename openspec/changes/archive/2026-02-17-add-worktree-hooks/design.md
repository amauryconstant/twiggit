## Context

Twiggit manages git worktrees across multiple projects. After creating a worktree, developers typically need to run setup commands to configure the development environment (e.g., `mise trust`, `npm install`, etc.). Currently, this requires manual execution.

This design introduces an opt-in hook system that executes configured commands after worktree creation. Hooks are defined in a `.twiggit.toml` file at the repository root, keeping configuration project-specific and version-controlled.

**Current State:**
- `WorktreeService.CreateWorktree` returns `(*domain.WorktreeInfo, error)`
- No hook mechanism exists
- Configuration lives in `~/.config/twiggit/config.toml` (global only)

**Constraints:**
- Must follow existing DDD-light architecture patterns
- Must use constructor injection for dependencies
- Must use existing `CommandExecutor` for command execution
- Must use koanf/toml for config parsing (consistent with ConfigManager)

## Goals / Non-Goals

**Goals:**
- Execute user-defined commands after worktree creation
- Support per-project hook configuration via `.twiggit.toml`
- Provide environment variables with execution context
- Report hook failures as warnings without blocking worktree creation
- Design for extensibility (future hook types: pre-create, post-delete, etc.)

**Non-Goals:**
- Global hooks (hooks only exist per-project via `.twiggit.toml`)
- Hook rollback/retry mechanisms
- Conditional hook execution based on branch patterns
- Parallel hook execution

## Decisions

### Decision 1: Configuration File Format

**Choice:** TOML file at repository root (`.twiggit.toml`)

**Rationale:**
- Consistent with existing config approach (ConfigManager uses TOML)
- Declarative and easy to parse/validate
- Version-controllable alongside project code
- Simple for the primary use case (list of commands)

**Format:**
```toml
[hooks.post-create]
commands = [
    "mise trust",
    "npm install",
]
```

**Alternatives Considered:**
- Executable script (`.twiggit/hooks`): More flexible but less discoverable, harder to validate
- Directory of scripts (`.twiggit/hooks/post-create.d/`): Overkill for simple command lists, more complex to implement

### Decision 2: Execution Context

**Choice:** Commands execute in the new worktree directory

**Rationale:**
- Most setup commands need to run inside the worktree (e.g., `mise trust` checks `.mise.toml` in cwd)
- Natural "cd into it and set up" workflow
- Main repo path available via environment variable if needed

**Environment Variables:**
| Variable | Description |
|----------|-------------|
| `TWIGGIT_WORKTREE_PATH` | Path to the new worktree (also cwd) |
| `TWIGGIT_PROJECT_NAME` | Project identifier |
| `TWIGGIT_BRANCH_NAME` | New branch name |
| `TWIGGIT_SOURCE_BRANCH` | Branch created from |
| `TWIGGIT_MAIN_REPO_PATH` | Main repository location |

### Decision 3: Failure Handling

**Choice:** Hook failures produce warnings; no rollback of worktree creation

**Rationale:**
- Worktree creation is the primary operation and has succeeded
- Hook failure indicates incomplete setup, not failed creation
- User can manually retry commands if needed
- Simpler implementation (no transaction/rollback logic)

**Behavior:**
- All commands in a hook run even if previous commands fail
- Failures collected and returned in `HookResult`
- CMD layer displays warnings after successful creation message

### Decision 4: Interface Location

**Choice:** `HookRunner` interface in `infrastructure/interfaces.go`

**Rationale:**
- Hook execution is an external integration concern (executing shell commands)
- Follows pattern of `GitClient`, `CommandExecutor` in infrastructure layer
- Service layer depends on interface, not implementation

### Decision 5: Result Type Design

**Choice:** New `CreateWorktreeResult` type wrapping both `WorktreeInfo` and `HookResult`

**Rationale:**
- `WorktreeInfo` is used across multiple operations (list, get, delete)
- Adding `HookResult` directly to `WorktreeInfo` would pollute all usages
- Only `CreateWorktree` needs hook results

**Type Structure:**
```go
type CreateWorktreeResult struct {
    Worktree   *WorktreeInfo
    HookResult *HookResult  // nil if no hooks configured or no hooks ran
}

type HookResult struct {
    HookType  HookType
    Executed  bool          // Were any commands configured?
    Success   bool          // All commands succeeded?
    Failures  []HookFailure // Empty if success
}

type HookFailure struct {
    Command  string
    ExitCode int
    Output   string
}
```

### Decision 6: Hook Extensibility

**Choice:** Design types to support future hook types without structural changes

**Rationale:**
- `post-create` is the initial implementation
- Future hooks likely: `pre-create`, `post-delete`, `post-prune`
- `HookType` enum allows easy addition

**Type Structure:**
```go
type HookType string

const (
    HookPostCreate HookType = "post-create"
    // Future: HookPreCreate, HookPostDelete, HookPostPrune
)

type HookConfig struct {
    PostCreate *HookDefinition `toml:"post-create" koanf:"post-create"`
    // Future hooks added here
}

type HookDefinition struct {
    Commands []string `toml:"commands" koanf:"commands"`
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Malicious commands in `.twiggit.toml` from cloned repos | Document that users should review `.twiggit.toml` before trusting a repo; hooks are opt-in (no file = no execution) |
| Long-running hooks delay worktree creation | Document best practices; future: add timeout config per-hook |
| Commands fail due to missing dependencies | Clear error output with command, exit code, and stdout/stderr |
| `.twiggit.toml` not found/symlinked | Silently skip hooks (opt-in behavior) |
