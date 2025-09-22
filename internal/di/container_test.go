package di

import (
	"testing"

	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/infrastructure/mise"
	"github.com/amaury/twiggit/internal/infrastructure/validation"
	"github.com/amaury/twiggit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainer(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	// Create container
	container := NewContainer(cfg)

	// Verify container is not nil
	require.NotNil(t, container, "Container should not be nil")

	// Verify config is set correctly
	assert.Equal(t, cfg, container.Config(), "Config should be the same instance")
}

func TestContainer_InfrastructureDependencies(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test infrastructure dependencies
	assert.NotNil(t, container.GitClient(), "GitClient should not be nil")
	assert.NotNil(t, container.FileSystem(), "FileSystem should not be nil")
	assert.NotNil(t, container.PathValidator(), "PathValidator should not be nil")

	// Verify types
	_, ok := container.GitClient().(*git.Client)
	assert.True(t, ok, "GitClient should be of type *git.Client")

	// FileSystem is already of type fs.FS, no type assertion needed

	_, ok = container.PathValidator().(*validation.PathValidatorImpl)
	assert.True(t, ok, "PathValidator should be of type *validation.PathValidatorImpl")
}

func TestContainer_InfrastructureService(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test infrastructure service
	infraService := container.InfrastructureService()
	require.NotNil(t, infraService, "InfrastructureService should not be nil")

	// Verify type
	_, ok := infraService.(*infrastructure.InfrastructureServiceImpl)
	assert.True(t, ok, "InfrastructureService should be of type *infrastructure.InfrastructureServiceImpl")
}

func TestContainer_ApplicationServices(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test validation service
	validationService := container.ValidationService()
	require.NotNil(t, validationService, "ValidationService should not be nil")
	assert.IsType(t, &services.ValidationService{}, validationService, "ValidationService should be of type *services.ValidationService")

	// Test discovery service
	discoveryService := container.DiscoveryService()
	require.NotNil(t, discoveryService, "DiscoveryService should not be nil")
	assert.IsType(t, &services.DiscoveryService{}, discoveryService, "DiscoveryService should be of type *services.DiscoveryService")

	// Test worktree services
	worktreeCreator := container.WorktreeCreator()
	require.NotNil(t, worktreeCreator, "WorktreeCreator should not be nil")
	assert.IsType(t, &services.WorktreeCreator{}, worktreeCreator, "WorktreeCreator should be of type *services.WorktreeCreator")

	worktreeRemover := container.WorktreeRemover()
	require.NotNil(t, worktreeRemover, "WorktreeRemover should not be nil")
	assert.IsType(t, &services.WorktreeRemover{}, worktreeRemover, "WorktreeRemover should be of type *services.WorktreeRemover")

	currentDirectoryDetector := container.CurrentDirectoryDetector()
	require.NotNil(t, currentDirectoryDetector, "CurrentDirectoryDetector should not be nil")
	assert.IsType(t, &services.CurrentDirectoryDetector{}, currentDirectoryDetector, "CurrentDirectoryDetector should be of type *services.CurrentDirectoryDetector")
}

func TestContainer_MiseIntegration(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test mise integration
	miseIntegration := container.MiseIntegration()
	require.NotNil(t, miseIntegration, "MiseIntegration should not be nil")

	// Verify type
	assert.IsType(t, &mise.MiseIntegration{}, miseIntegration, "MiseIntegration should be of type *mise.MiseIntegration")
}

func TestContainer_ServiceDependencies(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test that services have the correct dependencies
	validationService := container.ValidationService()
	infraService := container.InfrastructureService()

	// Validation service should depend on infrastructure service
	// This is a bit tricky to test without reflection, but we can verify
	// that the services are properly initialized
	assert.NotNil(t, validationService, "ValidationService should be initialized")
	assert.NotNil(t, infraService, "InfrastructureService should be initialized")

	// Discovery service should have the correct dependencies
	discoveryService := container.DiscoveryService()
	assert.NotNil(t, discoveryService, "DiscoveryService should be initialized")

	// Worktree services should have the correct dependencies
	worktreeCreator := container.WorktreeCreator()
	assert.NotNil(t, worktreeCreator, "WorktreeCreator should be initialized")

	worktreeRemover := container.WorktreeRemover()
	assert.NotNil(t, worktreeRemover, "WorktreeRemover should be initialized")

	currentDirectoryDetector := container.CurrentDirectoryDetector()
	assert.NotNil(t, currentDirectoryDetector, "CurrentDirectoryDetector should be initialized")
}

