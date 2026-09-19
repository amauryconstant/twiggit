# 02 — Command Design and UX

The primary reference for CLI UX is [clig.dev](https://clig.dev/) (Command Line Interface Guidelines), created by the co-creators of Docker Compose. This document distills those guidelines alongside patterns from Thoughtworks, Atlassian, and observed conventions from major Go CLIs.

This reference owns UX *decisions* (naming, help layout, destructive-op flow, color rules). The Cobra *API* that implements them — `Args` validators, the `PersistentPreRunE` hook chain (a child's hook replaces the parent's), `RegisterFlagCompletionFunc`, `MarkDeprecated`, `RunE`-not-`Run` — is owned by `samber/cc-skills-golang@golang-spf13-cobra`. Load it alongside.

---

## The Conversation Metaphor

Running a program usually involves more than one invocation. Each command is a turn in a conversation. Design accordingly:

- After `myapp init`, print: `Run 'myapp start' to begin.`
- After an error, suggest the fix.
- After a destructive operation, print how to undo it.

The CLI should never leave the user wondering "what now?"

---

## Command Hierarchy

### Naming Convention

Cobra's convention is `APPNAME VERB NOUN --FLAG` (`kubectl get pods`). An alternative is `APPNAME NOUN VERB` (`gh pr create`), which groups by resource and aids discoverability when you have many resource types.

Pick one and be consistent. For CLIs with many resource types, noun-first tends to scale better because the top-level help text groups related operations.

### Positional Arguments

**One argument is fine. Two are questionable. Three is too many.** Use flags for everything beyond the primary target:

```
# Good: clear what each part means
myapp deploy --env production --tag v1.2.3

# Bad: positional ambiguity
myapp deploy production v1.2.3
```

### Command Naming

Use common words. Don't invent synonyms for established concepts:

- `list`, `get`, `create`, `delete`, `update` — universally understood
- `init` — project/workspace initialization
- `status` — current state
- `login` / `logout` — authentication

Don't use `--ver` when `--version` exists. Don't use `remove` in some commands and `delete` in others.

---

## Help Text

### Examples Are the Most-Read Section

Put them prominently. Show common workflows, not just flag descriptions:

```
Examples:
  # Create a new project
  myapp project create my-project

  # Deploy to staging
  myapp deploy --env staging

  # Watch logs in real-time
  myapp logs --follow
```

### Description Length

Keep short descriptions to ~50-75 characters. Users skim. Use visual hierarchy (grouping, spacing) to aid scanning.

### Help Routing

`--help` output goes to **stdout** (the user asked for it — it's the primary output). Usage errors go to **stderr** (an error occurred, and help is supplementary).

### Progressive Disclosure

Don't dump everything in `--help`. Layer information:

1. **No args / wrong usage:** Brief usage + most common commands
2. **`--help`:** Full command list with descriptions
3. **`myapp <command> --help`:** Detailed flag descriptions + examples
4. **Advanced:** Environment variables and edge cases in docs or a `--help --verbose` extension

---

## Destructive Operations

### Confirmation Prompts

Require confirmation for anything irreversible: deletion, overwrite, force-push.
This is the destructive-op form of the _scriptable_ rule (`SKILL.md` §The
`scriptable` Rule) — prompt only on an interactive TTY; otherwise require
`--yes` / `--force`:

```go
func confirmOrFlag(io *iostreams.IOStreams, force bool, message string) error {
    if force {
        return nil
    }
    if !io.IsInteractive() { // stdout+stdin both TTY; see canonical IOStreams in SKILL.md
        return fmt.Errorf("%s (use --force to skip confirmation)", message)
    }
    // interactive prompt here
}
```

### Dry-Run Support

`--dry-run` shows what *would* happen without executing. The output structure SHOULD be identical to the real output, annotated with a dry-run marker.

### Undo Hints

After a destructive operation, print how to reverse it if possible:

```
Deleted 3 files. Undo with: myapp restore --snapshot 2024-01-15T10:30:00
```

---

## Color and Formatting

- Check terminal capability before using ANSI codes. Respect `NO_COLOR` env var ([no-color.org](https://no-color.org/)).
- Provide `--color=always|never|auto` (auto = TTY detection).
- Red for errors, yellow for warnings, green for success. Don't overuse.
- Don't rely on color alone to convey meaning — always pair with text (accessibility).
- Keep output greppable. Don't use emoji to replace searchable words.

---

## Cobra's "Did You Mean?"

Cobra has built-in typo suggestion: `myapp srver` → `Did you mean 'myapp server'?` This is powered by Levenshtein distance. Apply the same principle to flag names and argument values where feasible.

---

## Flag Deprecation

Cobra supports flag deprecation natively:

```go
cmd.Flags().StringVar(&old, "old-flag", "", "deprecated: use --new-flag")
cmd.Flags().MarkDeprecated("old-flag", "use --new-flag instead")
```

Deprecated flags still work but print a warning to stderr. Remove them after 2–3 minor versions with a clear migration path.

For command deprecation:

```go
&cobra.Command{
    Use:        "oldcmd",
    Deprecated: "use 'newcmd' instead",
}
```
