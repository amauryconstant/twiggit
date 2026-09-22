# Spec Delta: Infrastructure Path Utils

## REMOVED Requirements

### Requirement: `ExtractProjectFromWorktreePath`

**Reason**: The helper moves to the new `domain-path-utils` capability
as part of the layer inversion. The domain layer is the right home: the
function is filesystem-pure and depends only on stdlib. Previous
consumers reached it via `internal/infrastructure/` cross-layer imports.

**Migration**: Callers SHALL use the new `domain-path-utils` capability's
helper instead. The function signature returns plain `error`; callers
wrap the cause at the service boundary into the appropriate
`*ServiceError` per `domain-typed-errors`. The new `domain-path-utils`
capability owns the requirement after archive.

### Requirement: `NormalizePath`

**Reason**: Same as `ExtractProjectFromWorktreePath`. The helper moves
to `domain-path-utils` and returns plain `error`.

**Migration**: Callers SHALL use the new `domain-path-utils` capability's
helper instead. The new `domain-path-utils` capability owns the
requirement after archive.

### Requirement: `IsPathUnder`

**Reason**: Same as `ExtractProjectFromWorktreePath`. The helper moves
to `domain-path-utils` and returns plain `error`.

**Migration**: Callers SHALL use the new `domain-path-utils` capability's
helper instead. The new `domain-path-utils` capability owns the
requirement after archive.

### Requirement: Error wrapping

**Reason**: The path utility helpers move to `domain-path-utils` as part
of the layer inversion. Once they live in the domain layer, the mandate
to wrap failures via `domain.NewContextDetectionError` creates a
domain-to-domain self-wrap cycle, since `domain.ContextDetectionError`
is itself a domain type. Returning plain `error` keeps the helpers
pure; callers wrap at the service boundary into the appropriate
`*ServiceError` per `domain-typed-errors`.

**Migration**: Callers that previously relied on typed inspection of the
wrapping error SHALL switch to wrapping at the service boundary into the
appropriate `*ServiceError` (per `domain-typed-errors`) with the original
cause as the wrapped error. The unwrapped error chain preserves cause
for `errors.Is` / `errors.As` walks through the new wrapper.
