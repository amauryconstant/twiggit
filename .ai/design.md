# Twiggit Design Specification

## Purpose

A pragmatic tool for managing git worktrees with a focus on rebase workflows. Context-aware operations SHALL be provided for creating, listing, switching, and deleting worktrees across multiple projects.

## Directory Structure & Defaults

### Default Paths
- **Projects**: `$HOME/Projects/<project-name>/`
- **Workspaces**: `$HOME/Workspaces/<project-name>/<worktree-branch-name>/`

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
**Output format**: Simple text list with:
- Branch name
- Last commit information  
- Working directory status (clean/dirty)

**Context behavior**:
- From project folder: List worktrees for current project
- From workspace folder: List worktrees for current project
- From outside git: List all worktrees across all projects

**Flags**:
- `--all`: Show worktrees from all projects (overrides context)
- **Note**: `-C / --change-dir` flag SHOULD NOT be supported for list command

#### `create` - Create a new git worktree
**Required parameters**:
- Project name (inferred from context when possible)
- New branch name
- Source branch name (defaults to `main`)

**Parameter Inference Rules**:
1. **Project name** (inferred in this order):
   - From current directory if in project folder (`.git/` found)
   - From workspace path if in workspace folder (`$HOME/Workspaces/<project>/`)
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
- **Success**: Absolute path to worktree directory (e.g., `/home/user/Workspaces/project-name/feature-branch`)
- **Error**: Error message to stderr, exit with non-zero code

**Shell Wrapper**:
- SHALL intercept `twiggit cd` calls and change shell directory
- SHALL provide escape hatch with `builtin cd` for shell built-in
- SHALL warn when overriding shell built-in `cd`
- SHALL pass through all other commands unchanged
- SHALL be automatically installed via `twiggit setup-shell` command

**Context behavior**:
- From project folder: Change SHALL occur to specified worktree of current project
- From workspace folder: Change SHALL occur to different worktree of current project
- From outside git: Project and worktree specification SHALL be required

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

**Output Format**:
- **Success**: Installation instructions, escape hatch usage, and next steps
- **Error**: Specific error message about detection or installation failure

**Shell Wrapper Features**:
- SHALL intercept `twiggit cd` calls and change shell directory
- SHALL provide `builtin cd` for shell built-in access
- SHALL warn when overriding shell built-in `cd`
- SHALL pass through all other commands unchanged

### Help Command

#### `help` - Display help text
- Basic help text SHALL be returned
- Usage patterns and available commands SHALL be shown

## Context Detection & Behavior

### Context Detection Rules
1. **Project folder**: `.git/` directory found in current or parent directories
   - Directory tree WILL be traversed up until finding `.git/` or reaching filesystem root
   - First `.git/` found WILL be used (closest to current directory)
2. **Workspace folder**: Path matches `$HOME/Workspaces/<project>/<branch>/` pattern  
   - Exact pattern matching SHALL be used with configurable base directories
   - Alternative workspace detection patterns MAY be supported in future
   - Workspace SHALL be validated to contain valid git worktree
3. **Outside git**: No `.git/` found and not in workspace pattern

### Edge Case Handling
- **Nested directories**: Context SHALL be determined by closest valid parent directory
- **Multiple `.git` directories**: First one found during upward traversal SHALL be used
- **Invalid workspace pattern**: SHALL be treated as "outside git" context
- **Broken git repositories**: SHALL be detected as invalid context, error and exit SHALL occur

### Context Behavior
- **From project folder**: Command SHALL be applied to current project
- **From workspace folder**: Command SHALL be applied to current worktree or encapsulating project
- **From outside git**: Explicit project/worktree specification SHALL be required or `--all` SHALL be used

### Cross-context Operations
- Positional arguments SHALL be used for project and worktree specification
- `--all` flag SHALL be used for operations across all projects
- Git CLI patterns SHALL be followed for argument handling

## Error Handling

### Error Philosophy
- The system WILL fail fast when conflicts cannot be resolved
- Specific, actionable error messages SHALL be provided for debugging
- Interactive processes SHALL NOT be used - all information SHALL be available or failure SHALL occur

