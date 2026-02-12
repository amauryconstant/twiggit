# Audit Areas

Configurable audit categories for codebase quality review. Each area defines what to look for, examples, and severity guidelines.

## Extensibility

Add new audit areas by creating sections below. When adding new areas:

1. Define what the area checks for
2. Provide concrete examples of findings
3. Establish severity guidelines
4. Reference language-specific patterns if applicable
5. Consider if the area should use parallel execution or sequential

---

## Package Structure

Check for naming conventions, directory organization, layer separation, and architectural violations.

### What to Look For

- Package naming consistency (singular vs plural)
- Directory structure violations (deep nesting, inappropriate locations)
- Layer boundary violations (e.g., service logic in domain layer)
- Interface placement (application vs domain vs infrastructure)
- Naming conflicts (same name in different packages)
- File naming conventions (snake_case vs camelCase)
- Orphaned or misplaced files

### Examples

- **Package naming**: `services` package (plural) while all others use singular (`application`, `domain`, `infrastructure`)
- **Layer violations**: Service interfaces defined in `domain/interfaces.go` instead of `application/interfaces.go`
- **Interface placement**: `ContextService` interface in wrong layer (domain instead of application)
- **Naming conflicts**: Both `domain.ContextService` (interface) and `services.ContextService` (implementation) exist with same name
- **Test file naming**: `validators_test.go` when source is `validation.go`

### Severity Guidelines

- **CRITICAL**: Architectural violations that break layer separation or design principles
- **HIGH**: Package naming violations that affect consistency across codebase
- **MEDIUM**: File naming violations or minor structural issues
- **LOW**: Stylistic variations that don't affect functionality

### Language-Specific Notes

For Go projects:
- Package names should be singular (e.g., `service`, not `services`)
- Interfaces in `application/`, implementations in `services/` or `infrastructure/`
- Domain layer should not define service interfaces
- Test files match source: `validation.go` → `validation_test.go`

See [go-patterns.md](go-patterns.md) for Go-specific patterns.

---

## Duplicate Code

Check for repeated logic, similar functions, and consolidation opportunities.

### What to Look For

- Identical or very similar functions in different files
- Duplicate struct or interface definitions
- Repeated error handling patterns
- Similar validation logic
- Copied code blocks
- Configuration or initialization logic repeated

### Examples

- **Duplicate templates**: Shell wrapper templates (bash, zsh, fish) defined in both `domain/shell.go` and `infrastructure/shell_infra.go`
- **Duplicate validation**: `isValidShellType()` function exists in both `domain/shell.go` and `services/shell_service.go`
- **Duplicate file listing**: `ConfigFiles()` and `getConfigFiles()` return nearly identical shell config file lists
- **Duplicate project discovery**: Similar directory scanning and validation logic in `context_resolver.go` and `project_service.go`
- **Repeated error handling**: Same path validation error pattern repeated 5 times in `context_resolver.go`

### Severity Guidelines

- **CRITICAL**: Complete duplication of complex logic (>100 lines)
- **HIGH**: Duplicate functions or templates (>20 lines)
- **MEDIUM**: Similar logic or repeated error patterns (5-20 lines)
- **LOW**: Minor repetitions or similar patterns

---

## Interface Compliance

Check for implementation completeness, signature mismatches, and unused interfaces.

### What to Look For

- Interface methods not implemented
- Implementation methods with wrong signatures
- Unused or unimplemented interfaces
- Interface methods only used in tests
- Overly broad or overly narrow interfaces

### Examples

- **Dead interfaces**: `GitRepositoryInterface` and `ProjectRepository` defined but never implemented or used
- **Signature mismatches**: `DeleteWorktree` has `force` parameter in interface but `keepBranch` in documentation (if inconsistency exists)
- **Unused methods**: `ContextDetector.InvalidateCacheForRepo()` and `ClearCache()` only used in tests, not production
- **Missing implementations**: Interface defined but no struct satisfies it

