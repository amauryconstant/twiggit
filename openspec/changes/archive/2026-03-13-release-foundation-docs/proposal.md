## Why

Twiggit needs the minimal documentation and legal foundation required for a public open-source release. Currently, the project lacks a LICENSE file, has no CHANGELOG to track version history, and README.md lacks badges that signal project quality and health to potential users.

## What Changes

- Create `LICENSE` file with MIT license (copyright holder: Amaury Constant)
- Create `CHANGELOG.md` following Keep a Changelog format with git history
- Add project quality badges to `README.md` (Go Report Card, GoDoc, License, CI status)

## Capabilities

### New Capabilities

None. This change is documentation-only and does not introduce new code capabilities.

### Modified Capabilities

None. No existing specifications require modification - this is purely project infrastructure.

## Impact

- Repository root: New files (LICENSE, CHANGELOG.md)
- README.md: Badge additions after title
- GoReleaser: Already references `LICENSE*` in archives - will now have a file to include
- Users: Will have legal clarity on usage rights and visibility into version history
