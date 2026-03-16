# Config Path Expansion

## Purpose

The config loader expands environment variables and tilde in path fields during loading, before validation. This allows users to write portable configuration files using variables like `$HOME`, `${HOME}`, or `~` instead of hard-coded absolute paths.

## Requirements

### Requirement: Environment variable expansion in config paths

The config loader SHALL expand environment variables and tilde in path fields during loading, before validation.

#### Scenario: Dollar-sign variable expansion
- **WHEN** config contains `projects_dir = "$HOME/Projects"`
- **THEN** the loaded config SHALL have `projects_dir = "/home/user/Projects"` (expanded)

#### Scenario: Curly-brace variable expansion
- **WHEN** config contains `worktrees_dir = "${HOME}/Worktrees"`
- **THEN** the loaded config SHALL have `worktrees_dir = "/home/user/Worktrees"` (expanded)

#### Scenario: Tilde expansion
- **WHEN** config contains `projects_dir = "~/Projects"`
- **THEN** the loaded config SHALL have `projects_dir = "/home/user/Projects"` (expanded)

#### Scenario: Mixed variables in path
- **WHEN** config contains `worktrees_dir = "$XDG_DATA_HOME/worktrees"`
- **AND** `XDG_DATA_HOME` is set to `/home/user/.local/share`
- **THEN** the loaded config SHALL have `worktrees_dir = "/home/user/.local/share/worktrees"`

#### Scenario: Absolute path unchanged
- **WHEN** config contains `projects_dir = "/absolute/path/Projects"`
- **THEN** the loaded config SHALL have `projects_dir = "/absolute/path/Projects"` (unchanged)

#### Scenario: Empty env var leaves literal
- **WHEN** config contains `projects_dir = "$UNDEFINED_VAR/Projects"`
- **AND** `UNDEFINED_VAR` is not set in environment
- **THEN** the loaded config SHALL have `projects_dir = "/Projects"` (empty string expansion)

### Requirement: Expansion applies to all path fields

The config loader SHALL expand environment variables in all path-type config fields.

#### Scenario: Projects directory expansion
- **WHEN** config contains `projects_dir = "$HOME/Projects"`
- **THEN** the loaded config SHALL have expanded `projects_dir`

#### Scenario: Worktrees directory expansion
- **WHEN** config contains `worktrees_dir = "$HOME/Worktrees"`
- **THEN** the loaded config SHALL have expanded `worktrees_dir`

#### Scenario: Backup directory expansion
- **WHEN** config contains `[shell.wrapper] backup_dir = "~/.config/twiggit/backups"`
- **THEN** the loaded config SHALL have expanded `backup_dir`

### Requirement: Expansion occurs before validation

The config loader SHALL expand paths after loading TOML but before calling validation.

#### Scenario: Validation receives expanded paths
- **WHEN** config contains `projects_dir = "$HOME/Projects"`
- **THEN** validation SHALL receive `/home/user/Projects` (already expanded)
- **AND** `filepath.IsAbs()` SHALL return `true`
