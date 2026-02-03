//go:build e2e
// +build e2e

package fixtures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/internal/infrastructure"
	e2ehelpers "twiggit/test/e2e/helpers"
	"twiggit/test/helpers"
)

// E2ETestFixture provides comprehensive test setup for E2E tests
type E2ETestFixture struct {
	tempDir          string
	configHelper     *e2ehelpers.ConfigHelper
	gitHelper        *helpers.GitTestHelper
	repoHelper       *helpers.RepoTestHelper
	gitExecutor      infrastructure.CommandExecutor
	projects         map[string]string
	testID           *e2ehelpers.TestIDGenerator
	createdWorktrees []string
}

// NewE2ETestFixture creates a new E2E test fixture
func NewE2ETestFixture() *E2ETestFixture {
	tempDir := GinkgoT().TempDir()

	return &E2ETestFixture{
		tempDir:          tempDir,
		configHelper:     e2ehelpers.NewConfigHelper(),
		gitHelper:        helpers.NewGitTestHelper(&testing.T{}),
		repoHelper:       helpers.NewRepoTestHelper(&testing.T{}),
		gitExecutor:      infrastructure.NewDefaultCommandExecutor(30 * time.Second),
		projects:         make(map[string]string),
		testID:           e2ehelpers.NewTestIDGenerator(),
		createdWorktrees: make([]string, 0),
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
	project1Path := f.repoHelper.SetupTestRepo(f.testID.ProjectNameWithSuffix("1"))
	f.projects[f.testID.ProjectNameWithSuffix("1")] = project1Path

	// Project 2: Project with feature branches
	project2Path := f.repoHelper.SetupTestRepo(f.testID.ProjectNameWithSuffix("2"))
	f.projects[f.testID.ProjectNameWithSuffix("2")] = project2Path

	// Add feature branches to project2
	err = f.gitHelper.CreateBranch(project2Path, f.testID.BranchName("feature-a"))
	Expect(err).NotTo(HaveOccurred())
	err = f.gitHelper.CreateBranch(project2Path, f.testID.BranchName("feature-b"))
	Expect(err).NotTo(HaveOccurred())

	// Project 3: Project with develop as default branch
	project3Path := f.repoHelper.SetupTestRepo(f.testID.ProjectNameWithSuffix("3"))
	f.projects[f.testID.ProjectNameWithSuffix("3")] = project3Path

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

	projectsDir := filepath.Dir(projectPath)
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

// GetTestID returns the test ID generator
func (f *E2ETestFixture) GetTestID() *e2ehelpers.TestIDGenerator {
	return f.testID
}

// GetGitHelper returns the git helper
func (f *E2ETestFixture) GetGitHelper() *helpers.GitTestHelper {
	return f.gitHelper
}

// GetTempDir returns the temporary directory
func (f *E2ETestFixture) GetTempDir() string {
	return f.tempDir
}

// Build builds the configuration and returns the config directory
func (f *E2ETestFixture) Build() string {
	return f.configHelper.Build()
}

// removeWorktreeWithRetry removes a worktree with retry logic and force flag
func (f *E2ETestFixture) removeWorktreeWithRetry(worktreePath, mainRepoPath string) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		result, err := f.gitExecutor.Execute(
			context.Background(),
			mainRepoPath,
			"git", "worktree", "remove", "--force", worktreePath,
		)

		if err == nil && result != nil && result.ExitCode == 0 {
			return nil
		}

		if i < maxRetries-1 {
			GinkgoT().Logf("Retry %d/%d: removing worktree %s", i+1, maxRetries, worktreePath)
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		}
	}
	return fmt.Errorf("failed to remove worktree after %d attempts: %s", maxRetries, worktreePath)
}

