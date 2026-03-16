package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"twiggit/cmd"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
)

func main() {
	// Set up panic recovery for graceful handling of unexpected errors
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Internal error: %v\n", r)
			if os.Getenv("TWIGGIT_DEBUG") != "" {
				fmt.Fprintln(os.Stderr, "\nStack trace:")
				debug.PrintStack()
			}
			os.Exit(1)
		}
	}()

	// Initialize and load configuration
	configManager := infrastructure.NewConfigManager()
	config, err := configManager.Load()
	if err != nil {
		// Use functional error handling instead of panic
		cmd.HandleCLIError(err)
		os.Exit(1)
	}

	// Initialize infrastructure services in dependency order
	commandExecutor := infrastructure.NewDefaultCommandExecutor(30 * 1000000000) // 30 seconds
	goGitClient := infrastructure.NewGoGitClient(true)
	cliClient := infrastructure.NewCLIClient(commandExecutor)

	// Create composite GitClient that implements both interfaces
	gitClient := infrastructure.NewCompositeGitClient(goGitClient, cliClient)

	contextDetector := infrastructure.NewContextDetector(config)
	contextResolver := infrastructure.NewContextResolver(config, gitClient)

	// Initialize application services (contextService first as others depend on it)
	contextService := service.NewContextService(contextDetector, contextResolver, config)
	projectService := service.NewProjectService(gitClient, contextService, config)
	navigationService := service.NewNavigationService(projectService, contextService, config)
	hookRunner := infrastructure.NewHookRunner(commandExecutor)
	worktreeService := service.NewWorktreeService(gitClient, projectService, config, hookRunner)
	shellInfra := infrastructure.NewShellInfrastructure()
	shellService := service.NewShellService(shellInfra, config)

	// Create command configuration
	commandConfig := &cmd.CommandConfig{
		Config: config,
		Services: &cmd.ServiceContainer{
			ContextService:    contextService,
			ProjectService:    projectService,
			NavigationService: navigationService,
			WorktreeService:   worktreeService,
			ShellService:      shellService,
		},
	}

	// Use NewRootCommand to create a properly configured command tree
	rootCmd := cmd.NewRootCommand(commandConfig)

	// Execute CLI with functional error handling
	if err := rootCmd.Execute(); err != nil {
		// Pass rootCmd to respect quiet mode for hint suppression
		exitCode := cmd.HandleCLIErrorWithCommand(rootCmd, err)
		os.Exit(int(exitCode))
	}
}
