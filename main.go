package main

import (
	"os"

	"twiggit/cmd"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/internal/infrastructure/shell"
	"twiggit/internal/service"
	"twiggit/internal/services"
)

func main() {
	// Initialize and load configuration
	configManager := infrastructure.NewConfigManager()
	config, err := configManager.Load()
	if err != nil {
		// Use functional error handling instead of panic
		cmd.HandleCLIError(err)
		os.Exit(1)
	}

	// Set global configuration for simple access
	domain.SetGlobalConfig(config)

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
	projectService := services.NewProjectService(gitClient, contextService, config)
	navigationService := services.NewNavigationService(projectService, contextService, config)
	worktreeService := services.NewWorktreeService(gitClient, projectService, config)
	shellInfra := shell.NewShellService()
	shellService := services.NewShellService(shellInfra, config)

	// Create command configuration
	commandConfig := &cmd.CommandConfig{
		Config: config,
		Services: &cmd.ServiceContainer{
			ContextService:    contextService,
			ProjectService:    projectService,
			NavigationService: navigationService,
			WorktreeService:   worktreeService,
			ShellService:      shellService,
			GitClient:         gitClient,
		},
	}

	// Use NewRootCommand to create a properly configured command tree
	rootCmd := cmd.NewRootCommand(commandConfig)

	// Execute CLI with functional error handling
	if err := rootCmd.Execute(); err != nil {
		exitCode := cmd.HandleCLIError(err)
		os.Exit(int(exitCode))
	}
}
