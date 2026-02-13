//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
)

type PruneIntegrationTestSuite struct {
	suite.Suite
	executor   *infrastructure.DefaultCommandExecutor
	cliClient  infrastructure.CLIClient
	gitService infrastructure.GitClient
}

func (s *PruneIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("Skipping integration tests in short mode")
	}
	s.executor = infrastructure.NewDefaultCommandExecutor(30 * time.Second)
	s.cliClient = infrastructure.NewCLIClient(s.executor, 30)
	goGitClient := infrastructure.NewGoGitClient(true)
	s.gitService = infrastructure.NewCompositeGitClient(goGitClient, s.cliClient)
}

func TestPruneIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PruneIntegrationTestSuite))
}

func (s *PruneIntegrationTestSuite) setupTestRepo(projectName string) (repoPath string) {
	tempDir := s.T().TempDir()
	repoPath = filepath.Join(tempDir, projectName)
	s.Require().NoError(os.MkdirAll(repoPath, 0755))

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "init")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "config", "user.name", "Test User")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "config", "user.email", "test@example.com")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "commit", "--allow-empty", "-m", "Initial commit")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "branch", "-M", "main")
	s.Require().NoError(err)

	return repoPath
}

func (s *PruneIntegrationTestSuite) createWorktreeService(repoPath string) application.WorktreeService {
	config := domain.DefaultConfig()
	mockProjectService := &mockProjectServiceForPrune{
		project: &domain.ProjectInfo{
			Name:        "test-project",
			Path:        repoPath,
			GitRepoPath: repoPath,
		},
	}
	return service.NewWorktreeService(s.gitService, mockProjectService, config)
}

func (s *PruneIntegrationTestSuite) TestDeleteBranch_NonExistentBranch() {
	repoPath := s.setupTestRepo("test-repo")

	err := s.gitService.DeleteBranch(context.Background(), repoPath, "non-existent-branch")
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")
}

func (s *PruneIntegrationTestSuite) TestDeleteBranch_CurrentHEADBranch() {
	repoPath := s.setupTestRepo("test-repo")

	err := s.gitService.DeleteBranch(context.Background(), repoPath, "main")
	s.Require().Error(err)
	s.Contains(err.Error(), "cannot delete")
}

func (s *PruneIntegrationTestSuite) TestDeleteBranch_EmptyRepositoryPath() {
	err := s.gitService.DeleteBranch(context.Background(), "", "some-branch")
	s.Require().Error(err)
	s.Contains(err.Error(), "repository path cannot be empty")
}

func (s *PruneIntegrationTestSuite) TestDeleteBranch_EmptyBranchName() {
	repoPath := s.setupTestRepo("test-repo")

	err := s.gitService.DeleteBranch(context.Background(), repoPath, "")
	s.Require().Error(err)
	s.Contains(err.Error(), "branch name cannot be empty")
}

func (s *PruneIntegrationTestSuite) TestDeleteBranch_Success() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-to-delete")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	err = s.gitService.DeleteBranch(context.Background(), repoPath, "feature-to-delete")
	s.Require().NoError(err)
}

func (s *PruneIntegrationTestSuite) TestPruneMergedWorktrees_DryRun() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-dry-run")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-dry-run")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-dry-run", "main", worktreePath)
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         true,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted, "dry run should not delete")

	_, err = os.Stat(worktreePath)
	s.NoError(err, "worktree should still exist after dry run")
}

func (s *PruneIntegrationTestSuite) TestPruneMergedWorktrees_WorktreeOperations() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-test")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-test")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-test", "main", worktreePath)
	s.Require().NoError(err)

	worktrees, err := s.cliClient.ListWorktrees(context.Background(), repoPath)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(worktrees), 2, "should have main + feature worktree")

	err = s.cliClient.DeleteWorktree(context.Background(), repoPath, worktreePath, true)
	s.NoError(err)
}

func (s *PruneIntegrationTestSuite) TestPruneMergedWorktrees_FullLifecycle() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-merged")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-merged")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-merged", "main", worktreePath)
	s.Require().NoError(err)

	_, err = os.Stat(worktreePath)
	s.Require().NoError(err, "worktree should exist before prune")

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-merged", "--no-edit")
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted, "should delete 1 merged worktree")
	s.Len(result.DeletedWorktrees, 1)
	s.Equal("feature-merged", result.DeletedWorktrees[0].BranchName)

	_, err = os.Stat(worktreePath)
	s.True(os.IsNotExist(err), "worktree should be deleted after prune")
}

