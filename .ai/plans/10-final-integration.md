# Plan 10: Final Integration and Release Preparation

## 🎯 Purpose

Integrate all completed components into a cohesive, production-ready twiggit tool that meets all design requirements and performance targets for initial release.

## 📋 Current Status

### Completed Components
- ✅ Plan 1: Foundation (project structure, basic CLI)
- ✅ Plan 2: Configuration (Koanf integration, config management)
- ✅ Plan 3: Context detection (repository analysis, branch detection)
- ✅ Plan 4: Hybrid git operations (libgit2/go-git integration)
- ✅ Plan 5: Core services (worktree management, validation)
- ✅ Plan 6: CLI commands (list, create, delete, cd, setup-shell)
- ✅ Plan 7: Shell integration (bash/zsh/fish completion)
- ✅ Plan 8: Testing infrastructure (Ginkgo/Gomega, coverage)
- ✅ Plan 9: Performance optimization (caching, parallel operations)

### Integration Requirements
> "The system SHALL provide a cohesive user experience across all commands" - design.md
> "Performance targets: <100ms for list operations, <500ms for create operations" - implementation.md

## 🏗️ Integration Architecture

### Component Integration Matrix
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Core Services  │    │   Git Layer     │
│                 │    │                 │    │                 │
│ • Commands      │◄──►│ • WorktreeMgr   │◄──►│ • HybridGit     │
│ • Validation    │    │ • ContextDetect │    │ • Operations    │
│ • Output        │    │ • ConfigMgr     │    │ • Caching       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Data Layer    │
                    │                 │
                    │ • Config        │
                    │ • Cache         │
                    │ • State         │
                    └─────────────────┘
```

## 📝 Implementation Tasks

### Phase 1: End-to-End Integration Testing

#### 1.1 Workflow Integration Tests
**Priority**: Critical
**Files**: `test/integration/workflow_test.go`

**Test Scenarios**:
```go
Describe("Complete User Workflows", func() {
    Context("Rebase workflow", func() {
        It("should handle feature branch rebase preparation", func() {
            // Setup: Create main worktree, feature branch
            // Execute: twiggit create feature-branch
            // Validate: Worktree created, proper branch checked out
            // Cleanup: twiggit delete feature-branch
        })
        
        It("should handle multiple worktree management", func() {
            // Test: Create multiple worktrees, list, navigate
            // Validate: Proper isolation, no conflicts
        })
    })
    
    Context("Shell integration workflow", func() {
        It("should provide seamless navigation", func() {
            // Test: twiggit cd, shell completion
            // Validate: Proper directory changes, completion works
        })
    })
})
```

#### 1.2 Cross-Component Integration Tests
**Priority**: High
**Files**: `test/integration/component_test.go`

**Integration Points**:
- Configuration ↔ Core Services
- Context Detection ↔ Git Operations
- CLI Commands ↔ All Services
- Shell Integration ↔ CLI Commands

### Phase 2: Performance Validation

#### 2.1 Performance Benchmark Suite
**Priority**: Critical
**Files**: `test/performance/benchmark_test.go`

**Benchmark Targets** (from implementation.md):
```go
func BenchmarkListOperations(b *testing.B) {
    // Target: <100ms for up to 50 worktrees
}

func BenchmarkCreateOperations(b *testing.B) {
    // Target: <500ms for new worktree creation
}

func BenchmarkContextDetection(b *testing.B) {
    // Target: <50ms for repository analysis
}
```

#### 2.2 Load Testing
**Priority**: High
**Files**: `test/performance/load_test.go`

**Test Scenarios**:
- Large repositories (1000+ branches)
- Many worktrees (50+ concurrent)
- Deep directory structures
- Network latency conditions

### Phase 3: Cross-Platform Validation

#### 3.1 Platform Compatibility Tests
**Priority**: High
**Files**: `test/platform/compatibility_test.go`

**Target Platforms**:
- Linux (Ubuntu, CentOS, Alpine)
- macOS (Intel, Apple Silicon)
- Windows (WSL, native)

**Validation Areas**:
- File path handling
- Shell integration
- Git operations
- Permission handling

#### 3.2 Shell Integration Testing
**Priority**: Critical
**Files**: `test/shell/integration_test.go`

**Shell Support Matrix**:
```bash
# Bash 4.0+
bash -c "source <(twiggit setup-shell bash)"

