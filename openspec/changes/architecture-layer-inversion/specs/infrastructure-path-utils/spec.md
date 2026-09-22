# Spec Delta: Infrastructure Path Utils

## REMOVED Requirements

### Requirement: Error wrapping

**Reason**: The path utility helpers (`ExtractProjectFromWorktreePath`, `NormalizePath`, `IsPathUnder`) move to `internal/domain/pathutils.go` as part of the layer inversion. Once they live in the domain layer, the mandate to wrap failures via `domain.NewContextDetectionError` creates a domain-to-domain self-wrap cycle, since `domain.ContextDetectionError` is itself a domain type. Returning plain `error` keeps the helpers pure; callers in `internal/service/` wrap into the appropriate `domain.New*ServiceError()` per the existing service-layer error-wrapping contract.

**Migration**: Callers that previously relied on `errors.As(err, *domain.ContextDetectionError)` for typed inspection SHALL switch to wrapping at the service boundary into `domain.NewWorktreeServiceError`, `domain.NewProjectServiceError`, or `domain.NewNavigationServiceError` with the original cause as the wrapped error. The unwrapped error chain preserves cause for `errors.Is` / `errors.As` walks through the new wrapper.
