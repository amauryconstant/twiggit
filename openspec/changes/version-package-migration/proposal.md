## Why

Version variables currently live in `cmd/version.go`, mixing infrastructure-level data with presentation layer code. This change extracts version information into a dedicated `internal/version` package, aligning with the germinator pattern and establishing clean separation of concerns where build-time injected data lives in infrastructure.

## What Changes

- Create `internal/version/version.go` with Version, Commit, Date variables and String() function
- Update `cmd/version.go` to import from `internal/version` package
- Update `.goreleaser.yml` ldflags path from `twiggit/cmd.Version` to `twiggit/internal/version.Version`
- Update `.mise/config.toml` ldflags paths from `twiggit/cmd.version` to `twiggit/internal/version.Version` (both build and build:local tasks)

## Capabilities

### New Capabilities

- `version-package`: Separate internal/version package for build-time version injection with Version, Commit, Date variables and formatted String() output

### Modified Capabilities

## Impact

**Files Created:**
- `internal/version/version.go`

**Files Modified:**
- `cmd/version.go` - imports from internal/version
- `.goreleaser.yml` - updated ldflags paths
- `.mise/config.toml` - updated ldflags paths in build and build:local tasks