# Zsh 5.0+
zsh -c "source <(twiggit setup-shell zsh)"

# Fish 3.0+
fish -c "twiggit setup-shell fish | source"
```

### Phase 4: Documentation Completion

#### 4.1 User Documentation
**Priority**: Critical
**Files**: 
- `docs/user-guide.md`
- `docs/quick-start.md`
- `docs/faq.md`
- `README.md`

**Content Requirements**:
> "Documentation SHALL be user-focused and comprehensive" - documentation-design.md

#### 4.2 API Documentation
**Priority**: High
**Files**: `docs/api/README.md`

**Generated Documentation**:
```bash
# Generate Go documentation
go doc -all > docs/api/go-docs.txt

# Generate CLI documentation
twiggit --help > docs/api/cli-help.txt
twiggit list --help > docs/api/list-help.txt
# ... for all commands
```

### Phase 5: Build and Release Pipeline

#### 5.1 Build System Enhancement
**Priority**: Critical
**Files**: `Makefile`, `.github/workflows/release.yml`

**Build Targets**:
```makefile
# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -o bin/twiggit-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o bin/twiggit-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o bin/twiggit-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o bin/twiggit-windows-amd64.exe

# Release preparation
release-prep: test lint build-all docs
	@echo "Preparing release..."
```

#### 5.2 Release Automation
**Priority**: High
**Files**: `.github/workflows/release.yml`

**Release Process**:
1. Run full test suite
2. Validate performance benchmarks
3. Build for all platforms
4. Generate documentation
5. Create GitHub release
6. Publish to package managers (if applicable)

### Phase 6: Quality Assurance

#### 6.1 Code Quality Validation
**Priority**: Critical
**Files**: Various (all source files)

**Quality Checks**:
```bash
# Code formatting
go fmt ./...

# Linting
golangci-lint run

# Security scanning
gosec ./...

# Dependency checking
go list -u -m all
```

#### 6.2 Final Integration Testing
**Priority**: Critical
**Files**: `test/integration/final_test.go`

**Comprehensive Test Suite**:
- All commands with all flag combinations
- Error handling and edge cases
- Resource cleanup and memory leaks
- Concurrent access patterns

## 🧪 Testing Strategy

### Integration Test Categories

#### 1. Workflow Integration Tests
**Purpose**: Validate complete user stories from design.md
**Coverage**: All primary use cases
**Tools**: Ginkgo/Gomega, temporary test repositories

#### 2. Component Integration Tests
**Purpose**: Validate interaction between components
**Coverage**: All component boundaries
**Tools**: Mock interfaces, contract testing

#### 3. Performance Integration Tests
**Purpose**: Validate optimization targets
**Coverage**: All critical paths
**Tools**: Benchmark testing, profiling

#### 4. Platform Integration Tests
**Purpose**: Validate cross-platform compatibility
**Coverage**: All supported platforms/shells
**Tools**: CI matrix testing, virtualization

### Test Data Management
```go
// test/fixtures/repositories.go
var TestRepositories = struct {
    SimpleRepo      string
    ComplexRepo     string
    LargeRepo       string
    RemoteRepo      string
}{...}

