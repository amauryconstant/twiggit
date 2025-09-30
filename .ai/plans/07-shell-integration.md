# Shell Integration System Implementation Plan

## Overview

This plan details the implementation of the shell integration system for twiggit, focusing on the `setup-shell` command and supporting infrastructure. The system provides shell wrapper functions that intercept `twiggit cd` calls and enable seamless directory navigation across worktrees and projects.

## Design Requirements

From [design.md](../design.md):

> **Shell Wrapper**:
> - SHALL intercept `twiggit cd` calls and change shell directory
> - SHALL provide escape hatch with `builtin cd` for shell built-in
> - SHALL warn when overriding shell built-in `cd`
> - SHALL pass through all other commands unchanged
> - SHALL be automatically installed via `twiggit setup-shell` command

> **setup-shell Command**:
> - Current shell (bash, zsh, fish) SHALL be detected
> - Appropriate wrapper function with `builtin cd` escape hatch SHALL be generated
> - Wrapper SHALL be added to correct shell configuration file
> - Warning SHALL be provided about overriding shell built-in `cd` command
> - User SHALL be instructed to restart shell or source configuration

From [technology.md](../technology.md):

> **Shell Integration Strategy**:
> Carapace SHALL generate completion scripts. Directory navigation SHALL use shell wrapper functions that intercept command output.

> **Shell Integration Constraints**:
> - Carapace SHALL be used for shell completion generation
> - Shell wrapper functions SHALL be generated for directory navigation
> - Shell detection SHALL be performed automatically during setup

## Architecture

### Core Components

1. **ShellDetector Interface** - Detects current shell environment
2. **ShellIntegration Service** - Generates and installs shell wrappers
3. **Shell-Specific Syntax Handlers** - Handles alias generation for each shell
4. **Configuration File Manager** - Detects and modifies shell config files
5. **Setup-Shell Command** - CLI command implementation

### File Structure

```
internal/
├── domain/
│   ├── shell.go              # Shell types and interfaces
│   └── shell_test.go         # Shell domain tests
├── infrastructure/
│   ├── shell/
│   │   ├── detector.go       # Shell detection implementation
│   │   ├── detector_test.go  # Detection tests
│   │   ├── integration.go    # Shell integration service
│   │   ├── integration_test.go # Integration tests
│   │   ├── bash.go          # Bash-specific syntax
│   │   ├── zsh.go           # Zsh-specific syntax
│   │   ├── fish.go          # Fish-specific syntax
│   │   └── interfaces.go     # Shell infrastructure interfaces
│   └── config/
│       └── files.go         # Config file detection logic
cmd/
└── setup-shell.go           # Setup-shell command implementation
```

## Interface Definitions

### ShellDetector Interface

```go
// ShellDetector detects the current shell environment
type ShellDetector interface {
    // DetectCurrentShell identifies the shell from environment
    DetectCurrentShell() (Shell, error)
    
    // IsSupported checks if a shell type is supported
    IsSupported(shellType ShellType) bool
}

// Shell represents a shell environment
type Shell interface {
    Type() ShellType
    Path() string
    Version() string
    ConfigFiles() []string
}

// ShellType represents supported shell types
type ShellType string

const (
    ShellBash ShellType = "bash"
    ShellZsh  ShellType = "zsh"
    ShellFish ShellType = "fish"
)
```

### ShellIntegration Interface

```go
// ShellIntegration manages shell wrapper installation
type ShellIntegration interface {
    // GenerateWrapper creates shell-specific wrapper function
    GenerateWrapper(shell Shell) (string, error)
    
    // InstallWrapper adds wrapper to shell configuration
    InstallWrapper(shell Shell, wrapper string) error
    
    // DetectConfigFile finds appropriate config file for shell
    DetectConfigFile(shell Shell) (string, error)
    
    // BackupConfig creates backup before modification
    BackupConfig(configPath string) error
    
    // ValidateInstallation checks if wrapper is properly installed
    ValidateInstallation(shell Shell) error
}
```

## Shell-Specific Implementation

### Bash Wrapper Function

