# 08 — Shell Completion

---

## Cobra's Built-In Completion

Cobra generates completion scripts for bash, zsh, fish, and PowerShell via a `completion` subcommand. Static completions (commands, flags, `ValidArgs`) work out of the box.

**Dynamic completions** use `RegisterFlagCompletionFunc` and `ValidArgsFunction`:

```go
cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "table", "plain"}, cobra.ShellCompDirectiveNoFileComp
})

cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return fetchProjectNames(toComplete), cobra.ShellCompDirectiveNoFileComp
}
```

**File path completions:**

```go
cmd.MarkFlagFilename("config", "yaml", "yml", "json")
cmd.MarkFlagDirname("output-dir")
```

**ShellCompDirective flags:**

| Directive | Effect |
|-----------|--------|
| `ShellCompDirectiveDefault` | Default file completion after custom values |
| `ShellCompDirectiveNoFileComp` | Suppress file completion |
| `ShellCompDirectiveNoSpace` | No space after completion (for `key=` prefixes) |
| `ShellCompDirectiveFilterFileExt` | Filter files by extension |
| `ShellCompDirectiveFilterDirs` | Filter to directories only |

**Cobra's limitations:** No completion for elvish, nushell, xonsh. No built-in support for multi-part value completion (`--label=foo,bar,`). No caching for expensive completions.

---

## Carapace: The Power Option

[carapace-sh/carapace](https://github.com/carapace-sh/carapace) replaces Cobra's completion engine while keeping the command tree. Recommended when you need broad shell support or advanced completion features.

**Shell support:** bash, zsh, fish, elvish, oil, powershell, xonsh, nushell.

**Adoption:**

```go
import "github.com/carapace-sh/carapace"

var rootCmd = &cobra.Command{
    Use: "myapp",
    CompletionOptions: cobra.CompletionOptions{
        DisableDefaultCmd: true,
    },
}

func init() {
    carapace.Gen(rootCmd)
}
```

**Key features beyond Cobra:**

- **Positional argument completion** — explicit per-position with context access
- **ActionMultiParts** — complete comma/colon-separated values independently
- **ActionExecCommand** — shell out to external commands for completions
- **Caching** — avoid repeated API calls during rapid tab-completion
- **Bridge system** — consume completions from other frameworks (Cobra, Click, Clap) and native shell scripts

---

## Other Completion Libraries

For non-Cobra CLIs: `posener/complete` (standalone, used by HashiCorp tools) or a
framework's built-in generator (e.g. Kong from struct tags). See
[14-libraries.md](./14-libraries.md#shell-completion).

---

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

---

## Decision Matrix

| Need | Choice |
|------|--------|
| Standard Cobra, bash/zsh/fish, simple completions | Cobra built-in |
| Multi-part values, caching, broad shell support | Carapace |
| Non-Cobra framework | `posener/complete` or framework-specific |
| Plugin system needing completions | Carapace bridge system |

---

**See also:** `samber/cc-skills-golang@golang-spf13-cobra` — Cobra's `ValidArgsFunction`, `RegisterFlagCompletionFunc`, `ShellCompDirective` API, and completion command setup
