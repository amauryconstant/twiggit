# Shell Integration Services Implementation Plan

## Overview

This plan establishes the shell integration services layer that provides wrapper function generation and installation for twiggit. The services layer coordinates shell type handling, wrapper generation, and configuration file management while maintaining clean separation of concerns and functional programming principles.

**Context**: Foundation, configuration, context detection, hybrid git, core services, and CLI commands layers are established. This layer provides the shell integration functionality that enhances the user experience with seamless directory navigation.

## Foundation Principles

### TDD Approach
- **Test First**: Write failing tests, then implement minimal code to pass
- **Red-Green-Refactor**: Follow strict TDD cycle for each service
- **Minimal Implementation**: Implement only what's needed to pass current tests
- **Interface Contracts**: Test service interfaces before implementation

### Functional Programming Principles
- **Pure Functions**: Service operations SHALL be pure functions without side effects
- **Immutability**: Request/response structures SHALL be immutable
- **Function Composition**: Complex operations SHALL be composed from smaller functions
- **Error Handling**: SHALL use Result/Either patterns for error handling

### Clean Architecture
- **Thin Orchestration**: Services coordinate, don't contain business logic
- **Dependency Injection**: All external dependencies SHALL be injected via interfaces
- **Explicit Configuration**: Behavior SHALL be based on explicit user-provided shell type
- **Interface Segregation**: Services SHALL have focused, single-purpose interfaces

## Phase Boundaries

### Phase 7 Scope
- Shell service interface and implementation for explicit shell types
- Shell integration service for wrapper generation and installation
- Shell-specific syntax handlers for bash, zsh, and fish
- Setup-shell CLI command implementation
- Unit testing for service contracts
- Quality assurance configuration
- Functional programming patterns

### Deferred to Later Phases
- Performance optimization and caching (Phase 9)
- Advanced shell customization features (Phase 10)
- Integration/E2E testing (Phase 8)
- Multi-shell completion integration (Phase 10)

## Project Structure

Phase 7 minimal structure following Go standards:

```
internal/
├── domain/
│   ├── shell.go                # Shell types and interfaces
│   ├── shell_test.go           # Shell domain tests
│   ├── shell_requests.go       # Shell service request types
│   └── shell_errors.go         # Shell service errors
├── infrastructure/
│   └── shell/
│       ├── interfaces.go       # Shell infrastructure interfaces
│       ├── service.go          # Shell service implementation
│       ├── service_test.go     # Shell service tests
│       ├── bash.go            # Bash-specific syntax
│       ├── zsh.go             # Zsh-specific syntax
│       └── fish.go            # Fish-specific syntax
├── services/
│   ├── interfaces.go           # Service interfaces (extend existing)
│   └── shell_service.go        # ShellService implementation
cmd/
└── setup-shell.go              # Setup-shell command implementation
```

**Removed from Phase 7** (deferred to later phases):
- Advanced wrapper customization → Phase 10
- Performance monitoring → Phase 9
- Complex shell integration patterns → Phase 10
- Integration test fixtures → Phase 8

## Implementation Steps

### Step 1: Shell Domain Types and Interfaces

**Files to create:**
- `internal/domain/shell.go`
- `internal/domain/shell_requests.go`
- `internal/domain/shell_errors.go`

**Tests first:** `internal/domain/shell_test.go`

```go
func TestShellDomain_ContractCompliance(t *testing.T) {
    testCases := []struct {
        name        string
        shellType   ShellType
        expectValid bool
    }{
        {
            name:        "bash shell type is valid",
            shellType:   ShellBash,
            expectValid: true,
        },
        {
            name:        "zsh shell type is valid",
            shellType:   ShellZsh,
            expectValid: true,
        },
        {
            name:        "fish shell type is valid",
            shellType:   ShellFish,
            expectValid: true,
        },
        {
            name:        "unknown shell type is invalid",
            shellType:   ShellType("unknown"),
            expectValid: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            shell := NewShell(tc.shellType, "/bin/test", "1.0")
            
            if tc.expectValid {
                assert.NotNil(t, shell)
                assert.Equal(t, tc.shellType, shell.Type())
            } else {
                assert.Nil(t, shell)
            }
        })
    }
}
```

