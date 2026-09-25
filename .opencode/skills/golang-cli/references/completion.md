# Shell Completion

The completion API (`ValidArgsFunction`, `RegisterFlagCompletionFunc`, directives) → See `samber/cc-skills-golang@golang-spf13-cobra`.

## Contents

- [Cobra's built-in completion](#cobras-built-in-completion)
- [Carapace](#carapace-the-power-option)
- [Other completion libraries](#other-completion-libraries)
- [Distribution](#distribution-patterns)
- [Decision matrix](#decision-matrix)

## Cobra's Built-In Completion

Cobra generates completion scripts for bash, zsh, fish, and PowerShell via a `completion` subcommand. Static completions (commands, flags, `ValidArgs`) work out of the box.

**Dynamic completions** use `RegisterFlagCompletionFunc` and `ValidArgsFunction`:

```go
cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "table", "plain"}, cobra.ShellCompDirectiveNoFileComp
})

cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    client, err := f.Client() // completion funcs live in cmd/ and call adapters
    if err != nil {
        return nil, cobra.ShellCompDirectiveError
    }
    return client.BranchNames(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp
}
```

Completion runs on every Tab press, so keep these calls fast and silent: no prompts, nothing on stdout except candidates, diagnostics only under `MYAPP_DEBUG`.

**File path completions:**

```go
cmd.MarkFlagFilename("config", "yaml", "yml", "json")
cmd.MarkFlagDirname("output-dir")
```

**ShellCompDirective flags:**

| Directive | Effect |
| --- | --- |
| `ShellCompDirectiveDefault` | Default file completion after custom values |
| `ShellCompDirectiveNoFileComp` | Suppress file completion |
| `ShellCompDirectiveNoSpace` | No space after completion (for `key=` prefixes) |
| `ShellCompDirectiveFilterFileExt` | Filter files by extension |
| `ShellCompDirectiveFilterDirs` | Filter to directories only |

**Cobra's limitations:** No completion for elvish, nushell, xonsh. No built-in support for multi-part value completion (`--label=foo,bar,`). No caching for expensive completions.

## Carapace: The Power Option

[carapace-sh/carapace](https://github.com/carapace-sh/carapace) replaces Cobra's completion engine while keeping the command tree. Recommended when you need broad shell support or advanced completion features.

**Shell support:** bash, zsh, fish, elvish, oil, powershell, xonsh, nushell.

**Adoption** — in `NewRootCommand(f)`, after building the tree:

```go
root.CompletionOptions.DisableDefaultCmd = true // carapace provides `_carapace`
carapace.Gen(root)
```

**Key features beyond Cobra:**

- **Positional argument completion** — explicit per-position with context access
- **ActionMultiParts** — complete comma/colon-separated values independently
- **ActionExecCommand** — shell out to external commands for completions
- **Caching** — avoid repeated API calls during rapid tab-completion
- **Bridge system** — consume completions from other frameworks (Cobra, Click, Clap) and native shell scripts

## Other Completion Libraries

For non-Cobra CLIs: `posener/complete` (standalone, used by HashiCorp tools) or a framework's built-in generator (e.g. Kong from struct tags). See [libraries.md](./libraries.md#shell-completion).

## Distribution Patterns

**Runtime generation (recommended):** Provide a `myapp completion <shell>` command. Users source it in their shell rc:

```
eval "$(myapp completion zsh)"
```

This keeps completions in sync with the binary version.

**goreleaser integration:** Generate completion scripts as build artifacts. For Homebrew: `share/bash-completion/`, `share/zsh/site-functions/`, `share/fish/vendor_completions.d/`.

**XDG directories** for tools that manage their own completion installation:

- bash: `$XDG_DATA_HOME/bash-completion/completions/`
- fish: `$XDG_CONFIG_HOME/fish/completions/`

## Decision Matrix

| Need | Choice |
| --- | --- |
| Standard Cobra, bash/zsh/fish, simple completions | Cobra built-in |
| Multi-part values, caching, broad shell support | Carapace |
| Non-Cobra framework | `posener/complete` or framework-specific |
| Plugin system needing completions | Carapace bridge system |