### Severity Guidelines

- **CRITICAL**: Critical interface in wrong layer causing architectural violation
- **HIGH**: Unused interfaces (>30 lines of dead code)
- **MEDIUM**: Signature mismatches or missing implementations
- **LOW**: Unused methods on otherwise valid interfaces

---

## Test Patterns

Check for test organization, coverage gaps, and mock usage consistency.

### What to Look For

- Test file organization and naming
- Mock patterns (inline vs centralized)
- Coverage gaps (files without tests)
- Test pattern consistency (suite-based vs table-driven)
- Mock duplication
- TDD artifacts or outdated comments

### Examples

- **Inline mocks**: `mockShellIntegration` defined in `shell_service_test.go` instead of centralized `test/mocks/cmd_mocks.go`
- **Mock duplication**: `MockWorktreeService`, `MockProjectService`, `MockNavigationService` exist in both `test/mocks/cmd_mocks.go` and `test/mocks/services/`
- **Coverage gaps**: Error type files (`errors.go`, `service_errors.go`, `shell_errors.go`) lack tests
- **Pattern mixing**: `cli_client_test.go` uses both suite-based and table-driven tests inconsistently
- **Naming inconsistency**: `ContextResolverMock` should be `MockContextResolver` for consistency

### Severity Guidelines

- **CRITICAL**: All application service mocks duplicated (causes maintenance burden)
- **HIGH**: Critical files missing tests (error constructors, validation logic)
- **MEDIUM**: Inline mocks that should be centralized, test pattern inconsistency
- **LOW**: TDD comments or minor naming variations

---

## Documentation Accuracy

Check AGENTS.md and other documentation files against actual code implementation.

### What to Look For

- Missing fields in struct documentation
- Incorrect method signatures
- Wrong parameter types or names
- Undocumented types (error types, result types, domain types)
- Missing methods in interface documentation
- Outdated patterns (e.g., referencing removed fields or deprecated approaches)
- Misleading information (non-existent features)

### Examples

- **Missing fields**: Documentation shows `Force bool` field in `CreateWorktreeRequest` (missing in docs)
- **Incorrect types**: Docs reference `ProjectRepository` but actual uses `application.ProjectService`
- **Missing types**: All error types (`ServiceError`, `WorktreeServiceError`, `ProjectServiceError`, etc.) completely undocumented
- **Wrong package names**: Docs show `service_requests.CreateWorktreeRequest` but actual is `domain.CreateWorktreeRequest`
- **Undocumented methods**: `GetRepositoryInfo`, `ListRemotes`, `GetCommitInfo` on `GoGitClient` not documented
- **Missing context**: Shell interface methods (`Type()`, `Path()`, `Version()`) completely undocumented

### Severity Guidelines

- **CRITICAL**: Documentation completely missing for critical types (error types, result types)
- **HIGH**: Incorrect signatures or types affecting architectural understanding
- **MEDIUM**: Missing fields or methods (documentation is incomplete but usable)
- **LOW**: Typos or minor inaccuracies

---

## Import Consistency

Check for import ordering, circular dependencies, and dependency health.

### What to Look For

- Import ordering violations (stdlib, third-party, internal)
- Circular dependencies between packages
- Unused imports
- Import aliases (used appropriately)
- External dependency management
- Layer dependency violations (e.g., services importing domain only, not application)

### Examples

- **Circular dependencies**: A imports B, B imports C, C imports A (rare but critical)
- **Import ordering**: Files mixing stdlib and third-party imports randomly
- **Unused imports**: Import statement but package never referenced
- **Wrong layer dependencies**: Application layer importing infrastructure directly (should go through services)

### Severity Guidelines

- **CRITICAL**: Circular dependencies breaking compilation or design
- **HIGH**: Layer violations or dependency injection issues
- **MEDIUM**: Import ordering or unused imports
- **LOW**: Inconsistent alias usage