// Cleanup cleans up all test resources
func (f *E2ETestFixture) Cleanup() {
	if f == nil {
		return
	}

	mainRepoPath := ""
	if len(f.projects) > 0 {
		for _, repoPath := range f.projects {
			if repoPath != "" {
				mainRepoPath = repoPath
				break
			}
		}
	}

	if mainRepoPath == "" || len(f.createdWorktrees) == 0 {
		if f.repoHelper != nil {
			f.repoHelper.Cleanup()
		}
		if f.configHelper != nil {
			f.configHelper.Cleanup()
		}
		return
	}

	for _, wt := range f.createdWorktrees {
		if wt != "" {
			err := f.removeWorktreeWithRetry(wt, mainRepoPath)
			if err != nil {
				GinkgoT().Logf("Warning: %v", err)
			}
		}
	}

	if f.repoHelper != nil {
		f.repoHelper.Cleanup()
	}
	if f.configHelper != nil {
		f.configHelper.Cleanup()
	}
}

// CreateWorktreeSetup creates a project with worktrees for testing
func (f *E2ETestFixture) CreateWorktreeSetup(projectName string) *E2ETestFixture {
	projectPath := f.SetupSingleProject(projectName).GetProjectPath(projectName)

	worktreesDir := filepath.Join(f.tempDir, "worktrees", projectName)
	err := os.MkdirAll(worktreesDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	f.configHelper.WithWorktreesDir(filepath.Join(f.tempDir, "worktrees"))

	worktree1Path := filepath.Join(worktreesDir, f.testID.BranchName("feature-1"))
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", f.testID.BranchName("feature-1"), worktree1Path,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to create worktree for feature-1")
	f.createdWorktrees = append(f.createdWorktrees, worktree1Path)

	worktree2Path := filepath.Join(worktreesDir, f.testID.BranchName("feature-2"))
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", f.testID.BranchName("feature-2"), worktree2Path,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to create worktree for feature-2")
	f.createdWorktrees = append(f.createdWorktrees, worktree2Path)

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

// GetCreatedWorktrees returns all created worktree paths
func (f *E2ETestFixture) GetCreatedWorktrees() []string {
	return f.createdWorktrees
}

// Inspect returns a detailed snapshot of the fixture state for debugging
func (f *E2ETestFixture) Inspect() string {
	var sb strings.Builder

	sb.WriteString("=== E2ETestFixture State ===\n")
	sb.WriteString(fmt.Sprintf("TempDir: %s\n", f.tempDir))

	sb.WriteString("\n=== Projects ===\n")
	for name, path := range f.projects {
		exists := ""
		if _, err := os.Stat(path); err == nil {
			exists = "✓"
		} else {
			exists = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s: %s [%s]\n", name, path, exists))
	}

	sb.WriteString("\n=== Worktrees ===\n")
	for i, wt := range f.createdWorktrees {
		exists := ""
		if _, err := os.Stat(wt); err == nil {
			exists = "✓"
		} else {
			exists = "✗"
		}
		sb.WriteString(fmt.Sprintf("  [%d] %s [%s]\n", i, wt, exists))
	}

	sb.WriteString("\n=== Config ===\n")
	if f.configHelper != nil {
		sb.WriteString(fmt.Sprintf("  ConfigPath: %s\n", f.configHelper.GetConfigPath()))
		sb.WriteString(fmt.Sprintf("  ProjectsDir: %s\n", f.configHelper.GetProjectsDir()))
		sb.WriteString(fmt.Sprintf("  WorktreesDir: %s\n", f.configHelper.GetWorktreesDir()))
	}

	return sb.String()
}

// WithDebugLogging enables debug logging for this fixture
func (f *E2ETestFixture) WithDebugLogging(enabled bool) *E2ETestFixture {
	if enabled {
		GinkgoT().Log("Debug mode enabled\n", f.Inspect())
	}
	return f
}

// ValidateCleanup verifies that cleanup was successful
func (f *E2ETestFixture) ValidateCleanup() error {
	var validationErrors []error

	for _, wt := range f.createdWorktrees {
		if _, err := os.Stat(wt); err == nil {
			validationErrors = append(validationErrors, fmt.Errorf("worktree %s still exists after cleanup", wt))
		}
	}

	for _, repoPath := range f.projects {
		if _, err := os.Stat(repoPath); err == nil {
			validationErrors = append(validationErrors, fmt.Errorf("project %s still exists after cleanup", repoPath))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("cleanup validation failed: %v", validationErrors)
	}

	return nil
}
