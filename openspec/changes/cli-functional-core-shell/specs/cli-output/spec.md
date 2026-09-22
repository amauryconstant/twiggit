# Spec Delta

## Purpose

Defines the `--output` flag and the `Formatter` interface contract for `json`, `table`, `plain`, and `jsonl` rendering so every list-style command shares one formatting seam. The data-stream contract (stdout vs stderr split) lives in `cli-error-formatting`; the TTY-aware rendering surface lives in `cli-iostreams`.

## ADDED Requirements

### Requirement: --output accepts json, table, plain, jsonl values

The system SHALL accept `--output=<value>` (short `-o <value>`) on every list-style command. The accepted values SHALL be exactly `json`, `table`, `plain`, and `jsonl`. The `jsonl` value emits one JSON object per line for streaming consumers; `json` emits a single JSON document. The `table` value renders aligned columns with headers; `plain` emits raw human-readable rendering without table borders.

#### Scenario: --output json emits a single JSON document
- **WHEN** the user runs `twiggit list -o json`
- **THEN** the output is a single JSON document suitable for `jq`

#### Scenario: --output jsonl emits one JSON object per line
- **WHEN** the user runs `twiggit list -o jsonl | head`
- **THEN** each line is independently parseable JSON; the pipeline receives newline-delimited records

#### Scenario: --output table emits aligned columns
- **WHEN** the user runs `twiggit list -o table`
- **THEN** the output is an aligned table with headers and a row per worktree

#### Scenario: --output plain emits raw human-readable rendering
- **WHEN** the user runs `twiggit list -o plain`
- **THEN** the output is unbordered human-readable rendering without table borders or JSON quoting

#### Scenario: Unknown --output value is rejected
- **WHEN** the user runs `twiggit list -o xml`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

### Requirement: Formatter interface is a single Write method

The `output.Formatter` interface SHALL declare exactly one method: `Write(w io.Writer, data any) error`. Implementations SHALL NOT touch `iostreams.IOStreams` directly so they remain unit-testable without TTY detection. Implementations: `JSONFormatter`, `JSONLinesFormatter`, `TableFormatter`, `PlainFormatter`.

#### Scenario: Formatter tests use *bytes.Buffer
- **WHEN** a test invokes `formatter.Write(&buf, &data)`
- **THEN** the test reads `buf.String()` to verify the output without configuring IOStreams

#### Scenario: TableFormatter accepts structured data
- **WHEN** a command calls `formatter.Write(w, TableData{Headers: []string{"Name", "Branch"}, Rows: [][]string{{"feat/foo", "main"}, {"feat/bar", "main"}}})`
- **THEN** `TableFormatter` renders the aligned columns to `w`

#### Scenario: JSONLinesFormatter emits one object per line
- **WHEN** a command calls `formatter.Write(w, []core.Worktree{...})`
- **THEN** `JSONLinesFormatter` writes one JSON object per line with no enclosing array

### Requirement: Default (no --output) falls through to plain

When `--output` is not supplied (including the empty-string case), the system SHALL use `plain` for every command. This spec defines the global default only; per-command overrides (for example the `list` command overriding the default to `table`) are declared in their owning command spec.

#### Scenario: Empty --output falls back to plain
- **WHEN** the user runs `twiggit list --output ""` (or omits `--output` entirely) on a command that declares no override
- **THEN** the system renders the output as `plain` (the global default)

#### Scenario: Global default is plain for non-list commands
- **WHEN** the user runs `twiggit create feat/foo` without `--output`
- **THEN** the output SHALL be a plain human-readable success message
