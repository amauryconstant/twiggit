package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// GitTestSuite provides common git repository setup for tests
type GitTestSuite struct {
	suite.Suite
	Repo *GitRepo
}

// SetupTest creates a new git repository for each test
func (s *GitTestSuite) SetupTest() {
	s.Repo = NewGitRepo(s.T(), "twiggit-test-*")
}

// TearDownTest cleans up the git repository after each test
func (s *GitTestSuite) TearDownTest() {
	if s.Repo != nil {
		s.Repo.Cleanup()
	}
}

// TempDir creates a temporary directory with automatic cleanup
func (s *GitTestSuite) TempDir(pattern string) (string, func()) {
	return TempDir(s.T(), pattern)
}

// MustMkdirAll creates directories with proper error handling
func (s *GitTestSuite) MustMkdirAll(path string, perm uint32) {
	MustMkdirAll(s.T(), path, 0755)
}

// IntegrationTestSuite provides setup for integration tests with workspace structure
type IntegrationTestSuite struct {
	suite.Suite
	WorkspaceDir string
	MainRepo     *IntegrationRepo
	cleanup      func()
}

// IntegrationRepo wraps GitRepo with additional integration test helpers
type IntegrationRepo struct {
	*GitRepo
	TempDir      string
	cleanupFuncs []func()
}

// NewIntegrationRepo creates a repository setup for integration testing
func NewIntegrationRepo(t *testing.T) *IntegrationRepo {
	tempDir, cleanup := TempDir(t, "twiggit-integration-*")

	// Create subdirectory for the actual repo
	repoDir := tempDir + "/test-repo"
	MustMkdirAll(t, repoDir, 0755)

	// Create git repo and move it to our structure
	repo := NewGitRepo(t, "twiggit-integration-*")

	// Copy the repo content to our desired location
	err := os.Rename(repo.Path, repoDir)
	if err != nil {
		t.Fatalf("Failed to move repo: %v", err)
	}

	repo.Path = repoDir

	integrationRepo := &IntegrationRepo{
		GitRepo:      repo,
		TempDir:      tempDir,
		cleanupFuncs: []func(){repo.Cleanup, cleanup},
	}

	return integrationRepo
}

// Cleanup cleans up all resources
func (r *IntegrationRepo) Cleanup() {
	for _, cleanup := range r.cleanupFuncs {
		if cleanup != nil {
			cleanup()
		}
	}
}

// RepoDir returns the repository directory for backward compatibility
func (r *IntegrationRepo) RepoDir() string {
	return r.Path
}

// SetupTest creates workspace structure for integration tests
func (s *IntegrationTestSuite) SetupTest() {
	tempDir, cleanup := TempDir(s.T(), "twiggit-integration-*")
	s.WorkspaceDir = tempDir
	s.cleanup = cleanup

	// Create main repository in workspace
	s.MainRepo = NewIntegrationRepo(s.T())
}

// TearDownTest cleans up integration test resources
func (s *IntegrationTestSuite) TearDownTest() {
	if s.MainRepo != nil {
		s.MainRepo.Cleanup()
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

// CreateProject creates a new project repository in the workspace
func (s *IntegrationTestSuite) CreateProject(name string) *GitRepo {
	repo := NewGitRepo(s.T(), "project-*")
	return repo
}