### Error Scenarios
- Invalid git repository SHALL be detected and reported with actionable error message
- Network errors SHALL be caught and reported with connection details
- Permission issues SHALL be detected and reported with specific path information
- Ambiguous contexts SHALL be identified and resolved with explicit user guidance
- Missing directories SHALL be detected and reported with path details

### Return Codes
- The system WILL use standard POSIX return codes
- `0`: Success
- `1`: General error
- `2`: Misuse/invalid arguments
- Other codes SHALL be used as appropriate for specific error conditions

## Configuration

### Configuration System
- The system WILL use XDG Base Directory specification for config file location
- TOML format SHALL be supported exclusively
- Configuration SHALL be applied in priority order: defaults → config file → environment variables → command flags
- Configuration validation SHALL occur during startup

### Configurable Options
- **Directory paths**: Defaults for projects and workspaces directories SHALL be overridden
- **Default source branch**: Default `main` branch for create command SHALL be overridden
- See [implementation.md](./implementation.md) for detailed configuration examples and file location

## CLI Features

### Help System
- Basic help text format SHALL be provided for all commands
- Usage examples and flag descriptions SHALL be included
- Context-aware help SHALL be provided when possible

### Version Information
- `--version` flag SHALL display version information
- Semantic versioning SHALL be followed

### Shell Completion
- Support for bash, zsh, and fish shell completion MAY be provided
- Completion scripts MAY be generated for common shells

### Verbosity
- Default POSIX output behavior SHALL be used
- Additional logging or verbosity controls MAY NOT be provided in first iteration

## Future Features

### Maintenance Commands
- The system MAY support maintenance commands like prune, clean, status, repair

### Advanced Features
- Post-creation command execution MAY be supported
- Git hooks integration MAY be provided
- Advanced configuration options MAY be available

## Command Requirements

### List Command Requirements
- Worktrees SHALL be displayed in tabular format with branch, commit, and status columns
- Project context SHALL be inferred from current directory when possible
- --all flag SHALL override context detection
- -C/--change-dir flag SHALL NOT be supported for list operations

### Create Command Requirements
- New branch name SHALL be required as positional argument
- Project name SHALL be inferred using the specified priority order
- Source branch SHALL default to 'main' unless overridden
- Worktree creation SHALL NOT proceed if target directory already exists

### Delete Command Requirements
- Safety checks SHALL be performed before deletion
- Uncommitted changes SHALL prevent deletion unless --force flag is used
- Current worktree SHALL NOT be deletable
- Branch deletion SHALL occur by default unless --keep-branch flag is used

### Setup-Shell Command Requirements
- Current shell detection SHALL be performed automatically
- Wrapper function SHALL be generated for detected shell
- Shell configuration file SHALL be modified appropriately
- User SHALL be instructed to restart shell or source configuration

## Error Handling Requirements
- Non-zero exit codes SHALL be used for all error conditions
- Specific, actionable error messages SHALL be provided
- Interactive error recovery SHALL NOT be implemented
- All inputs SHALL be validated before execution
- POSIX-compliant exit codes SHALL be used (0=success, 1=general error, 2=misuse)

## Configuration Requirements
- TOML configuration format SHALL be supported exclusively
- Config file SHALL be located using XDG Base Directory specification
- Configuration SHALL be applied in priority order: defaults → config file → environment variables → command flags
- All configuration values SHALL be validated during startup

## Validation Requirements
- Project names SHALL follow GitHub repository naming rules
- Branch names SHALL follow git branch naming rules
- Directory paths SHALL be validated to exist and be accessible
- Source branches SHALL be validated to exist in repository
- All user inputs SHALL be sanitized and validated

## Summary

A pragmatic tool that provides context-aware git worktree management with a focus on rebase workflows. Core commands (`list`, `create`, `cd`, `delete`) adapt their behavior based on the user's current context, with robust error handling and a simple configuration system. User experience and practical functionality are prioritized over technical complexity.