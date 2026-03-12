# Twiggit Fish Integration

Shell integration for twiggit with completions and directory navigation.

## Installation

### conf.d (Recommended)

Copy to your Fish conf.d directory:

```fish
cp /path/to/twiggit/contrib/fish/twiggit.fish ~/.config/fish/conf.d/
```

This will be automatically sourced on shell startup.

### Oh My Fish

```fish
omf install https://gitlab.com/amoconst/twiggit.git --path=contrib/fish
```

### Standalone

Add to your `~/.config/fish/config.fish`:

```fish
source /path/to/twiggit/contrib/fish/twiggit.fish
```

## What It Does

1. Checks if `twiggit` is installed (bails silently if not)
2. Sources the shell wrapper for `twiggit cd` navigation
3. Loads completions for all twiggit commands

## Manual Alternative

Add to your `~/.config/fish/config.fish`:

```fish
if type -q twiggit
    twiggit init fish | source
    twiggit _carapace fish | source
end
```