```bash
# Twiggit bash wrapper function
twiggit() {
    if [[ "$1" == "cd" ]]; then
        # Handle cd command with directory change
        local target_dir
        target_dir=$(command twiggit "${@:2}")
        if [[ $? -eq 0 && -n "$target_dir" ]]; then
            builtin cd "$target_dir"
        else
            return $?
        fi
    elif [[ "$1" == "cd" && "$2" == "--help" ]]; then
        # Show help for cd command
        command twiggit "$@"
    else
        # Pass through all other commands
        command twiggit "$@"
    fi
}

# Warning about overriding built-in cd
echo "twiggit: bash wrapper installed - use 'builtin cd' for shell built-in"
```

### Zsh Wrapper Function

```zsh
# Twiggit zsh wrapper function
twiggit() {
    if [[ "$1" == "cd" ]]; then
        # Handle cd command with directory change
        local target_dir
        target_dir=$(command twiggit "${@:2}")
        if [[ $? -eq 0 && -n "$target_dir" ]]; then
            builtin cd "$target_dir"
        else
            return $?
        fi
    elif [[ "$1" == "cd" && "$2" == "--help" ]]; then
        # Show help for cd command
        command twiggit "$@"
    else
        # Pass through all other commands
        command twiggit "$@"
    fi
}

# Warning about overriding built-in cd
echo "twiggit: zsh wrapper installed - use 'builtin cd' for shell built-in"
```

### Fish Wrapper Function

```fish
# Twiggit fish wrapper function
function twiggit
    if test (count $argv) -gt 0 -a "$argv[1]" = "cd"
        # Handle cd command with directory change
        set target_dir (command twiggit $argv[2..])
        if test $status -eq 0 -a -n "$target_dir"
            builtin cd "$target_dir"
        else
            return $status
        end
    else if test (count $argv) -gt 1 -a "$argv[1]" = "cd" -a "$argv[2]" = "--help"
        # Show help for cd command
        command twiggit $argv
    else
        # Pass through all other commands
        command twiggit $argv
    end
end

# Warning about overriding built-in cd
echo "twiggit: fish wrapper installed - use 'builtin cd' for shell built-in"
```

## Configuration File Detection

### Detection Logic

```go
// DetectConfigFile finds the appropriate configuration file for a shell
func (s *ShellIntegrationService) DetectConfigFile(shell Shell) (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("failed to get home directory: %w", err)
    }

    // Check shell-specific config files in order of preference
    var configFiles []string
    switch shell.Type() {
    case ShellBash:
        configFiles = []string{
            filepath.Join(homeDir, ".bashrc"),
            filepath.Join(homeDir, ".bash_profile"),
            filepath.Join(homeDir, ".profile"),
        }
    case ShellZsh:
        configFiles = []string{
            filepath.Join(homeDir, ".zshrc"),
            filepath.Join(homeDir, ".zprofile"),
            filepath.Join(homeDir, ".profile"),
        }
    case ShellFish:
        configFiles = []string{
            filepath.Join(homeDir, ".config", "fish", "config.fish"),
            filepath.Join(homeDir, ".fishrc"),
        }
    }

    // Return the first existing config file
    for _, configFile := range configFiles {
        if _, err := os.Stat(configFile); err == nil {
            return configFile, nil
        }
    }

    // If no config file exists, return the preferred one for creation
    return configFiles[0], nil
}
```

### Configuration File Modification

```go
// InstallWrapper adds the wrapper function to the shell configuration
func (s *ShellIntegrationService) InstallWrapper(shell Shell, wrapper string) error {
    configPath, err := s.DetectConfigFile(shell)
    if err != nil {
        return fmt.Errorf("failed to detect config file: %w", err)
    }

    // Create backup before modification
    if err := s.BackupConfig(configPath); err != nil {
        return fmt.Errorf("failed to backup config: %w", err)
    }

    // Read existing content
    content, err := os.ReadFile(configPath)
    if err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    // Check if wrapper is already installed
    if strings.Contains(string(content), "# Twiggit shell wrapper") {
        return fmt.Errorf("twiggit wrapper already installed in %s", configPath)
    }

    // Prepare wrapper content with markers
    wrapperContent := fmt.Sprintf(`
# Twiggit shell wrapper - installed on %s
%s
# End twiggit wrapper
`, time.Now().Format("2006-01-02 15:04:05"), wrapper)

    // Write updated content
    newContent := string(content) + wrapperContent
    if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }

    return nil
}
```

## Setup-Shell Command Implementation

### Command Structure