func (s *PruneIntegrationTestSuite) TestPruneMergedWorktrees_WithDeleteBranches() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-with-branch")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-with-branch")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-with-branch", "main", worktreePath)
	s.Require().NoError(err)

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-with-branch", "--no-edit")
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          false,
		DeleteBranches: true,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted, "should delete 1 worktree")
	s.Equal(1, result.TotalBranchesDeleted, "should delete 1 branch")
	s.Len(result.DeletedWorktrees, 1)
	s.True(result.DeletedWorktrees[0].BranchDeleted)

	_, err = os.Stat(worktreePath)
	s.True(os.IsNotExist(err), "worktree should be deleted")

	output, err := s.executor.Execute(context.Background(), repoPath, "git", "branch", "--list", "feature-with-branch")
	s.Require().NoError(err)
	s.NotContains(output.Stdout, "feature-with-branch", "branch should be deleted")
}

func (s *PruneIntegrationTestSuite) TestProtectedBranch_Skipped() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "develop")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-develop")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "develop", "main", worktreePath)
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          true,
		DeleteBranches: true,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted, "should not delete protected branch worktree")
	s.Equal(1, len(result.ProtectedSkipped), "should skip 1 protected branch")
}

func (s *PruneIntegrationTestSuite) TestMergeStatus_UnmergedSkipped() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-unmerged")
	s.Require().NoError(err)
	testFile := filepath.Join(repoPath, "uncommitted.txt")
	s.Require().NoError(os.WriteFile(testFile, []byte("change"), 0644))
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "add", "uncommitted.txt")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "commit", "-m", "Unique change")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-unmerged")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-unmerged", "main", worktreePath)
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted, "should not delete unmerged worktree")
	s.Equal(1, len(result.UnmergedSkipped), "should skip 1 unmerged branch")
}

func (s *PruneIntegrationTestSuite) TestPruneErrorHandling_InvalidWorktreeFormat() {
	repoPath := s.setupTestRepo("test-repo")

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		SpecificWorktree: "invalid-format",
		Context:          &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "must be in format project/branch")
}

func (s *PruneIntegrationTestSuite) TestPruneConfig_DefaultProtectedBranches() {
	config := domain.DefaultConfig()
	s.Contains(config.Validation.ProtectedBranches, "main")
	s.Contains(config.Validation.ProtectedBranches, "master")
	s.Contains(config.Validation.ProtectedBranches, "develop")
}

func (s *PruneIntegrationTestSuite) TestPruneConfig_CustomProtectedBranches() {
	config := domain.DefaultConfig()
	config.Validation.ProtectedBranches = []string{"main", "custom-protected"}

	s.Contains(config.Validation.ProtectedBranches, "main")
	s.Contains(config.Validation.ProtectedBranches, "custom-protected")
	s.NotContains(config.Validation.ProtectedBranches, "develop")
}

func (s *PruneIntegrationTestSuite) TestUncommittedChanges_SkippedWithoutForce() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-uncommitted")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-uncommitted", "--no-edit")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-uncommitted")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-uncommitted", "main", worktreePath)
	s.Require().NoError(err)

	uncommittedFile := filepath.Join(worktreePath, "uncommitted.txt")
	s.Require().NoError(os.WriteFile(uncommittedFile, []byte("uncommitted changes"), 0644))

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted, "should not delete worktree with uncommitted changes")

	var found bool
	for _, skipped := range result.SkippedWorktrees {
		if skipped.BranchName == "feature-uncommitted" {
			s.Contains(skipped.SkipReason, "uncommitted changes")
			found = true
			break
		}
	}
	s.True(found, "worktree should be in skipped list with uncommitted changes reason")

	_, err = os.Stat(worktreePath)
	s.NoError(err, "worktree should still exist")
}

func (s *PruneIntegrationTestSuite) TestUncommittedChanges_ForceBypasses() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-force-uncommitted")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-force-uncommitted", "--no-edit")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-force-uncommitted")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-force-uncommitted", "main", worktreePath)
	s.Require().NoError(err)

	uncommittedFile := filepath.Join(worktreePath, "uncommitted.txt")
	s.Require().NoError(os.WriteFile(uncommittedFile, []byte("uncommitted changes"), 0644))

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted, "should delete worktree with --force")

	_, err = os.Stat(worktreePath)
	s.True(os.IsNotExist(err), "worktree should be deleted")
}

