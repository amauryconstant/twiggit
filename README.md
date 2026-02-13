# twiggit

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

## Setup

### Enable Shell Completions

Completions allow TAB-autocomplete for all twiggit commands and flags.

**Bash:**
```bash
echo 'source <(twiggit completion bash)' >> ~/.bashrc
source ~/.bashrc
```

**Zsh:**
```bash
# Standard zsh
twiggit completion zsh > ~/.local/share/zsh/site-functions/_twiggit

# Or with plugin managers (antidote, oh-my-zsh, etc.)
twiggit completion zsh > ~/.config/zsh/.zfunctions/_twiggit
# Reload: autoload -Uz compinit && compinit
```

**Fish:**
```bash
twiggit completion fish > ~/.config/fish/completions/twiggit.fish
```

### Enable Directory Navigation

The `twiggit cd` command requires a shell wrapper to change directories:

```bash
twiggit init                    # Auto-detect shell and config file
# or
twiggit init ~/.zshrc           # Specify config file explicitly
# or
twiggit init --shell=zsh        # Specify shell explicitly
```

This installs a wrapper that:
- Intercepts `twiggit cd` and changes to the target directory
- Preserves `builtin cd` for normal navigation
- Passes through all other twiggit commands

Restart your shell after running `init`.

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

## Documentation

For development and usage documentation, see [AGENTS.md](AGENTS.md).

## Troubleshooting

**twiggit command not found**
- Ensure the installation directory is in your PATH
- Run: `which twiggit` to locate it
- Add installation directory to PATH if needed

**TAB completion not working**
- Verify the completion script is in your shell's fpath
- Run the appropriate completion setup command from [Setup](#setup)
- Restart your shell after installing completions

**twiggit cd doesn't change directory**
- Run: `twiggit init` (auto-detects shell)
- Or run: `twiggit init --shell=<your-shell>` (explicit)
- Restart your shell after running init

**Permission denied during installation**
- Try with sudo: `sudo bash <(curl -fsSL https://gitlab.com/amoconst/twiggit/-/raw/main/install.sh)`
- Or create directory manually: `mkdir -p ~/.local/bin`
- Or install to a writable directory and add to PATH
