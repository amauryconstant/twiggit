# Spec Delta

## REMOVED Requirements

### Requirement: Text output (default)

**Reason**: The rule "tabular text by default for commands that produce listings" was a global claim that contradicted the per-command default behavior (`list` renders bespoke `"branch -> path (modified)(detached)"` for empty `--output`). The new contract: each command owns its interactive default; per-command specs declare them.

**Migration**: See ADDED Requirement "Per-command interactive default".

### Requirement: JSON output via `--output`

**Reason**: The vocabulary drops `jsonl` (the `JSONLinesFormatter` type is removed). The JSON shape pin moves from envelope (`{"worktrees":[…]}`) to bare array (`[…]`) for collections. The `Formatter` interface contract moves to its own requirement in `cli-output`. The new contract: `--output` accepts `json|table|plain`; collections emit bare JSON arrays; the `Formatter` interface and constructor live in `cli-output`.

**Migration**: See ADDED Requirement "JSON output shape and vocabulary" (new vocabulary and shape) and the Formatter interface and constructor requirements in `cli-output`. Scripts consuming `{"worktrees":[…]}` must switch to bare-array jq (`jq '.[]'`). Scripts using `--output jsonl` must switch to `jq -c .` over `--output json`.

## ADDED Requirements

### Requirement: Per-command interactive default

When `--output` is not supplied (the empty string), each list-style command SHALL own its interactive default. Per-command interactive defaults are declared in the owning command's spec. For non-list commands without `--output`, the default SHALL be `plain`. This spec no longer prescribes a global "tabular text by default for listings" rule.

#### Scenario: list with no --output uses its bespoke default

- **WHEN** the user runs `twiggit list` with no `--output` flag
- **THEN** the rendering matches the per-command default declared in `cli-list` (not a global rule in this spec)

#### Scenario: Non-list command without --output uses plain

- **WHEN** the user runs a non-list command (e.g., `twiggit create feature/x`) without `--output`
- **THEN** the output is a plain human-readable success message on stdout

### Requirement: JSON output shape and vocabulary

The system SHALL accept `--output=<value>` (short `-o <value>`) on every list-style command. The accepted values SHALL be exactly `json`, `table`, and `plain`. The `json` value emits a single JSON document; collections emit bare JSON arrays. The `table` value renders aligned columns with headers; `plain` emits headerless TSV. The system SHALL NOT accept any other value; values outside this set (including the legacy `text` value, and the dropped `jsonl` value) SHALL cause the command to return a `core.UsageError` and exit with code 2.

#### Scenario: JSON list output is a bare array

- **WHEN** the user runs `twiggit list -o json` and three worktrees exist
- **THEN** the system SHALL emit a bare JSON array on stdout with shape:
  `[{branch, path, status}, ...]`
- **AND** SHALL exit with status 0

#### Scenario: Empty worktrees yields empty array

- **WHEN** the user runs `twiggit list -o json` and no worktrees exist
- **THEN** the stdout payload is `[]` (an empty JSON array)

#### Scenario: Unknown format rejected

- **WHEN** the user passes `--output=xml` (unsupported)
- **THEN** the system SHALL return `core.UsageError` naming supported formats (`json`, `table`, `plain`)
- **AND** SHALL exit with code 2 (usage error)

#### Scenario: Legacy format text rejected

- **WHEN** the user runs `twiggit list -o text`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

#### Scenario: Dropped format jsonl rejected

- **WHEN** the user runs `twiggit list -o jsonl`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

The Formatter interface contract (single `Write(w io.Writer, data any) error` method, no `IOStreams` access) is owned by `cli-output`.

### Requirement: Output values are shell-completable

The `--output` flag (and its short form `-o`) SHALL be shell-completable. The completion candidate list SHALL be exactly `json`, `table`, `plain`. File-path completion SHALL NOT be offered for the value position. The Formatter interface contract and the `NewFormatter` constructor behavior are owned by `cli-output`.

#### Scenario: Tab completion offers the three accepted values

- **WHEN** the user invokes shell completion after typing `--output ` or `-o `
- **THEN** the shell SHALL offer `json`, `table`, `plain` as candidates
- **AND** SHALL NOT offer file-path completion