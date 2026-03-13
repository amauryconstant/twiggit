# twiggit

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/amoconst/twiggit)](https://goreportcard.com/report/gitlab.com/amoconst/twiggit)
[![GoDoc](https://pkg.go.dev/badge/gitlab.com/amoconst/twiggit)](https://pkg.go.dev/gitlab.com/amoconst/twiggit)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitLab CI](https://gitlab.com/amoconst/twiggit/badges/main/pipeline.svg)](https://gitlab.com/amoconst/twiggit/-/pipelines)

Pragmatic git worktree management tool with focus on rebase workflows.

## Installation

### Quick Install (Linux/macOS)
```bash
curl -fsSL https://gitlab.com/amoconst/twiggit/-/raw/main/install.sh | bash
```

The install script will prompt you to:
- Install the twiggit binary
- Enable shell completions (recommended)
- Set up directory navigation for `twiggit cd`

### Manual Install
Download from: https://gitlab.com/amoconst/twiggit/-/releases

 After manual installation, run:
 ```bash
 # Enable completions
 twiggit completion zsh > ~/.local/share/zsh/site-functions/_twiggit  # zsh
 # or
 echo 'source <(twiggit completion bash)' >> ~/.bashrc  # bash

 # Enable directory navigation
 twiggit init                    # Auto-detects shell and config file
 # or
 twiggit init ~/.zshrc           # Specify config file explicitly
 # or
 twiggit init --shell=zsh        # Specify shell explicitly
 ```

## Shell Integration

Shell integration enables:
- **Directory navigation**: `twiggit cd <branch>` changes to the worktree
- **Completions**: TAB-autocomplete for all commands and flags

### Using Plugin Files (Recommended)

Shell plugins are available in `contrib/` for easy integration with plugin managers.

See [contrib/zsh/README.md](contrib/zsh/README.md), [contrib/bash/README.md](contrib/bash/README.md), or [contrib/fish/README.md](contrib/fish/README.md) for detailed instructions.

### Manual Setup

If you prefer manual configuration:

**Zsh** (add to `~/.zshrc`):
```zsh
if (( $+commands[twiggit] )); then
  eval "$(twiggit init zsh)"
  source <(twiggit _carapace zsh)
fi
```

**Bash** (add to `~/.bashrc`):
```bash
if command -v twiggit &>/dev/null; then
  eval "$(twiggit init bash)"
  source <(twiggit _carapace bash)
fi
```

**Fish** (add to `~/.config/fish/config.fish`):
```fish
if type -q twiggit
    twiggit init fish | source
    twiggit _carapace fish | source
end
```

Restart your shell after adding the configuration.

## Quick Start

```bash
# Verify installation
twiggit version

# List worktrees in current project
twiggit list

# Create a new worktree
twiggit create feature/my-new-feature

# Navigate to a worktree (requires setup-shell)
twiggit cd feature/my-new-feature

# Delete a worktree
twiggit delete feature/old-feature

# Prune merged worktrees
twiggit prune --dry-run              # Preview what would be deleted
twiggit prune                        # Delete merged worktrees in current project
twiggit prune --all                  # Prune across all projects
```

## Post-Create Hooks

Twiggit can execute commands automatically after creating a worktree. This is useful for running project setup commands like `mise trust` or `npm install`.

### Configuration

Create a `.twiggit.toml` file in your repository root:

```toml
[hooks.post-create]
commands = [
    "mise trust",
    "npm install",
]
```

When you run `twiggit create`, these commands will execute in the new worktree directory with the following environment variables:

| Variable | Description |
|----------|-------------|
| `TWIGGIT_WORKTREE_PATH` | Path to the new worktree |
| `TWIGGIT_PROJECT_NAME` | Project identifier |
| `TWIGGIT_BRANCH_NAME` | Name of the new branch |
| `TWIGGIT_SOURCE_BRANCH` | Branch the worktree was created from |
| `TWIGGIT_MAIN_REPO_PATH` | Path to the main repository |

### Behavior

- Commands run sequentially in the worktree directory
- If a command fails, remaining commands continue to execute
- Failures are displayed as warnings (worktree creation still succeeds)
- No `.twiggit.toml` file = no hooks executed

### Security Warning

**Important**: The `.twiggit.toml` file can execute arbitrary commands on your system. Always review this file before trusting a repository:

```bash
# Check for hooks before creating worktrees in a new repo
cat .twiggit.toml
```

Hooks are opt-in only—without a `.twiggit.toml` file, no commands are executed.
