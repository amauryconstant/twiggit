# Twiggit Zsh Plugin

Lazy-loaded shell integration for twiggit with Oh My Zsh, antidote, zinit, or standalone sourcing.

## Features

- **Lazy completions**: Completions load only on first TAB press, keeping shell startup fast
- **Directory navigation**: Enables `twiggit cd` to change directories
- **Framework agnostic**: Works with any zsh plugin manager

## Installation

### Oh My Zsh

```zsh
git clone https://gitlab.com/amoconst/twiggit.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/twiggit
cd ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/twiggit
git sparse-checkout set contrib/zsh
```

Add `twiggit` to your plugins array in `~/.zshrc`:

```zsh
plugins=(... twiggit)
```

### Antidote

Add to your `.zsh_plugins.txt`:

```
https://gitlab.com/amoconst/twiggit.git contrib/zsh
```

### Zinit

```zsh
zinit light-mode for"contrib/zsh" https://gitlab.com/amoconst/twiggit.git
```

### Znap

```zsh
znap source https://gitlab.com/amoconst/twiggit.git contrib/zsh
```

### Standalone

Add to your `~/.zshrc`:

```zsh
source /path/to/twiggit/contrib/zsh/twiggit.plugin.zsh
```

## What It Does

1. Checks if `twiggit` is installed (bails silently if not)
2. Sources the shell wrapper for `twiggit cd` navigation
3. Sets up lazy-loaded completions (only loads on first TAB)

## Manual Alternative

If you prefer not to use a plugin, you can add this to your `~/.zshrc`:

```zsh
if (( $+commands[twiggit] )); then
  eval "$(twiggit init zsh)"
  source <(twiggit _carapace zsh)
fi
```
