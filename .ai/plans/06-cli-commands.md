# CLI Commands Layer Implementation Plan

## Overview

This plan implements the user-facing CLI commands layer using the Cobra framework, providing a clean interface to the underlying services while maintaining context awareness and consistent user experience.

## Context from Previous Layers

> From technology.md: "Cobra for CLI structure - provides consistent command structure, flag handling, and help generation"

> From design.md: "Core commands: list, create, cd, delete, setup-shell, context, help. Context-aware command behavior based on detected context"

> From implementation.md: "Error handling with POSIX exit codes. Commands should be thin wrappers around services"

## Architecture

### Directory Structure
```
cmd/
├── root.go              # Root command and global flags
├── list.go              # list command implementation
├── create.go            # create command implementation
├── cd.go                # cd command implementation
├── delete.go            # delete command implementation
├── setup_shell.go       # setup-shell command implementation
└── context.go           # context command implementation
```

### Command Design Principles
1. **Thin wrappers**: Commands delegate to services layer
2. **Context awareness**: Detect context before execution
3. **Consistent flags**: Standard flag set across commands
4. **Error handling**: POSIX exit codes with clear messages
5. **Help adaptation**: Context-sensitive help text

## Implementation Steps

### Step 1: Root Command and Global Configuration

**File**: `cmd/root.go`

**Requirements**:
- Global flags: --dry-run, --explain, --verbose, --config
- Version information
- Context detection integration
- Dependency injection setup

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

type GlobalOptions struct {
    DryRun  bool
    Explain bool
    Verbose bool
    Config  string
}

var globalOpts = &GlobalOptions{}

var rootCmd = &cobra.Command{
    Use:   "twiggit",
    Short: "Pragmatic git worktree management tool",
    Long:  `Twiggit manages git worktrees with focus on rebase workflows.`,
    Version: "v1.0.0",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Context detection before any command
        // Service initialization
        // Configuration loading
        return nil
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().BoolVar(&globalOpts.DryRun, "dry-run", false, "Show what would be done without executing")
    rootCmd.PersistentFlags().BoolVar(&globalOpts.Explain, "explain", false, "Explain the reasoning behind decisions")
    rootCmd.PersistentFlags().BoolVar(&globalOpts.Verbose, "verbose", false, "Enable verbose output")
    rootCmd.PersistentFlags().StringVar(&globalOpts.Config, "config", "", "Path to configuration file")
}
```

### Step 2: List Command Implementation

**File**: `cmd/list.go`

**Requirements from design.md**:
- "tabular format, context-aware scope, --all flag for unlimited display"

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

var listAll bool

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List worktrees",
    Long:  `List worktrees in tabular format with context-aware scope.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get services from dependency injection
        worktreeService := services.GetWorktreeService()
        contextService := services.GetContextService()
        
        // Detect current context
        ctx, err := contextService.DetectContext()
        if err != nil {
            return fmt.Errorf("context detection failed: %w", err)
        }
        
        // List worktrees with context awareness
        worktrees, err := worktreeService.ListWorktrees(ctx, listAll)
        if err != nil {
            return fmt.Errorf("failed to list worktrees: %w", err)
        }
        
        // Display in tabular format
        displayWorktrees(worktrees, globalOpts.Verbose)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(listCmd)
    listCmd.Flags().BoolVar(&listAll, "all", false, "Show unlimited worktrees without pagination")
}
```

### Step 3: Create Command Implementation

**File**: `cmd/create.go`

**Requirements from design.md**:
- "<project>/<branch> format, --source flag, --cd flag, context-aware project inference"

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

var (
    createSource string
    createCd     bool
)

var createCmd = &cobra.Command{
    Use:   "create <project>/<branch>",
    Short: "Create a new worktree",
    Long:  `Create a new worktree with project/branch format and context-aware project inference.`,
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        worktreeService := services.GetWorktreeService()
        contextService := services.GetContextService()
        
        // Detect context for project inference
        ctx, err := contextService.DetectContext()
        if err != nil {
            return fmt.Errorf("context detection failed: %w", err)
        }
        
        // Parse project/branch specification
        project, branch, err := parseProjectBranch(args[0], ctx)
        if err != nil {
            return fmt.Errorf("invalid project/branch format: %w", err)
        }
        
        // Create worktree
        worktree, err := worktreeService.CreateWorktree(ctx, project, branch, createSource)
        if err != nil {
            return fmt.Errorf("failed to create worktree: %w", err)
        }
        
        // Handle --cd flag - output path for shell wrapper (Phase 07)
        if createCd {
            fmt.Println(worktree.Path)
        }
        
        fmt.Printf("Created worktree: %s\n", worktree.Path)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(createCmd)
    createCmd.Flags().StringVar(&createSource, "source", "", "Source branch or commit")
    createCmd.Flags().BoolVar(&createCd, "cd", false, "Change to worktree directory after creation")
}
```

