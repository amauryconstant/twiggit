# Documentation Design for twiggit

## Core Architecture

This project uses a structured documentation system designed for AI agent consumption. The architecture separates concerns across specialized files with clear responsibilities.

### File Responsibilities

- **AGENTS.md**: Entry point with essential knowledge and keyword reference
- **design.md**: User requirements and system specifications
- **technology.md**: Technology choices and integration patterns
- **implementation.md**: Implementation standards, testing, configuration, error handling
- **testing.md**: Testing philosophy and framework usage
- **code-style-guide.md**: Concrete Go patterns and examples
- **documentation-design.md**: Keyword reference and maintenance guidelines

## Keyword System

### Keyword Philosophy

Formal specification keywords eliminate ambiguity in documentation by providing precise, machine-interpretable requirements. Each keyword type serves a distinct purpose in the documentation architecture, enabling AI agents to understand implementation obligations, recommendations, and commitments.

### Keyword Definitions

#### SHALL - Mandatory Requirements
Non-negotiable requirements that MUST be implemented exactly as specified. Used for critical functionality, security constraints, and core behavior.

#### SHALL NOT - Absolute Prohibitions  
Behaviors that MUST NOT occur under any circumstances. Used for security boundaries and architectural constraints.

#### SHOULD - Recommended Practices
Best practices that improve quality and maintainability. Deviation requires justification.

#### SHOULD NOT - Discouraged Practices
Practices that should be avoided unless there's compelling reason. Usage requires documented justification.

#### WILL - System Commitments
Factual statements about how the system will behave. Describes system properties and guarantees.

#### WILL NOT - Absence Guarantees
Declares what the system will not do. Provides guarantees about system limitations.

#### MAY - Optional Features
Features that are completely at implementation discretion.

#### MAY NOT - Optional Restrictions
Restrictions that implementation may choose to enforce.

## Usage Guidelines

### When to Use Each Keyword Type

**SHALL/SHALL NOT**: Use when the requirement is absolutely critical to system functionality, security, or core behavior. These create the foundation of the system's contract and MUST be testable.
- **Use SHALL for**: Non-negotiable features, security requirements, core functionality, input validation, error handling
- **Use SHALL NOT for**: Security boundaries, architectural constraints, safety requirements, prohibited operations

**SHOULD/SHOULD NOT**: Use for guidance that improves quality, performance, or maintainability. These provide best practices while allowing flexibility when justified.
- **Use SHOULD for**: Best practices, optimization recommendations, user experience improvements, code quality standards
- **Use SHOULD NOT for**: Anti-patterns, performance pitfalls, maintenance risks, discouraged practices (requires justification)

**WILL/WILL NOT**: Use for declaring facts about system behavior or properties. These are commitments about how the system will operate, not requirements to be tested.
- **Use WILL for**: System behavior guarantees, design commitments, operational facts, processing guarantees
- **Use WILL NOT for**: Absence guarantees, scope boundaries, privacy commitments, system limitations

**MAY/MAY NOT**: Use for features or restrictions that are entirely at the implementation team's discretion. These provide maximum flexibility for future decisions.
- **Use MAY for**: Optional features, future enhancements, extensibility points, alternative approaches
- **Use MAY NOT for**: Configurable restrictions, discretionary prohibitions, optional limitations, implementation choices

### Decision Framework: Choosing Between Similar Keywords

**SHALL vs WILL**: 
- Use **SHALL** when it's a requirement that must be implemented and tested
- Use **WILL** when it's a statement about how the system behaves (not a testable requirement)

**SHOULD NOT vs MAY NOT**:
- Use **SHOULD NOT** when it's discouraged but allowed with justification (anti-pattern)
- Use **MAY NOT** when it's completely optional whether to enforce the restriction

**SHOULD vs MAY**:
- Use **SHOULD** when it's a best practice that should be followed unless there's good reason not to
- Use **MAY** when it's truly optional and implementation has complete discretion

## Consistency Rules

- Keywords SHALL be used consistently across all documentation files
- SHALL/SHALL NOT requirements MUST be testable and verifiable
- SHOULD/SHOULD NOT recommendations MUST include justification for deviations
- WILL/WILL NOT statements MUST accurately reflect system behavior
- MAY/MAY NOT options MUST be truly optional and not required

## Maintenance Guidelines

### Quality Assurance

- Keywords SHALL be used consistently across all files
- File boundaries SHALL be respected - no cross-contamination of concerns
- Requirements SHALL be testable where applicable
- Documentation SHALL be kept current with implementation

### Update Procedures

- Update file content according to its specific responsibility
- Verify keyword usage consistency
- Test that requirements remain implementable

### Keyword Optimization Checklist

When reviewing documentation for keyword usage:
- [ ] All critical requirements use SHALL/SHALL NOT
- [ ] All best practices use SHOULD/SHOULD NOT  
- [ ] All system behaviors use WILL/WILL NOT
- [ ] All optional features use MAY/MAY NOT
- [ ] No SHALL requirements are untestable
- [ ] No WILL statements are actually requirements
- [ ] SHOULD NOT guidelines have clear justification paths
- [ ] MAY NOT restrictions are truly optional

## Summary

This documentation system provides clear, consistent, and maintainable specifications for AI agents. By using formal specification keywords consistently and maintaining clear file boundaries, we ensure that documentation remains actionable, unambiguous, and aligned with implementation needs.