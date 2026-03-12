# Twiggit Bash Integration

Shell integration for twiggit with completions and directory navigation.

## Installation

### Standalone

Add to your `~/.bashrc`:

```bash
source /path/to/twiggit/contrib/bash/twiggit.bash
```

### With Bash-It

```bash
ln -s /path/to/twiggit/contrib/bash/twiggit.bash ~/.bash_it/aliases/enabled/twiggit.aliases.bash
```

## What It Does

1. Checks if `twiggit` is installed (bails silently if not)
2. Sources the shell wrapper for `twiggit cd` navigation
3. Loads completions for all twiggit commands

## Manual Alternative

Add to your `~/.bashrc`:

```bash
if command -v twiggit &>/dev/null; then
  eval "$(twiggit init bash)"
  source <(twiggit _carapace bash)
fi
```
