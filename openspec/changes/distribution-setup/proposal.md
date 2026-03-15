## Why

Users need easy installation methods beyond the current install script. Homebrew is the standard package manager for macOS and Linux users. GitHub release pages provide broader discoverability while GitLab remains the canonical source for artifacts. Additionally, contributors lack clear documentation on development setup and contribution guidelines.

## What Changes

- Add Homebrew tap configuration to `.goreleaser.yml` for `amoconst/homebrew-tap`
- Configure GitLab as canonical artifact publisher with GitHub release pages for discoverability
- Create hook script to generate GitHub release pages pointing to GitLab downloads
- Create `CONTRIBUTING.md` with complete contributor guide
- Update `README.md` installation section with Homebrew instructions and GitHub Releases link

## Capabilities

### New Capabilities

- `homebrew-distribution`: Homebrew formula generation and tap publishing configuration
- `github-discoverability`: GitHub release pages for broader visibility (artifacts remain on GitLab)
- `contributor-guide`: Developer documentation for contributors

### Modified Capabilities

- None (documentation and build configuration only, no spec-level behavior changes)

## Impact

- **Build Configuration**: `.goreleaser.yml` updated with `brews` section and hook script for GitHub releases
- **Documentation**: New `CONTRIBUTING.md`, updated `README.md`
- **Release Process**: Automated Homebrew formula publishing, GitLab artifacts with GitHub discoverability
- **External Dependencies**: Requires `amoconst/homebrew-tap` repository to exist
- **Artifact Strategy**: Single source of truth on GitLab, GitHub provides visibility
