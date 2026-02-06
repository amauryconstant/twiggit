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

	git "github.com/go-git/go-git/v5"
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
	repoHelper       *helpers.RepoTestHelper
	gitExecutor      infrastructure.CommandExecutor
	projects         map[string]*ProjectInfo
	testID           *e2ehelpers.TestIDGenerator
	createdWorktrees []string
	worktreeOwners   map[string]string
	mu               sync.Mutex
}

type ProjectInfo struct {
	Name    string
	Path    string
	Fixture *RepoFixture
}

// NewE2ETestFixture creates a new E2E test fixture
func NewE2ETestFixture() *E2ETestFixture {
	tempDir := GinkgoT().TempDir()

	return &E2ETestFixture{
		tempDir:          tempDir,
		configHelper:     e2ehelpers.NewConfigHelper().WithTempDir(tempDir),
		gitHelper:        helpers.NewGitTestHelper(&testing.T{}),
		repoHelper:       helpers.NewRepoTestHelper(&testing.T{}),
		gitExecutor:      infrastructure.NewDefaultCommandExecutor(30 * time.Second),
		projects:         make(map[string]*ProjectInfo),
		testID:           e2ehelpers.NewTestIDGenerator(),
		createdWorktrees: make([]string, 0),
		worktreeOwners:   make(map[string]string),
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
	err := os.MkdirAll(projectsDir, FilePermAll)
	Expect(err).NotTo(HaveOccurred())

	// Project 1: Simple project with main branch using fixture
	project1Name := f.testID.ProjectNameWithSuffix("1")
	project1Path := filepath.Join(projectsDir, project1Name)
	err = extractRepoFixtureToDir("single-branch", project1Path)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract project 1")
	f.projects[project1Name] = &ProjectInfo{
		Name:    project1Name,
		Path:    project1Path,
		Fixture: nil,
	}

	// Project 2: Project with feature branches using fixture
	project2Name := f.testID.ProjectNameWithSuffix("2")
	project2Path := filepath.Join(projectsDir, project2Name)
	err = extractRepoFixtureToDir("multi-branch", project2Path)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract project 2")
	f.projects[project2Name] = &ProjectInfo{
		Name:    project2Name,
		Path:    project2Path,
		Fixture: nil,
	}

	// Project 3: Project with develop as default branch using fixture
	project3Name := f.testID.ProjectNameWithSuffix("3")
	project3Path := filepath.Join(projectsDir, project3Name)
	err = extractRepoFixtureToDir("single-branch", project3Path)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract project 3")
	f.projects[project3Name] = &ProjectInfo{
		Name:    project3Name,
		Path:    project3Path,
		Fixture: nil,
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
		Name:    name,
		Path:    projectPath,
		Fixture: nil,
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

	if f.repoHelper != nil {
		f.repoHelper.Cleanup()
	}
	if f.configHelper != nil {
		f.configHelper.Cleanup()
	}
}

// TrackWorktree registers a worktree with its owning repository
func (f *E2ETestFixture) TrackWorktree(worktreePath, repoPath string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.worktreeOwners[worktreePath] = repoPath
}

// CreateWorktreeSetup creates a project with worktrees for testing
func (f *E2ETestFixture) CreateWorktreeSetup(projectName string) *E2ETestFixture {
	projectsDir := filepath.Join(f.tempDir, "projects")
	err := os.MkdirAll(projectsDir, FilePermAll)
	Expect(err).NotTo(HaveOccurred())

	projectPath := filepath.Join(projectsDir, projectName)
	err = extractRepoFixtureToDir("single-branch", projectPath)
	Expect(err).NotTo(HaveOccurred(), "Failed to extract fixture to project directory")

	f.projects[projectName] = &ProjectInfo{
		Name:    projectName,
		Path:    projectPath,
		Fixture: nil,
	}

	f.configHelper.WithProjectsDir(projectsDir)

	worktreesDir := filepath.Join(f.tempDir, "worktrees")
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

	return f
}

// CreateCustomBranchSetup creates a project with custom default branch

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

	// Note: Projects are not validated here because they're cleaned up by Ginkgo's TempDir
	// Only worktrees are validated as they require explicit git worktree remove

	if len(validationErrors) > 0 {
		return fmt.Errorf("cleanup validation failed: %v", validationErrors)
	}

	return nil
}

// CreateWorktree creates a new worktree manually using git CLI
func (f *E2ETestFixture) CreateWorktree(projectPath, worktreePath, branch string) error {
	result, err := f.gitExecutor.Execute(
		context.Background(),
		projectPath,
		"git", "worktree", "add", "-b", branch, worktreePath,
	)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w, output: %s", err, result.Stdout)
	}
	f.createdWorktrees = append(f.createdWorktrees, worktreePath)
	return nil
}

// RemoveWorktree removes a worktree using git CLI
func (f *E2ETestFixture) RemoveWorktree(worktreePath string) error {
	// Find owning repository
	var repoPath string
	for _, info := range f.projects {
		if info != nil && info.Path != "" {
			repoPath = info.Path
			break
		}
	}

	if repoPath == "" {
		return fmt.Errorf("no main repo found")
	}

	result, err := f.gitExecutor.Execute(
		context.Background(),
		repoPath,
		"git", "worktree", "remove", worktreePath,
	)
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w, output: %s", err, result.Stdout)
	}
	return nil
}

// CreateFileAndCommit creates a file with content, adds it to the worktree, and commits with the given message
func (f *E2ETestFixture) CreateFileAndCommit(worktreePath, filename, content, commitMsg string) error {
	filePath := filepath.Join(worktreePath, filename)
	if err := os.WriteFile(filePath, []byte(content), FilePermReadWrite); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	repo, err := f.gitHelper.PlainOpen(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to open repo at %s: %w", worktreePath, err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if _, err := wt.Add(filename); err != nil {
		return fmt.Errorf("failed to add file %s: %w", filename, err)
	}

	if _, err := wt.Commit(commitMsg, &git.CommitOptions{}); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// GetWorktreePath returns the full worktree path for a given project and branch
func (f *E2ETestFixture) GetWorktreePath(projectName, branch string) string {
	worktreesDir := f.configHelper.GetWorktreesDir()
	return filepath.Join(worktreesDir, projectName, branch)
}
