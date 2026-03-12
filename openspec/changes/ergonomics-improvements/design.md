## Context

The twiggit CLI is used frequently for worktree management. Current UX has friction points:
- `twiggit list` and `twiggit delete` are verbose compared to Unix conventions (`ls`, `rm`)
- `prune` command always prompts for confirmation on bulk operations, even when user is certain
- `list --all` lacks the common `-a` short form
- `create` and `init` Long descriptions duplicate flag information Cobra already displays
- `list` and `delete` lack practical examples in help text

## Goals / Non-Goals

**Goals:**
- Reduce typing for common commands via Unix-style aliases
- Provide auto-confirmation option for scripts and confident users
- Improve help text clarity by removing duplication
- Add practical examples to guide users

**Non-Goals:**
- Changing any existing flag behavior
- Adding new commands or features
- Modifying service layer or domain logic

## Decisions

### Decision 1: Command Aliases via Cobra

**Choice:** Use Cobra's built-in `Aliases` field on commands.

**Rationale:** Cobra natively supports aliases and automatically includes them in help text. No custom logic needed.

**Alternatives considered:**
- Shell aliases: Requires user setup, not portable
- Symlinks: Creates binary management complexity
- Wrapper script: Unnecessary indirection

### Decision 2: --yes/-y Flag Distinction from --force

**Choice:** Add separate `--yes/-y` flag for auto-confirmation, distinct from `--force`.

**Rationale:** Clear semantic distinction:
- `--force` = bypass safety checks (uncommitted changes, merged status)
- `--yes` = auto-confirm prompts (keep safety checks)

**Alternatives considered:**
- Reuse `--force` for both: Confusing, loses semantic clarity
- Environment variable: Less discoverable, not per-command

### Decision 3: Help Text Cleanup via Examples Sections

**Choice:** Remove flag descriptions from Long text, add `Examples:` sections like `prune` has.

**Rationale:** Cobra already displays flags in the `Flags:` section. Duplicating in Long is redundant. Examples provide practical guidance without redundancy.

**Alternatives considered:**
- Keep current format: Maintains duplication problem
- Comprehensive flag documentation: Belongs in AGENTS.md, not command help

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| `ls`/`rm` aliases conflict with shell builtins | Aliases only apply when invoked as `twiggit ls` - shell builtins take precedence in direct usage |
| `--yes` may encourage unsafe bulk operations | Safety checks still apply; `--yes` only skips prompts, `--force` still required for dirty worktrees |
| Help text changes may surprise existing users | Changes are additive (aliases, -a flag, --yes) or clarifying (examples) - no removals |
