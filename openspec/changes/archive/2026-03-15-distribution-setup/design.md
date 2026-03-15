## Context

Twiggit currently releases to GitLab only, with a simple install script for users. To improve discoverability and ease of installation, we need to:
1. Support Homebrew installation (standard for macOS/Linux developers)
2. Publish release pages to GitHub for broader visibility (artifacts remain on GitLab)
3. Provide contributor documentation

The current `.goreleaser.yml` handles GitLab releases. GoReleaser supports multiple release targets and Homebrew formula generation natively.

## Goals / Non-Goals

**Goals:**
- Configure GoReleaser to publish Homebrew formulas to `amoconst/homebrew-tap`
- Configure GitLab as canonical artifact source, GitHub for discoverability only
- Create GitHub release pages that point to GitLab downloads (not duplicate artifacts)
- Create comprehensive `CONTRIBUTING.md`
- Update `README.md` with Homebrew installation instructions

**Non-Goals:**
- Changes to application code
- Changes to existing GitLab release process
- Automated tap repository creation (manual prerequisite)
- Artifact duplication on GitHub (maintain GitLab as single source of truth)

## Decisions

### Decision 1: GoReleaser Brews Configuration
**Choice:** Use GoReleaser's native `brews` section to generate and publish formula.

**Rationale:** GoReleaser handles formula generation, commit, and push automatically. No manual formula maintenance required.

**Alternatives Considered:**
- Manual formula in tap repo: Requires manual updates per release
- Third-party Homebrew tools: Additional complexity, GoReleaser is already in use

### Decision 2: GitHub for Discoverability Only
**Choice:** Configure GitLab as primary artifact publisher, create GitHub release pages for discoverability.

**Rationale:** GitLab remains the canonical source for artifacts while GitHub provides visibility. This avoids artifact duplication and maintains single source of truth. A hook script creates GitHub release pages with links to GitLab downloads.

**Alternatives Considered:**
- Full dual-publishing: Would duplicate artifacts, increase complexity, create version synchronization challenges
- GitHub only: Would break existing GitLab release workflow and GitLab users
- Manual GitHub releases: Additional manual step per release, error-prone

### Decision 3: CONTRIBUTING.md Structure
**Choice:** Follow standard open-source conventions with sections for setup, testing, code style, and PR process.

**Rationale:** Familiar structure for contributors. Reference existing `AGENTS.md` for project-specific details.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Tap repository doesn't exist | Document as prerequisite; verify with `mise run release:dry-run` |
| GitHub remote not configured | Verify with `git remote -v` before release |
| Formula test command fails | Use simple `twiggit version` as test |
| Contributors miss project conventions | Link to AGENTS.md from CONTRIBUTING.md |
| Users confused by GitHub release page not having downloads | Clear messaging in release notes pointing to GitLab |
