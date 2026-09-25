# Command Design and UX

UX decisions — naming, help, destructive-operation flow, deprecation — distilled from [clig.dev](https://clig.dev/) and the conventions of `gh`, `kubectl`, and `docker`. The Cobra API that implements them → See `samber/cc-skills-golang@golang-spf13-cobra`.

## Contents

- [The conversation](#the-conversation)
- [Command hierarchy](#command-hierarchy)
- [Help text](#help-text)
- [Destructive operations](#destructive-operations)
- [Deprecation](#deprecation)

## The conversation

Each invocation is a turn in a conversation, so every output ends with the user knowing the next step:

- After `myapp init`: `Run 'myapp start' to begin.`
- After an error: the fix, in `Suggestions`.
- After a destructive operation: how to undo it.

## Command hierarchy

- **Order**: `APP VERB NOUN` (`kubectl get pods`) or `APP NOUN VERB` (`gh pr create`). Noun-first scales better with many resource types because help groups related operations. Pick one and use it everywhere.
- **Verbs**: `list`, `get`, `create`, `delete`, `update`, `init`, `status`, `login`/`logout` — one word per concept across the whole tool (always `delete`, never a `remove` elsewhere).
- **Positional arguments**: one primary target; everything else is a named flag, so each value states its meaning.

```text
myapp deploy --env production --tag v1.2.3   # each value is labeled
```

- **Flag names**: full words with a short alias for frequent ones (`--output`/`-o`, `--verbose`/`-v`).

## Help text

- **Examples are the most-read section**: show common workflows first.

```text
Examples:
  # Deploy to staging
  myapp deploy --env staging

  # Follow logs
  myapp logs --follow
```

- `Short` descriptions stay around 50–75 characters, since users skim.
- Layer the detail: bare `myapp` shows the common commands, `myapp --help` lists everything, `myapp <cmd> --help` has flags and examples, and docs cover environment variables and edge cases.

## Destructive operations

Irreversible actions (delete, overwrite, force-push) follow the _scriptable_ rule: confirm on an interactive TTY, otherwise require `--yes`/`--force`:

```go
func confirmOrForce(ios *iostreams.IOStreams, force bool, message string) error {
    if force {
        return nil
    }
    if !ios.IsInteractive() {
        return &core.UsageError{Message: message + " (pass --force to confirm)"}
    }
    ok, err := output.Confirm(ios, message)
    if err != nil {
        return err
    }
    if !ok {
        return &core.OperationError{Op: "confirm", Message: "aborted"}
    }
    return nil
}
```

- `--dry-run` prints what would happen in the same structure as the real output, marked as a dry run.
- Afterwards, print the undo path when one exists: `Deleted 3 files. Undo with: myapp restore --snapshot 2024-01-15T10:30:00`.
- For operations over many items, use plan → confirm → execute: show the full plan, confirm once, then act.

## Deprecation

```go
cmd.Flags().StringVar(&old, "old-flag", "", "deprecated: use --new-flag")
_ = cmd.Flags().MarkDeprecated("old-flag", "use --new-flag instead")

&cobra.Command{Use: "oldcmd", Deprecated: "use 'newcmd' instead"}
```

Deprecated flags and commands keep working and print a warning to stderr. For a renamed command, keep the old name as its own command with `Deprecated` set — a silent alias gives scripts no warning before removal. The lifecycle — announce, warn, remove after 2–3 minor versions — lives in `versioning.md`.