**Domain definitions:**
```go
// internal/domain/shell.go
type ShellType string

const (
    ShellBash ShellType = "bash"
    ShellZsh  ShellType = "zsh"
    ShellFish ShellType = "fish"
)

type Shell interface {
    Type() ShellType
    Path() string
    Version() string
    ConfigFiles() []string
    WrapperTemplate() string
}

type shell struct {
    shellType ShellType
    path      string
    version   string
}

func NewShell(shellType ShellType, path, version string) (Shell, error) {
    if !isValidShellType(shellType) {
        return nil, fmt.Errorf("unsupported shell type: %s", shellType)
    }
    
    return &shell{
        shellType: shellType,
        path:      path,
        version:   version,
    }, nil
}
```

### Step 2: Shell Service

**Tests first:** `internal/infrastructure/shell/service_test.go`

```go
func TestShellService_GenerateWrapper_Success(t *testing.T) {
    testCases := []struct {
        name        string
        shellType   ShellType
        expectError bool
        validate    func(t *testing.T, wrapper string)
    }{
        {
            name:      "generate bash wrapper",
            shellType: ShellBash,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "twiggit() {")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
                assert.Contains(t, wrapper, "# Twiggit bash wrapper")
            },
        },
        {
            name:      "generate zsh wrapper",
            shellType: ShellZsh,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "twiggit() {")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
                assert.Contains(t, wrapper, "# Twiggit zsh wrapper")
            },
        },
        {
            name:      "generate fish wrapper",
            shellType: ShellFish,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "function twiggit")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
                assert.Contains(t, wrapper, "# Twiggit fish wrapper")
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := NewShellService()
            wrapper, err := service.GenerateWrapper(tc.shellType)
            
            if tc.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, wrapper)
                tc.validate(t, wrapper)
            }
        })
    }
}
```

**Implementation sketch:**
```go
// internal/infrastructure/shell/service.go
type ShellService interface {
    GenerateWrapper(shellType ShellType) (string, error)
    DetectConfigFile(shellType ShellType) (string, error)
    InstallWrapper(shellType ShellType, wrapper string) error
    ValidateInstallation(shellType ShellType) error
}

type shellService struct{}

func NewShellService() ShellService {
    return &shellService{}
}

func (s *shellService) GenerateWrapper(shellType ShellType) (string, error) {
    template := s.getWrapperTemplate(shellType)
    if template == "" {
        return nil, fmt.Errorf("unsupported shell type: %s", shellType)
    }
    
    // Pure function composition for wrapper generation
    return s.composeWrapper(template, shellType), nil
}

func (s *shellService) getWrapperTemplate(shellType ShellType) string {
    switch shellType {
    case ShellBash:
        return s.bashWrapperTemplate()
    case ShellZsh:
        return s.zshWrapperTemplate()
    case ShellFish:
        return s.fishWrapperTemplate()
    default:
        return ""
    }
}

func (s *shellService) composeWrapper(template string, shellType ShellType) string {
    // Pure function: no side effects, deterministic output
    replacements := map[string]string{
        "{{SHELL_TYPE}}": string(shellType),
        "{{TIMESTAMP}}":  time.Now().Format("2006-01-02 15:04:05"),
    }
    
    result := template
    for key, value := range replacements {
        result = strings.ReplaceAll(result, key, value)
    }
    
    return result
}
```

### Step 3: Shell Service Layer

**Tests first:** `internal/services/shell_service_test.go`

```go
func TestShellService_SetupShell_Success(t *testing.T) {
    testCases := []struct {
        name        string
        shellType   ShellType
        expectError bool
        validate    func(t *testing.T, wrapper string)
    }{
        {
            name:      "generate bash wrapper",
            shellType: ShellBash,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "twiggit() {")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
                assert.Contains(t, wrapper, "# Twiggit bash wrapper")
            },
        },
        {
            name:      "generate zsh wrapper",
            shellType: ShellZsh,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "twiggit() {")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
                assert.Contains(t, wrapper, "# Twiggit zsh wrapper")
            },
        },
        {
            name:      "generate fish wrapper",
            shellType: ShellFish,
            validate: func(t *testing.T, wrapper string) {
                assert.Contains(t, wrapper, "function twiggit")
                assert.Contains(t, wrapper, "builtin cd")
                assert.Contains(t, wrapper, "command twiggit")
                assert.Contains(t, wrapper, "# Twiggit fish wrapper")
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := setupTestShellService()
            
            request := &SetupShellRequest{
                ShellType: tc.shellType,
                Force:     false,
                DryRun:    true,
            }
            result, err := service.SetupShell(context.Background(), request)
            
            if tc.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, wrapper)
                tc.validate(t, wrapper)
            }
        })
    }
}
```

