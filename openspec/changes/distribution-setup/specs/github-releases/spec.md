## ADDED Requirements

### Requirement: Dual release targets configured
The `.goreleaser.yml` configuration SHALL publish releases to both GitLab and GitHub.

#### Scenario: GitLab release (primary)
- **WHEN** GoReleaser runs during release
- **THEN** release artifacts SHALL be published to GitLab at `amoconst/twiggit`

#### Scenario: GitHub release (secondary)
- **WHEN** GoReleaser runs during release
- **THEN** release artifacts SHALL be published to GitHub at `amauryconstant/twiggit`

### Requirement: Same artifacts on both platforms
Both release targets SHALL publish identical artifacts (archives, checksums, SBOMs).

#### Scenario: Artifact consistency
- **WHEN** comparing releases on GitLab and GitHub
- **THEN** the same version SHALL have identical archives, checksums, and changelog
