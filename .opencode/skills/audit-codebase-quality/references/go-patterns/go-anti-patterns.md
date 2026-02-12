# Go Anti-Patterns

Common Go anti-patterns, code smells, and security issues to watch for during codebase auditing.

## Code Smells

- **Long parameter lists**: Functions with >5 parameters indicate need for struct
- **Deeply nested code**: >3 levels of nesting indicates need for extraction
- **God functions**: Functions that do too many things (>50 lines, multiple switch statements)
- **Magic numbers**: Unexplained numeric literals (use constants)
- **Flag parameters**: Boolean parameters that change function behavior (should be separate functions)

## Security Anti-Patterns

- **SQL injection**: String concatenation in database queries
- **Path traversal**: Not validating user-provided file paths
- **Command injection**: Passing user input directly to exec.Command
- **Hardcoded secrets**: API keys, passwords in source code

## Performance Anti-Patterns

- **N+1 in loops**: Querying database inside loop instead of single query with IN clause
- **String concatenation in loops**: Using `+` instead of strings.Builder
- **Unnecessary allocations**: Creating slices/strings in loops instead of pre-allocating

## Audit-Specific Patterns

### For Package Structure Audits

- Check that `domain/` has no internal dependencies
- Verify all service interfaces are in `application/interfaces.go`
- Ensure infrastructure depends only on `domain/`
- Verify services depends on appropriate layers (application, domain, infrastructure)

### For Duplicate Code Audits

- Look for identical function signatures across files
- Check for similar logic with minor variations
- Identify template-like code (copy-paste with small changes)
- Find duplicate struct definitions

### For Interface Compliance Audits

- Verify all interface methods have implementations
- Check for unused interface definitions
- Ensure interface methods match signatures exactly
- Look for interface methods only called in tests

### For Test Pattern Audits

- Verify all test files follow `<source>_test.go` naming
- Check for inline mocks that should be in `test/mocks/`
- Identify duplicate mock implementations
- Find files without corresponding tests (if logic exists)

### For Documentation Accuracy Audits

- Compare AGENTS.md struct definitions with actual code
- Check interface method signatures match implementations
- Look for missing field documentation
- Find undocumented types and methods

### For Import Consistency Audits

- Verify import ordering: stdlib (alphabetical), third-party (alphabetical), internal (alphabetical)
- Check for circular dependencies (try `go build ./...`)
- Look for unused imports (run `golangci-lint`)
- Verify layer dependencies are correct