// test/fixtures/workflows.go
var TestWorkflows = struct {
    RebaseWorkflow  func(*testing.T)
    FeatureWorkflow func(*testing.T)
    HotfixWorkflow  func(*testing.T)
}{...}
```

## 📊 Validation Criteria

### Functional Validation
- [ ] All design.md requirements implemented and tested
- [ ] All commands work with all flag combinations
- [ ] Error handling covers all edge cases
- [ ] Shell integration works across bash/zsh/fish

### Performance Validation
- [ ] List operations <100ms (≤50 worktrees)
- [ ] Create operations <500ms
- [ ] Context detection <50ms
- [ ] Memory usage <50MB for typical operations

### Quality Validation
- [ ] Test coverage >80%
- [ ] No linting errors
- [ ] No security vulnerabilities
- [ ] Documentation complete and accurate

### Platform Validation
- [ ] Linux compatibility verified
- [ ] macOS compatibility verified
- [ ] Windows (WSL) compatibility verified
- [ ] All shell integrations functional

## 🚀 Release Preparation

### Pre-Release Checklist
1. **Code Quality**
   - [ ] All tests passing
   - [ ] Linting clean
   - [ ] Security scan clean
   - [ ] Documentation complete

2. **Performance Validation**
   - [ ] Benchmarks meet targets
   - [ ] Memory usage acceptable
   - [ ] Load testing successful

3. **Platform Testing**
   - [ ] All platforms tested
   - [ ] Shell integration verified
   - [ ] Installation procedures tested

4. **Documentation**
   - [ ] User guide complete
   - [ ] API documentation generated
   - [ ] Installation instructions clear
   - [ ] Troubleshooting guide ready

### Release Artifacts
```
release/
├── bin/
│   ├── twiggit-linux-amd64
│   ├── twiggit-darwin-amd64
│   ├── twiggit-darwin-arm64
│   └── twiggit-windows-amd64.exe
├── docs/
│   ├── user-guide.md
│   ├── quick-start.md
│   ├── api/
│   └── faq.md
├── scripts/
│   ├── install.sh
│   └── post-install.sh
└── checksums.txt
```

## 📈 Success Metrics

### Technical Metrics
- Test coverage: >80%
- Performance targets: 100% achieved
- Platform compatibility: 100% supported
- Code quality: Zero critical issues

### User Experience Metrics
- Installation time: <2 minutes
- Learning curve: <30 minutes for basic usage
- Error recovery: Clear error messages and recovery paths
- Documentation completeness: All scenarios covered

## 🔄 Continuous Integration

### CI Pipeline Enhancements
```yaml
# .github/workflows/integration.yml
name: Integration Testing
on: [push, pull_request]

jobs:
  integration:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        shell: [bash, zsh, fish]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - name: Run integration tests
        run: mise run test:integration
      - name: Validate performance
        run: mise run test:performance
```

## 📝 Documentation Updates

### User Guide Structure
```markdown
# User Guide

## Quick Start
## Installation
## Basic Usage
## Advanced Workflows
## Shell Integration
## Configuration
## Troubleshooting
## FAQ
```

### API Documentation
```markdown
# API Reference

## Commands
## Configuration
## Exit Codes
## Error Messages
## Examples
```

## ✅ Acceptance Criteria

### Must-Have (Release Blockers)
- [ ] All integration tests passing
- [ ] Performance targets met
- [ ] Cross-platform compatibility verified
- [ ] Documentation complete and reviewed
- [ ] Release pipeline functional

### Should-Have (Release Goals)
- [ ] User feedback incorporated
- [ ] Performance optimizations validated
- [ ] Security audit completed
- [ ] Installation scripts tested

### Could-Have (Future Enhancements)
- [ ] Additional shell support
- [ ] Performance monitoring
- [ ] Advanced configuration options
- [ ] Plugin system foundation

## 🎯 Next Steps

### Post-Release Activities
1. Monitor user feedback and issues
2. Collect performance data from real usage
3. Plan feature enhancements based on user needs
4. Establish maintenance and update schedule

### Long-term Roadmap
1. Advanced workflow automation
2. Integration with other development tools
3. Performance optimizations for large repositories
4. Enhanced shell integration features

---

## 📚 References

- [Design Requirements](../design.md)
- [Implementation Guide](../implementation.md)
- [Testing Strategy](../testing.md)
- [Documentation Design](../documentation-design.md)
- [Code Style Guide](../code-style-guide.md)

---

## Advanced Shell Integration Features (Deferred from Phase 7)

### Phase 7: Advanced Shell Customization

#### 7.1 Custom Wrapper Templates

**File**: `internal/domain/shell_templates.go`

```go
package domain

