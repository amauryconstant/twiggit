//go:build e2e
// +build e2e

package fixtures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/internal/infrastructure"
	e2ehelpers "twiggit/test/e2e/helpers"
	"twiggit/test/helpers"
)

const (
	// File permissions for test files
	FilePermReadWrite = 0644 // read/write for owner, read for others
	FilePermAll       = 0755 // read/write/execute for all (or use git default)
)

// E2ETestFixture provides comprehensive test setup for E2E tests
type E2ETestFixture struct {
	tempDir          string
	configHelper     *e2ehelpers.ConfigHelper
	gitHelper        *helpers.GitTestHelper
	gitExecutor      infrastructure.CommandExecutor
	projects         map[string]*ProjectInfo
	testID           *e2ehelpers.TestIDGenerator
	createdWorktrees []string
	worktreeOwners   map[string]string
	mu               sync.Mutex
}

type ProjectInfo struct {
	Name string
	Path string
}

// CreateWorktreeSetupResult contains branch names created by CreateWorktreeSetup
type CreateWorktreeSetupResult struct {
	Feature1Branch string
	Feature2Branch string
}

// NewE2ETestFixture creates a new E2E test fixture
func NewE2ETestFixture() *E2ETestFixture {
	tempDir := GinkgoT().TempDir()

	return &E2ETestFixture{
		tempDir:          tempDir,
		configHelper:     e2ehelpers.NewConfigHelper().WithTempDir(tempDir),
		gitHelper:        helpers.NewGitTestHelper(&testing.T{}),
		gitExecutor:      infrastructure.NewDefaultCommandExecutor(30 * time.Second),
		projects:         make(map[string]*ProjectInfo),
		testID:           e2ehelpers.NewTestIDGenerator(),
		createdWorktrees: make([]string, 0),
		worktreeOwners:   make(map[string]string),
	}
}

// SetupMultiProject creates multiple test projects with different configurations
func (f *E2ETestFixture) SetupMultiProject() *E2ETestFixture {
	// Create projects directory
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, FilePermAll)
	Expect(err).NotTo(HaveOccurred())

	// Project 1: Simple project with main branch using fixture
	project1Name := f.testID.ProjectNameWithSuffix("1")
	project1Path := filepath.Join(projectsDir, project1Name)
	err = extractRepoFixtureToDir("single-branch", project1Path)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract project 1")
	f.projects[project1Name] = &ProjectInfo{
		Name: project1Name,
		Path: project1Path,
	}

	// Project 2: Project with feature branches using fixture
	project2Name := f.testID.ProjectNameWithSuffix("2")
	project2Path := filepath.Join(projectsDir, project2Name)
	err = extractRepoFixtureToDir("multi-branch", project2Path)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract project 2")
	f.projects[project2Name] = &ProjectInfo{
		Name: project2Name,
		Path: project2Path,
	}

	// Project 3: Project with develop as default branch using fixture
	project3Name := f.testID.ProjectNameWithSuffix("3")
	project3Path := filepath.Join(projectsDir, project3Name)
	err = extractRepoFixtureToDir("single-branch", project3Path)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract project 3")
	f.projects[project3Name] = &ProjectInfo{
		Name: project3Name,
		Path: project3Path,
	}

	// Rename default branch to develop in project3
	err = f.gitHelper.CreateBranch(project3Path, "develop")
	Expect(err).NotTo(HaveOccurred())

	// Update config to use projects and worktrees directories
	worktreesDir := filepath.Join(f.tempDir, "worktrees")
	f.configHelper.WithProjectsDir(projectsDir)
	f.configHelper.WithWorktreesDir(worktreesDir)

	return f
}

// SetupSingleProject creates a single test project
func (f *E2ETestFixture) SetupSingleProject(name string) *E2ETestFixture {
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, FilePermAll)
	Expect(err).NotTo(HaveOccurred())

	projectPath := filepath.Join(projectsDir, name)
	err = extractRepoFixtureToDir("single-branch", projectPath)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract fixture to project directory")

	f.projects[name] = &ProjectInfo{
		Name: name,
		Path: projectPath,
	}

	worktreesDir := filepath.Join(f.tempDir, "worktrees")
	f.configHelper.WithProjectsDir(projectsDir)
	f.configHelper.WithWorktreesDir(worktreesDir)

	return f
}

// GetProjectPath returns the path for a specific project
func (f *E2ETestFixture) GetProjectPath(projectName string) string {
	info, exists := f.projects[projectName]
	Expect(exists).To(BeTrue(), "Project %s not found", projectName)
	return info.Path
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

	// Cleanup worktrees using their owning repos (fixes multi-project bug)
	for worktreePath, repoPath := range f.worktreeOwners {
		if worktreePath != "" {
			err := f.removeWorktreeWithRetry(worktreePath, repoPath)
			if err != nil {
				GinkgoT().Logf("Warning: %v", err)
			}
		}
	}
}

