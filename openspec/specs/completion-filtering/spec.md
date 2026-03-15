## Purpose

Pattern-based exclusion of branches and projects from completion suggestions, plus fuzzy matching for partial input that doesn't match exact prefixes.

## Requirements

### Requirement: Fuzzy Matching for Completion

The system SHALL support fuzzy matching for completion suggestions when exact prefix matching yields no results or when fuzzy matching is explicitly enabled via configuration.

#### Scenario: Fuzzy match branches with subsequence

- **WHEN** user types `twiggit cd f1<tab>` from a project with branch `feature-1`
- **AND** no branch starts with exact prefix "f1"
- **THEN** system SHALL suggest `feature-1` as a fuzzy match
- **AND** match SHALL be case-insensitive

#### Scenario: Fuzzy match projects with subsequence

- **WHEN** user types `twiggit cd twi<tab>` from outside git context
- **AND** a project named `twiggit` exists
- **THEN** system SHALL suggest `twiggit` as a fuzzy match

#### Scenario: Fuzzy match disabled by config

- **WHEN** `navigation.fuzzy_matching` config is `false`
- **THEN** system SHALL NOT perform fuzzy matching
- **AND** system SHALL only return exact prefix matches

#### Scenario: Exact prefix match takes priority

- **WHEN** user types `twiggit cd feat<tab>` with branches `feature-1` and `feat-api`
- **THEN** system SHALL return `feat-api` as exact prefix match first
- **AND** system SHALL include `feature-1` only if fuzzy matching is enabled

### Requirement: Branch Exclusion Patterns

The system SHALL filter completion suggestions based on configured glob patterns to exclude noisy or irrelevant branches.

#### Scenario: Exclude branches matching glob pattern

- **WHEN** config includes `completion.exclude_branches = ["dependabot/*"]`
- **AND** repository has branches `dependabot/npm-1.2.3`, `feature-1`, `main`
- **THEN** system SHALL NOT suggest `dependabot/npm-1.2.3`
- **AND** system SHALL suggest `feature-1` and `main`

#### Scenario: Multiple exclusion patterns

- **WHEN** config includes `completion.exclude_branches = ["dependabot/*", "renovate/*", "gh-pages"]`
- **THEN** system SHALL exclude branches matching any pattern

#### Scenario: Exclusion applies to all contexts

- **WHEN** exclusion patterns are configured
- **THEN** patterns SHALL apply to suggestions from project context
- **AND** patterns SHALL apply to suggestions from worktree context
- **AND** patterns SHALL apply to cross-project completion

#### Scenario: Empty exclusion list allows all

- **WHEN** `completion.exclude_branches` is empty or not configured
- **THEN** system SHALL suggest all branches without filtering

### Requirement: Project Exclusion Patterns

The system SHALL filter project suggestions based on configured glob patterns.

#### Scenario: Exclude projects matching glob pattern

- **WHEN** config includes `completion.exclude_projects = ["archive/*"]`
- **AND** projects directory contains `archive/old-project`, `active-project`
- **THEN** system SHALL NOT suggest `archive/old-project`
- **AND** system SHALL suggest `active-project`

#### Scenario: Project exclusion from outside git context

- **WHEN** user requests completion from outside git context
- **AND** project exclusion patterns are configured
- **THEN** system SHALL filter project suggestions by exclusion patterns

#### Scenario: Project exclusion from project context

- **WHEN** user requests completion from project context
- **AND** project exclusion patterns are configured
- **THEN** system SHALL filter cross-project suggestions by exclusion patterns
