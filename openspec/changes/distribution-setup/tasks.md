## 1. Homebrew Tap Setup

- [ ] 1.1 Add `brews` section to `.goreleaser.yml` with tap repository configuration for `amoconst/homebrew-tap`
- [ ] 1.2 Update README.md with Homebrew installation instructions

## 2. GitHub Releases

- [ ] 2.1 Add GitHub as secondary release target in `.goreleaser.yml` alongside existing GitLab configuration

## 3. Contributor Documentation

- [ ] 3.1 Create `CONTRIBUTING.md` with development setup, testing, code style, and PR process sections

## 4. README Installation Section Update

- [ ] 4.1 Update README.md installation section with Homebrew, GitHub Releases, and verification step

## 5. Verification

- [ ] 5.1 Run `mise run check` to verify all tests pass
- [ ] 5.2 Run `mise run release:dry-run` to verify GoReleaser configuration
