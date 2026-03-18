## Why

CLI output verification requires reliable snapshot testing that handles complex multi-line output. Golden file testing provides easy updates via UPDATE_GOLDEN flag, clear diffs on failure, and better maintainability than hard-coded expectations.

## What Changes

- Create test/helpers/golden.go with CompareGolden function
- Create test/golden/ directory structure (list/, errors/)
- Add UPDATE_GOLDEN environment variable support for updating golden files
- Add mise tasks test:golden and test:golden:update
- Create golden file tests for list command (text and JSON output)
- Create golden file tests for error formatting (validation, service, not found errors)

## Capabilities

### New Capabilities

- `golden-file-testing`: Golden file infrastructure for CLI output verification with UPDATE_GOLDEN support

### Modified Capabilities

None - this is new infrastructure.

## Impact

**Files Created:**
- test/helpers/golden.go
- test/golden/list/*.golden
- test/golden/errors/*.golden


**Files Modified:**
- mise/config.toml (new tasks)
- test/AGENTS.md (documentation)
- test/helpers/AGENTS.md (documentation)