import (
    "fmt"
    "text/template"
)

type CustomTemplate struct {
    Name        string            `toml:"name"`
    ShellType   ShellType         `toml:"shell_type"`
    Template    string            `toml:"template"`
    Variables   map[string]string `toml:"variables"`
    Description string            `toml:"description"`
}

type TemplateManager interface {
    LoadCustomTemplates(configPath string) error
    GetCustomTemplate(name string, shellType ShellType) (*CustomTemplate, error)
    RenderTemplate(template *CustomTemplate, data map[string]interface{}) (string, error)
    ValidateTemplate(template string) error
}

type templateManager struct {
    templates map[string]map[ShellType]*CustomTemplate // name -> shell_type -> template
}

func NewTemplateManager() TemplateManager {
    return &templateManager{
        templates: make(map[string]map[ShellType]*CustomTemplate),
    }
}

func (tm *templateManager) LoadCustomTemplates(configPath string) error {
    // Load custom templates from user configuration
    // Implementation would read from ~/.twiggit/templates.toml or similar
    return nil
}

func (tm *templateManager) GetCustomTemplate(name string, shellType ShellType) (*CustomTemplate, error) {
    if shellTemplates, exists := tm.templates[name]; exists {
        if template, exists := shellTemplates[shellType]; exists {
            return template, nil
        }
    }
    return nil, fmt.Errorf("template '%s' not found for shell type '%s'", name, shellType)
}

func (tm *templateManager) RenderTemplate(template *CustomTemplate, data map[string]interface{}) (string, error) {
    tmpl, err := template.New(template.Name).Parse(template.Template)
    if err != nil {
        return "", fmt.Errorf("failed to parse template: %w", err)
    }
    
    // Merge template variables with provided data
    mergedData := make(map[string]interface{})
    for k, v := range template.Variables {
        mergedData[k] = v
    }
    for k, v := range data {
        mergedData[k] = v
    }
    
    var buf strings.Builder
    if err := tmpl.Execute(&buf, mergedData); err != nil {
        return "", fmt.Errorf("failed to execute template: %w", err)
    }
    
    return buf.String(), nil
}

func (tm *templateManager) ValidateTemplate(templateStr string) error {
    _, err := template.New("validation").Parse(templateStr)
    return err
}
```

#### 7.2 Multi-Shell Completion Integration

**File**: `internal/infrastructure/shell/completion.go`

```go
package shell

import (
    "context"
    "fmt"
    "strings"
)

type CompletionGenerator interface {
    GenerateCompletion(shell Shell) (string, error)
    InstallCompletion(shell Shell, completion string) error
    SupportedShells() []ShellType
}

type carapaceCompletionGenerator struct {
    binPath string
}

func NewCarapaceCompletionGenerator(binPath string) CompletionGenerator {
    return &carapaceCompletionGenerator{
        binPath: binPath,
    }
}

func (c *carapaceCompletionGenerator) GenerateCompletion(shell Shell) (string, error) {
    var args []string
    
    switch shell.Type() {
    case ShellBash:
        args = []string{"_carapace", "bash"}
    case ShellZsh:
        args = []string{"_carapace", "zsh"}
    case ShellFish:
        args = []string{"_carapace", "fish"}
    default:
        return "", fmt.Errorf("unsupported shell type for completion: %s", shell.Type())
    }
    
    // Execute carapace completion generation
    ctx := context.Background()
    completion, err := c.executeCompletionCommand(ctx, args)
    if err != nil {
        return "", fmt.Errorf("failed to generate completion: %w", err)
    }
    
    return c.postProcessCompletion(completion, shell.Type()), nil
}

