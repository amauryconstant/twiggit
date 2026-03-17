## 1. Create Version Package

- [x] 1.1 Create internal/version/version.go with Version, Commit, Date variables (defaults: "dev", "", "")
- [x] 1.2 Implement String() function returning formatted output: `<version> (<full-commit>) <date>` (complete), `<version> () ` (empty commit), or `<version> (<commit>) ` (empty date with commit present) - note: String() does NOT include "twiggit " prefix, command layer adds it

## 2. Migrate Version Command

- [x] 2.1 Update cmd/version.go to import internal/version and call version.String()
- [x] 2.2 Remove version variable declarations from cmd/version.go

## 3. Update Build Configuration

- [x] 3.1 Update .goreleaser.yml ldflags from `twiggit/cmd.Version` to `twiggit/internal/version.Version`
- [x] 3.2 Update .goreleaser.yml ldflags for Commit and Date paths

## 4. Verify and Document

- [x] 4.1 Run mise run build and verify version command output unchanged
- [x] 4.2 Run mise run test:full to verify no regressions
