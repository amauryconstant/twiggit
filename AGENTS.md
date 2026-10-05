# Twiggit - OpenCode Reference

Pragmatic git worktree CLI. Functional-core + I/O-adapter layout.

## Architecture

| Package | Role |
|---|---|
| `internal/core/` | Pure value objects, errors, validation. No I/O. |
| `internal/git/` | I/O adapter: client, writer, hook runner, context resolver. |
| `internal/config/` | koanf config loader. |
| `internal/output/` | Formatters + error renderer. |
| `internal/iostreams/` | TTY, IOStreams, slog singleton. |
| `internal/cmdutil/` | Factory, ExitCodeFor, persistent flags. |
| `cmd/` + `main.go` | Cobra tree + composition root. |

Depguard owns the import rules (see `.golangci.yml`).

## Commands

| Task | Command |
|---|---|
| Verify everything | `mise run verify` |
| Run all tests | `mise run test` |
| Run e2e | `mise run test:e2e` |
| Lint | `mise run lint:check` |
| Format | `mise run format` |
| Build | `mise run build` |
| Update goldens | `UPDATE_GOLDEN=true mise run test:e2e` |

## CLI Surface

| Command | Alias | Purpose |
|---|---|---|
| `list` | `ls` | List worktrees |
| `create` | - | New worktree + post-create hooks |
| `delete` | `rm` | Remove worktree (+ optional branch) |
| `prune` | - | Delete merged worktrees |
| `cd` | - | Print worktree path |
| `init` | - | Shell wrapper |
| `version` | - | Build version |

## Conventions

- **Bool prefix.** `is/has/can` on every bool field. `Was*` on past-tense bools (`WasDeleted`).
- **Single-handling.** Adapter returns wrapped errors only; logging at cmd boundary via `opts.IO.Logger`.
- **Sentinels.** Nine `core.Err*` live in `internal/core/sentinels.go`. Match via `errors.Is`, never by `Op` string.
- **Factory roles.** No per-role lazy fields. Callers do `f.GitClient()` then assign to a role-typed local.
- **Exit codes.** 0=ok, 1=error, 2=usage. `cmdutil.ExitCodeFor` is the single mapping.

## Subdirectory Guides

| File | Purpose |
|---|---|
| [cmd/AGENTS.md](cmd/AGENTS.md) | Cobra patterns, command tree |
| [openspec/AGENTS.md](openspec/AGENTS.md) | OpenSpec workflow + spec rules |
| [test/AGENTS.md](test/AGENTS.md) | Test organization |
| [test/integration/AGENTS.md](test/integration/AGENTS.md) | Testify suites |
| [test/e2e/AGENTS.md](test/e2e/AGENTS.md) | Ginkgo/Gomega CLI testing |
| [test/concurrent/AGENTS.md](test/concurrent/AGENTS.md) | Race detector patterns |
| [docs/release/AGENTS.md](docs/release/AGENTS.md) | Release process, cosign verification |
| [internal/version/AGENTS.md](internal/version/AGENTS.md) | Build-time ldflags |