func (c *carapaceCompletionGenerator) InstallCompletion(shell Shell, completion string) error {
    configPath, err := c.detectCompletionConfigFile(shell)
    if err != nil {
        return fmt.Errorf("failed to detect completion config file: %w", err)
    }
    
    return c.appendCompletionToConfig(configPath, completion, shell.Type())
}

func (c *carapaceCompletionGenerator) SupportedShells() []ShellType {
    return []ShellType{ShellBash, ShellZsh, ShellFish}
}

func (c *carapaceCompletionGenerator) executeCompletionCommand(ctx context.Context, args []string) (string, error) {
    // Implementation would execute the binary with carapace completion args
    return "", nil
}

func (c *carapaceCompletionGenerator) postProcessCompletion(completion string, shellType ShellType) string {
    // Add twiggit-specific completion enhancements
    switch shellType {
    case ShellBash:
        return c.enhanceBashCompletion(completion)
    case ShellZsh:
        return c.enhanceZshCompletion(completion)
    case ShellFish:
        return c.enhanceFishCompletion(completion)
    }
    return completion
}

func (c *carapaceCompletionGenerator) enhanceBashCompletion(completion string) string {
    // Add bash-specific enhancements
    enhancements := `
# Twiggit bash completion enhancements
_twiggit_cd_completion() {
    local cur prev words cword
    _init_completion || return
    
    # Complete branch names and project names
    if [[ ${cword} -eq 2 ]]; then
        local branches
        branches=$(twiggit list --format=short 2>/dev/null | awk '{print $1}')
        COMPREPLY=($(compgen -W "$branches" -- "$cur"))
    fi
}

complete -F _twiggit_cd_completion twiggit
`
    
    return completion + enhancements
}

func (c *carapaceCompletionGenerator) enhanceZshCompletion(completion string) string {
    // Add zsh-specific enhancements
    enhancements := `
# Twiggit zsh completion enhancements
_twiggit_cd() {
    local -a branches
    branches=($(twiggit list --format=short 2>/dev/null | awk '{print $1}'))
    _describe 'branches' branches
}

compdef _twiggit_cd twiggit
`
    
    return completion + enhancements
}

func (c *carapaceCompletionGenerator) enhanceFishCompletion(completion string) string {
    // Add fish-specific enhancements
    enhancements := `
# Twiggit fish completion enhancements
complete -c twiggit -n '__fish_use_subcommand' -a cd -d 'Change to worktree directory'
complete -c twiggit -n '__fish_seen_subcommand_from cd' -f -a '(twiggit list --format=short | awk \'{print $1}\')' -d 'Branch name'
`
    
    return completion + enhancements
}
```

#### 7.3 Advanced Shell Configuration

**File**: `internal/domain/shell_config.go`

```go
package domain

type AdvancedShellConfig struct {
    CustomTemplates    map[string]*CustomTemplateConfig `toml:"custom_templates"`
    Completion         *CompletionConfig                `toml:"completion"`
    AdvancedWrappers   *AdvancedWrapperConfig           `toml:"advanced_wrappers"`
    ShellIntegration   *ShellIntegrationConfig          `toml:"shell_integration"`
}

type CustomTemplateConfig struct {
    Name        string            `toml:"name"`
    ShellType   ShellType         `toml:"shell_type"`
    Template    string            `toml:"template"`
    Variables   map[string]string `toml:"variables"`
    Description string            `toml:"description"`
    Enabled     bool              `toml:"enabled"`
}

type CompletionConfig struct {
    Enabled     bool     `toml:"enabled"`
    AutoInstall bool     `toml:"auto_install"`
    Shells      []string `toml:"shells"`
    CustomHooks []string `toml:"custom_hooks"`
}

