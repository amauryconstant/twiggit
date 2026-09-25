# Prompts and TUI

Interactivity is added per command without restructuring the CLI: a TUI is another way to collect input, and results still flow through the command's normal Respond step.

## Contents

- [Plain CLI or TUI](#plain-cli-or-tui)
- [Flags first, prompt as fallback](#flags-first-prompt-as-fallback)
- [Confirmation](#confirmation)
- [Wizards](#wizards)
- [Bubble Tea](#bubble-tea)
- [gum for shell scripts](#gum-for-shell-scripts)

## Plain CLI or TUI

| TUI earns its cost for                  | Plain CLI covers       |
| --------------------------------------- | ---------------------- |
| Wizards, dashboards, file browsers      | Single-action commands |
| Long-running monitoring                 | Batch processing       |
| Multi-step flows with decisions mid-way | Scripted use, CI/CD    |

## Flags first, prompt as fallback

The pattern behind _scriptable_: every input is a flag; when one is missing **and** the session is interactive, prompt; otherwise return a usage error.

```go
func runCreate(opts *CreateOptions) error {
    if opts.Name == "" && opts.IO.IsInteractive() {
        if err := huh.NewInput().Title("Project name").Value(&opts.Name).Run(); err != nil {
            return err
        }
    }
    if opts.Name == "" {
        return &core.UsageError{Message: "--name is required when not running in a terminal"}
    }
    // proceed
}
```

A `--no-input` flag forces the non-interactive path even on a terminal.

## Confirmation

Shared prompt helpers live in `internal/output/prompt.go`; prompts render on stderr so stdout stays clean:

```go
func Confirm(ios *iostreams.IOStreams, message string) (bool, error) {
    var ok bool
    err := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title(message).Value(&ok))).
        WithOutput(ios.ErrOut).Run()
    return ok, err
}
```

The destructive-operation flow built on it → `command-ux.md`.

## Wizards

Setup commands (`init`) can run a full `huh` form. Gate the wizard on `IsInteractive()` and give every field a flag, so `myapp init --name x --lang go` works in CI.

## Bubble Tea

Bubble Tea implements the Elm Architecture: a **Model** (state), **Update** (message → new model + commands), and **View** (model → string). Models nest; parents forward `Update` to children, which gives natural view-to-view state machines.

- Use standalone `huh` (`form.Run()`) for prompts inside an otherwise plain command.
- Reserve full Bubble Tea for persistent screens: dashboards, monitors, browsers.
- TUI commands still receive `IOStreams` and hand their result to the same formatters.

The Charm stack: `bubbletea` (core), `bubbles` (inputs, lists, spinners, tables), `lipgloss` (styling), `glamour` (markdown), `huh` (forms). Alternatives → `libraries.md`.

## gum for shell scripts

`gum` offers prompts as standalone binaries for shell scripts, no Go needed:

```bash
NAME=$(gum input --placeholder "Project name")
gum confirm "Create $NAME?" && create_project "$NAME"
```
