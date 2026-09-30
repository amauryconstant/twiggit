# Spec Delta

## REMOVED Requirements

### Requirement: --output accepts json, table, plain, jsonl values

**Reason**: The implementation removes `jsonl` from the vocabulary (delete `JSONLinesFormatter` type). The "plain" output is redefined as headerless TSV rather than raw human-readable rendering. The `text` value, previously accepted as a legacy quirk, is now rejected as a usage error. The surviving contract is captured in the ADDED requirement below.

**Migration**: See ADDED Requirement "Output format vocabulary". Scripts using `--output jsonl` must switch to `jq -c .` over `--output json`. Scripts using `--output text` must switch to `--output plain`. Pipelines consuming the previous per-element `PlainFormatter` output (fmt.Sprintln) must adopt headerless TSV.

### Requirement: Default (no --output) falls through to plain

**Reason**: The constructor's behavior for the empty format changes from "return PlainFormatter so the global default renders as plain" to "return (nil, nil) so each command picks its own human default." The new rule is captured in the ADDED requirement below; per-command interactive defaults live in the owning command's spec.

**Migration**: See ADDED Requirement "Formatter constructor returns Formatter or UsageError" (specifically the Empty format returns nil scenario) and ADDED Requirement "Empty format defers to per-command spec". For `list`, see the deferred `cli-list` delta for the bespoke interactive default.

## MODIFIED Requirements

### Requirement: Formatter interface is a single Write method

The `output.Formatter` interface SHALL declare exactly one method: `Write(w io.Writer, data any) error`. Implementations SHALL NOT touch `iostreams.IOStreams` directly so they remain unit-testable without TTY detection. Implementations: `JSONFormatter`, `TableFormatter`, `PlainFormatter`. `JSONLinesFormatter` SHALL NOT be defined; streaming consumers SHALL pipe `jq -c .` over `--output json` for the NDJSON shape.

#### Scenario: Formatter tests use *bytes.Buffer

- **WHEN** a test invokes `formatter.Write(&buf, &data)`
- **THEN** the test reads `buf.String()` to verify the output without configuring IOStreams

#### Scenario: TableFormatter accepts structured data

- **WHEN** a command calls `formatter.Write(w, tabularAdapter)` where `tabularAdapter` implements `Tabular` with `Header() = {"BRANCH","PATH"}` and `Rows() = [][]string{{"feat/foo","/tmp/feat/foo"}}`
- **THEN** `TableFormatter` renders the aligned columns to `w` via `text/tabwriter`

#### Scenario: JSONLinesFormatter emits one object per line

- **WHEN** the `output` package is read
- **THEN** `JSONLinesFormatter` SHALL NOT be defined; consumers wanting NDJSON SHALL pipe `jq -c .` over `--output json`

#### Scenario: Formatter implementations are limited to three types

- **WHEN** the `output` package is read
- **THEN** the only types implementing `Formatter` SHALL be `JSONFormatter`, `TableFormatter`, `PlainFormatter`
- **AND** `JSONLinesFormatter` SHALL NOT exist

## ADDED Requirements

### Requirement: Output format vocabulary

The system SHALL accept `--output=<value>` (short `-o <value>`) on every list-style command. The accepted values SHALL be exactly `json`, `table`, and `plain`. The `json` value emits a single JSON document; collections emit bare JSON arrays. The `table` value renders aligned columns with headers; `plain` emits headerless TSV. The system SHALL NOT accept any other value; values outside this set (including the legacy `text` value, and the dropped `jsonl` value) SHALL cause the command to return a `core.UsageError` and exit with code 2.

#### Scenario: --output json emits a single JSON document

- **WHEN** the user runs `twiggit list -o json`
- **THEN** the output is a single JSON document suitable for `jq`

#### Scenario: --output table emits aligned columns

- **WHEN** the user runs `twiggit list -o table`
- **THEN** the output is an aligned table with headers and a row per worktree

#### Scenario: --output plain emits headerless TSV

- **WHEN** the user runs `twiggit list -o plain`
- **THEN** the output is tab-separated rows (one record per line, no header row, no borders) suitable for `cut` / `awk` consumption

