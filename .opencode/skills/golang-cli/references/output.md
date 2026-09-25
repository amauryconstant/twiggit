# Output and Composability

## Contents

- [The Unix contract](#the-unix-contract)
- [TTY detection and adaptive output](#tty-detection-and-adaptive-output)
- [The --output flag](#the---output-flag)
- [JSON shape](#json-shape)
- [Stdin input](#stdin-input)
- [Long-running operations](#long-running-operations)
- [Color](#color)

## The Unix contract

| Stream | Carries | Contract |
| --- | --- | --- |
| stdout | The command's result | Machine-parseable when piped, no decoration |
| stderr | Errors, progress, prompts, verbose and debug output | For humans; never parsed downstream |
| Exit code | Success or failure | 0 / 1 / 2 / 130 (→ `errors.md`) |

Every line that is not the result goes to stderr, which is what makes `myapp list -o json | jq .` work. Help requested with `--help` is the result and goes to stdout; usage printed after an error goes to stderr.

## TTY detection and adaptive output

`IOStreams` captures TTY state once (`golang.org/x/term.IsTerminal`) and exposes `IsStdoutTTY()` and `IsInteractive()`. Route every decision through it so tests control the answer.

| Context | Behavior |
| --- | --- |
| stdout is a TTY | Colors, aligned tables, spinners, progress bars |
| stdout is a pipe | Bare data (TSV or `--output` format), no ANSI, no spinners |
| stdin is piped | Read input from stdin |
| stdin is a TTY and no input given | Prompt (when interactive) or fail fast with a usage error |

Offer `--color=auto|always|never` and `--no-input` to override detection.

## The --output flag

Query commands take `--output json|table|plain` (`-o`). [formatter.go](../assets/examples/formatter.go):

- `NewFormatter(format)` returns a `Formatter` or a `UsageError` for unknown values — resolve it **before** doing any work ([list.go](../assets/examples/list.go)).
- `Formatter.Write(w io.Writer, data any)` takes a writer, not `IOStreams`; pass `opts.IO.Out`.
- `table` and `plain` need data implementing `Tabular` (`Header()`, `Rows()`); build that adapter in `cmd/` so presentation stays out of the core.
- No flag means the human default: styled, TTY-aware output written by the command itself.

| Format | Audience | Shape |
| --- | --- | --- |
| `json` | Programs | Indented JSON of the result value |
| `table` | Humans reading columns | Header + aligned columns (`text/tabwriter`) |
| `plain` | `cut`, `awk`, `grep` | Headerless TSV, one record per line |

For richer tables, see `libraries.md`.

## JSON shape

| Result                          | Emit                             |
| ------------------------------- | -------------------------------- |
| One item                        | A single object                  |
| A small collection              | A JSON array                     |
| A large or streaming collection | JSON Lines — one object per line |

| Property                 | JSON array               | JSON Lines             |
| ------------------------ | ------------------------ | ---------------------- |
| Consumer simplicity      | Simpler                  | Per-line parsing       |
| Streaming                | Buffers the whole output | Each line stands alone |
| `head` / `tail` / `grep` | Breaks the document      | Works                  |

## Stdin input

`-` as a filename means stdin:

```go
func openInput(path string, stdinPiped bool) (io.ReadCloser, error) {
    if path != "" && path != "-" {
        return os.Open(path)
    }
    if !stdinPiped {
        return nil, &core.UsageError{Message: "no input: pass a file or pipe data on stdin"}
    }
    return io.NopCloser(os.Stdin), nil
}
```

Detect piped stdin with `fi, _ := os.Stdin.Stat(); fi.Mode()&os.ModeCharDevice == 0`. When stdin is a terminal and no file was given, fail fast with that usage error so the command never sits waiting for input.

## Long-running operations

1. Unknown duration: a spinner.
2. Known total: switch to a progress bar.
3. Parallel work: one bar per task.
4. Phased work (connecting → downloading → done): update one status line.

All of it goes to stderr and only on a TTY; in a pipe, print one plain status line per step or nothing. Library picks: `libraries.md`.

## Color

- Respect `NO_COLOR` ([no-color.org](https://no-color.org/)) and disable color on a stream that is not a TTY — `NewStyles(false)` makes every style the identity. Stdout and stderr are checked separately: `Styles()` for data, `ErrStyles()` for errors and progress.
- Red for errors, yellow for warnings, green for success, used sparingly.
- Pair every color with text, so meaning survives without color and stays greppable.
