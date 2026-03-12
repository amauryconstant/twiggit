## Why

Users need easy installation methods beyond the current install script. Homebrew is the standard package manager for macOS and Linux users, and GitHub Releases provides broader discoverability. Additionally, contributors lack clear documentation on development setup and contribution guidelines.

## What Changes

- Add Homebrew tap configuration to `.goreleaser.yml` for `amoconst/homebrew-tap`
- Configure dual release targets (GitLab primary, GitHub secondary) in `.goreleaser.yml`
- Create `CONTRIBUTING.md` with complete contributor guide
- Update `README.md` installation section with Homebrew instructions and GitHub Releases link

## Capabilities

### New Capabilities

- `homebrew-distribution`: Homebrew formula generation and tap publishing configuration
- `github-releases`: GitHub as secondary release target alongside GitLab
- `contributor-guide`: Developer documentation for contributors

### Modified Capabilities

- None (documentation and build configuration only, no spec-level behavior changes)

## Impact

- **Build Configuration**: `.goreleaser.yml` updated with `brews` section and GitHub release config
- **Documentation**: New `CONTRIBUTING.md`, updated `README.md`
- **Release Process**: Automated Homebrew formula publishing, dual-platform releases
- **External Dependencies**: Requires `amoconst/homebrew-tap` repository to exist
