# Spec Delta

## MODIFIED Requirements

### Requirement: Builtin max is not shadowed

The golden-file test helper SHALL NOT define a local
`func max(a, b int) int` that shadows the Go 1.21+ builtin
`max`. The helper SHALL either inline the comparison at each
use site or define a distinct helper named `maxInt` (or
similar) that does not collide with the builtin.

#### Scenario: Helper uses unshadowed max
- **WHEN** the golden test helper needs to compute a maximum
  of two integers
- **THEN** it uses either the builtin `max(a, b)` or a
  helper named `maxInt`; no local function named `max`
  exists in the test helpers

## ADDED Requirements

### Requirement: Empty-line diffs are not hidden

The golden-file comparison helper SHALL compare the actual
output against the golden file byte-for-byte, including blank
lines. The previous behavior of skipping lines where the
expected value was empty (`if expectedLine != ""`) SHALL NOT
appear; an empty expected line that differs from an empty
actual line is a real diff and SHALL be reported.

#### Scenario: Blank-line diff surfaces
- **WHEN** the actual output contains an extra blank line
  compared to the golden file
- **THEN** the diff includes the blank line and the test
  fails until `UPDATE_GOLDEN=1` regenerates the golden
