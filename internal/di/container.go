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
	"io/fs"
	"os"
)

// Container manages all application dependencies
type Container struct {
	// Infrastructure dependencies
	gitClient     domain.GitClient
	config        *config.Config
	fileSystem    fs.FS
	pathValidator domain.PathValidator

	// Infrastructure services
	infraService infrastructure.InfrastructureService

	// Application services
	validationService *services.ValidationService
	discoveryService  *services.DiscoveryService
	operationsService *services.OperationsService
	miseIntegration   *mise.MiseIntegration
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
	c.fileSystem = os.DirFS("/")
	c.pathValidator = validation.NewPathValidator()

	// Create infrastructure services
	c.infraService = infrastructure.NewInfrastructureService(
		c.gitClient,
		c.fileSystem,
		c.pathValidator,
	)

	// Create mise integration
	c.miseIntegration = mise.NewMiseIntegration()
}

// initializeServices sets up application services
func (c *Container) initializeServices() {
	// Create validation service (depends on infrastructure service)
	c.validationService = services.NewValidationService(c.infraService)

	// Create discovery service (depends on infrastructure)
	c.discoveryService = services.NewDiscoveryService(
		c.gitClient,
		c.config,
		c.fileSystem,
		c.pathValidator,
	)

	// Create operations service (depends on validation and discovery)
	c.operationsService = services.NewOperationsService(
		c.gitClient,
		c.discoveryService,
		c.validationService,
		c.miseIntegration,
	)
}

// Getters for services

// GitClient returns the Git client instance
func (c *Container) GitClient() domain.GitClient {
	return c.gitClient
}

// Config returns the configuration instance
func (c *Container) Config() *config.Config {
	return c.config
}

// FileSystem returns the file system instance
func (c *Container) FileSystem() fs.FS {
	return c.fileSystem
}

// PathValidator returns the path validator instance
func (c *Container) PathValidator() domain.PathValidator {
	return c.pathValidator
}

// InfrastructureService returns the infrastructure service instance
func (c *Container) InfrastructureService() infrastructure.InfrastructureService {
	return c.infraService
}

// ValidationService returns the validation service instance
func (c *Container) ValidationService() *services.ValidationService {
	return c.validationService
}

// DiscoveryService returns the discovery service instance
func (c *Container) DiscoveryService() *services.DiscoveryService {
	return c.discoveryService
}

// OperationsService returns the operations service instance
func (c *Container) OperationsService() *services.OperationsService {
	return c.operationsService
}

// MiseIntegration returns the mise integration instance
func (c *Container) MiseIntegration() *mise.MiseIntegration {
	return c.miseIntegration
}
