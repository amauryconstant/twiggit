# Documentation Design for twiggit

## Core Architecture

This project uses a structured documentation system designed for AI agent consumption. The architecture separates concerns across specialized files with clear responsibilities.

### File Responsibilities

- **AGENTS.md**: Entry point with essential knowledge and keyword reference
- **design.md**: User requirements and system specifications
- **technology.md**: Technology choices and integration patterns
- **implementation.md**: Implementation standards and constraints
- **testing.md**: Testing philosophy and quality standards
- **code-style-guide.md**: Concrete Go code examples and patterns

## Keyword System

### Keyword Philosophy

Formal specification keywords eliminate ambiguity in documentation by providing precise, machine-interpretable requirements. Each keyword type serves a distinct purpose in the documentation architecture, enabling AI agents to understand implementation obligations, recommendations, and commitments.

### Detailed Keyword Definitions

#### SHALL - Mandatory Requirements
**Purpose**: Defines non-negotiable requirements that MUST be implemented exactly as specified
**Characteristics**: 
- Creates testable, verifiable requirements
- Non-compliance results in implementation rejection
- Used for critical functionality, security constraints, and core behavior
**Examples in Context**: "Configuration SHALL be loaded using the configured provider", "Error messages SHALL be written to stderr"

#### SHALL NOT - Absolute Prohibitions
**Purpose**: Defines behaviors that MUST NOT occur under any circumstances
**Characteristics**:
- Creates security boundaries and architectural constraints
- Violation results in immediate rejection
- Used for security limits, performance boundaries, and anti-patterns
**Examples in Context**: "Shell commands SHALL NOT be used for git operations", "Sensitive information SHALL NOT be logged"

#### SHOULD - Recommended Practices
**Purpose**: Provides guidance on best practices that improve quality and maintainability
**Characteristics**:
- Not mandatory but strongly recommended
- Deviation requires justification
- Used for optimization, user experience, and code quality
**Examples in Context**: "Error messages SHOULD include suggested actions", "Test coverage SHOULD exceed 80%"

#### SHOULD NOT - Discouraged Practices
**Purpose**: Warns against practices that should be avoided unless there's compelling reason
**Characteristics**:
- Not prohibited but strongly discouraged
- Usage requires documented justification
- Used for anti-patterns, performance pitfalls, and maintenance risks
**Examples in Context**: "Dependencies SHOULD NOT be added without approval", "Global variables SHOULD NOT be used"

#### WILL - System Commitments
**Purpose**: Declares factual statements about how the system will behave
**Characteristics**:
- Not a requirement to test, but a commitment about behavior
- Describes system properties and guarantees
- Used for behavioral declarations and design commitments
**Examples in Context**: "The system WILL validate all inputs before processing", "Configuration WILL be applied in priority order"

#### WILL NOT - Absence Guarantees
**Purpose**: Declares what the system will not do
**Characteristics**:
- Not a requirement but a commitment about absence
- Provides guarantees about system limitations
**Used for**: Security guarantees, privacy commitments, and scope boundaries
**Examples in Context**: "The system WILL NOT store sensitive data in logs", "The CLI WILL NOT require internet connectivity"

#### MAY - Optional Features
**Purpose**: Indicates features that are completely at implementation discretion
**Characteristics**:
- Completely optional, not required
- Implementation has complete flexibility
**Used for**: Future enhancements, extensibility points, and optional functionality
**Examples in Context**: "Additional validation rules MAY be added in future", "Custom error handlers MAY be provided"

#### MAY NOT - Optional Restrictions
**Purpose**: Indicates restrictions that implementation may choose to enforce
**Characteristics**:
- Completely optional restriction
- Implementation may choose to enforce or ignore
**Used for**: Configurable constraints and discretionary prohibitions
**Examples in Context**: "The system MAY NOT allow concurrent access", "Users MAY NOT be able to modify system configuration"

### Keyword Usage Guidelines

#### When to Use Each Keyword Type

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

#### Decision Framework: Choosing Between Similar Keywords

**SHALL vs WILL**: 
- Use **SHALL** when it's a requirement that must be implemented and tested
- Use **WILL** when it's a statement about how the system behaves (not a testable requirement)
- *Example*: "Configuration SHALL be loaded from file" (requirement) vs "System WILL apply configuration in priority order" (behavior commitment)

**SHOULD NOT vs MAY NOT**:
- Use **SHOULD NOT** when it's discouraged but allowed with justification (anti-pattern)
- Use **MAY NOT** when it's completely optional whether to enforce the restriction
- *Example*: "Dependencies SHOULD NOT be added without approval" (discouraged) vs "Verbosity controls MAY NOT be provided" (optional restriction)

**SHOULD vs MAY**:
- Use **SHOULD** when it's a best practice that should be followed unless there's good reason not to
- Use **MAY** when it's truly optional and implementation has complete discretion
- *Example*: "Error messages SHOULD include suggested actions" (best practice) vs "Alternative detection methods MAY be supported" (optional feature)

#### Common Usage Patterns

