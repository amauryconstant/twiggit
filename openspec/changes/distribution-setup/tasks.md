## 1. Homebrew Tap Setup

- [x] 1.1 Add `brews` section to `.goreleaser.yml` with tap repository configuration for `amoconst/homebrew-tap`
- [x] 1.2 Update README.md with Homebrew installation instructions

## 2. GitHub Releases

- [x] 2.1 Add GitHub as secondary release target in `.goreleaser.yml` alongside existing GitLab configuration

## 3. Contributor Documentation

- [x] 3.1 Create `CONTRIBUTING.md` with development setup, testing, code style, and PR process sections

## 4. README Installation Section Update

- [x] 4.1 Update README.md installation section with Homebrew, GitHub Releases, and verification step

## 5. Verification

- [x] 5.1 Run `mise run check` to verify all tests pass
- [x] 5.2 Run `mise run release:dry-run` to verify GoReleaser configuration