**Implementation sketch:**
```go
// internal/services/shell_service.go
type ShellService interface {
    SetupShell(ctx context.Context, req *SetupShellRequest) (*SetupShellResult, error)
    ValidateInstallation(ctx context.Context, shellType ShellType) error
}

type shellService struct {
    integration ShellIntegration
    config      *domain.Config
}

func NewShellService(
    integration ShellIntegration,
    config *domain.Config,
) ShellService {
    return &shellService{
        integration: integration,
        config:      config,
    }
}

func (s *shellService) SetupShell(ctx context.Context, req *SetupShellRequest) (*SetupShellResult, error) {
    // Pure function: validate request first
    if err := s.validateSetupRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Check existing installation
    if !req.Force {
        if err := s.integration.ValidateInstallation(req.ShellType); err == nil {
            return &SetupShellResult{
                ShellType: req.ShellType,
                Installed: true,
                Skipped:   true,
                Message:   "Shell wrapper already installed",
            }, nil
        }
    }
    
    // Generate wrapper
    wrapper, err := s.integration.GenerateWrapper(req.ShellType)
    if err != nil {
        return nil, fmt.Errorf("failed to generate wrapper: %w", err)
    }
    
    // Handle dry run
    if req.DryRun {
        return &SetupShellResult{
            ShellType:     req.ShellType,
            Installed:     false,
            DryRun:        true,
            WrapperContent: wrapper,
            Message:       "Dry run completed",
        }, nil
    }
    
    // Install wrapper
    if err := s.integration.InstallWrapper(req.ShellType, wrapper); err != nil {
        return nil, fmt.Errorf("failed to install wrapper: %w", err)
    }
    
    return &SetupShellResult{
        ShellType: req.ShellType,
        Installed: true,
        Message:   "Shell wrapper installed successfully",
    }, nil
}
```

### Step 4: Shell Service Layer

**Tests first:** `internal/services/shell_service_test.go`

```go
func TestShellService_SetupShell_Success(t *testing.T) {
    testCases := []struct {
        name         string
        request      *SetupShellRequest
        expectError  bool
        errorMessage string
    }{
        {
            name: "valid setup request",
            request: &SetupShellRequest{
                Force:  false,
                DryRun: false,
            },
            expectError: false,
        },
        {
            name: "dry run request",
            request: &SetupShellRequest{
                Force:  false,
                DryRun: true,
            },
            expectError: false,
        },
        {
            name: "force reinstall request",
            request: &SetupShellRequest{
                Force:  true,
                DryRun: false,
            },
            expectError: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            service := setupTestShellService()
            result, err := service.SetupShell(context.Background(), tc.request)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                
                if tc.request.DryRun {
                    assert.NotEmpty(t, result.WrapperContent)
                    assert.False(t, result.Installed)
                } else {
                    assert.True(t, result.Installed)
                }
            }
        })
    }
}
```

**Implementation sketch:**
```go
// internal/services/shell_service.go
type ShellService interface {
    SetupShell(ctx context.Context, req *SetupShellRequest) (*SetupShellResult, error)
    ValidateInstallation(ctx context.Context, shellType ShellType) error
}

type shellService struct {
    integration  ShellIntegration
    config       *domain.Config
}

func NewShellService(
    integration ShellIntegration,
    config *domain.Config,
) ShellService {
    return &shellService{
        integration: integration,
        config:      config,
    }
}

func (s *shellService) SetupShell(ctx context.Context, req *SetupShellRequest) (*SetupShellResult, error) {
    // Pure function: validate request first
    if err := s.validateSetupRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Check existing installation
    if !req.Force {
        if err := s.integration.ValidateInstallation(req.ShellType); err == nil {
            return &SetupShellResult{
                ShellType: req.ShellType,
                Installed: true,
                Skipped:   true,
                Message:   "Shell wrapper already installed",
            }, nil
        }
    }
    
    // Generate wrapper
    wrapper, err := s.integration.GenerateWrapper(req.ShellType)
    if err != nil {
        return nil, fmt.Errorf("failed to generate wrapper: %w", err)
    }
    
    // Handle dry run
    if req.DryRun {
        return &SetupShellResult{
            ShellType:     shell.Type(),
            Installed:     false,
            DryRun:        true,
            WrapperContent: wrapper,
            Message:       "Dry run completed",
        }, nil
    }
    
    // Install wrapper
    if err := s.integration.InstallWrapper(shell, wrapper); err != nil {
        return nil, fmt.Errorf("failed to install wrapper: %w", err)
    }
    
    return &SetupShellResult{
        ShellType: shell.Type(),
        Installed: true,
        Message:   "Shell wrapper installed successfully",
    }, nil
}
```

