# 04 — Output and Composability

---

## The Unix Contract

| Stream | Content | Contract |
|--------|---------|----------|
| stdout | Primary data output | Machine-parseable when piped. No decorations. |
| stderr | Errors, diagnostics, progress, verbose/debug | Human-readable. Never parsed by downstream tools. |
| Exit code | Success/failure signal | 0 = success. Non-zero = failure with semantic codes. |

**The critical rule:** Never mix diagnostic output into stdout. Every `fmt.Printf` that isn't primary data output MUST go to stderr. This is what makes `myapp list | jq .` work.

---

## TTY Detection

```go
import "golang.org/x/term"

isStdoutTTY := term.IsTerminal(int(os.Stdout.Fd()))
isStdinPipe := !term.IsTerminal(int(os.Stdin.Fd()))
```

For more reliable stdin pipe detection:

```go
fi, _ := os.Stdin.Stat()
isInputFromPipe := fi.Mode()&os.ModeCharDevice == 0
```

---

## Adaptive Output

| Context | Behavior |
|---------|----------|
| stdout is TTY | Human output: colors, tables, spinners, progress bars |
| stdout is pipe | Machine output: JSON, TSV, or plain text. No ANSI. No spinners. |
| stdin has data | Read from stdin (piped input) |
| stdin is TTY + no args | Either prompt interactively or show help |

Gate all color output, interactive prompts, progress bars, and spinner animations on TTY detection. Provide `--color=always|never|auto` and `--no-input` flags as overrides.

---

## Output Format Strategy

### The Formatter Interface

The output formatter implements the **Respond** concern (see [01-architecture.md](./01-architecture.md)). It takes an `io.Writer` — not `IOStreams` — so it stays decoupled from terminal concerns:

```go
type Formatter interface {
    Write(w io.Writer, data any) error
}
```

Implementations for JSON, table, and plain text, selected by the `--output json|table|plain` flag value and TTY state via `output.NewFormatter(format)` (see `SKILL.md`). This is the ecosystem-standard flag (`samber/cc-skills-golang@golang-cli`).

### JSON Output

- Single JSON object for single-item results
- JSON array for collections
- Or: one JSON object per line (JSON Lines / NDJSON) for streaming

**Choosing between JSON array and JSON Lines:**

| Property | JSON Array | JSON Lines |
|----------|-----------|------------|
| Consumer simplicity | Simpler | Requires per-line parsing |
| Streaming | Must buffer entire output | Each line independently parseable |
| Works with `head`/`tail`/`grep` | No | Yes |
| Recommendation for CLIs | Small collections | Large/streaming datasets |

### TSV

Good default for piped context. No quoting issues, easy to `cut`/`awk`. The `gh` CLI's `tableprinter` renders aligned columns to terminals and tab-separated values to pipes automatically.

### Table Output

Use `text/tabwriter` (stdlib) for basic tab-aligned text with zero dependencies. For richer tables, see [14-libraries.md](./14-libraries.md).

---

## Stdin Composability

### File-or-Stdin Pattern

The most common composability pattern. Convention: `-` as a filename means "read from stdin":

```go
func getInput(path string) (io.ReadCloser, error) {
    if path == "-" || path == "" {
        if isInputFromPipe() {
            return io.NopCloser(os.Stdin), nil
        }
        return nil, fmt.Errorf("no input: provide a file or pipe data via stdin")
    }
    return os.Open(path)
}
```

**Don't hang on empty stdin.** If stdin is a TTY (no pipe) and no file argument was given, print help or error immediately. Don't silently block waiting for input.

---

## Long-Running Operations

### Progressive Disclosure Pattern

1. **Unknown duration:** Start with a spinner
2. **Known total:** Switch to a progress bar when total is discovered
3. **Multi-task:** Use concurrent progress bars for parallel work

### TTY-Aware Progress

In non-TTY mode: skip spinners and progress bars entirely. Print machine-parseable output (JSON lines, one status per line). Don't use ANSI escape codes.

### Status Line with Context

For operations with phases (connecting → downloading → processing → done), update a single status line. See [14-libraries.md](./14-libraries.md) for spinner and progress bar libraries.