#### Scenario: Unknown --output value is rejected

- **WHEN** the user runs `twiggit list -o xml`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

#### Scenario: Legacy --output text is rejected

- **WHEN** the user runs `twiggit list -o text`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

#### Scenario: Dropped --output jsonl is rejected

- **WHEN** the user runs `twiggit list -o jsonl`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

### Requirement: Empty format defers to per-command spec

When `--output` is not supplied (the empty string), the `NewFormatter` constructor SHALL return `(nil, nil)` so each command may pick its own human default. Per-command interactive defaults are declared in the owning command's spec (see `cli-list` for the `list` command).

#### Scenario: Empty format returns nil from NewFormatter

- **WHEN** a command calls `NewFormatter("")`
- **THEN** the constructor returns `(nil, nil)` so the command may choose its own rendering

#### Scenario: list with no --output uses its bespoke default

- **WHEN** the user runs `twiggit list` with no `--output` flag
- **THEN** the rendering matches the per-command default declared in `cli-list` (not a global rule in this spec)

### Requirement: Tabular projection interface

The system SHALL define a `Tabular` interface exposing two methods: `Header() []string` returning the column titles, and `Rows() [][]string` returning the data rows. Both `TableFormatter` and `PlainFormatter` SHALL accept `Tabular` data. A non-`Tabular` argument SHALL cause the formatter to return a `*core.UsageError` and exit 2. Each list-style command SHALL provide a per-command adapter that projects its domain slice (`[]*core.WorktreeInfo`, etc.) to a `Tabular` shape. The first column SHALL be the most-stable identifier (e.g., `BRANCH`); subsequent columns SHALL follow the same order in `Header()` and `Rows()`.

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

When the result of a list-style command is a collection, the system SHALL emit it on stdout as a bare JSON array for `--output json`, not wrapped in a top-level object. One item SHALL be a single JSON object; a small collection SHALL be a JSON array. Consumers using `jq '.[0]'` over a `list -o json` payload SHALL receive the first collection element directly.

#### Scenario: list -o json over three worktrees

- **WHEN** the user runs `twiggit list -o json` and three worktrees exist
- **THEN** the stdout payload is a JSON array of three objects, with no top-level wrapper

#### Scenario: list -o json over zero worktrees

- **WHEN** the user runs `twiggit list -o json` and no worktrees exist
- **THEN** the stdout payload is `[]` (an empty JSON array)

#### Scenario: list -o json consumed by jq

- **WHEN** the user pipes `twiggit list -o json | jq '.[0].branch'`
- **THEN** `jq` returns the branch name of the first worktree without addressing `.worktrees`

### Requirement: Formatter constructor returns Formatter or UsageError

The `NewFormatter(format string) (Formatter, error)` constructor SHALL resolve a format name to its `Formatter`. For format values outside the accepted set (`json`, `table`, `plain`) the constructor SHALL return a nil `Formatter` and a `*core.UsageError`. For the empty string the constructor SHALL return `(nil, nil)` so each command may pick its own human default. Callers SHALL resolve the formatter before performing any work so a bad value fails fast as a usage error.

#### Scenario: Unknown format returns UsageError

- **WHEN** a command calls `NewFormatter("xml")`
- **THEN** the constructor returns `(nil, *core.UsageError)` and the caller exits 2

#### Scenario: Legacy format "text" returns UsageError

- **WHEN** a command calls `NewFormatter("text")`
- **THEN** the constructor returns `(nil, *core.UsageError)` so the legacy value does not silently alias to plain

#### Scenario: Empty format returns nil

- **WHEN** a command calls `NewFormatter("")`
- **THEN** the constructor returns `(nil, nil)` so the command may pick its own human default

### Requirement: Output values are shell-completable

The `--output` flag (and its short form `-o`) SHALL be shell-completable. The completion candidate list SHALL be exactly `json`, `table`, `plain`. The completion SHALL NOT offer file-path completion for the value position.

#### Scenario: Tab completion offers the three accepted values

- **WHEN** the user invokes shell completion after typing `--output ` or `-o `
- **THEN** the shell SHALL offer `json`, `table`, `plain` as candidates
- **AND** SHALL NOT offer file-path completion