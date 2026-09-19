# 12 — TUI and Interactive Patterns

This section covers interactive CLI patterns. The architecture described here is designed to be **compatible** with the non-interactive architecture from [01-architecture.md](./01-architecture.md) — you can add interactivity to individual commands without restructuring the rest of your CLI.

---

## When to Use TUI vs. Plain CLI

If you just execute a function and produce output, plain CLI is fine. TUI adds visual richness but costs development time and complexity.

**TUI is justified for:** Interactive workflows (wizards, dashboards, file browsers), long-running monitoring, multi-step processes where the user needs to make decisions mid-flow.

**Plain CLI is sufficient for:** Single-action commands, batch processing, scripted usage, CI/CD.

---

## Flags-First, Prompt-as-Fallback

The pattern that makes a prompting command _scriptable_ (`SKILL.md` §The
`scriptable` Rule): define every input as a flag; if a required one is missing
**and** stdin is an interactive TTY, prompt; otherwise (pipe, CI) fail with a
clear error rather than blocking.

```go
func runCreate(opts *CreateOptions) error {
    name := opts.Name

    if name == "" && opts.IO.IsInteractive() {
        if err := huh.NewInput().Title("Project name").Value(&name).Run(); err != nil {
            return err
        }
    }
    if name == "" {
        return fmt.Errorf("--name is required (or run interactively in a terminal)")
    }
    // proceed with name
}
```

This ensures the CLI works in both interactive and scripted contexts. The `--non-interactive` / `--yes` flag provides an explicit opt-out from prompts.

---

## Wizard Mode

For complex setup commands (like `init`), run a full interactive wizard via `huh` forms. Gate the entire wizard behind TTY detection:

```go
func runInit(io *iostreams.IOStreams) error {
    if !io.IsInteractive() {
        return fmt.Errorf("init requires an interactive terminal (or provide all flags)")
    }
    
    var config InitConfig
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().Title("Project name").Value(&config.Name),
            huh.NewSelect[string]().Title("Language").
                Options(huh.NewOption("Go", "go"), huh.NewOption("Rust", "rust")).
                Value(&config.Language),
        ),
    )
    if err := form.Run(); err != nil {
        return err
    }
    return createProject(config)
}
```

---

## Bubble Tea Architecture

Bubble Tea implements the Elm Architecture (TEA):

- **Model** — application state (a struct)
- **Update** — handles messages, returns new model + commands
- **View** — renders model to string

```go
type Model interface {
    Init() Cmd
    Update(msg Msg) (Model, Cmd)
    View() string
}
```

**Composability:** Models can contain sub-models. Propagate `Update()` to children for complex multi-view apps. The architecture naturally supports state machines for navigation between views.

### Ecosystem

The Charm ecosystem composes: `bubbletea` (Elm-Architecture core), `bubbles`
(pre-built components — text input, lists, spinners, viewports, tables),
`lipgloss` (CSS-like styling), `glamour` (terminal markdown), and `huh`
(forms, standalone or embedded in Bubble Tea). See
[14-libraries.md](./14-libraries.md#tui-frameworks) and
[14-libraries.md](./14-libraries.md#output-formatting) for the full comparison.

### Integration with Non-TUI Architecture

`huh` can run standalone (`form.Run()`) or be embedded as a Bubble Tea model. For commands that need both static output and interactive prompts, use standalone `huh`. Reserve full Bubble Tea for commands that need persistent TUI state (dashboards, monitoring, file browsers).

The key compatibility rule: TUI commands still receive `IOStreams` (or the Factory at Tier 3) and produce results through the same output formatting pipeline. The TUI is an alternative *input* mechanism, not a replacement for the Respond step.

---

## Interactive Prompt Libraries

Recommended: `charmbracelet/huh` (forms + prompts, standalone or embedded in
Bubble Tea, generics, accessible mode). See
[14-libraries.md](./14-libraries.md#interactive-prompts) for the full comparison
(`survey`, `promptui`, `go-prompt`, `gum`).

---

## gum for Shell Scripts

`gum` provides interactive prompts (confirm, choose, input, filter, spin) as standalone binaries callable from shell scripts. It bridges the gap between shell and TUI without requiring Go code:

```bash
NAME=$(gum input --placeholder "Project name")
LANG=$(gum choose "Go" "Rust" "Python")
gum confirm "Create project $NAME in $LANG?" && create_project "$NAME" "$LANG"
```
