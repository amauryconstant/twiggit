# Spec Delta

## MODIFIED Requirements

### Requirement: --output accepts json, table, plain, jsonl values

The system SHALL accept `--output=<value>` (short `-o <value>`) on every list-style command. The accepted values SHALL be exactly `json`, `table`, `plain`, and `jsonl`. The `jsonl` value emits one JSON object per line for streaming consumers; `json` emits a single JSON document. The `table` value renders aligned columns with headers; `plain` emits headerless TSV. The system SHALL NOT accept any other value; values outside this set (including the legacy `text` value) SHALL cause the command to return a `core.UsageError` and exit with code 2.

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
- **THEN** the output is tab-separated rows (one record per line, no header row, no borders) suitable for `cut` / `awk` consumption

#### Scenario: Unknown --output value is rejected

- **WHEN** the user runs `twiggit list -o xml`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

#### Scenario: Legacy --output text is rejected

- **WHEN** the user runs `twiggit list -o text`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

### Requirement: Formatter interface is a single Write method

The `output.Formatter` interface SHALL declare exactly one method: `Write(w io.Writer, data any) error`. Implementations SHALL NOT touch `iostreams.IOStreams` directly so they remain unit-testable without TTY detection. Implementations: `JSONFormatter`, `JSONLinesFormatter`, `TableFormatter`, `PlainFormatter`. The `NewFormatter(format string) (Formatter, error)` constructor SHALL resolve a format name to its `Formatter`; for format values outside the accepted set the constructor SHALL return a nil `Formatter` and a `*core.UsageError`; for the empty string the constructor SHALL return `(PlainFormatter{}, nil)` to honor the global default. Callers SHALL resolve the formatter before performing any work so a bad value fails fast as a usage error.

#### Scenario: Formatter tests use *bytes.Buffer

- **WHEN** a test invokes `formatter.Write(&buf, &data)`
- **THEN** the test reads `buf.String()` to verify the output without configuring IOStreams

#### Scenario: TableFormatter accepts structured data

- **WHEN** a command calls `formatter.Write(w, worktreeRows)` where `worktreeRows` is a `Tabular` with `Header() = {"BRANCH","PATH"}` and `Rows() = [][]string{{"feat/foo","/tmp/feat/foo"}}`
- **THEN** `TableFormatter` renders the header row followed by the aligned data rows

#### Scenario: JSONLinesFormatter emits one object per line

- **WHEN** a command calls `formatter.Write(w, []core.Worktree{...})`
- **THEN** `JSONLinesFormatter` writes one JSON object per line with no enclosing array

#### Scenario: Unknown format returns UsageError

- **WHEN** a command calls `NewFormatter("xml")`
- **THEN** the constructor returns `(nil, *core.UsageError)` and the caller exits 2

#### Scenario: Legacy format "text" returns UsageError

- **WHEN** a command calls `NewFormatter("text")`
- **THEN** the constructor returns `(nil, *core.UsageError)` so the legacy value does not silently alias to plain

#### Scenario: Empty format returns PlainFormatter

- **WHEN** a command calls `NewFormatter("")`
- **THEN** the constructor returns `(PlainFormatter{}, nil)` to honor the global default

## ADDED Requirements

### Requirement: Tabular projection interface

The system SHALL define a `Tabular` interface in `internal/output/tabular.go` with exactly two methods: `Header() []string` returning the column titles, and `Rows() [][]string` returning the data rows. Both `TableFormatter` and `PlainFormatter` SHALL accept `Tabular` data. A non-`Tabular` argument SHALL cause the formatter to return a `*core.UsageError` and exit 2. Each list-style command SHALL provide a per-command adapter that projects its domain slice (`[]*core.WorktreeInfo`, etc.) to a `Tabular` shape. The first column SHALL be the most-stable identifier (e.g., `BRANCH`); subsequent columns SHALL follow the same order in `Header()` and `Rows()`.

#### Scenario: TableFormatter renders header + aligned rows

- **WHEN** a command calls `formatter.Write(w, worktreeRows)` where `worktreeRows` is a `Tabular` with `Header() = {"BRANCH","PATH","STATUS"}`
- **THEN** `TableFormatter` writes a header row followed by one row per worktree, columns aligned with `text/tabwriter`

#### Scenario: PlainFormatter renders headerless TSV

- **WHEN** a command calls `formatter.Write(w, worktreeRows)` where `worktreeRows` is a `Tabular`
- **THEN** `PlainFormatter` writes one tab-separated row per record with no header line

#### Scenario: Non-tabular data rejected

- **WHEN** a command calls `formatter.Write(w, "raw string")` against `TableFormatter` or `PlainFormatter`
- **THEN** the formatter returns `(nil, *core.UsageError)` and the caller exits 2

### Requirement: --output json emits a bare collection

When the result of a list-style command is a collection, the system SHALL emit it on stdout as a bare JSON array for `--output json`, not wrapped in a top-level object. One item SHALL be a single JSON object; a small collection SHALL be a JSON array; a streaming or large collection SHALL fall through to `--output jsonl`. Consumers using `jq '.[0]'` over a `list -o json` payload SHALL receive the first collection element directly.

#### Scenario: list -o json over three worktrees

- **WHEN** the user runs `twiggit list -o json` and three worktrees exist
- **THEN** the stdout payload is a JSON array of three objects, with no top-level wrapper

#### Scenario: list -o json over zero worktrees

- **WHEN** the user runs `twiggit list -o json` and no worktrees exist
- **THEN** the stdout payload is `[]` (an empty JSON array)

#### Scenario: list -o json consumed by jq

- **WHEN** the user pipes `twiggit list -o json | jq '.[0].branch'`
- **THEN** `jq` returns the branch name of the first worktree without addressing `.worktrees`
