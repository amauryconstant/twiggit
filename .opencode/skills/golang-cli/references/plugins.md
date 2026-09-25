# Plugin Architecture

How to choose an extension model and wire PATH plugins. Library picks per model → `libraries.md`.

## Four Models

### Model A: Git-Style PATH Plugins

**Used by:** `gh`, Git, Helm, kubectl (via krew)

The host CLI looks for executables named `<prefix>-<plugin>` on `$PATH`. When the user runs `myapp my-plugin`, the host execs `myapp-my-plugin` as a subprocess. Communication is via stdin/stdout/stderr and exit codes.

**Pros:** Zero coupling, any language, trivial to develop and distribute, plugins can't crash the host.

**Cons:** No structured data exchange (just text/JSON on stdout), no callbacks into host, no shared state, limited to "run and exit" semantics.

### Model B: HashiCorp go-plugin (RPC/gRPC)

**Used by:** Terraform, Vault, Consul, Nomad, Packer, Boundary, Waypoint

Plugins are separate binaries that launch as subprocesses and communicate over local RPC (net/rpc or gRPC). The host and plugin share interface definitions. The host `Dispenses` a plugin by name and gets back an RPC client implementing the agreed interface.

**Pros:** Typed interface contracts, bidirectional communication (plugins can call back into the host), protocol versioning, crash isolation, language-agnostic via gRPC, stdout/stderr syncing, built-in logging.

**Cons:** Significant boilerplate (RPC scaffolding, protobuf for gRPC), localhost-only, more complex development and testing. ~30-50μs overhead per RPC call vs. in-process.

### Model C: Go's `plugin` Package (Shared Libraries)

Host and plugin must be built with the identical toolchain and identical versions of every shared dependency, and Windows is unsupported. In-process speed rarely outweighs that fragility; prefer Model A or B.

### Model D: Embedded Interpreters

Options: Lua (`gopher-lua`), JavaScript, WASM, CEL (`google/cel-go`).

**Use case:** users who are not Go developers extending the app in small ways — filters, transforms, policy expressions. Full plugin systems use Model A or B.

## Decision Framework

| Need | Model |
| --- | --- |
| Simple command extensions, any language | Git-style PATH |
| Rich typed interface, bidirectional calls | HashiCorp go-plugin |
| User-defined scripting/rules | Embedded interpreter (Lua, WASM, CEL) |

## PATH Plugin Implementation

Discovery:

```go
func discoverPlugins(prefix string) []string {
    var plugins []string
    for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
        entries, _ := os.ReadDir(dir)
        for _, e := range entries {
            if strings.HasPrefix(e.Name(), prefix+"-") {
                plugins = append(plugins, strings.TrimPrefix(e.Name(), prefix+"-"))
            }
        }
    }
    return plugins
}
```

Execution:

```go
func execPlugin(ctx context.Context, ios *iostreams.IOStreams, name string, args []string) error {
    cmd := exec.CommandContext(ctx, "myapp-"+name, args...)
    cmd.Stdin, cmd.Stdout, cmd.Stderr = ios.In, ios.Out, ios.ErrOut
    return cmd.Run() // *exec.ExitError carries the plugin's exit code
}
```

To pass the plugin's own exit code through, add a registry entry (→ `errors.md`) that maps `*exec.ExitError` to `exitErr.ExitCode()` and skip formatting it — the plugin already printed its error.

The `gh` CLI registers discovered extensions as Cobra subcommands so they appear in help text and benefit from shell completion.

## Completion for Plugins

Carapace's bridge system can consume completions from plugin binaries that use different CLI frameworks (Cobra, Click, Clap) or native shell completion scripts. This is the strongest argument for carapace if you have a plugin system.

→ See `samber/cc-skills-golang@golang-design-patterns` for lifecycle and shutdown patterns in a plugin host.
