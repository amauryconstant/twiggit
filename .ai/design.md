# Twiggit Design Specification

## Purpose

A pragmatic tool for managing git worktrees with a focus on rebase workflows. Context-aware operations WILL be provided for creating, listing, navigating, and deleting worktrees across multiple projects.

## Directory Structure & Defaults

### Default Paths
- **Projects**: `$HOME/Projects/<project-name>/`
- **Worktrees**: `$HOME/Worktrees/<project-name>/<worktree-branch-name>/`

### Directory Creation
- Default directories SHOULD exist before execution
- Missing directories SHOULD cause termination with error
- Required directories SHOULD be created manually
- Automatic directory creation SHOULD NOT occur

### Naming Conventions
- **Project names**: SHOULD follow GitHub repository naming rules
- **Branch names**: SHOULD follow git branch naming rules
- Additional configurable restrictions MAY be added in future

## Commands

### Core Commands

#### `list` - List git worktrees
**Output format**: Tabular format with columns for branch name, last commit information, and working directory status (clean/dirty)

**Context behavior**:
- From project folder: List worktrees for current project
- From worktree folder: List worktrees for current project
- From outside git: List all worktrees across all projects

**Flags**:
- `--all`: Show worktrees from all projects (overrides context)
- Note: `-C / --change-dir` flag SHOULD NOT be supported for list command

#### `create` - Create a new git worktree
**Required parameters**:
- Project name (inferred from context when possible)
- New branch name
- Source branch name (defaults to `main`)

**Parameter Inference Rules**:
1. **Project name** (inferred in this order):
   - From current directory if in project folder (`.git/` found)
   - From worktree path if in worktree folder (`$HOME/Worktrees/<project>/`)
   - Required as positional argument if outside git context

2. **Source branch** (defaults in this order):
   - `--source` flag if provided
   - `default_source_branch` from config file
   - `main` as final fallback

3. **New branch name**: SHALL be required as positional argument (cannot be inferred)

**Behavior**:
- Worktree SHALL be created from specified source branch
- If branch exists but worktree doesn't: Worktree SHALL be created only
- If worktree already exists: Error and exit SHALL occur
- Current directory SHALL be maintained after creation (unless `-C` flag used)

**Flags**:
- `--source <branch>`: Specify source branch (default: main)
- `-C / --change-dir`: Change to new worktree directory after creation

#### `cd` - Change directory to a git worktree
**Behavior**:
- Shell working directory SHALL be changed to target worktree via shell wrapper
- Target path SHALL be output to stdout for wrapper function consumption
- Specific error message SHALL be provided and non-zero exit SHALL occur if worktree doesn't exist or is inaccessible
- Explicit worktree specification SHALL be required when context is ambiguous

**Usage**: `twiggit cd feature-branch` (requires shell wrapper setup)

**Output Format**:
- **Success**: Absolute path to worktree directory (e.g., `/home/user/Worktrees/project-name/feature-branch`)
- **Error**: Error message to stderr, exit with non-zero code

**Shell Wrapper**:
- SHALL intercept `twiggit cd` calls and change shell directory
- SHALL provide escape hatch with `builtin cd` for shell built-in
- SHALL warn when overriding shell built-in `cd`
- SHALL pass through all other commands unchanged
- SHALL be automatically installed via `twiggit setup-shell` command

#### `delete` - Delete a git worktree
**Safety checks (always enforced)**:
- Uncommitted changes SHALL be checked - abort SHALL occur if found
- Current worktree status SHALL be checked - abort SHALL occur if active

**Default behavior (rebase workflow optimized)**:
- Worktree directory SHALL be removed
- Git branch SHALL be deleted
- Current directory SHALL be maintained after deletion (unless `-C` flag used)

**Flags**:
- `--keep-branch`: Branch SHALL be preserved after removing worktree
- `--force`: Uncommitted changes safety check SHALL be bypassed
- `--merged-only`: Deletion SHALL only be allowed if branch is merged
- `-C / --change-dir`: Change SHALL occur to main project directory after deletion

### Setup Command

#### `setup-shell` - Install shell wrapper
**Behavior**:
- Current shell (bash, zsh, fish) SHALL be detected
- Appropriate wrapper function with `builtin cd` escape hatch SHALL be generated
- Wrapper SHALL be added to correct shell configuration file
- Warning SHALL be provided about overriding shell built-in `cd` command
- User SHALL be instructed to restart shell or source configuration

### Help Command

#### `help` - Display help text
- Basic help text SHALL be returned
- Usage patterns and available commands SHALL be shown

## Context Detection & Behavior

### Context-Aware Navigation System

**Key Behaviors**:
- **Project Context**: `cd <branch>` SHALL navigate to worktree, `cd <project>` SHALL navigate to different project, `cd <project>/<branch>` SHALL navigate to cross-project worktree
- **Worktree Context**: `cd main` SHALL navigate to main project, `cd <branch>` SHALL navigate to different worktree, `cd <project>` SHALL navigate to different project
- **Outside Git Context**: `cd <project>` SHALL navigate to project main, `cd <project>/<branch>` SHALL navigate to cross-project worktree

### Identifier Resolution

| Context | Input | Target |
|---------|-------|--------|
| Project | `<branch>` | Current project worktree |
| Project | `<project>` | Different project main |
| Project | `<project>/<branch>` | Cross-project worktree |
| Worktree | `<branch>` | Different worktree same project |
| Worktree | `main` | Current project main |
| Worktree | `<project>` | Different project main |
| Worktree | `<project>/<branch>` | Cross-project worktree |
| Outside | `<branch>` | Invalid - requires project context |
| Outside | `<project>` | Project main directory |
| Outside | `<project>/<branch>` | Cross-project worktree |

### Context Detection Rules
1. **Project folder**: `.git/` directory found in current or parent directories
2. **Worktree folder**: Path matches `$HOME/Worktrees/<project>/<branch>/` pattern  
3. **Outside git**: No `.git/` found and not in worktree pattern

### Context Detection Priority
- Context detection SHALL be performed before identifier resolution
- Worktree folder detection SHALL take precedence over project folder detection when both patterns match
- Context type SHALL be determined using the following priority:
  1. **Worktree folder** (if path matches worktree pattern and contains valid worktree)
  2. **Project folder** (if `.git/` found and path doesn't match worktree pattern)
  3. **Outside git** (neither condition met)

### Context Behavior
- **From project folder**: Command SHALL be applied to current project
- **From worktree folder**: Command SHALL be applied to current worktree or encapsulating project
- **From outside git**: Explicit project/worktree specification SHALL be required or `--all` SHALL be used

## Future Features

### Maintenance Commands
- The system MAY support maintenance commands like prune, clean, status, repair

### Advanced Features
- Post-creation command execution MAY be supported
- Git hooks integration MAY be provided
- Advanced configuration options MAY be available

## Summary

A pragmatic tool that provides context-aware git worktree management with a focus on rebase workflows. Core commands (`list`, `create`, `cd`, `delete`) adapt their behavior based on the user's current context, with robust error handling and a simple configuration system. User experience and practical functionality are prioritized over technical complexity.