func (s *PruneIntegrationTestSuite) TestCurrentWorktree_Skipped() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-current")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-current", "--no-edit")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-current")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-current", "main", worktreePath)
	s.Require().NoError(err)

	originalWd, err := os.Getwd()
	s.Require().NoError(err)
	defer os.Chdir(originalWd)

	err = os.Chdir(worktreePath)
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted, "should not delete current worktree")
	s.Len(result.CurrentWorktreeSkipped, 1)
	s.Equal("feature-current", result.CurrentWorktreeSkipped[0].BranchName)
	s.Contains(result.CurrentWorktreeSkipped[0].SkipReason, "cannot prune current worktree")

	_, err = os.Stat(worktreePath)
	s.NoError(err, "worktree should still exist")
}

func (s *PruneIntegrationTestSuite) TestNavigationOutput_SingleWorktreePrune() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-nav")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-nav", "--no-edit")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath := filepath.Join(tempDir, "wt-feature-nav")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-nav", "main", worktreePath)
	s.Require().NoError(err)

	config := domain.DefaultConfig()
	config.ProjectsDirectory = tempDir
	mockProjectService := &mockProjectServiceForPrune{
		project: &domain.ProjectInfo{
			Name:        "test-repo",
			Path:        repoPath,
			GitRepoPath: repoPath,
		},
	}
	worktreeService := service.NewWorktreeService(s.gitService, mockProjectService, config)

	req := &domain.PruneWorktreesRequest{
		SpecificWorktree: "test-repo/feature-nav",
		Context:          &domain.Context{Type: domain.ContextProject, ProjectName: "test-repo", Path: repoPath},
		DryRun:           false,
		Force:            false,
		DeleteBranches:   false,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted)

	s.NotEmpty(result.NavigationPath, "navigation path should be set for single worktree prune")
	s.Equal(repoPath, result.NavigationPath)
}

func (s *PruneIntegrationTestSuite) TestOperationSummary_CorrectTotals() {
	repoPath := s.setupTestRepo("test-repo")

	_, err := s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-summary-1")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "merge", "feature-summary-1", "--no-edit")
	s.Require().NoError(err)

	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-summary-2")
	s.Require().NoError(err)
	testFile := filepath.Join(repoPath, "unique-change.txt")
	s.Require().NoError(os.WriteFile(testFile, []byte("unique content"), 0644))
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "add", "unique-change.txt")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "commit", "-m", "Unique change on feature-summary-2")
	s.Require().NoError(err)
	_, err = s.executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
	s.Require().NoError(err)

	tempDir := filepath.Dir(repoPath)
	worktreePath1 := filepath.Join(tempDir, "wt-feature-summary-1")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-summary-1", "main", worktreePath1)
	s.Require().NoError(err)

	worktreePath2 := filepath.Join(tempDir, "wt-feature-summary-2")
	err = s.cliClient.CreateWorktree(context.Background(), repoPath, "feature-summary-2", "main", worktreePath2)
	s.Require().NoError(err)

	worktreeService := s.createWorktreeService(repoPath)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: repoPath},
		DryRun:         false,
		Force:          true,
		DeleteBranches: true,
	}

	result, err := worktreeService.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)

	s.Equal(1, result.TotalDeleted, "should delete 1 merged worktree")
	s.Equal(1, result.TotalBranchesDeleted, "should delete 1 branch")
	s.Equal(1, result.TotalSkipped, "should skip 1 (unmerged)")

	s.Len(result.DeletedWorktrees, 1)
	s.Equal("feature-summary-1", result.DeletedWorktrees[0].BranchName)
	s.True(result.DeletedWorktrees[0].BranchDeleted)

	s.Len(result.UnmergedSkipped, 1, "should have 1 unmerged skipped")
	s.Equal("feature-summary-2", result.UnmergedSkipped[0].BranchName)
}

type mockProjectServiceForPrune struct {
	project *domain.ProjectInfo
}

func (m *mockProjectServiceForPrune) DiscoverProject(ctx context.Context, name string, context *domain.Context) (*domain.ProjectInfo, error) {
	return m.project, nil
}

func (m *mockProjectServiceForPrune) GetProjectInfo(ctx context.Context, path string) (*domain.ProjectInfo, error) {
	return m.project, nil
}

func (m *mockProjectServiceForPrune) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	if m.project != nil {
		return []*domain.ProjectInfo{m.project}, nil
	}
	return []*domain.ProjectInfo{}, nil
}

func (m *mockProjectServiceForPrune) ValidateProject(ctx context.Context, path string) error {
	return nil
}
