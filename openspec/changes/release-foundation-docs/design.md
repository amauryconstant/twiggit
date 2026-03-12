## Context

Twiggit is preparing for public open-source release. The codebase is stable, but lacks the standard documentation files expected in open-source projects: a license, changelog, and visible quality badges.

**Current state:**
- `.goreleaser.yml` references `LICENSE*` but no LICENSE file exists
- No CHANGELOG.md to track version history
- README.md lacks quality signal badges

**Constraints:**
- Documentation-only change (no code modifications)
- Must follow standard open-source conventions
- Should integrate with existing tooling (GoReleaser)

## Goals / Non-Goals

**Goals:**
- Add MIT license with correct copyright holder
- Create changelog with complete git history
- Add quality badges to README for discoverability

**Non-Goals:**
- Code changes of any kind
- Modifying build or release configuration
- Creating additional documentation (CONTRIBUTING, CODE_OF_CONDUCT, etc.)

## Decisions

1. **MIT License**
   - **Choice:** Use MIT license
   - **Rationale:** Most common for Go CLIs; permissive; well-understood
   - **Alternatives:** Apache 2.0 (more verbose, patent clause), GPL (copyleft, not suitable for CLI tool)

2. **Keep a Changelog Format**
   - **Choice:** Follow keepachangelog.com format
   - **Rationale:** Industry standard; machine-readable; easy to maintain
   - **Alternatives:** Simple version list (loses context), GitHub Releases only (not in repo)

3. **Badge Selection**
   - **Choice:** Go Report Card, GoDoc, License, CI status
   - **Rationale:** These are the essential quality signals for Go projects
   - **Alternatives:** More badges (coverage, downloads) - can be added later

## Risks / Trade-offs

- **Badge link rot** → Use stable badge services (goreportcard, pkg.go.dev)
- **Changelog maintenance burden** → Use `openspec-generate-changelog` skill to automate from archived changes
