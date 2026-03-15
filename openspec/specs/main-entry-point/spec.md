# Main Entry Point

## Purpose

The main entry point handles application initialization, configuration loading, and graceful error handling for the entire application lifecycle.

## Requirements

### Requirement: Config load failure handling
The main entry point SHALL handle configuration load failures gracefully.

#### Scenario: Invalid config file
- **WHEN** configuration file contains invalid YAML
- **THEN** system exits with non-zero status code
- **AND** error message describes the configuration problem

#### Scenario: Missing config directory
- **WHEN** config directory does not exist and cannot be created
- **THEN** system exits with non-zero status code
- **AND** error message indicates configuration issue

### Requirement: Successful execution path
The main entry point SHALL initialize all services and execute commands successfully.

#### Scenario: Normal execution with valid config
- **WHEN** valid configuration exists
- **AND** command is valid (e.g., twiggit list)
- **THEN** system initializes all services
- **AND** command executes successfully
- **AND** system exits with status code 0

#### Scenario: Help command execution
- **WHEN** user runs help command
- **THEN** system displays help text
- **AND** system exits with status code 0

### Requirement: Service initialization failure handling
The main entry point SHALL handle service initialization failures gracefully.

#### Scenario: Git client initialization failure
- **WHEN** git client cannot be initialized
- **THEN** system exits with non-zero status code
- **AND** error message describes the initialization problem

#### Scenario: Command execution failure
- **WHEN** command execution fails (e.g., invalid worktree name)
- **THEN** system exits with appropriate non-zero status code
- **AND** error message is formatted for user consumption
