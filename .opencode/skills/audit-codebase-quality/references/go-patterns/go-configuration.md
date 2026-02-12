# Go Configuration & Dependency Patterns

Go-specific patterns for configuration and dependency management in codebase auditing.

## Configuration

### Patterns

- **Environment variables**: Use `os.Getenv()` with defaults
- **Config files**: Use well-known locations (`~/.config/app/config.toml`)
- **Flags**: Use flag package for CLI configuration
- **Priority order**: Defaults → config file → environment variables → flags

### Common Patterns

```go
// Load configuration with defaults
type Config struct {
    ProjectsDirectory string `toml:"projects_dir"`
    WorktreesDirectory string `toml:"worktrees_dir"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{
        ProjectsDirectory: "~/projects",  // Default
        WorktreesDirectory: "~/worktrees",
    }

    // Load from file
    if err := koanf.UnmarshalWithConf(toml.Parser(), &cfg); err != nil {
        return nil, err
    }

    // Override with environment
    if env := os.Getenv("PROJECTS_DIR"); env != "" {
        cfg.ProjectsDirectory = env
    }

    return cfg, nil
}
```

### Anti-Patterns

- **Hardcoded paths**: Configuration values embedded in code
- **Multiple config systems**: Mixing file-based, env vars, and flags without clear priority
- **Global config**: Mutable global configuration variable

## Dependency Management

### Go Module Practices

- **Semantic versioning**: Use semantic versions in go.mod (v1.2.3)
- **Minimal dependencies**: Prefer stdlib and well-maintained packages
- **Indirect dependencies**: Keep indirect dependencies minimal
- **Vendor directory**: Avoid vendoring unless necessary

### Common Patterns

```
module github.com/user/project

go 1.21

require (
    github.com/gorilla/mux v1.8.0
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/gorilla/mux v1.8.0 // Direct
    github.com/stretchr/testify v1.8.4 // Indirect
)
```

### Anti-Patterns

- **Unnecessary dependencies**: Adding packages for functionality in stdlib
- **Outdated dependencies**: Using old versions with known vulnerabilities
- **Pinning to master**: Using `master` branch instead of semantic version tags

## Audit-Specific Patterns

### For Configuration Audits

- Check for hardcoded configuration values in code
- Identify missing environment variable support
- Look for global configuration variables
- Verify clear priority order for config sources
- Check for mutable global state (should be thread-safe)

### For Dependency Audits

- Verify semantic versioning in go.mod
- Check for unnecessary dependencies (stdlib replacements)
- Identify outdated dependencies with known vulnerabilities
- Look for master branch pinning instead of version tags
- Verify indirect dependencies are minimal
- Check for unnecessary vendoring