func TestContainer_SingletonBehavior(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test that multiple calls to getters return the same instance (singleton)
	gitClient1 := container.GitClient()
	gitClient2 := container.GitClient()
	assert.Same(t, gitClient1, gitClient2, "GitClient should be a singleton")

	config1 := container.Config()
	config2 := container.Config()
	assert.Same(t, config1, config2, "Config should be a singleton")

	fileSystem1 := container.FileSystem()
	fileSystem2 := container.FileSystem()
	assert.Equal(t, fileSystem1, fileSystem2, "FileSystem should be a singleton")

	pathValidator1 := container.PathValidator()
	pathValidator2 := container.PathValidator()
	assert.Same(t, pathValidator1, pathValidator2, "PathValidator should be a singleton")

	infraService1 := container.InfrastructureService()
	infraService2 := container.InfrastructureService()
	assert.Same(t, infraService1, infraService2, "InfrastructureService should be a singleton")

	validationService1 := container.ValidationService()
	validationService2 := container.ValidationService()
	assert.Same(t, validationService1, validationService2, "ValidationService should be a singleton")

	discoveryService1 := container.DiscoveryService()
	discoveryService2 := container.DiscoveryService()
	assert.Same(t, discoveryService1, discoveryService2, "DiscoveryService should be a singleton")

	worktreeCreator1 := container.WorktreeCreator()
	worktreeCreator2 := container.WorktreeCreator()
	assert.Same(t, worktreeCreator1, worktreeCreator2, "WorktreeCreator should be a singleton")

	worktreeRemover1 := container.WorktreeRemover()
	worktreeRemover2 := container.WorktreeRemover()
	assert.Same(t, worktreeRemover1, worktreeRemover2, "WorktreeRemover should be a singleton")

	currentDirectoryDetector1 := container.CurrentDirectoryDetector()
	currentDirectoryDetector2 := container.CurrentDirectoryDetector()
	assert.Same(t, currentDirectoryDetector1, currentDirectoryDetector2, "CurrentDirectoryDetector should be a singleton")

	miseIntegration1 := container.MiseIntegration()
	miseIntegration2 := container.MiseIntegration()
	assert.Same(t, miseIntegration1, miseIntegration2, "MiseIntegration should be a singleton")
}

func TestContainer_WithNilConfig(t *testing.T) {
	// Test that container handles nil config gracefully
	container := NewContainer(nil)

	// Container should still be created
	require.NotNil(t, container, "Container should be created even with nil config")

	// Config should be nil
	assert.Nil(t, container.Config(), "Config should be nil when passed nil")

	// Other services should still be initialized
	assert.NotNil(t, container.GitClient(), "GitClient should still be initialized")
	assert.NotNil(t, container.FileSystem(), "FileSystem should still be initialized")
	assert.NotNil(t, container.PathValidator(), "PathValidator should still be initialized")
}

func TestContainer_Interfaces(t *testing.T) {
	cfg := &config.Config{
		ProjectsPath:   "/test/projects",
		WorkspacesPath: "/test/workspaces",
		Workspace:      "/test/workspaces",
	}

	container := NewContainer(cfg)

	// Test that all services implement their expected interfaces
	// GitClient is already of type domain.GitClient, no type assertion needed
	assert.NotNil(t, container.GitClient(), "GitClient should implement domain.GitClient")

	// PathValidator is already of type domain.PathValidator, no type assertion needed
	assert.NotNil(t, container.PathValidator(), "PathValidator should implement domain.PathValidator")

	// InfrastructureService is already an interface, so we just check it's not nil
	assert.NotNil(t, container.InfrastructureService(), "InfrastructureService should not be nil")
}
