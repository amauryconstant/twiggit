## Context

Twiggit currently releases to GitLab only, with a simple install script for users. To improve discoverability and ease of installation, we need to:
1. Support Homebrew installation (standard for macOS/Linux developers)
2. Publish releases to GitHub for broader visibility
3. Provide contributor documentation

The current `.goreleaser.yml` handles GitLab releases. GoReleaser supports multiple release targets and Homebrew formula generation natively.

## Goals / Non-Goals

**Goals:**
- Configure GoReleaser to publish Homebrew formulas to `amoconst/homebrew-tap`
- Configure dual release targets (GitLab primary, GitHub secondary)
- Create comprehensive `CONTRIBUTING.md`
- Update `README.md` with Homebrew installation instructions

**Non-Goals:**
- Changes to application code
- Changes to existing GitLab release process
- Automated tap repository creation (manual prerequisite)

## Decisions

### Decision 1: GoReleaser Brews Configuration
**Choice:** Use GoReleaser's native `brews` section to generate and publish formula.

**Rationale:** GoReleaser handles formula generation, commit, and push automatically. No manual formula maintenance required.

**Alternatives Considered:**
- Manual formula in tap repo: Requires manual updates per release
- Third-party Homebrew tools: Additional complexity, GoReleaser is already in use

### Decision 2: Dual Release Targets
**Choice:** Configure GitLab as primary and GitHub as secondary release target.

**Rationale:** GitLab remains the canonical source while GitHub provides discoverability. GoReleaser supports multiple `release` configurations.

**Alternatives Considered:**
- GitHub only: Would break existing GitLab release workflow
- Separate pipelines: More complex, error-prone

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