```go
// cmd/setup-shell.go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/infrastructure/shell"
)

func NewSetupShellCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "setup-shell",
        Short: "Install shell wrapper for directory navigation",
        Long: `Install shell wrapper functions that intercept 'twiggit cd' calls
and enable seamless directory navigation between worktrees and projects.

The wrapper provides:
- Automatic directory change on 'twiggit cd'
- Escape hatch with 'builtin cd' for shell built-in
- Pass-through for all other commands

Supported shells: bash, zsh, fish`,
        RunE: runSetupShell,
    }

    cmd.Flags().Bool("force", false, "force reinstall even if already installed")
    cmd.Flags().Bool("dry-run", false, "show what would be done without making changes")

    return cmd
}

func runSetupShell(cmd *cobra.Command, args []string) error {
    force, _ := cmd.Flags().GetBool("force")
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // Initialize shell detector
    detector := shell.NewShellDetector()
    currentShell, err := detector.DetectCurrentShell()
    if err != nil {
        return fmt.Errorf("failed to detect current shell: %w", err)
    }

    // Initialize shell integration service
    integration := shell.NewShellIntegrationService(detector)

    // Check if already installed
    if !force {
        if err := integration.ValidateInstallation(currentShell); err == nil {
            fmt.Printf("twiggit wrapper already installed for %s\n", currentShell.Type())
            fmt.Printf("Use --force to reinstall\n")
            return nil
        }
    }

    // Generate wrapper
    wrapper, err := integration.GenerateWrapper(currentShell)
    if err != nil {
        return fmt.Errorf("failed to generate wrapper: %w", err)
    }

    if dryRun {
        fmt.Printf("Would install wrapper for %s:\n", currentShell.Type())
        fmt.Printf("Wrapper function:\n%s\n", wrapper)
        return nil
    }

    // Install wrapper
    if err := integration.InstallWrapper(currentShell, wrapper); err != nil {
        return fmt.Errorf("failed to install wrapper: %w", err)
    }

    // Show success message
    configPath, _ := integration.DetectConfigFile(currentShell)
    fmt.Printf("✓ twiggit wrapper installed for %s\n", currentShell.Type())
    fmt.Printf("✓ Added to: %s\n", configPath)
    fmt.Printf("\nTo activate the wrapper:\n")
    fmt.Printf("  1. Restart your shell, or\n")
    fmt.Printf("  2. Run: source %s\n", configPath)
    fmt.Printf("\nUsage:\n")
    fmt.Printf("  twiggit cd <branch>     # Change to worktree\n")
    fmt.Printf("  builtin cd <path>       # Use shell built-in cd\n")

    return nil
}
```

## Testing Strategy

### Unit Tests

#### Shell Detection Tests

```go
func TestShellDetector_DetectCurrentShell(t *testing.T) {
    testCases := []struct {
        name           string
        envSHELL       string
        expectedType   shell.ShellType
        expectedError  bool
    }{
        {
            name:         "detect bash",
            envSHELL:     "/bin/bash",
            expectedType: shell.ShellBash,
            expectedError: false,
        },
        {
            name:         "detect zsh",
            envSHELL:     "/usr/bin/zsh",
            expectedType: shell.ShellZsh,
            expectedError: false,
        },
        {
            name:         "detect fish",
            envSHELL:     "/usr/local/bin/fish",
            expectedType: shell.ShellFish,
            expectedError: false,
        },
        {
            name:         "unsupported shell",
            envSHELL:     "/bin/sh",
            expectedError: true,
        },
        {
            name:         "empty shell",
            envSHELL:     "",
            expectedError: true,
        },
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            // Set environment variable
            oldShell := os.Getenv("SHELL")
            defer os.Setenv("SHELL", oldShell)
            
            if tt.envSHELL != "" {
                os.Setenv("SHELL", tt.envSHELL)
            } else {
                os.Unsetenv("SHELL")
            }

            detector := shell.NewShellDetector()
            detectedShell, err := detector.DetectCurrentShell()

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedType, detectedShell.Type())
            }
        })
    }
}
```

#### Wrapper Generation Tests

