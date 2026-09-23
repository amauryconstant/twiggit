# Capability: Path Utilities

## Purpose

Low-level filesystem path helpers used by context detection and
project discovery. Symlink-aware, cross-platform-safe.

## Requirements

### Requirement: No behavior owned; canonical home is core-paths

The `infrastructure-path-utils` capability SHALL own no behavior. All
path-utility behavior lives in `core-paths` after the
`2026-09-22-architecture-layer-inversion` change removed the four
historical requirements. This spec is preserved only so legacy
cross-references remain resolvable.

#### Scenario: Reader follows the delegation pointer

- **WHEN** a reader seeks the path-utility contract
- **THEN** they are redirected to `core-paths`; this spec carries no
  testable behavior of its own