### Step 4: CD Command Implementation

**File**: `cmd/cd.go`

**Requirements from design.md**:
- "target resolution with context awareness, shell integration support"

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

var cdCmd = &cobra.Command{
    Use:   "cd [target]",
    Short: "Change to worktree directory",
    Long:  `Change to worktree directory with context-aware target resolution.`,
    Args:  cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        worktreeService := services.GetWorktreeService()
        contextService := services.GetContextService()
        
        // Detect context for target resolution
        ctx, err := contextService.DetectContext()
        if err != nil {
            return fmt.Errorf("context detection failed: %w", err)
        }
        
        var target string
        if len(args) > 0 {
            target = args[0]
        } else {
            // Use context-aware default
            target = ctx.DefaultWorktree
        }
        
        // Resolve target identifier using context service
        resolution, err := contextService.ResolveIdentifier(ctx, target)
        if err != nil {
            return fmt.Errorf("failed to resolve target: %w", err)
        }
        
        // Get worktree from resolution
        worktree, err := worktreeService.GetWorktree(resolution.WorktreePath)
        if err != nil {
            return fmt.Errorf("failed to get worktree: %w", err)
        }
        
        // Output target path for shell wrapper consumption (Phase 07)
        fmt.Println(worktree.Path)
        
        return nil
    },
}

func init() {
    rootCmd.AddCommand(cdCmd)
}
```

### Step 5: Delete Command Implementation

**File**: `cmd/delete.go`

**Requirements from design.md**:
- "safety checks, --force flag, --keep-branch option"

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

var (
    deleteForce     bool
    deleteKeepBranch bool
)

var deleteCmd = &cobra.Command{
    Use:   "delete <worktree>",
    Short: "Delete a worktree",
    Long:  `Delete a worktree with safety checks and branch preservation options.`,
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        worktreeService := services.GetWorktreeService()
        contextService := services.GetContextService()
        
        // Detect context
        ctx, err := contextService.DetectContext()
        if err != nil {
            return fmt.Errorf("context detection failed: %w", err)
        }
        
        // Resolve target identifier using context service
        resolution, err := contextService.ResolveIdentifier(ctx, args[0])
        if err != nil {
            return fmt.Errorf("failed to resolve target: %w", err)
        }
        
        // Get worktree from resolution
        worktree, err := worktreeService.GetWorktree(resolution.WorktreePath)
        if err != nil {
            return fmt.Errorf("failed to get worktree: %w", err)
        }
        
        // Safety checks
        if !deleteForce {
            if err := confirmDeletion(worktree); err != nil {
                return err
            }
        }
        
        // Delete worktree
        if err := worktreeService.DeleteWorktree(ctx, worktree, deleteKeepBranch); err != nil {
            return fmt.Errorf("failed to delete worktree: %w", err)
        }
        
        fmt.Printf("Deleted worktree: %s\n", worktree.Path)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(deleteCmd)
    deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip safety checks")
    deleteCmd.Flags().BoolVar(&deleteKeepBranch, "keep-branch", false, "Keep the branch after deletion")
}
```

### Step 6: Setup-Shell Command Implementation

**File**: `cmd/setup_shell.go`

**Requirements from design.md**:
- "alias generation, shell detection, configuration file modification"

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

