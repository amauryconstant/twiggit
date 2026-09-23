# Spec Delta

## MODIFIED Requirements

### Requirement: Text output (default)

The system SHALL emit tabular text by default for commands that produce listings (`list`, etc.). When `--output` is not supplied, individual commands MAY override the default to `table` for explicit list views; the override is documented in the per-command spec. For every other command without `--output`, the default SHALL be `plain`.

#### Scenario: Default text output
- **WHEN** user runs `twiggit list` without `--output`
- **THEN** system SHALL emit a table on stdout
- **AND** errors and progress SHALL go to stderr

### Requirement: JSON output via `--output`

The system SHALL accept `--output=<value>` (short `-o <value>`) on every list-style command. The accepted values SHALL be exactly `json`, `table`, `plain`, and `jsonl`. The `jsonl` value emits one JSON object per line for streaming consumers; `json` emits a single JSON document. The `table` value renders aligned columns with headers; `plain` emits raw human-readable rendering without table borders.

#### Scenario: JSON list output
- **WHEN** user runs `twiggit list -o json`
- **THEN** system SHALL emit JSON with shape:
  `{"worktrees": [{"branch": "...", "path": "...", "status": "clean|modified|detached"}]}`
- **AND** SHALL exit with status 0

#### Scenario: JSON Lines output
- **WHEN** user runs `twiggit list -o jsonl | head`
- **THEN** each line of the output SHALL be independently parseable JSON

#### Scenario: Unknown format rejected
- **WHEN** user passes `--output=xml` (unsupported)
- **THEN** system SHALL return `core.UsageError` naming supported formats (`json`, `table`, `plain`, `jsonl`)
- **AND** SHALL exit with code 2 (usage error)

#### Scenario: Formatter interface contract
- **WHEN** the `output.Formatter` interface is read
- **THEN** it SHALL declare exactly one method `Write(w io.Writer, data any) error`
- **AND** implementations SHALL NOT touch `iostreams.IOStreams` directly

### Requirement: Stream separation

The system SHALL send all data output to stdout and all diagnostic output (errors, progress, verbose messages) to stderr so that `twiggit list -o json | jq` works in a pipeline.

#### Scenario: Pipeable JSON
- **WHEN** user runs `twiggit list -o json | jq .`
- **THEN** `jq` SHALL receive only the JSON document on stdin
- **AND** progress messages SHALL NOT corrupt the JSON stream
