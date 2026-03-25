// Package domain contains the core business logic and types.
//
// This package must have zero external dependencies (enforced by depguard).
//
// Key types:
//   - Result[T] - Functional error handling
//   - ValidationPipeline[T] - Composable validators
//   - Domain errors with immutable builders
//
// All domain types should be testable without external dependencies.
package domain