### Language-Specific Notes

For Go projects:
- Order: stdlib (alphabetical), third-party (alphabetical), internal (alphabetical)
- No circular dependencies (verified via `go build`)
- Minimal external dependencies
- Clean layer separation: `services` → `application`, `domain`, `infrastructure`

See [go-patterns.md](go-patterns.md) for Go-specific patterns.

---

## Error Handling

Check for error wrapping patterns, error type usage, and message consistency.

### What to Look For

- Mixed error wrapping styles (fmt.Errorf vs domain error types)
- Inconsistent error type selection
- Error message format inconsistency (casing, prefixes)
- Loss of error context (not using %w)
- Error chain breaks (converting errors to nil returns)
- String-based error detection (strings.Contains instead of errors.As)
- Panic in production code
- Inconsistent boundary error handling

### Examples

- **Mixed wrapping**: Infrastructure layer sometimes returns `fmt.Errorf()` and sometimes `domain.NewGitRepositoryError()`
- **Plain errors**: Domain constructors using `errors.New()` instead of `ValidationError`
- **Context loss**: `context_detector.go` returns plain error without wrapping underlying cause
- **Message casing**: Some errors use "failed to", others use "Failed to"
- **Chain breaks**: Errors converted to nil returns, then checked separately
- **String detection**: `strings.Contains(err.Error(), "worktree not found")` instead of type checking

### Severity Guidelines

- **CRITICAL**: Panic in production code, complete loss of error context
- **HIGH**: String-based error detection (fragile and error-prone)
- **MEDIUM**: Mixed wrapping styles, error chain breaks
- **LOW**: Message casing inconsistency or minor format variations

---

## Mock Centralization

Check for inline vs centralized mocks, duplicate implementations, and missing mocks.

### What to Look For

- Inline mock definitions in test files
- Duplicate mock implementations for same interface
- Missing centralized mocks for infrastructure interfaces
- Mock naming inconsistencies
- Mocks with incomplete method implementations

### Examples

- **Inline mocks**: `mockShellIntegration` in `shell_service_test.go`, `mockCommandExecutor` in `command_executor_test.go`
- **Duplicate mocks**: `MockWorktreeService`, `MockProjectService`, `MockNavigationService` exist in both `test/mocks/cmd_mocks.go` and `test/mocks/services/`
- **Missing mocks**: No `MockShellInfrastructure` in centralized mocks
- **Naming inconsistency**: `ContextResolverMock` should be `MockContextResolver` for consistent naming pattern

### Severity Guidelines

- **CRITICAL**: All application service mocks duplicated (maintenance burden)
- **HIGH**: Inline mocks for commonly used interfaces
- **MEDIUM**: Missing centralized mocks, naming inconsistencies
- **LOW**: Mocks for niche or rarely-used interfaces

---
 
## Cross-References Between Audit Areas

When documenting or implementing audit areas, reference related areas to help users understand relationships:

### Common Cross-References

- **Error handling** ↔ **Import consistency**: Mixed error wrapping often causes import issues
- **Duplicate code** ↔ **Interface compliance**: Duplicate interfaces may not implement all methods
- **Test patterns** ↔ **Mock centralization**: Test mock usage issues relate to mock organization
- **Documentation accuracy** ↔ **Package structure**: Incorrect documentation often reflects structural issues

### Audit Area Relationships

| Area | Related To | Rationale |
|-------|-----------|-----------|
| package-structure | Interface compliance, Duplicate code, Test patterns | Layer structure affects interface placement and code organization |
| duplicate-code | Interface compliance, Documentation accuracy | Duplicated code may not match documented interfaces |
| interface-compliance | Package structure, Mock centralization | Interface placement affects implementation patterns |
| test-patterns | Mock centralization, Import consistency | Test organization relates to mock and import patterns |
| documentation-accuracy | Package structure, Interface compliance | Documentation should match code structure |
| import-consistency | Package structure, Error handling | Import layering affects error propagation |
| error-handling | Package structure, Documentation accuracy, Test patterns | Error handling depends on architectural decisions |
| mock-centralization | Test patterns, Interface compliance | Mock usage is part of test organization |

