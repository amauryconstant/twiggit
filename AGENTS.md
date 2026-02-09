# Twiggit - OpenCode Reference

## Essential Commands

| Command | Purpose |
|---------|---------|
| mise run test | Quick tests |
| mise run test:full | Full test suite (unit/integration/e2e/race) |
| mise run lint:fix | Lint + format |
| mise run check | All validation |
| mise run build | Build binary |
| mise tasks | List all tasks |

## Release

| Command | Purpose |
|---------|---------|
| mise run release:validate | Clean tree check |
| mise run release:tag patch | Tag v0.5.0 |
| mise run release:dry-run | Test GoReleaser |

## Pre-Commit Hooks

Setup: `mise install && pre-commit install`
Run: `pre-commit run --all-files`
Skip: `git commit -m "msg" --no-verify`

## Specification Keywords
| Keyword | Meaning | Usage |
|---------|---------|-------|
| SHALL | Mandatory | Critical functionality |
| SHALL NOT | Absolute prohibition | Security boundaries |
| SHOULD | Recommended | Conventional patterns |
| SHOULD NOT | Discouraged | Anti-patterns |
| WILL/WILL NOT | System facts | Behavior declarations |
| MAY/MAY NOT | Optional | Extensibility points |

## Location-Specific Guides
cmd/AGENTS.md              # CLI commands, Cobra patterns
internal/application/AGENTS.md # Interface definitions
internal/services/AGENTS.md # Service implementation patterns
internal/domain/AGENTS.md   # Domain model conventions
internal/infrastructure/AGENTS.md # Git client routing, config
test/integration/AGENTS.md  # Testify suite patterns
test/e2e/AGENTS.md          # Ginkgo/Gomega CLI testing
test/e2e/fixtures/AGENTS.md # E2E fixture usage
test/helpers/AGENTS.md      # Test utilities (repo, git, shell)
