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

	configManager := infrastructure.NewConfigManager()
	config, err := configManager.Load()
	if err != nil {
		os.Exit(int(cmd.HandleCLIError(err)))
	}

	cliTimeout := time.Duration(config.Git.CLITimeout) * time.Second
	commandExecutor := infrastructure.NewCommandExecutor(cliTimeout)
	goGitClient, err := infrastructure.NewGoGitClient(true)
	if err != nil {
		os.Exit(int(cmd.HandleCLIError(fmt.Errorf("init go-git client: %w", err))))
	}
	cliClient := infrastructure.NewCLIClient(commandExecutor, config.Git.CLITimeout)

	contextDetector, err := infrastructure.NewContextDetector(config)
	if err != nil {
		os.Exit(int(cmd.HandleCLIError(fmt.Errorf("init context detector: %w", err))))
	}
	contextResolver := infrastructure.NewContextResolver(config, goGitClient, cliClient)
	repoFinder := infrastructure.NewRepoFinder(goGitClient)

	contextService := service.NewContextService(contextDetector, contextResolver)
	projectService := service.NewProjectService(goGitClient, cliClient, repoFinder, config)
	navigationService := service.NewNavigationService(projectService, contextService, config)
	hookRunner := infrastructure.NewHookRunner(commandExecutor, config.Shell.HookTimeout)
	worktreeService := service.NewWorktreeService(goGitClient, cliClient, projectService, config, hookRunner)
	shellInfra := infrastructure.NewShellInfrastructure()
	shellService := service.NewShellService(shellInfra, config)

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

	rootCmd := cmd.NewRootCommand(commandConfig)

	if err := rootCmd.Execute(); err != nil {
		exitCode := cmd.HandleCLIErrorWithCommand(rootCmd, err)
		os.Exit(int(exitCode))
	}
}