```go
func TestShellIntegration_GenerateWrapper(t *testing.T) {
    testCases := []struct {
        name        string
        shellType   shell.ShellType
        expectError bool
        validate    func(t *testing.T, wrapper string)
    }{
        {
            name:      "bash wrapper",
            shellType: shell.ShellBash,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "twiggit() {")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
            },
        },
        {
            name:      "zsh wrapper",
            shellType: shell.ShellZsh,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "twiggit() {")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
            },
        },
        {
            name:      "fish wrapper",
            shellType: shell.ShellFish,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "function twiggit")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
            },
        },
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            detector := shell.NewShellDetector()
            integration := shell.NewShellIntegrationService(detector)
            
            mockShell := &mockShell{shellType: tt.shellType}
            wrapper, err := integration.GenerateWrapper(mockShell)

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                tt.validate(t, wrapper)
            }
        })
    }
}
```

### Integration Tests

#### Shell Installation Tests

```go
//go:build integration
// +build integration

func TestShellIntegration_InstallWrapper(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    testCases := []struct {
        name      string
        shellType shell.ShellType
        configDir string
    }{
        {
            name:      "bash installation",
            shellType: shell.ShellBash,
            configDir: "bash",
        },
        {
            name:      "zsh installation",
            shellType: shell.ShellZsh,
            configDir: "zsh",
        },
        {
            name:      "fish installation",
            shellType: shell.ShellFish,
            configDir: "fish",
        },
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            // Create temporary config directory
            tempDir := t.TempDir()
            configPath := filepath.Join(tempDir, tt.configDir+"rc")
            
            // Create initial config file
            if err := os.WriteFile(configPath, []byte("# Initial config"), 0644); err != nil {
                t.Fatalf("Failed to create config file: %v", err)
            }

            detector := shell.NewShellDetector()
            integration := shell.NewShellIntegrationService(detector)
            
            mockShell := &mockShell{
                shellType:   tt.shellType,
                configFiles: []string{configPath},
            }

            // Generate wrapper
            wrapper, err := integration.GenerateWrapper(mockShell)
            require.NoError(t, err)

            // Install wrapper
            err = integration.InstallWrapper(mockShell, wrapper)
            require.NoError(t, err)

            // Verify installation
            content, err := os.ReadFile(configPath)
            require.NoError(t, err)
            
            assert.Contains(t, string(content), "# Twiggit shell wrapper")
            assert.Contains(t, string(content), "twiggit()")

            // Test duplicate installation
            err = integration.InstallWrapper(mockShell, wrapper)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), "already installed")
        })
    }
}
```

### E2E Tests

#### Setup-Shell Command Tests

```go
//go:build e2e
// +build e2e

func TestSetupShellCommand(t *testing.T) {
    // Build CLI binary for testing
    binPath := buildTestBinary(t)
    defer os.Remove(binPath)

    testCases := []struct {
        name        string
        args        []string
        env         map[string]string
        expectError bool
        validate    func(t *testing.T, output string)
    }{
        {
            name: "setup bash shell",
            args: []string{"setup-shell", "--dry-run"},
            env:  map[string]string{"SHELL": "/bin/bash"},
            validate: func(t *testing.T, output string) {
                assert.Contains(t, output, "Would install wrapper for bash")
                assert.Contains(t, output, "twiggit() {")
            },
        },
        {
            name: "setup zsh shell",
            args: []string{"setup-shell", "--dry-run"},
            env:  map[string]string{"SHELL": "/usr/bin/zsh"},
            validate: func(t *testing.T, output string) {
                assert.Contains(t, output, "Would install wrapper for zsh")
                assert.Contains(t, output, "twiggit() {")
            },
        },
        {
            name:        "unsupported shell",
            args:        []string{"setup-shell"},
            env:         map[string]string{"SHELL": "/bin/sh"},
            expectError: true,
        },
    }

    for _, tt := range testCases {
        t.Run(tt.name, func(t *testing.T) {
            // Set up environment
            env := []string{}
            for k, v := range tt.env {
                env = append(env, fmt.Sprintf("%s=%s", k, v))
            }

            // Run command
            session := gexec.Start(gexec.Command(binPath, tt.args...), gexec.NewBuffer(), gexec.NewBuffer())
            session.Wait()

            if tt.expectError {
                assert.NotEqual(t, 0, session.ExitCode())
            } else {
                assert.Equal(t, 0, session.ExitCode())
                if tt.validate != nil {
                    tt.validate(t, string(session.Out.Contents()))
                }
            }
        })
    }
}
```