1. **Core Functionality**: SHALL for essential features, SHALL NOT for security boundaries
2. **Quality Standards**: SHOULD for best practices, SHOULD NOT for anti-patterns  
3. **System Behavior**: WILL for behavioral guarantees, WILL NOT for absence guarantees
4. **Future Flexibility**: MAY for optional features, MAY NOT for configurable restrictions

#### Keyword Optimization Checklist

When reviewing documentation for keyword usage:
- [ ] All critical requirements use SHALL/SHALL NOT
- [ ] All best practices use SHOULD/SHOULD NOT  
- [ ] All system behaviors use WILL/WILL NOT
- [ ] All optional features use MAY/MAY NOT
- [ ] No SHALL requirements are untestable
- [ ] No WILL statements are actually requirements
- [ ] SHOULD NOT guidelines have clear justification paths
- [ ] MAY NOT restrictions are truly optional

#### Keyword Consistency Rules

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

## Keyword Optimization Workflows

### Systematic Keyword Review Process

When performing comprehensive keyword optimization:

#### Phase 1: Best Practices Optimization
- [ ] Identify all current best practices and recommendations
- [ ] Convert appropriate statements to use SHOULD keywords
- [ ] Ensure SHOULD NOT is used for anti-patterns and discouraged practices
- [ ] Verify that all SHOULD recommendations have clear justification paths
- [ ] Confirm that SHOULD NOT guidelines allow for documented exceptions

#### Phase 2: System Behavior Optimization  
- [ ] Identify statements describing system behavior and guarantees
- [ ] Convert behavioral commitments to use WILL keywords
- [ ] Convert absence guarantees to use WILL NOT keywords
- [ ] Ensure WILL statements describe actual behavior, not requirements
- [ ] Confirm that WILL NOT statements provide clear absence guarantees

#### Phase 3: Optional Features Optimization
- [ ] Identify optional features and future enhancements
- [ ] Convert truly optional features to use MAY keywords
- [ ] Convert optional restrictions to use MAY NOT keywords
- [ ] Ensure MAY features are completely discretionary
- [ ] Confirm that MAY NOT restrictions are configurable/enforceable at implementation discretion

#### Post-Optimization Verification
- [ ] Review all keyword changes for consistency with AGENTS.md definitions
- [ ] Verify that no requirements were weakened inappropriately
- [ ] Confirm that all SHALL requirements remain testable and critical
- [ ] Check that keyword usage provides appropriate implementation flexibility
- [ ] Validate that the documentation maintains clear separation of concerns

### Common Keyword Anti-Patterns to Avoid

**Overusing SHALL**: 
- Problem: Using SHALL for non-critical requirements that should be SHOULD
- Solution: Reserve SHALL for security, core functionality, and testable requirements

**Confusing WILL with SHALL**:
- Problem: Using WILL for requirements that should be tested
- Solution: Use SHALL for testable requirements, WILL for behavioral commitments

**Underusing MAY/MAY NOT**:
- Problem: Using SHOULD/SHOULD NOT for truly optional features
- Solution: Use MAY/MAY NOT for completely optional features and restrictions

**Inconsistent Keyword Usage**:
- Problem: Similar concepts using different keywords across files
- Solution: Apply keyword optimization checklist consistently across all documentation

## Agent Checklists

### When Updating Documentation

#### Before Making Changes
- [ ] Identify which file(s) need updating
- [ ] Verify the change aligns with file responsibilities
- [ ] Plan keyword usage for new requirements
- [ ] Review keyword decision framework for guidance

#### During Updates
- [ ] Use keywords consistently with AGENTS.md definitions
- [ ] Maintain separation of concerns between files
- [ ] Ensure requirements are clear and unambiguous
- [ ] Keep language concise and actionable
- [ ] Apply keyword optimization checklist for new content

#### After Updates
- [ ] Verify keyword usage consistency across files
- [ ] Check for unintended cross-references
- [ ] Ensure all requirements remain testable
- [ ] Confirm file boundaries are still respected
- [ ] Run keyword optimization checklist on updated content
- [ ] Validate that SHALL requirements are implementable and testable
- [ ] Confirm that WILL statements describe actual system behavior

### When Creating New Documentation

#### Planning Phase
- [ ] Determine if new content fits existing file responsibilities
- [ ] Decide if new file is needed or if content belongs in existing file
- [ ] Define clear purpose and scope for new documentation
- [ ] Plan keyword usage strategy using decision framework
- [ ] Identify which requirements are critical vs best practices vs optional

#### Implementation Phase
- [ ] Create file with clear, focused responsibility
- [ ] Use consistent keyword system from AGENTS.md
- [ ] Ensure content is actionable for AI agents
- [ ] Apply keyword optimization patterns during writing
- [ ] Use decision framework to choose between similar keywords

#### Validation Phase
- [ ] Verify no duplication with existing documentation
- [ ] Test that content can be implemented by AI agents
- [ ] Check that file boundaries are maintained
- [ ] Confirm keyword usage is consistent
- [ ] Run complete keyword optimization checklist
- [ ] Validate that keyword choices align with documentation purpose