### Step 4: Setup-Shell CLI Command

**Tests first:** `cmd/setup-shell_test.go`

```go
func TestSetupShellCommand_Success(t *testing.T) {
    testCases := []struct {
        name        string
        args        []string
        expectError bool
        validate    func(t *testing.T, output string)
    }{
        {
            name: "setup bash shell dry run",
            args: []string{"setup-shell", "--shell=bash", "--dry-run"},
            validate: func(t *testing.T, output string) {
                assert.Contains(t, output, "Dry run completed")
                assert.Contains(t, output, "twiggit() {")
            },
        },
        {
            name: "setup zsh with force flag",
            args: []string{"setup-shell", "--shell=zsh", "--force"},
            validate: func(t *testing.T, output string) {
                assert.Contains(t, output, "Shell wrapper installed successfully")
            },
        },
        {
            name:        "missing shell flag fails",
            args:        []string{"setup-shell"},
            expectError: true,
        },
        {
            name:        "invalid shell type fails",
            args:        []string{"setup-shell", "--shell=invalid"},
            expectError: true,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Set up test environment
            config := setupTestCommandConfig()
            cmd := NewSetupShellCmd(config)
            
            // Execute command
            buf := new(bytes.Buffer)
            cmd.SetOut(buf)
            cmd.SetErr(buf)
            
            err := cmd.Execute()
            
            if tc.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                if tc.validate != nil {
                    tc.validate(t, buf.String())
                }
            }
        })
    }
}
```

**Implementation sketch:**
```go
// cmd/setup-shell.go
func NewSetupShellCmd(config *CommandConfig) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "setup-shell",
        Short: "Install shell wrapper for directory navigation",
        Long: `Install shell wrapper functions that intercept 'twiggit cd' calls
and enable seamless directory navigation between worktrees and projects.

The wrapper provides:
- Automatic directory change on 'twiggit cd'
- Escape hatch with 'builtin cd' for shell built-in
- Pass-through for all other commands

Supported shells: bash, zsh, fish

Usage: twiggit setup-shell --shell=bash`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runSetupShell(cmd, config)
        },
    }

    cmd.Flags().String("shell", "", "shell type (bash|zsh|fish) [required]")
    cmd.Flags().Bool("force", false, "force reinstall even if already installed")
    cmd.Flags().Bool("dry-run", false, "show what would be done without making changes")

    _ = cmd.MarkFlagRequired("shell")

    return cmd
}

func runSetupShell(cmd *cobra.Command, config *CommandConfig) error {
    shellTypeStr, _ := cmd.Flags().GetString("shell")
    force, _ := cmd.Flags().GetBool("force")
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // Validate shell type
    shellType := ShellType(shellTypeStr)
    if !isValidShellType(shellType) {
        return fmt.Errorf("unsupported shell type: %s (supported: bash, zsh, fish)", shellTypeStr)
    }

    // Create request
    request := &SetupShellRequest{
        ShellType: shellType,
        Force:     force,
        DryRun:    dryRun,
    }

    // Execute service
    result, err := config.ShellService.SetupShell(context.Background(), request)
    if err != nil {
        return fmt.Errorf("setup failed: %w", err)
    }

    // Display results
    return displaySetupResults(result, dryRun)
}

func displaySetupResults(result *SetupShellResult, dryRun bool) error {
    if result.Skipped {
        fmt.Printf("✓ Shell wrapper already installed for %s\n", result.ShellType)
        fmt.Printf("Use --force to reinstall\n")
        return nil
    }

    if dryRun {
        fmt.Printf("Would install wrapper for %s:\n", result.ShellType)
        fmt.Printf("Wrapper function:\n%s\n", result.WrapperContent)
        return nil
    }

    if result.Installed {
        fmt.Printf("✓ Shell wrapper installed for %s\n", result.ShellType)
        fmt.Printf("✓ %s\n", result.Message)
        
        fmt.Printf("\nTo activate the wrapper:\n")
        fmt.Printf("  1. Restart your shell, or\n")
        fmt.Printf("  2. Run: source ~/.bashrc (or ~/.zshrc, etc.)\n")
        fmt.Printf("\nUsage:\n")
        fmt.Printf("  twiggit cd <branch>     # Change to worktree\n")
        fmt.Printf("  builtin cd <path>       # Use shell built-in cd\n")
    }

    return nil
}
```

