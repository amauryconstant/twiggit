---
description: Comprehensive code quality audit for consistency, coherency, and de-duplication
---

Follow the 'audit-codebase-quality' skill workflow to perform a comprehensive codebase audit.

## Purpose

Audits codebase for architectural issues, duplicate code patterns, test gaps, documentation accuracy, and quality standards. Uses parallel processing to analyze multiple quality areas simultaneously, then consolidates findings into a prioritized action plan.

## Audit Areas

- **Package structure**: Naming conventions, directory organization, layer separation, architectural violations
- **Duplicate code**: Repeated logic, similar functions, consolidation opportunities, copied code
- **Interface compliance**: Implementation completeness, signature mismatches, unused interfaces, method gaps
- **Test patterns**: Test organization, coverage gaps, mock usage consistency, test naming conventions
- **Documentation accuracy**: AGENTS.md vs actual code, missing types/methods, incorrect signatures
- **Import consistency**: Import ordering, circular dependencies, unused imports, external dependencies
- **Error handling**: Wrapping patterns, error type usage, message consistency, error chain breaks
- **Mock centralization**: Inline vs centralized mocks, duplicate mock implementations
- **Security**: Common security vulnerabilities, secret handling, input validation
- **Performance**: Performance anti-patterns, inefficient algorithms, resource leaks
- **Dead code**: Unused code, unreachable code, commented-out production code

## Severity Levels

- **CRITICAL**: Architectural violations, security issues, dead code in production
- **HIGH**: Large duplications (>50 lines), naming conflicts, unused interfaces
- **MEDIUM**: Test gaps, error handling inconsistencies, consolidation needs
- **LOW**: Cosmetic issues, style improvements, minor inconsistencies

## Output

Generates comprehensive markdown report with:
- Executive summary with metrics
- Findings grouped by severity
- Prioritized recommendations
- Complete findings appendix with file locations

@.opencode/skills/audit-codebase-quality/SKILL.md
