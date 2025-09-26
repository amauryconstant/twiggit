// Package di provides dependency injection container for managing application services
package di

import (
	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/infrastructure/mise"
	"github.com/amaury/twiggit/internal/infrastructure/validation"
	"github.com/amaury/twiggit/internal/services"
)

// Container manages all application dependencies
type Container struct {
	// Infrastructure dependencies
	gitClient         infrastructure.GitClient
	config            *config.Config
	fileSystem        infrastructure.FileSystem
	pathValidator     infrastructure.PathValidator
	fileSystemChecker domain.FileSystemChecker

	// Application services
	validationService        *services.ValidationService
	discoveryService         *services.DiscoveryService
	worktreeCreator          *services.WorktreeCreator
	worktreeRemover          *services.WorktreeRemover
	currentDirectoryDetector *services.CurrentDirectoryDetector
	miseIntegration          infrastructure.MiseIntegration
}

// NewContainer creates and initializes the dependency container
func NewContainer(cfg *config.Config) *Container {
	container := &Container{
		config: cfg,
	}

	container.initializeInfrastructure()
	container.initializeServices()

	return container
}

// initializeInfrastructure sets up infrastructure dependencies
func (c *Container) initializeInfrastructure() {
	// Create core infrastructure dependencies
	c.gitClient = git.NewClient()
	c.fileSystem = infrastructure.NewOSFileSystem()
	c.pathValidator = validation.NewPathValidator()
	c.fileSystemChecker = c.fileSystem // Use the same instance for both dependencies

	// Create mise integration
	c.miseIntegration = mise.NewMiseIntegration(mise.WithFileSystem(c.fileSystem))
}

// initializeServices sets up application services
func (c *Container) initializeServices() {
	// Create validation service (depends on filesystem)
	c.validationService = services.NewValidationService(c.fileSystem)

	// Create discovery service (depends on infrastructure)
	c.discoveryService = services.NewDiscoveryService(
		c.gitClient,
		c.config,
		c.fileSystem,
	)

	// Create specialized worktree services
	c.worktreeCreator = services.NewWorktreeCreator(
		c.gitClient,
		c.validationService,
		c.miseIntegration,
	)

	c.worktreeRemover = services.NewWorktreeRemover(
		c.gitClient,
	)

	c.currentDirectoryDetector = services.NewCurrentDirectoryDetector(
		c.gitClient,
	)
}

// Getters for services

// GitClient returns the Git client instance
func (c *Container) GitClient() infrastructure.GitClient {
	return c.gitClient
}

// Config returns the configuration instance
func (c *Container) Config() *config.Config {
	return c.config
}

// FileSystem returns the file system instance
func (c *Container) FileSystem() infrastructure.FileSystem {
	return c.fileSystem
}

// PathValidator returns the path validator instance
func (c *Container) PathValidator() infrastructure.PathValidator {
	return c.pathValidator
}

// ValidationService returns the validation service instance
func (c *Container) ValidationService() *services.ValidationService {
	return c.validationService
}

// DiscoveryService returns the discovery service instance
func (c *Container) DiscoveryService() *services.DiscoveryService {
	return c.discoveryService
}

// WorktreeCreator returns the worktree creator service instance
func (c *Container) WorktreeCreator() *services.WorktreeCreator {
	return c.worktreeCreator
}

// WorktreeRemover returns the worktree remover service instance
func (c *Container) WorktreeRemover() *services.WorktreeRemover {
	return c.worktreeRemover
}

// CurrentDirectoryDetector returns the current directory detector service instance
func (c *Container) CurrentDirectoryDetector() *services.CurrentDirectoryDetector {
	return c.currentDirectoryDetector
}

// MiseIntegration returns the mise integration instance
func (c *Container) MiseIntegration() infrastructure.MiseIntegration {
	return c.miseIntegration
}

// FileSystemChecker returns the filesystem checker instance
func (c *Container) FileSystemChecker() domain.FileSystemChecker {
	return c.fileSystemChecker
}
