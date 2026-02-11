## ADDED Requirements

None - this change is an implementation refactoring and documentation update only. No new capabilities are introduced.

## MODIFIED Requirements

None - this change does not modify any existing specification-level requirements. Only implementation details, flag naming, and documentation are updated.

## REMOVED Requirements

None - no capabilities are being deprecated or removed.

## Summary

This change harmonizes flag usage and updates documentation to match implemented reality without changing specification-level behavior. All changes are internal to implementation and documentation layers:

- Flag naming: `-C, --cd` standardized across create/delete
- Output format: Path-only when `-C` flag is set
- Shell wrapper: Enhanced to handle create/delete with `-C` flag
- Documentation: Aligned with actual implemented flags
- Testing: Comprehensive coverage added for new behavior

No user-facing behavioral changes to existing command workflows.
