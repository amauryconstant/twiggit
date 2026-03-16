## Purpose

The output-formats capability provides machine-readable output formats for CLI commands, enabling reliable integration with scripts and automation tools.

---

## Requirements

### Requirement: Output format flag controls output structure
The system SHALL provide an `--output/-o` flag on supported commands that accepts `text` (default) and `json` values to control output format.

#### Scenario: Default text output
- **WHEN** user runs a command without `--output` flag
- **THEN** output is formatted as human-readable text
- **AND** behavior matches existing text output format

#### Scenario: JSON output format
- **WHEN** user runs `list --output json`
- **THEN** output is valid JSON to stdout
- **AND** JSON structure contains `worktrees` array
- **AND** each worktree has `branch`, `path`, and `status` fields

#### Scenario: Invalid output format
- **WHEN** user runs a command with `--output invalid`
- **THEN** command returns an error
- **AND** error message indicates valid formats are `text` and `json`

### Requirement: JSON worktree list output structure
The system SHALL output worktree lists in a structured JSON format when JSON output is requested.

#### Scenario: JSON list with worktrees
- **WHEN** user runs `list --output json` and worktrees exist
- **THEN** output is valid JSON with structure:
  ```json
  {
    "worktrees": [
      {
        "branch": "<branch-name>",
        "path": "<absolute-path>",
        "status": "<clean|modified|detached>"
      }
    ]
  }
  ```
- **AND** `status` is `"clean"` when worktree has no uncommitted changes
- **AND** `status` is `"modified"` when worktree has uncommitted changes
- **AND** `status` is `"detached"` when worktree is in detached HEAD state

#### Scenario: JSON list with no worktrees
- **WHEN** user runs `list --output json` and no worktrees exist
- **THEN** output is valid JSON: `{"worktrees": []}`
- **AND** output is a single line (no pretty printing)

#### Scenario: JSON output is parseable
- **WHEN** user runs `list --output json`
- **THEN** output can be parsed by standard JSON tools (jq, python json module)
- **AND** output contains no trailing content after closing brace

### Requirement: Output formatter interface
The system SHALL provide an OutputFormatter interface for consistent output formatting across commands.

#### Scenario: Formatter interface contract
- **WHEN** a new output format is needed
- **THEN** formatter implements the OutputFormatter interface
- **AND** formatter methods accept domain types and return formatted strings

#### Scenario: TextFormatter implementation
- **WHEN** TextFormatter formats worktrees
- **THEN** output matches existing text output format
- **AND** format is `"<branch> -> <path> (<status>)"`

#### Scenario: JSONFormatter implementation
- **WHEN** JSONFormatter formats worktrees
- **THEN** output is valid JSON
- **AND** output is compact (no extra whitespace)
- **AND** special characters in paths are properly escaped

### Requirement: JSON output goes to stdout
The system SHALL write JSON output to stdout to enable piping to other commands.

#### Scenario: JSON output piped to jq
- **WHEN** user runs `list --output json | jq '.worktrees[0].branch'`
- **THEN** jq receives valid JSON on stdin
- **AND** jq outputs the first worktree's branch name

#### Scenario: JSON output with verbose flag
- **WHEN** user runs `list --output json -v`
- **THEN** JSON output goes to stdout
- **AND** verbose messages go to stderr
- **AND** JSON output is not mixed with verbose messages
