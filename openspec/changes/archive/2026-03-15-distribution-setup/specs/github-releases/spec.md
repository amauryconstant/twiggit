## ADDED Requirements

### Requirement: Dual release targets configured
The `.goreleaser.yml` configuration SHALL publish releases to GitLab (canonical source) and create GitHub release pages for discoverability.

#### Scenario: GitLab release (primary)
- **WHEN** GoReleaser runs during release
- **THEN** release artifacts SHALL be published to GitLab at `amoconst/twiggit`
- **AND** archives, checksums, and SBOMs SHALL be available for download

#### Scenario: GitHub release (discoverability)
- **WHEN** GoReleaser runs during release
- **THEN** a GitHub release page SHALL be created at `amauryconstant/twiggit`
- **AND** the release page SHALL contain a link to GitLab downloads
- **AND** artifacts SHALL NOT be duplicated on GitHub (to maintain GitLab as canonical source)

### Requirement: GitHub release page points to GitLab
The GitHub release page SHALL direct users to GitLab for artifact downloads.

#### Scenario: Release notes contain GitLab link
- **WHEN** GitHub release page is created
- **THEN** release notes SHALL include a link to GitLab downloads
- **AND** the link SHALL use the format: `https://gitlab.com/amoconst/twiggit/-/releases/<TAG>`

#### Scenario: Download redirection
- **WHEN** user views GitHub release
- **THEN** clear instructions SHALL guide them to download from GitLab
- **AND** artifacts SHALL NOT be available for download from GitHub
