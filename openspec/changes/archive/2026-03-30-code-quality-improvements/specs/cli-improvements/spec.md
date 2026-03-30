# CLI Improvements

## Purpose

Improve CLI ergonomics with consistent flag conventions and better user feedback for destructive operations.

## ADDED Requirements

### Requirement: Create command SHALL use configured default source branch

The create command SHALL use config.Validation.DefaultSourceBranch when --source flag is not specified.

#### Scenario: Create without --source uses config default
- **WHEN** user runs twiggit create feature-x without --source flag
- **THEN** config.Validation.DefaultSourceBranch is used as source branch
- **AND** when no config set, "main" is used as default

### Requirement: Prune SHALL show preview before confirmation

The prune command with --all flag SHALL display list of worktrees to be pruned before prompting for confirmation.

#### Scenario: Prune all shows preview before confirmation
- **WHEN** user runs twiggit prune --all
- **THEN** list of worktrees to be pruned is displayed
- **AND** user is prompted for confirmation before deletion proceeds

### Requirement: Critical flags SHALL have short forms

The --merged-only flag SHALL have short form -m and --delete-branches SHALL have a short form.

#### Scenario: Delete with short merged-only flag
- **WHEN** user runs twiggit delete worktree -m
- **THEN** short flag -m is recognized as --merged-only
