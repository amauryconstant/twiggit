## 1. Create Version Package

- [ ] 1.1 Create internal/version/version.go with Version, Commit, Date variables (defaults: "dev", "", "")
- [ ] 1.2 Implement String() function returning formatted output: `<version> (<full-commit>) <date>` (complete), `<version> () ` (empty commit), or `<version> (<commit>) ` (empty date with commit present) - note: String() does NOT include "twiggit " prefix, command layer adds it

## 2. Migrate Version Command

- [ ] 2.1 Update cmd/version.go to import internal/version and call version.String()
- [ ] 2.2 Remove version variable declarations from cmd/version.go

## 3. Update Build Configuration

- [ ] 3.1 Update .goreleaser.yml ldflags from `twiggit/cmd.Version` to `twiggit/internal/version.Version`
- [ ] 3.2 Update .goreleaser.yml ldflags for Commit and Date paths

## 4. Verify and Document

- [ ] 4.1 Run mise run build and verify version command output unchanged
- [ ] 4.2 Run mise run test:full to verify no regressions
