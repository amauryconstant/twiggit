## Context

Currently, twiggit outputs human-readable text only, making scripting difficult:
- `list` output must be parsed with regex/awk
- Success messages clutter script logs
- Bulk operations like `prune --all` provide no progress feedback

The existing `logv()` function in `cmd/util.go` handles verbose output to stderr. We extend this pattern for JSON output and quiet mode.

## Goals / Non-Goals

**Goals:**
- Add `--output/-o` flag for JSON output format on `list` command
- Add global `--quiet/-q` flag to suppress non-essential output
- Add progress reporting during `prune --all` bulk operations
- Ensure output is predictable and parseable for scripting
- Maintain backward compatibility (text output remains default)

**Non-Goals:**
- Adding JSON output to all commands (start with `list`, extend later)
- Complex progress bars with spinners (simple text progress is sufficient)
- External dependencies for progress reporting (use stdlib)

## Decisions

### Decision 1: Output Formatter Interface

**Choice:** Create `OutputFormatter` interface in `cmd/` with `TextFormatter` and `JSONFormatter` implementations.

**Rationale:** Interface allows easy extension to other formats (YAML, TOML) and commands without changing service layer.

**Alternatives:**
- Template-based approach: Less type-safe, harder to ensure JSON validity
- Service layer formatting: Violates layer separation (service returns domain types, cmd formats)

```go
type OutputFormatter interface {
    FormatWorktrees(worktrees []*domain.WorktreeInfo) string
}

type TextFormatter struct{}
type JSONFormatter struct{}
```

### Decision 2: Quiet Mode Implementation

**Choice:** Global persistent flag `--quiet/-q` on root command, checked by output functions.

**Rationale:** Consistent with Unix conventions (`grep -q`, `curl -s`). Global flag applies to all commands uniformly.

**Alternatives:**
- Per-command quiet flag: Inconsistent, more boilerplate
- Environment variable: Less discoverable, conflicts with flags

**Behavior:**
- Suppresses success messages
- Suppresses hint messages
- Still outputs essential data (paths for `-C` mode)
- Still outputs errors to stderr
- `--verbose` wins over `--quiet` (verbose takes priority)

### Decision 3: Progress Reporting Design

**Choice:** Simple `ProgressReporter` struct in `cmd/util.go` with text-based progress output.

**Rationale:** No external dependencies, easy to test, sufficient for bulk operations. Progress goes to stderr to not interfere with stdout parsing.

**Alternatives:**
- External library (e.g., mpb): Adds dependency, overkill for simple use case
- Channel-based progress: More complex, requires goroutine coordination

```go
type ProgressReporter struct {
    quiet bool
    out   io.Writer
}

func (p *ProgressReporter) Report(format string, args ...interface{})
func (p *ProgressReporter) ReportProgress(current, total int, item string)
```

### Decision 4: JSON Output Structure

**Choice:** Structured JSON with `worktrees` array containing objects with `branch`, `path`, `status` fields.

**Rationale:** Common JSON structure, easy to parse with `jq` or standard JSON libraries.

```json
{
  "worktrees": [
    {
      "branch": "feature-1",
      "path": "/home/user/worktrees/myproject/feature-1",
      "status": "modified"
    }
  ]
}
```

**Status values:** `"clean"`, `"modified"`, `"detached"`

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| JSON format changes break scripts | Version the JSON structure, document stability guarantees |
| Quiet mode hides useful output | Document clearly what is suppressed, verbose wins |
| Progress output to stderr may surprise users | Document behavior, consistent with verbose pattern |
| Formatter interface adds indirection | Keep interface minimal, only essential methods |

## Migration Plan

1. Add new flags with defaults preserving current behavior
2. JSON output is opt-in (`--output json`)
3. Quiet mode is opt-in (`--quiet`)
4. Progress output is always-on for bulk ops (suppressed by `--quiet`)
5. No breaking changes to existing commands

## Open Questions

- Should JSON output include additional fields (commit hash, last modified)? → **Defer to future enhancement, keep initial implementation minimal**
- Should `--quiet` also suppress progress? → **Yes, quiet suppresses all non-essential output including progress**