type AdvancedWrapperConfig struct {
    EnableAliases      bool              `toml:"enable_aliases"`
    CustomAliases      map[string]string `toml:"custom_aliases"`
    EnableHooks        bool              `toml:"enable_hooks"`
    PreCommandHooks    []string          `toml:"pre_command_hooks"`
    PostCommandHooks   []string          `toml:"post_command_hooks"`
    EnableTelemetry    bool              `toml:"enable_telemetry"`
    TelemetryEndpoint  string            `toml:"telemetry_endpoint"`
}

type ShellIntegrationConfig struct {
    AutoUpdate         bool          `toml:"auto_update"`
    UpdateInterval     time.Duration `toml:"update_interval"`
    BackupConfig       bool          `toml:"backup_config"`
    BackupLocation     string        `toml:"backup_location"`
    EnableValidation   bool          `toml:"enable_validation"`
    ValidationLevel    string        `toml:"validation_level"` // strict, lenient, disabled
}
```

#### 7.4 Enhanced Setup-Shell Command

**File**: `cmd/setup-shell-advanced.go`

```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

func NewSetupShellAdvancedCmd(config *CommandConfig) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "setup-shell-advanced",
        Short: "Advanced shell integration setup with customization options",
        Long: `Advanced shell integration setup with support for custom templates,
completion integration, and enhanced wrapper functionality.

This command provides:
- Custom wrapper template installation
- Shell completion setup
- Advanced configuration options
- Multi-shell support with customization`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runSetupShellAdvanced(cmd, config)
        },
    }

    cmd.Flags().String("template", "", "custom template name to use")
    cmd.Flags().Bool("completion", false, "install shell completion")
    cmd.Flags().Bool("advanced-wrapper", false, "install advanced wrapper with hooks")
    cmd.Flags().String("config-file", "", "path to custom configuration file")
    cmd.Flags().Bool("backup", true, "backup existing configuration")
    cmd.Flags().String("validation-level", "strict", "validation level (strict, lenient, disabled)")

    return cmd
}

func runSetupShellAdvanced(cmd *cobra.Command, config *CommandConfig) error {
    templateName, _ := cmd.Flags().GetString("template")
    installCompletion, _ := cmd.Flags().GetBool("completion")
    advancedWrapper, _ := cmd.Flags().GetBool("advanced-wrapper")
    configFile, _ := cmd.Flags().GetString("config-file")
    backup, _ := cmd.Flags().GetBool("backup")
    validationLevel, _ := cmd.Flags().GetString("validation-level")

    // Create advanced setup request
    request := &AdvancedSetupShellRequest{
        TemplateName:     templateName,
        InstallCompletion: installCompletion,
        AdvancedWrapper:  advancedWrapper,
        ConfigFile:       configFile,
        Backup:           backup,
        ValidationLevel:  validationLevel,
    }

    // Execute advanced setup
    result, err := config.ShellService.SetupShellAdvanced(context.Background(), request)
    if err != nil {
        return fmt.Errorf("advanced setup failed: %w", err)
    }

    // Display comprehensive results
    return displayAdvancedSetupResults(result)
}

func displayAdvancedSetupResults(result *AdvancedSetupShellResult) error {
    fmt.Printf("✓ Advanced shell integration completed\n")
    fmt.Printf("✓ Shell type: %s\n", result.ShellType)
    
    if result.TemplateInstalled {
        fmt.Printf("✓ Custom template '%s' installed\n", result.TemplateName)
    }
    
    if result.CompletionInstalled {
        fmt.Printf("✓ Shell completion installed\n")
    }
    
    if result.AdvancedWrapperInstalled {
        fmt.Printf("✓ Advanced wrapper with hooks installed\n")
    }
    
    if result.BackupCreated {
        fmt.Printf("✓ Configuration backed up to: %s\n", result.BackupPath)
    }
    
    if len(result.Warnings) > 0 {
        fmt.Printf("\nWarnings:\n")
        for _, warning := range result.Warnings {
            fmt.Printf("  ⚠️  %s\n", warning)
        }
    }
    
    if len(result.NextSteps) > 0 {
        fmt.Printf("\nNext steps:\n")
        for _, step := range result.NextSteps {
            fmt.Printf("  → %s\n", step)
        }
    }
    
    return nil
}
```

### Phase 8: Shell Integration Testing Extensions

#### 8.1 Advanced Shell Integration Tests

**File**: `test/integration/advanced_shell_test.go`

```go
//go:build integration
// +build integration