// TrackWorktree registers a worktree with its owning repository
func (f *E2ETestFixture) TrackWorktree(worktreePath, repoPath string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.worktreeOwners[worktreePath] = repoPath
}

// CreateWorktreeSetup creates a project with worktrees for testing
func (f *E2ETestFixture) CreateWorktreeSetup(projectName string) *CreateWorktreeSetupResult {
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, FilePermAll)
	Expect(err).NotTo(HaveOccurred())

	projectPath := filepath.Join(projectsDir, projectName)
	err = extractRepoFixtureToDir("single-branch", projectPath)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract fixture to project directory")

	f.projects[projectName] = &ProjectInfo{
		Name: projectName,
		Path: projectPath,
	}

	worktreesDir := filepath.Join(f.tempDir, "worktrees")

	// Update in-memory config paths
	f.configHelper.WithProjectsDir(projectsDir)
	f.configHelper.WithWorktreesDir(worktreesDir)

	// Generate branch names once to ensure consistency across the method
	// TestIDGenerator generates a new random ID for each BranchName() call,
	// so we must call it once and reuse the result
	feature1Branch := f.testID.BranchName("feature-1")
	feature2Branch := f.testID.BranchName("feature-2")

	worktree1Path := filepath.Join(worktreesDir, projectName, feature1Branch)
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", feature1Branch, worktree1Path,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to create worktree for feature-1")
	f.createdWorktrees = append(f.createdWorktrees, worktree1Path)
	f.TrackWorktree(worktree1Path, projectPath)

	worktree2Path := filepath.Join(worktreesDir, projectName, feature2Branch)
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", feature2Branch, worktree2Path,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to create worktree for feature-2")
	f.createdWorktrees = append(f.createdWorktrees, worktree2Path)
	f.TrackWorktree(worktree2Path, projectPath)

	// Rebuild config file to ensure worktrees directory is configured
	f.configHelper.Build()

	return &CreateWorktreeSetupResult{
		Feature1Branch: feature1Branch,
		Feature2Branch: feature2Branch,
	}
}

// CreateMergedWorktreeSetup creates a project with a worktree whose branch is already merged to main
// This is used for prune tests that need worktrees which will actually be deleted
func (f *E2ETestFixture) CreateMergedWorktreeSetup(projectName string) *CreateWorktreeSetupResult {
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, FilePermAll)
	Expect(err).NotTo(HaveOccurred())

	projectPath := filepath.Join(projectsDir, projectName)
	err = extractRepoFixtureToDir("single-branch", projectPath)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract fixture to project directory")

	f.projects[projectName] = &ProjectInfo{
		Name: projectName,
		Path: projectPath,
	}

	worktreesDir := filepath.Join(f.tempDir, "worktrees")

	// Update in-memory config paths
	f.configHelper.WithProjectsDir(projectsDir)
	f.configHelper.WithWorktreesDir(worktreesDir)

	// Generate branch names
	feature1Branch := f.testID.BranchName("merged-feature-1")
	feature2Branch := f.testID.BranchName("merged-feature-2")

	// Create worktree 1 with feature branch
	worktree1Path := filepath.Join(worktreesDir, projectName, feature1Branch)
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", feature1Branch, worktree1Path,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to create worktree for merged-feature-1")
	f.createdWorktrees = append(f.createdWorktrees, worktree1Path)
	f.TrackWorktree(worktree1Path, projectPath)

	// Create worktree 2 with feature branch
	worktree2Path := filepath.Join(worktreesDir, projectName, feature2Branch)
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", feature2Branch, worktree2Path,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to create worktree for merged-feature-2")
	f.createdWorktrees = append(f.createdWorktrees, worktree2Path)
	f.TrackWorktree(worktree2Path, projectPath)

	// Merge feature branches into main so they're "merged" and can be pruned
	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "merge", feature1Branch, "--no-edit",
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to merge feature-1 into main")

	_, err = f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "merge", feature2Branch, "--no-edit",
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to merge feature-2 into main")

	// Rebuild config file
	f.configHelper.Build()

	return &CreateWorktreeSetupResult{
		Feature1Branch: feature1Branch,
		Feature2Branch: feature2Branch,
	}
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
	for name, info := range f.projects {
		exists := ""
		if _, err := os.Stat(info.Path); err == nil {
			exists = "✓"
		} else {
			exists = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s: %s [%s]\n", name, info.Path, exists))
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

// ValidateCleanup verifies that cleanup was successful
func (f *E2ETestFixture) ValidateCleanup() error {
	var validationErrors []error

	for _, wt := range f.createdWorktrees {
		if _, err := os.Stat(wt); err == nil {
			validationErrors = append(validationErrors, fmt.Errorf("worktree %s still exists after cleanup", wt))
		}
	}

	// Note: Projects are not validated here because they're cleaned up by Ginkgo's TempDir
	// Only worktrees are validated as they require explicit git worktree remove

	if len(validationErrors) > 0 {
		return fmt.Errorf("cleanup validation failed: %v", validationErrors)
	}

	return nil
}
