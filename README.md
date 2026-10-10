# twiggit

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

Download from:
- **GitLab Releases**: https://gitlab.com/amoconst/twiggit/-/releases
- **GitHub Releases**: https://github.com/amauryconstant/twiggit/releases

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

### Verify Installation

After installation, verify everything works:

```bash
twiggit version
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

# Rebase a worktree onto its tracked base
twiggit rebase                       # Rebase current worktree
twiggit rebase myproject/feature     # Rebase a specific worktree
twiggit rebase --all                 # Rebase every worktree in the project
twiggit rebase --fetch               # Fetch the base, then rebase
twiggit rebase --continue feature    # Resume a paused rebase
twiggit rebase --abort feature       # Abort a paused rebase
twiggit rebase --set-base develop    # Persist a new tracked base

# Sync a project's tracking refs from its remote
twiggit sync                         # Fetch origin for current project
twiggit sync --all                   # Sync every project
twiggit sync --remote upstream       # Fetch from a non-default remote
twiggit sync --rebase                # Fetch and rebase every worktree
twiggit sync --fetch-only            # Skip the rebase walk

# Show the diagnostic view of every worktree: ahead/behind counts
# against the tracked base, merge readiness, dirty state, last-commit
# date, and a stale flag. Pipeable via --output json|table|plain.
twiggit status                       # Status for the current project
```

## status

Show the diagnostic view of every worktree: ahead/behind counts against
the tracked base, merge readiness, dirty state, last-commit date, and a
stale flag. Output is human-readable by default and pipeable via
`--output json|table|plain`.

```sh
twiggit status                       # Status for the current project
twiggit status --all                 # Status across every project
twiggit status --output json         # JSON for scripts
twiggit status --stale-behind 5      # Override the stale threshold
```

The columns are `BRANCH`, `PATH`, `AHEAD`, `BEHIND`, `BASE`, `MERGED`,
`DIRTY`, `STALE`. A worktree is `stale` when its `Behind` count meets
`status.stale_behind` (default 20) or its last commit is older than
`status.stale_days` (default 30). Set both to `0` to disable the
heuristic.

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

### Pre-rebase, post-rebase, and post-sync hooks

Twiggit also runs hooks before rebase, after a clean rebase, and
after a sync. Pre-rebase hooks that exit non-zero abort the rebase
(useful for stashing or sanity checks); post-rebase and post-sync
hook failures are recorded as warnings.

```toml
[hooks.pre-rebase]
commands = ["git stash"]

[hooks.post-rebase]
commands = ["go build ./..."]

[hooks.post-sync]
commands = ["echo synced"]
```

The rebase and sync hooks also receive the variables above plus:

| Variable | Description |
|----------|-------------|
| `TWIGGIT_REBASE_BASE` | Branch being rebased onto (pre/post-rebase) |
| `TWIGGIT_REBASE_OLD_TIP` | Pre-rebase tip (post-rebase) |
| `TWIGGIT_REBASE_NEW_TIP` | Post-rebase tip (post-rebase) |
| `TWIGGIT_REBASE_RESULT` | "clean" / "conflicted" / "nothing-to-do" (post-rebase) |
| `TWIGGIT_SYNC_REMOTE` | Remote that was fetched (post-sync) |
| `TWIGGIT_SYNC_BRANCH` | Branch that was fetched (post-sync) |

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

## Debugging

`twiggit` honours the standard `TWIGGIT_DEBUG=1` env var to enable structured
debug logs (Go's `slog.Default` writes to the binary's stderr). For deeper
runtime diagnostics, the standard `GODEBUG` knob is also available and useful
when investigating `twiggit` itself:

| Env var | Purpose |
| --- | --- |
| `GODEBUG=gctrace=1` | Print GC traces to stderr at every collection; useful when `twiggit` is slow and you suspect GC pressure. |
| `GODEBUG=schedtrace=10000` | Emit a scheduler trace every 10 000 ms; pairs well with `schedtrace=` to profile goroutine scheduling under concurrent worktree operations. |
| `GODEBUG=asyncpreemptoff=1` | Disable async preemption; surfaces data races that preemption would otherwise hide when running `go test -race`. |

Example: `TWIGGIT_DEBUG=1 GODEBUG=gctrace=1 ./twiggit list` prints both the
structured debug channel and GC traces on stderr.