## Testing Strategy

Phase 7 focuses exclusively on unit testing for service contracts.

### Unit Tests Only
- **Framework**: Testify with table-driven tests
- **Coverage**: >80% for service logic (realistic for orchestration layer)
- **Location**: `*_test.go` files alongside implementation
- **Focus**: Interface contracts, error handling, shell type validation

### Test Organization
- **Domain Tests**: Test shell types and validation logic
- **Infrastructure Tests**: Test shell wrapper generation and installation
- **Service Tests**: Test shell service orchestration
- **Command Tests**: Test CLI command behavior with mocked services

### Deferred Testing Types
- **Integration Tests**: Phase 8 (when real shell config files exist)
- **E2E Tests**: Phase 8 (when complete CLI workflow exists)
- **Performance Tests**: Phase 9

## Quality Gates

### Pre-commit Requirements
- All tests pass: `go test ./...`
- Linting passes: `golangci-lint run`
- Coverage >80%: `go test -cover ./...`
- Interface compliance tests pass

### CI Requirements
- Unit tests pass
- Linting passes
- Build succeeds on target platforms
- Functional programming principles verified

## Key Principles

### TDD Approach
- **Write failing test first**
- **Implement minimal code to pass**
- **Refactor while keeping tests green**
- **Repeat for next service**

### Functional Programming
- **Pure functions**: No side effects in service operations
- **Immutability**: Immutable request/response structures
- **Composition**: Build complex operations from simple functions
- **Error handling**: Use Result patterns for predictable error flow

### Clean Code
- **Interface segregation**: Small, focused interfaces
- **Dependency injection**: All dependencies injected
- **Single responsibility**: Each service has one clear purpose
- **Consistent error handling**: Same error pattern throughout

## Configuration Integration

### Shell Configuration Extensions

Since Phase 02 (Configuration) is already implemented, the following shell-specific configuration SHALL be added in Phase 07:

```toml
# config.toml additions for shell integration
[shell]
default_shell = "bash"
backup_config = true

[shell.wrapper]
enable_warning = true
warning_message = "twiggit: shell wrapper installed - use 'builtin cd' for shell built-in"

[shell.config_files]
bash_preferred = [".bashrc", ".bash_profile", ".profile"]
zsh_preferred = [".zshrc", ".zprofile", ".profile"]
fish_preferred = ["config.fish", ".fishrc"]
```

### Configuration Implementation
**File:** `internal/domain/config.go` (extend existing)

```go
// ShellConfig holds shell-specific configuration
type ShellConfig struct {
    DefaultShell string `koanf:"default_shell"`
    BackupConfig bool   `koanf:"backup_config"`
}

type ShellWrapperConfig struct {
    EnableWarning   bool   `koanf:"enable_warning"`
    WarningMessage  string `koanf:"warning_message"`
}

type ShellConfigFiles struct {
    BashPreferred []string `koanf:"bash_preferred"`
    ZshPreferred  []string `koanf:"zsh_preferred"`
    FishPreferred []string `koanf:"fish_preferred"`
}
```

## Success Criteria

1. ✅ Shell domain types (Shell, ShellType) with comprehensive validation
2. ✅ Shell service with explicit shell type handling
3. ✅ Shell integration service with wrapper generation and installation
4. ✅ Shell service layer with functional composition patterns
5. ✅ Setup-shell CLI command with required --shell flag
6. ✅ Shell configuration extensions integrated with existing config system
7. ✅ Unit tests for service contracts pass with >80% coverage
8. ✅ Basic linting passes without errors
9. ✅ Clean service structure following Go standards
10. ✅ Quality gates enforce functional programming principles

## Incremental Development Strategy

Phase 7 follows strict incremental development:

1. **Write Test**: Create failing test for shell domain type
2. **Define Interface**: Add shell interface with method signatures
3. **Implement**: Add minimal code to make test pass
4. **Refactor**: Apply functional programming patterns while keeping tests green
5. **Repeat**: Move to next service component

**No detailed implementation, no premature optimization, no future-proofing.** Each service builds only what's needed for that phase.

## Next Phases

Phase 7 provides the shell integration services needed for enhanced user experience:

1. **Phase 8**: Comprehensive testing infrastructure with integration and E2E tests
2. **Phase 9**: Performance optimization and caching for shell operations
3. **Phase 10**: Final integration validation and advanced shell features

This shell integration services layer provides the essential functionality for seamless directory navigation while following true TDD principles, functional programming patterns, and maintaining clean phase boundaries.