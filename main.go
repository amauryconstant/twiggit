package main

import (
	"twiggit/cmd"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

func main() {
	// Initialize and load configuration
	configManager := infrastructure.NewConfigManager()
	config, err := configManager.Load()
	if err != nil {
		// For now, panic on config errors - this will be improved in CLI phase
		panic(err)
	}

	// Set global configuration for simple access
	domain.SetGlobalConfig(config)

	// Execute CLI
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
