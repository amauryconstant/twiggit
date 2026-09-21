# Spec Delta

## ADDED Requirements

### Requirement: Verbose log writer is injectable

The verbose-log helper (`logv`) SHALL accept an `io.Writer`
parameter so that tests can assert on verbose output without
writing to `os.Stderr`. When invoked from a cobra command, the
caller passes `cmd.ErrOrStderr()`; the helper signature SHALL
NOT default to `os.Stderr` directly.

#### Scenario: Test captures verbose output via injected writer
- **WHEN** a test invokes the verbose-log helper with a
  `bytes.Buffer` as the writer
- **THEN** the buffer contains the expected verbose line and
  no data is written to `os.Stderr`

#### Scenario: Cobra command uses cmd.ErrOrStderr
- **WHEN** a leaf command calls the verbose-log helper
- **THEN** the writer passed is `cmd.ErrOrStderr()` so
  `cmd.SetErr` (in tests) redirects the output correctly

### Requirement: Quiet mode suppresses verbose log output

When `--quiet`/`-q` is set, the verbose-log helper SHALL be a
no-op (no bytes written to the injected writer). When
`--verbose` and `--quiet` are both set, `--verbose` wins (see
`cli-command-options-pattern` for the precedence rule).

#### Scenario: Quiet flag suppresses verbose output
- **WHEN** a leaf command runs with `--quiet` and calls the
  verbose-log helper
- **THEN** the injected writer receives zero bytes

#### Scenario: Verbose wins over quiet
- **WHEN** a leaf command runs with both `--quiet` and `--verbose`
- **THEN** the injected writer receives the verbose line