var setupShellCmd = &cobra.Command{
    Use:   "setup-shell",
    Short: "Setup shell integration",
    Long:  `Setup shell integration with alias generation and configuration file modification.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        shellService := services.GetShellService()
        
        // Detect current shell
        shell, err := shellService.DetectShell()
        if err != nil {
            return fmt.Errorf("failed to detect shell: %w", err)
        }
        
        // Generate aliases
        aliases, err := shellService.GenerateAliases(shell)
        if err != nil {
            return fmt.Errorf("failed to generate aliases: %w", err)
        }
        
        // Modify configuration file
        if err := shellService.SetupShellIntegration(shell, aliases); err != nil {
            return fmt.Errorf("failed to setup shell integration: %w", err)
        }
        
        fmt.Printf("Shell integration setup complete for %s\n", shell)
        fmt.Println("Restart your shell or run: source ~/.bashrc")
        return nil
    },
}

func init() {
    rootCmd.AddCommand(setupShellCmd)
}
```

### Step 7: Context Command Implementation

**File**: `cmd/context.go`

**Requirements from design.md**:
- "context command for context inspection and --explain functionality"

**Code Structure**:
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)

var contextCmd = &cobra.Command{
    Use:   "context",
    Short: "Show current context",
    Long:  `Show current context and explain reasoning behind context detection.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        contextService := services.GetContextService()
        
        // Detect context
        ctx, err := contextService.DetectContext()
        if err != nil {
            return fmt.Errorf("context detection failed: %w", err)
        }
        
        // Display context
        displayContext(ctx, globalOpts.Explain, globalOpts.Verbose)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(contextCmd)
}
```

## Common Utilities

### Error Handling
```go
package cmd

import (
    "fmt"
    "os"
)

// Exit codes following POSIX conventions
const (
    ExitSuccess = 0
    ExitGeneral = 1
    ExitUsage   = 2
    ExitNoInput = 3
)

func handleError(err error, message string) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %s: %v\n", message, err)
        os.Exit(ExitGeneral)
    }
}

func handleUsageError(message string) {
    fmt.Fprintf(os.Stderr, "Usage Error: %s\n", message)
    os.Exit(ExitUsage)
}
```

### Display Utilities
```go
package cmd

import (
    "fmt"
    "text/tabwriter"
    "os"
)

func displayWorktrees(worktrees []Worktree, verbose bool) {
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintln(w, "PATH\tBRANCH\tPROJECT\tSTATUS")
    
    for _, wt := range worktrees {
        status := "clean"
        if wt.Dirty {
            status = "dirty"
        }
        
        fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", wt.Path, wt.Branch, wt.Project, status)
    }
    
    w.Flush()
}

func displayContext(ctx Context, explain bool, verbose bool) {
    fmt.Printf("Current Context:\n")
    fmt.Printf("  Project: %s\n", ctx.Project)
    fmt.Printf("  Branch: %s\n", ctx.Branch)
    fmt.Printf("  Worktree: %s\n", ctx.Worktree)
    
    if explain {
        fmt.Printf("\nExplanation:\n")
        fmt.Printf("  %s\n", ctx.Explanation)
    }
    
    if verbose {
        fmt.Printf("\nVerbose Information:\n")
        fmt.Printf("  Detection Method: %s\n", ctx.DetectionMethod)
        fmt.Printf("  Confidence: %d\n", ctx.Confidence)
    }
}
```

## Integration Points

### Services Integration
```go
// In cmd/root.go PersistentPreRunE
func initializeServices() error {
    // Initialize all services with proper configuration
    config := loadConfiguration(globalOpts.Config)
    
    services.InitializeWorktreeService(config.Worktree)
    services.InitializeContextService(config.Context)
    services.InitializeShellService(config.Shell)
    
    return nil
}
```

### Context Detection Integration
```go
// Common pattern for all commands
func executeWithContext(cmdFunc func(Context) error) error {
    contextService := services.GetContextService()
    ctx, err := contextService.DetectContext()
    if err != nil {
        return fmt.Errorf("context detection failed: %w", err)
    }
    
    if globalOpts.Explain {
        fmt.Printf("Context: %s\n", ctx.Explanation)
    }
    
    return cmdFunc(ctx)
}

// Common pattern for identifier resolution
func resolveIdentifier(ctx Context, target string) (*domain.Resolution, error) {
    contextService := services.GetContextService()
    return contextService.ResolveIdentifier(ctx, target)
}
```

## Testing Strategy

### Unit Tests
- Test each command in isolation
- Mock services for predictable behavior
- Test flag parsing and validation
- Test error handling and exit codes

### Integration Tests
- Test command execution with real services
- Test context detection integration
- Test shell integration functionality
- Test configuration loading

### Example Test Structure
```go
// cmd/list_test.go
package cmd

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
    // Setup mocks
    // Execute command
    // Verify output
    // Verify service calls
}
```

## Quality Assurance

### Code Review Checklist
- [ ] Consistent flag usage across commands
- [ ] Proper error handling with exit codes
- [ ] Context detection before execution
- [ ] Help text is clear and accurate
- [ ] Integration with services layer
- [ ] POSIX compliance

### Linting and Formatting
- Use `gofmt` for code formatting
- Use `golint` for style checking
- Use `go vet` for static analysis
- Ensure all tests pass

## Dependencies

### Required Imports
```go
import (
    "fmt"
    "os"
    "text/tabwriter"
    
    "github.com/spf13/cobra"
    "github.com/twiggit/twiggit/internal/services"
)
```

### Service Dependencies
- WorktreeService for worktree operations
- ContextService for context detection and identifier resolution
- ShellService for shell integration
- ConfigurationService for config management

## Success Criteria

1. All core commands implemented with Cobra
2. Context awareness integrated into all commands
3. Consistent flag usage and help text
4. Proper error handling with POSIX exit codes
5. Integration with services layer
6. Comprehensive test coverage
7. Documentation and help text complete

## Service Integration Patterns

### WorktreeService Integration

The WorktreeService provides the core worktree operations that CLI commands SHALL use:

```go
// Service integration pattern for worktree operations
type WorktreeService interface {
    CreateWorktree(ctx context.Context, req *CreateWorktreeRequest) (*WorktreeInfo, error)
    DeleteWorktree(ctx context.Context, req *DeleteWorktreeRequest) error
    ListWorktrees(ctx context.Context, req *ListWorktreesRequest) ([]*WorktreeInfo, error)
    GetWorktreeStatus(ctx context.Context, worktreePath string) (*WorktreeStatus, error)
    ValidateWorktree(ctx context.Context, worktreePath string) error
}

// CLI command integration example
func (cmd *createCmd) runWorktreeCreation(args []string) error {
    // Build request from CLI arguments and flags
    req := &CreateWorktreeRequest{
        ProjectName:  cmd.parseProject(args[0]),
        BranchName:   cmd.parseBranch(args[0]),
        SourceBranch: cmd.sourceFlag,
        ChangeDir:    cmd.cdFlag,
        Force:        cmd.forceFlag,
        Context:      cmd.detectContext(),
    }
    
    // Call service
    result, err := cmd.worktreeService.CreateWorktree(context.Background(), req)
    if err != nil {
        return fmt.Errorf("create worktree failed: %w", err)
    }
    
    // Handle --cd flag output
    if req.ChangeDir {
        fmt.Println(result.Path)
    }
    
    return nil
}
```

### ProjectService Integration

The ProjectService provides project discovery and validation:

```go
// Service integration pattern for project operations
type ProjectService interface {
    DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*ProjectInfo, error)
    ValidateProject(ctx context.Context, projectPath string) error
    ListProjects(ctx context.Context) ([]*ProjectInfo, error)
    GetProjectInfo(ctx context.Context, projectPath string) (*ProjectInfo, error)
}

// CLI command integration example
func (cmd *listCmd) resolveProjectScope() (*ProjectInfo, error) {
    if cmd.allFlag {
        return nil, nil // List all projects
    }
    
    if cmd.projectFlag != "" {
        return cmd.projectService.DiscoverProject(
            context.Background(), 
            cmd.projectFlag, 
            cmd.currentContext,
        )
    }
    
    // Use context to infer project
    return cmd.projectService.DiscoverProject(
        context.Background(), 
        "", 
        cmd.currentContext,
    )
}
```

### NavigationService Integration

The NavigationService provides path resolution for the `cd` command:

```go
// Service integration pattern for navigation operations
type NavigationService interface {
    ResolvePath(ctx context.Context, req *ResolvePathRequest) (*domain.ResolutionResult, error)
    ValidatePath(ctx context.Context, path string) error
    GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error)
}

// CLI command integration example
func (cmd *cdCmd) resolveTarget(target string) (string, error) {
    req := &ResolvePathRequest{
        Target:      target,
        Context:     cmd.currentContext,
        CurrentPath: cmd.originalPath,
    }
    
    result, err := cmd.navigationService.ResolvePath(context.Background(), req)
    if err != nil {
        return "", fmt.Errorf("path resolution failed: %w", err)
    }
    
    return result.ResolvedPath, nil
}
```

## Context-Aware Command Behavior

### Context Detection Integration

All CLI commands SHALL detect context before execution:

```go
// Context detection pattern for CLI commands
func (cmd *BaseCommand) detectContext() (*domain.Context, error) {
    context, err := cmd.contextService.GetCurrentContext()
    if err != nil {
        return nil, fmt.Errorf("context detection failed: %w", err)
    }
    
    // Store for command use
    cmd.currentContext = context
    return context, nil
}

// Context-aware help adaptation
func (cmd *BaseCommand) adaptHelpText() {
    switch cmd.currentContext.Type {
    case domain.ContextProject:
        cmd.Short = fmt.Sprintf("%s (project: %s)", cmd.baseShort, cmd.currentContext.ProjectName)
    case domain.ContextWorktree:
        cmd.Short = fmt.Sprintf("%s (worktree: %s/%s)", cmd.baseShort, cmd.currentContext.ProjectName, cmd.currentContext.BranchName)
    case domain.ContextOutsideGit:
        cmd.Short = fmt.Sprintf("%s (outside git)", cmd.baseShort)
    }
}
```

### Request Adaptation Patterns

Commands SHALL adapt their requests based on detected context:

```go
// Context-aware request adaptation
func (cmd *createCmd) adaptRequest(req *CreateWorktreeRequest) *CreateWorktreeRequest {
    switch cmd.currentContext.Type {
    case domain.ContextProject:
        // Infer project from context if not specified
        if req.ProjectName == "" {
            req.ProjectName = cmd.currentContext.ProjectName
        }
    case domain.ContextWorktree:
        // Infer project from current worktree
        if req.ProjectName == "" {
            req.ProjectName = cmd.currentContext.ProjectName
        }
        // Default source to current branch if not specified
        if req.SourceBranch == "" {
            req.SourceBranch = cmd.currentContext.BranchName
        }
    }
    
    return req
}
```

## Validation and Error Handling

### Command-Level Validation

CLI commands SHALL perform validation before calling services:

```go
// Command validation pattern
func (cmd *createCmd) validateArgs(args []string) error {
    if len(args) != 1 {
        return fmt.Errorf("create requires exactly one argument: <project>/<branch>")
    }
    
    parts := strings.Split(args[0], "/")
    if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
        return fmt.Errorf("invalid format: expected <project>/<branch>, got %s", args[0])
    }
    
    // Validate branch name
    if strings.HasPrefix(parts[1], "-") {
        return fmt.Errorf("branch name cannot start with '-'")
    }
    
    return nil
}
```

### Error Response Formatting

Commands SHALL format service errors for CLI display:

```go
// Error formatting for CLI
func (cmd *BaseCommand) formatServiceError(err error) error {
    switch e := err.(type) {
    case *domain.WorktreeExistsError:
        return fmt.Errorf("worktree already exists at %s. Use --force to override", e.Path)
    case *domain.ProjectNotFoundError:
        return fmt.Errorf("project '%s' not found. Check project name or context", e.Name)
    case *domain.UnsafeOperationError:
        return fmt.Errorf("unsafe operation: %s. Use --force to override", e.Reason)
    default:
        return fmt.Errorf("operation failed: %w", err)
    }
}
```

## Next Steps

1. Implement each command file following the structures above
2. Add comprehensive unit and integration tests
3. Update main.go to use the new command structure
4. Test CLI functionality end-to-end
5. Update documentation with command examples
6. Performance testing and optimization