## Implementation Steps

### Phase 1: Core Infrastructure (RED)

1. **Create interfaces and domain types**
   - Define `ShellDetector` interface
   - Define `ShellIntegration` interface
   - Create `Shell` and `ShellType` domain types

2. **Implement shell detection**
   - Create `ShellDetector` implementation
   - Add environment variable parsing
   - Add shell path validation

3. **Write failing tests**
   - Unit tests for shell detection
   - Unit tests for wrapper generation
   - Integration tests for config file detection

### Phase 2: Shell-Specific Implementation (GREEN)

1. **Implement shell-specific syntax**
   - Create bash wrapper generator
   - Create zsh wrapper generator
   - Create fish wrapper generator

2. **Implement configuration management**
   - Add config file detection logic
   - Add backup functionality
   - Add installation logic

3. **Make tests pass**
   - Implement minimum functionality to satisfy tests
   - Focus on core wrapper generation

### Phase 3: CLI Command Integration (GREEN)

1. **Implement setup-shell command**
   - Create command structure with Cobra
   - Add flag handling (force, dry-run)
   - Add user-friendly output

2. **Integrate with existing CLI**
   - Register command in root command
   - Add help text and examples

3. **Add E2E tests**
   - Test complete command workflow
   - Test error scenarios

### Phase 4: Refinement and Polish (REFACTOR)

1. **Improve error handling**
   - Add context-specific error messages
   - Add recovery suggestions

2. **Enhance user experience**
   - Improve output formatting
   - Add better installation instructions

3. **Performance optimization**
   - Optimize config file operations
   - Add caching where appropriate

## Integration with Existing System

### Context System Integration

The shell integration system works with the existing context detection system:

```go
// The cd command outputs paths that the shell wrapper consumes
func (cmd *CDCommand) RunE(c *cobra.Command, args []string) error {
    // Context detection and resolution
    context := cmd.contextDetector.DetectCurrentContext()
    targetPath, err := cmd.resolver.ResolveTarget(context, args)
    if err != nil {
        return err
    }
    
    // Output path for shell wrapper consumption
    fmt.Println(targetPath)
    return nil
}
```

### Configuration Integration

The shell integration respects the existing configuration system:

```go
// Shell detection can be overridden by configuration
func (s *ShellDetector) DetectCurrentShell() (Shell, error) {
    // Check configuration override first
    if configShell := s.config.GetString("shell.default"); configShell != "" {
        return s.createShellFromType(configShell)
    }
    
    // Fall back to environment detection
    return s.detectFromEnvironment()
}
```

## Quality Assurance

### Code Coverage Requirements

- **Shell detection**: 100% coverage for all supported shells
- **Wrapper generation**: 100% coverage for all shell types
- **Config file operations**: 95% coverage including error paths
- **CLI command**: 90% coverage via E2E tests

### Performance Requirements

- **Shell detection**: <10ms for environment detection
- **Wrapper generation**: <50ms for all shell types
- **Config file operations**: <100ms for typical config files
- **Setup command**: <500ms total execution time

### Security Requirements

- **Path validation**: All file paths validated before use
- **Permission checks**: Config file permissions verified
- **Backup safety**: Backups created before any modification
- **Input sanitization**: Shell wrapper content properly escaped

## Future Enhancements

### Optional Features

1. **Custom wrapper templates**: Allow users to customize wrapper functions
2. **Multiple shell support**: Add support for additional shells (powershell, etc.)
3. **Automatic updates**: Check for and update wrapper functions
4. **Integration with completion**: Combine wrapper installation with completion setup

### Extension Points

1. **Plugin system**: Allow third-party shell integrations
2. **Template system**: Customizable wrapper function templates
3. **Configuration options**: More granular control over wrapper behavior

## Summary

This implementation plan provides a comprehensive approach to shell integration that:

- Follows the established TDD methodology with RED/GREEN/REFACTOR phases
- Integrates seamlessly with existing context detection and configuration systems
- Provides robust testing coverage across unit, integration, and E2E levels
- Maintains the project's architectural patterns and quality standards
- Supports all required shells (bash, zsh, fish) with proper escape hatch functionality

The implementation prioritizes user experience while maintaining code quality and testability, ensuring reliable shell integration that enhances the twiggit workflow experience.