package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/twiggit/twiggit/internal/domain"
    "github.com/twiggit/twiggit/internal/infrastructure/shell"
)

func TestAdvancedShellIntegration_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    testCases := []struct {
        name         string
        templateName string
        shellType    domain.ShellType
        expectError  bool
    }{
        {
            name:         "custom bash template",
            templateName: "enhanced-bash",
            shellType:    domain.ShellBash,
            expectError:  false,
        },
        {
            name:         "custom zsh template",
            templateName: "enhanced-zsh",
            shellType:    domain.ShellZsh,
            expectError:  false,
        },
        {
            name:         "non-existent template",
            templateName: "non-existent",
            shellType:    domain.ShellBash,
            expectError:  true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            templateManager := shell.NewTemplateManager()
            
            // Load custom templates
            err := templateManager.LoadCustomTemplates("test/fixtures/templates.toml")
            require.NoError(t, err)
            
            // Get custom template
            template, err := templateManager.GetCustomTemplate(tc.templateName, tc.shellType)
            
            if tc.expectError {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.NotNil(t, template)
                assert.Equal(t, tc.templateName, template.Name)
                assert.Equal(t, tc.shellType, template.ShellType)
                
                // Test template rendering
                data := map[string]interface{}{
                    "ShellType": string(tc.shellType),
                    "Timestamp":  "2024-01-01 12:00:00",
                }
                
                rendered, err := templateManager.RenderTemplate(template, data)
                require.NoError(t, err)
                assert.NotEmpty(t, rendered)
                assert.Contains(t, rendered, "twiggit()")
            }
        })
    }
}
```

#### 8.2 Completion Integration Tests

**File**: `test/integration/completion_test.go`

```go
//go:build integration
// +build integration

package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/twiggit/twiggit/internal/infrastructure/shell"
)

func TestCompletionIntegration_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    testCases := []struct {
        name      string
        shellType domain.ShellType
        expectError bool
    }{
        {
            name:      "bash completion generation",
            shellType: domain.ShellBash,
            expectError: false,
        },
        {
            name:      "zsh completion generation",
            shellType: domain.ShellZsh,
            expectError: false,
        },
        {
            name:      "fish completion generation",
            shellType: domain.ShellFish,
            expectError: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            generator := shell.NewCarapaceCompletionGenerator("/usr/local/bin/twiggit")
            
            mockShell := &mockShell{shellType: tc.shellType}
            
            completion, err := generator.GenerateCompletion(mockShell)
            
            if tc.expectError {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.NotEmpty(t, completion)
                
                // Verify shell-specific completion features
                switch tc.shellType {
                case domain.ShellBash:
                    assert.Contains(t, completion, "_twiggit_cd_completion")
                case domain.ShellZsh:
                    assert.Contains(t, completion, "_twiggit_cd")
                case domain.ShellFish:
                    assert.Contains(t, completion, "complete -c twiggit")
                }
            }
        })
    }
}
```

These advanced shell integration features provide extensive customization options, enhanced completion support, and sophisticated configuration management while maintaining compatibility with the existing shell integration system.

*This plan ensures twiggit is production-ready with comprehensive testing, performance validation, complete documentation, and advanced shell integration features for a successful initial release.*