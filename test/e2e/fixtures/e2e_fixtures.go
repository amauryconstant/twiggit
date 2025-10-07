//go:build e2e
// +build e2e

package fixtures

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e2ehelpers "twiggit/test/e2e/helpers"
	"twiggit/test/helpers"
)

// E2ETestFixture provides comprehensive test setup for E2E tests
type E2ETestFixture struct {
	tempDir      string
	configHelper *e2ehelpers.ConfigHelper
	gitHelper    *helpers.GitTestHelper
	repoHelper   *helpers.RepoTestHelper
	projects     map[string]string
}

// NewE2ETestFixture creates a new E2E test fixture
func NewE2ETestFixture() *E2ETestFixture {
	tempDir := GinkgoT().TempDir()

	return &E2ETestFixture{
		tempDir:      tempDir,
		configHelper: e2ehelpers.NewConfigHelper(),
		gitHelper:    helpers.NewGitTestHelper(&testing.T{}),
		repoHelper:   helpers.NewRepoTestHelper(&testing.T{}),
		projects:     make(map[string]string),
	}
}

// WithConfig sets up the configuration with custom settings
func (f *E2ETestFixture) WithConfig(configFunc func(*e2ehelpers.ConfigHelper)) *E2ETestFixture {
	configFunc(f.configHelper)
	return f
}

// SetupMultiProject creates multiple test projects with different configurations
func (f *E2ETestFixture) SetupMultiProject() *E2ETestFixture {
	// Create projects directory
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	// Project 1: Simple project with main branch
	project1Path := f.repoHelper.SetupTestRepo("project1")
	f.projects["project1"] = project1Path

	// Project 2: Project with feature branches
	project2Path := f.repoHelper.SetupTestRepo("project2")
	f.projects["project2"] = project2Path

	// Add feature branches to project2
	err = f.gitHelper.CreateBranch(project2Path, "feature-a")
	Expect(err).NotTo(HaveOccurred())
	err = f.gitHelper.CreateBranch(project2Path, "feature-b")
	Expect(err).NotTo(HaveOccurred())

	// Project 3: Project with develop as default branch
	project3Path := f.repoHelper.SetupTestRepo("project3")
	f.projects["project3"] = project3Path

	// Rename default branch to develop
	err = f.gitHelper.CreateBranch(project3Path, "develop")
	Expect(err).NotTo(HaveOccurred())

	// Update config to use the projects directory
	f.configHelper.WithProjectsDir(projectsDir)

	return f
}

// SetupSingleProject creates a single test project
func (f *E2ETestFixture) SetupSingleProject(name string) *E2ETestFixture {
	projectPath := f.repoHelper.SetupTestRepo(name)
	f.projects[name] = projectPath

	// Update config to use the temp directory as projects directory
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	f.configHelper.WithProjectsDir(projectsDir)

	return f
}

// GetProjectPath returns the path for a specific project
func (f *E2ETestFixture) GetProjectPath(projectName string) string {
	path, exists := f.projects[projectName]
	Expect(exists).To(BeTrue(), "Project %s not found", projectName)
	return path
}

// GetConfigHelper returns the config helper
func (f *E2ETestFixture) GetConfigHelper() *e2ehelpers.ConfigHelper {
	return f.configHelper
}

// Build builds the configuration and returns the config directory
func (f *E2ETestFixture) Build() string {
	return f.configHelper.Build()
}

// Cleanup cleans up all test resources
func (f *E2ETestFixture) Cleanup() {
	f.repoHelper.Cleanup()
	f.configHelper.Cleanup()
}

// CreateWorktreeSetup creates a project with worktrees for testing
func (f *E2ETestFixture) CreateWorktreeSetup(projectName string) *E2ETestFixture {
	// Create main project
	projectPath := f.SetupSingleProject(projectName).GetProjectPath(projectName)

	// Create worktrees directory
	worktreesDir := filepath.Join(f.tempDir, "worktrees", projectName)
	err := os.MkdirAll(worktreesDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	// Update config
	f.configHelper.WithWorktreesDir(filepath.Join(f.tempDir, "worktrees"))

	// Create some feature branches for worktree testing
	err = f.gitHelper.CreateBranch(projectPath, "feature-1")
	Expect(err).NotTo(HaveOccurred())
	err = f.gitHelper.CreateBranch(projectPath, "feature-2")
	Expect(err).NotTo(HaveOccurred())

	return f
}

// CreateCustomBranchSetup creates a project with custom default branch
func (f *E2ETestFixture) CreateCustomBranchSetup(projectName, defaultBranch string) *E2ETestFixture {
	projectPath := f.SetupSingleProject(projectName).GetProjectPath(projectName)

	// Create custom default branch
	err := f.gitHelper.CreateBranch(projectPath, defaultBranch)
	Expect(err).NotTo(HaveOccurred())

	// Update config to use custom default branch
	f.configHelper.WithDefaultSourceBranch(defaultBranch)

	return f
}

// GetProjects returns all created project names
func (f *E2ETestFixture) GetProjects() []string {
	var names []string
	for name := range f.projects {
		names = append(names, name)
	}
	return names
}