### Guidelines

- When adding a new audit area, review existing areas and add cross-references
- When documenting findings, check if related findings exist in other areas and reference them
- Cross-references help users understand the interconnected nature of code quality issues

### Examples

- **Example 1**: When documenting `import-consistency`, add cross-reference to `package-structure` since poor import organization often reflects layer violations
- **Example 2**: When documenting `duplicate-code`, add cross-reference to `interface-compliance` since duplicated code often means interface design issues
- **Example 3**: When documenting `error-handling`, add cross-reference to `documentation-accuracy` since missing error types cause inconsistent error handling

### Severity Guidelines

- **CRITICAL**: Cross-references that affect architectural integrity
- **HIGH**: Cross-references that impact maintainability
- **MEDIUM**: Cross-references that improve understanding
- **LOW**: Optional cross-references for additional context

---


Check for common security vulnerabilities, secret handling, and input validation.

### What to Look For

- Hardcoded secrets or API keys
- SQL injection vulnerabilities
- Command injection in shell/command execution
- Path traversal vulnerabilities
- Missing input validation
- Insecure random number generation
- Unencrypted sensitive data storage

### Examples

- **Hardcoded secrets**: AWS keys, database passwords in source code
- **SQL injection**: String concatenation in SQL queries instead of parameterized queries
- **Command injection**: User input directly passed to shell/exec without sanitization
- **Path traversal**: `filepath.Join()` with unvalidated user paths allowing directory traversal

### Severity Guidelines

- **CRITICAL**: Hardcoded secrets, injection vulnerabilities, credential leaks
- **HIGH**: Missing input validation on user-controlled data
- **MEDIUM**: Weak cryptography or insecure defaults
- **LOW**: Security best practice violations (no immediate vulnerability)

---

## Performance

Check for common performance anti-patterns, inefficient algorithms, and resource leaks.

### What to Look For

- N+1 queries in loops (N+1 problem)
- Inefficient string operations in loops
- Unnecessary memory allocations
- Missing database indexes (inferred)
- Resource leaks (unclosed files, connections)
- Synchronous operations that could be async
- Inefficient data structures for common operations

### Examples

- **N+1 query**: Querying database inside loop instead of single query with IN clause
- **String concatenation**: Using `+` in loop instead of `strings.Builder`
- **Memory leak**: File opened but never closed
- **Inefficient search**: O(n²) algorithm when O(n) solution exists

### Severity Guidelines

- **CRITICAL**: Resource leaks causing gradual degradation
- **HIGH**: O(n²) algorithm where O(n) exists and affects common operations
- **MEDIUM**: Inefficient operations in hot paths
- **LOW**: Suboptimal data structure choice or minor performance issues

---

## Dead Code

Check for unused code, unreachable code, and commented-out production code.

### What to Look For

- Unused functions or methods
- Unreachable code after return statements
- Commented-out production code (large blocks)
- Unused imports
- Unused variables or constants
- Unreferenced files or entire modules

### Examples

- **Unused interfaces**: `GitRepositoryInterface` and `ProjectRepository` defined but never used
- **Unreachable code**: Code after `return` or infinite loop prevention
- **Commented code**: Large blocks of functional code commented out with `//` or `/* */`
- **Dead functions**: Helper functions defined but never called

### Severity Guidelines

- **CRITICAL**: Dead code in critical paths affecting functionality
- **HIGH**: Entire unused files or modules (>50 lines)
- **MEDIUM**: Unused functions or unreachable code blocks (10-50 lines)
- **LOW**: Unused variables, small helper functions, or commented debug code
