package infrastructure

// This file was intentionally left minimal after migrating all interface
// definitions to internal/application/interfaces.go. The interfaces that
// were previously defined here (GitClient, GoGitClient, CLIClient, HookRunner,
// ShellInfrastructure) are now defined in the application package.
//
// Infrastructure implementations now implement interfaces from the application
// package and include compile-time checks in their respective files.
