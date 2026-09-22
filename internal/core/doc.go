// Package core provides the functional core of twiggit: pure value objects,
// validation, business rules, and the canonical error type hierarchy.
//
// Layering rule: this package imports only the Go standard library and
// github.com/samber/lo. No filesystem, network, exec, or git access. All I/O
// adapters live in sibling packages (internal/git, internal/config,
// internal/output, internal/iostreams, internal/cmdutil) and reach into core
// via the types and rules declared here.
package core
