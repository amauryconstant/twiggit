# cmd/ — Cobra commands reference

Per the `cli-functional-core-shell` change, `cmd/` is the **composition
layer**: each command wires `internal/core`, `internal/git`, `internal/config`,
`internal/output`, and `internal/iostreams` through the `cmdutil.Factory`
without holding service-layer state of its own. The factory pattern is the
same one `gh`, `kubectl`, and `docker` use.

## Composition root

`main.go` constructs a `cmdutil.Factory`, defers panic recovery, sets up
`signal.NotifyContext` for SIGINT/SIGTERM, then dispatches into
`cmd.NewRootCommand(f).Execute()`. Exit code is `cmdutil.ExitCodeFor(err)`;
signal cancellation exits 130/143 directly (bypassing the formatter).

## Command shape

Each command follows the **Options + `runF`** pattern:

```go
type ListOptions struct {
    IO     *iostreams.IOStreams
    Config func() (*core.Config, error)
    Client func() (*git.Client, error)
    Ctx    context.Context
    All    bool
    Output string
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
    opts := &ListOptions{}
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List worktrees",
        Args:  cobra.NoArgs,
        RunE: func(cmd *cobra.Command, _ []string) error {
            opts.IO = f.IOStreams
            opts.Config = f.Config
            opts.Client = f.GitClient
            opts.Ctx = cmd.Context()
            opts.All, _ = cmd.Flags().GetBool("all")
            opts.Output, _ = cmd.Flags().GetString("output")
            if runF != nil {
                return runF(opts)
            }
            return runList(opts)
        },
    }
    cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Show all projects")
    return cmd
}

func runList(opts *ListOptions) error {
    // validate → execute → respond; pure core, then I/O adapter, then formatter.
}
```

`cmd.CommandConfig` is a type alias for `*cmdutil.Factory` so callers that
historically referenced `CommandConfig` compile unchanged
(`type CommandConfig = cmdutil.Factory` at `cmd/root.go:16`).

## Persistent flags

`cmdutil.AddPersistentFlags` registers `--output`, `--quiet`, and
`--verbose` on the root command so every subcommand inherits them. The
values land on `*cmdutil.GlobalOptions`; `root.PersistentPreRunE` mirrors
them onto `iostreams.IOStreams.Verbose` / `Quiet` so the verbose gate and
quiet-aware formatter see user-supplied values without re-parsing flags.

## Verbose output

`cmd.verbosef(ios, format, args...)` is a thin wrapper around
`iostreams.IOStreams.Verbosef` that tolerates a nil ios for early-startup
call sites. **There is no level distinction**: the previous `-v` (level 1)
and `-vv` (level 2) scheme is collapsed into a single boolean Verbose flag
per `cli-verbose-output`; `-vv` is treated as `-v` and the level-2
two-space indentation prefix is removed from message strings. New code
should call `opts.IO.Verbosef(...)` directly; the `cmd.verbosef` helper
exists only for nil-safety.

## Quiet mode

Global `--quiet/-q` flag suppresses non-essential output for scripting
scenarios. `cmd.isQuiet(cmd)` reads the bound flag value or falls back to
`ios.Quiet`. `ProgressReporter` automatically respects quiet mode.

## Error handling

`output.FormatError(ios.ErrOut, err, ios)` is the single formatter entry
point; dispatch is `errors.As`-driven across the four `core.Error` types
(`ValidationError`, `NotFoundError`, `OperationError`, `UsageError`).
Exit-code mapping is centralised in `cmdutil.ExitCodeFor(err)`:

| Error class | Exit code |
| ----------- | --------- |
| `*core.UsageError` (incl. cobra wrap) | `cmdutil.ExitUsage` (2) |
| `*core.ValidationError`, `*core.NotFoundError`, anything else | `cmdutil.ExitError` (1) |
| nil | `cmdutil.ExitOK` (0) |
| SIGINT (signal context cancelled) | 130 (bypasses formatter) |
| SIGTERM (signal context cancelled) | 143 (bypasses formatter) |

For command-level cobra usage failures, wrap with `core.UsageWrap(err)`
inside `SetFlagErrorFunc` or `Args` validators so dispatch lands on
`ExitUsage` instead of the default `ExitError`.

## Context-aware behavior

Commands adapt to the current context (`ContextProject`, `ContextWorktree`,
`ContextOutsideGit`) detected by `internal/git/context_resolver.go`. See
that package for detection rules and resolution order. The factory exposes
`f.GitContext()` as the lazy accessor.

## Shell completion

Carapace integration provides shell completion for all commands. Hidden
command: `twiggit _carapace <shell>` generates completion scripts.

| Shells | bash, zsh, fish, nushell, elvish, powershell, tcsh, oil, xonsh, cmd-clink |
| ------ | ---------------------------------------------------------------------------- |

**Wiring pattern:**

```go
carapace.Gen(cmd).PositionalCompletion(
    actionWorktreeTarget(f, WithExistingOnly()),
)
```

Helpers live in `cmd/suggestions.go`; see that file for `actionWorktreeTarget`,
`actionBranches`, and the `Cache(5s)` / `Timeout(...)` directives.

## Command specifications

### list
Alias: `ls` (Unix-style shortcut)
Output: Tabular format with branch, last commit, status (clean/dirty) or JSON for scripting
Flags:
- `--all/-a` (show all projects, override context)
- `--output/-o <format>`: `json` (default empty flag falls through to plain)

### create
Required: Project name (inferred), branch name, source branch (default: main)
Flags: `--source <branch>`, `-C, --cd`
Behavior: Create worktree, execute post-create hooks if `.twiggit.toml` configured, display hook failure warnings

### delete
Alias: `rm`
Safety checks: Uncommitted changes, current worktree status
Flags: `-f, --force`, `-m, --merged-only`, `-C, --cd`
Default behavior: Remove worktree + delete branch

### cd
Output: Absolute path to worktree (for shell wrapper)
Flags: None (target required)

### init
Default: Print shell wrapper to stdout (eval-safe, no metadata)
Optional: `[shell]` (bash|zsh|fish, auto-detected from $SHELL if omitted)
Flags: `-i, --install`, `-c, --config <path>`, `-f, --force`

### prune
Purpose: Delete merged worktrees for post-merge cleanup
Args: `[project/branch]` (optional)
Flags: `-n, --dry-run`, `-f, --force`, `-y, --yes`, `-d, --delete-branches`, `-a, --all`

## Testing

- **Unit tests**: `cmd/<command>_test.go` exercise `runX(opts)` directly via
  the `runF` injection point.
- **E2E tests**: `test/e2e/<command>_test.go` drive the built binary with
  Ginkgo/Gomega.
- **Integration tests**: `test/integration/` exercise I/O adapters against
  real git/config fixtures.

The `cmdutil.Factory` test seam (`runF`) is the canonical mock surface;
prefer it over hand-rolled mocks.
