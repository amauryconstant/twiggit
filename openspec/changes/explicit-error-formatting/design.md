## Context

The current error formatter in `cmd/error_formatter.go` uses reflection-based type dispatch via a `map[reflect.Type]formatterFunc`. This pattern:
- Requires runtime type inspection overhead
- Obscures the mapping between error types and formatters
- Makes debugging difficult when wrong formatter is selected
- Requires maintaining reflection-based type registration

The proposed change replaces this with an explicit strategy pattern using `errors.As()` for type matching.

## Goals / Non-Goals

**Goals:**
- Replace reflection-based type dispatch with explicit `errors.As()` matching
- Define clear `matcherFunc` and `formatterFunc` types
- Create explicit matcher functions for each error type
- Maintain identical error output behavior (no user-facing changes)

**Non-Goals:**
- Changing error message format or content
- Adding new error types or formatters
- Modifying error handling in other layers

## Decisions

### Decision 1: Strategy Pattern with Function Types

**Choice:** Define `matcherFunc func(error) bool` and `formatterFunc func(error) string` types, register as ordered pairs.

**Rationale:** Explicit functions make the type-to-formatter mapping clear at registration site. No reflection needed for dispatch.

**Alternatives:**
- Type switch: Would require all error types in same package, violates domain isolation
- Interface-based matching: Would require adding `Format()` method to domain errors, violates separation of concerns

### Decision 2: Ordered Matcher Iteration

**Choice:** Store matchers in registration order, iterate during `Format()` until match found.

**Rationale:** Allows explicit priority control (e.g., validation errors before generic service errors). Order-dependent matching is intentional and documented.

**Alternatives:**
- Priority field: Overkill for 4-5 matchers, adds complexity
- Map-based with priority key: Still requires ordering logic

### Decision 3: Formatter Functions Without Receiver

**Choice:** Formatters accept `error` directly, not `ErrorFormatter` receiver.

**Rationale:** Pure functions are easier to test and reason about. No need for formatter state access.

**Alternatives:**
- Keep receiver: Unnecessary coupling, formatters don't need formatter state

## Risks / Trade-offs

Risk: Matcher order determines which formatter handles wrapped errors
→ **Mitigation:** Document order requirement in code comments, registration order: ValidationError → WorktreeServiceError → ProjectServiceError → ServiceError

**Risk:** Performance difference from reflection vs iteration
→ **Mitigation:** Negligible for error formatting (cold path), 4-5 iterations max

**Trade-off:** Explicit matchers require manual addition for new error types
→ **Acceptance:** This is intentional - makes new error types visible at registration site
