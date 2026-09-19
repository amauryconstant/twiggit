# Capability: Golden-File Testing

## Purpose

Snapshot testing using golden files for CLI output, error formatting,
and other text-based outputs. The `UPDATE_GOLDEN` env var rewrites
golden files in place.

## Requirements

### Requirement: `CompareGolden` helper

The `test/helpers` package SHALL provide a `CompareGolden(actual,
goldenFile)` function that compares `actual` against the golden file
content.

#### Scenario: Match

- **WHEN** `CompareGolden` is called with actual matching the file
- **THEN** the test SHALL pass

#### Scenario: Mismatch

- **WHEN** `CompareGolden` is called with actual not matching
- **THEN** the test SHALL fail with a unified diff

#### Scenario: Golden file missing

- **WHEN** `CompareGolden` is called and the golden file does not exist
- **THEN** the test SHALL fail with an error naming the missing file

### Requirement: `UPDATE_GOLDEN` rewrites files

When `UPDATE_GOLDEN` is set, `CompareGolden` SHALL write the actual
content to the golden file (creating it if missing) and SHALL return
nil.

#### Scenario: Update existing golden

- **WHEN** `UPDATE_GOLDEN=true` and `CompareGolden` is called with a
  mismatching actual
- **THEN** the golden file SHALL be overwritten with the actual
- **AND** the test SHALL pass

#### Scenario: Create new golden

- **WHEN** `UPDATE_GOLDEN=true` and the golden file does not exist
- **THEN** the file SHALL be created with the actual content
- **AND** the test SHALL pass

### Requirement: Path resolution

Golden files SHALL be resolved relative to `test/golden/` with the
`.golden` extension. The `goldenFile` argument SHALL NOT include the
directory or extension.

#### Scenario: Path resolution

- **WHEN** `CompareGolden` is called with `goldenFile = "list/basic-text"`
- **THEN** the system SHALL resolve to `test/golden/list/basic-text.golden`

### Requirement: Coverage priorities

Golden-file tests SHALL cover at minimum:

- `twiggit list` text output
- `twiggit list --output json` JSON output
- Validation, service, and not-found error formatting



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `mise` task integration

The project SHALL provide `mise run test:golden` (run tests, fail on
mismatch) and `mise run test:golden:update` (run with
`UPDATE_GOLDEN=true`).

#### Scenario: Run golden tests

- **WHEN** developer runs `mise run test:golden`
- **THEN** all golden-file tests SHALL run
- **AND** tests SHALL fail on mismatch

#### Scenario: Update goldens

- **WHEN** developer runs `mise run test:golden:update`
- **THEN** all golden-file tests SHALL run with `UPDATE_GOLDEN=true`
- **AND** goldens SHALL be updated in place
