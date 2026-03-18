# Suggestions for explicit-error-formatting

## 2026-03-18 - PHASE2 Verification

- [ ] **[docs]** Document explicit error formatter pattern in cmd/AGENTS.md
  - Location: cmd/AGENTS.md
  - Impact: Low
  - Notes: Proposal.md lists cmd/AGENTS.md as a modified file for documenting the explicit error formatter pattern, but no task exists for this. The implementation is complete and functional, but documentation would help future developers understand the pattern.
  - Recommendation: Add a brief section in cmd/AGENTS.md explaining the explicit strategy pattern with errors.As() matching used in error_formatter.go, including the matcher-formatter pair registration approach and ordered iteration logic.
