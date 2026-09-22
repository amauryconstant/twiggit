package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"twiggit/cmd"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
)

func main() {
	slogLevel := slog.LevelInfo
	if os.Getenv("TWIGGIT_DEBUG") != "" {
		slogLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})))

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
		os.Exit(int(cmd.HandleCLIError(err)))
	}

	// Initialize infrastructure services in dependency order
	cliTimeout := time.Duration(config.Git.CLITimeout) * time.Second
	commandExecutor := infrastructure.NewCommandExecutor(cliTimeout)
	goGitClient, err := infrastructure.NewGoGitClient(true)
	if err != nil {
		os.Exit(int(cmd.HandleCLIError(fmt.Errorf("init go-git client: %w", err))))
	}
	cliClient := infrastructure.NewCLIClient(commandExecutor, config.Git.CLITimeout)

	// Create composite GitClient that implements both interfaces
	gitClient := infrastructure.NewCompositeGitClient(goGitClient, cliClient)

	contextDetector, err := infrastructure.NewContextDetector(config)
	if err != nil {
		os.Exit(int(cmd.HandleCLIError(fmt.Errorf("init context detector: %w", err))))
	}
	contextResolver := infrastructure.NewContextResolver(config, goGitClient, cliClient)

	// Initialize application services (contextService first as others depend on it)
	contextService := service.NewContextService(contextDetector, contextResolver, config)
	projectService := service.NewProjectService(gitClient, contextService, config)
	navigationService := service.NewNavigationService(projectService, contextService, config)
	hookRunner := infrastructure.NewHookRunner(commandExecutor, config.Shell.HookTimeout)
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
