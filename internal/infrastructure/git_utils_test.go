package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"twiggit/test/mocks"

	"github.com/stretchr/testify/suite"
)

type GitUtilsTestSuite struct {
	suite.Suite
	tmpDir string
}

func TestGitUtils(t *testing.T) {
	suite.Run(t, new(GitUtilsTestSuite))
}

func (s *GitUtilsTestSuite) SetupTest() {
	tmpDir, err := os.MkdirTemp("", "git-utils-test-*")
	s.Require().NoError(err)
	s.tmpDir = tmpDir
}

func (s *GitUtilsTestSuite) TearDownTest() {
	if s.tmpDir != "" {
		os.RemoveAll(s.tmpDir)
	}
}

func (s *GitUtilsTestSuite) TestFindGitRepositories() {
	s.Run("non_existent_directory_returns_empty", func() {
		result, err := FindGitRepositories("/nonexistent/path", nil)
		s.Require().NoError(err)
		s.Empty(result)
	})

	s.Run("empty_directory_returns_empty", func() {
		emptyDir := filepath.Join(s.tmpDir, "empty")
		s.Require().NoError(os.MkdirAll(emptyDir, 0755))

		result, err := FindGitRepositories(emptyDir, nil)
		s.Require().NoError(err)
		s.Empty(result)
	})

	s.Run("filters_out_non_directories", func() {
		testDir := filepath.Join(s.tmpDir, "mixed")
		s.Require().NoError(os.MkdirAll(testDir, 0755))
		s.Require().NoError(os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))
		s.Require().NoError(os.MkdirAll(filepath.Join(testDir, "subdir"), 0755))

		result, err := FindGitRepositories(testDir, nil)
		s.Require().NoError(err)
		s.Len(result, 1)
		s.Equal("subdir", result[0].Name)
	})

	s.Run("with_nil_git_service_includes_all_dirs", func() {
		testDir := filepath.Join(s.tmpDir, "nogit")
		s.Require().NoError(os.MkdirAll(testDir, 0755))
		s.Require().NoError(os.MkdirAll(filepath.Join(testDir, "dir1"), 0755))
		s.Require().NoError(os.MkdirAll(filepath.Join(testDir, "dir2"), 0755))

		result, err := FindGitRepositories(testDir, nil)
		s.Require().NoError(err)
		s.Len(result, 2)
		names := []string{result[0].Name, result[1].Name}
		s.Contains(names, "dir1")
		s.Contains(names, "dir2")
	})

	s.Run("with_git_service_filters_invalid_repos", func() {
		testDir := filepath.Join(s.tmpDir, "withvalidation")
		s.Require().NoError(os.MkdirAll(testDir, 0755))
		s.Require().NoError(os.MkdirAll(filepath.Join(testDir, "valid-repo"), 0755))
		s.Require().NoError(os.MkdirAll(filepath.Join(testDir, "invalid-repo"), 0755))

		mockClient := mocks.NewMockGoGitClient()
		mockClient.On("ValidateRepository", filepath.Join(testDir, "valid-repo")).Return(nil)
		mockClient.On("ValidateRepository", filepath.Join(testDir, "invalid-repo")).Return(os.ErrNotExist)

		result, err := FindGitRepositories(testDir, mockClient)
		s.Require().NoError(err)
		s.Len(result, 1)
		s.Equal("valid-repo", result[0].Name)
		s.Equal(filepath.Join(testDir, "valid-repo"), result[0].Path)
		mockClient.AssertExpectations(s.T())
	})

	s.Run("returns_correct_git_dir_structure", func() {
		testDir := filepath.Join(s.tmpDir, "structure")
		s.Require().NoError(os.MkdirAll(testDir, 0755))
		s.Require().NoError(os.MkdirAll(filepath.Join(testDir, "myproject"), 0755))

		result, err := FindGitRepositories(testDir, nil)
		s.Require().NoError(err)
		s.Len(result, 1)
		s.Equal("myproject", result[0].Name)
		s.Equal(filepath.Join(testDir, "myproject"), result[0].Path)
	})
}

func (s *GitUtilsTestSuite) TestFindGitRepositories_Errors() {
	s.Run("unreadable_directory_returns_error", func() {
		if os.Getuid() == 0 {
			s.T().Skip("Skipping permission test as root")
		}

		testDir := filepath.Join(s.tmpDir, "noperm")
		s.Require().NoError(os.MkdirAll(testDir, 0755))
		s.Require().NoError(os.Chmod(testDir, 0000))
		defer os.Chmod(testDir, 0755)

		_, err := FindGitRepositories(testDir, nil)
		s.Require().Error(err)
		s.Contains(err.Error(), "failed to read directory")
	})
}

func (s *GitUtilsTestSuite) TestFindMainRepoByTraversal() {
	s.Run("finds_main_repo_at_current_level", func() {
		repoDir := filepath.Join(s.tmpDir, "mainrepo")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		result := FindMainRepoByTraversal(repoDir)
		s.Equal(repoDir, result)
	})

	s.Run("finds_main_repo_by_traversing_up", func() {
		repoDir := filepath.Join(s.tmpDir, "mainrepo2")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		nestedDir := filepath.Join(repoDir, "subdir", "nested")
		s.Require().NoError(os.MkdirAll(nestedDir, 0755))

		result := FindMainRepoByTraversal(nestedDir)
		s.Equal(repoDir, result)
	})

	s.Run("returns_empty_when_no_main_repo_found", func() {
		testDir := filepath.Join(s.tmpDir, "norepo")
		s.Require().NoError(os.MkdirAll(testDir, 0755))

		result := FindMainRepoByTraversal(testDir)
		s.Empty(result)
	})

	s.Run("skips_worktree_repos", func() {
		worktreeDir := filepath.Join(s.tmpDir, "worktree")
		gitDir := filepath.Join(worktreeDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))
		gitdirFile := filepath.Join(gitDir, "gitdir")
		s.Require().NoError(os.WriteFile(gitdirFile, []byte("/path/to/main/.git/worktrees/branch"), 0644))

		result := FindMainRepoByTraversal(worktreeDir)
		s.Empty(result)
	})

	s.Run("finds_main_repo_above_worktree", func() {
		mainRepo := filepath.Join(s.tmpDir, "mainrepo3")
		mainGitDir := filepath.Join(mainRepo, ".git")
		s.Require().NoError(os.MkdirAll(mainGitDir, 0755))

		worktreeDir := filepath.Join(mainRepo, "subdir")
		s.Require().NoError(os.MkdirAll(worktreeDir, 0755))
		worktreeGitFile := filepath.Join(worktreeDir, ".git")
		s.Require().NoError(os.WriteFile(worktreeGitFile, []byte("gitdir: /path/to/main/.git/worktrees/branch"), 0644))

		nestedInWorktree := filepath.Join(worktreeDir, "nested")
		s.Require().NoError(os.MkdirAll(nestedInWorktree, 0755))

		result := FindMainRepoByTraversal(nestedInWorktree)
		s.Equal(mainRepo, result)
	})
}

func (s *GitUtilsTestSuite) TestFindGitDirByTraversal() {
	s.Run("finds_git_dir_at_current_level", func() {
		repoDir := filepath.Join(s.tmpDir, "gitdir1")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		result := FindGitDirByTraversal(repoDir)
		s.Require().NotNil(result)
		s.Equal(repoDir, *result)
	})

	s.Run("finds_git_dir_by_traversing_up", func() {
		repoDir := filepath.Join(s.tmpDir, "gitdir2")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		nestedDir := filepath.Join(repoDir, "a", "b", "c")
		s.Require().NoError(os.MkdirAll(nestedDir, 0755))

		result := FindGitDirByTraversal(nestedDir)
		s.Require().NotNil(result)
		s.Equal(repoDir, *result)
	})

	s.Run("returns_nil_when_no_git_dir_found", func() {
		testDir := filepath.Join(s.tmpDir, "nogitdir")
		s.Require().NoError(os.MkdirAll(testDir, 0755))

		result := FindGitDirByTraversal(testDir)
		s.Nil(result)
	})

	s.Run("finds_git_file_in_worktree", func() {
		worktreeDir := filepath.Join(s.tmpDir, "worktreegit")
		s.Require().NoError(os.MkdirAll(worktreeDir, 0755))
		gitFile := filepath.Join(worktreeDir, ".git")
		s.Require().NoError(os.WriteFile(gitFile, []byte("gitdir: /path/to/main/.git/worktrees/branch"), 0644))

		result := FindGitDirByTraversal(worktreeDir)
		s.Require().NotNil(result)
		s.Equal(worktreeDir, *result)
	})

	s.Run("finds_git_dir_in_deeply_nested_structure", func() {
		repoDir := filepath.Join(s.tmpDir, "deep")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		deepPath := filepath.Join(repoDir, "src", "internal", "infrastructure", "service")
		s.Require().NoError(os.MkdirAll(deepPath, 0755))

		result := FindGitDirByTraversal(deepPath)
		s.Require().NotNil(result)
		s.Equal(repoDir, *result)
	})
}

func (s *GitUtilsTestSuite) TestIsMainRepo() {
	s.Run("returns_true_for_main_repo_with_git_directory", func() {
		repoDir := filepath.Join(s.tmpDir, "ismain1")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		result := IsMainRepo(repoDir)
		s.True(result)
	})

	s.Run("returns_false_when_git_does_not_exist", func() {
		repoDir := filepath.Join(s.tmpDir, "ismain2")
		s.Require().NoError(os.MkdirAll(repoDir, 0755))

		result := IsMainRepo(repoDir)
		s.False(result)
	})

	s.Run("returns_false_when_git_is_file", func() {
		repoDir := filepath.Join(s.tmpDir, "ismain3")
		s.Require().NoError(os.MkdirAll(repoDir, 0755))
		gitFile := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.WriteFile(gitFile, []byte("gitdir: /path/to/main/.git/worktrees/branch"), 0644))

		result := IsMainRepo(repoDir)
		s.False(result)
	})

	s.Run("returns_false_when_gitdir_file_exists", func() {
		repoDir := filepath.Join(s.tmpDir, "ismain4")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))
		gitdirFile := filepath.Join(gitDir, "gitdir")
		s.Require().NoError(os.WriteFile(gitdirFile, []byte("/path/to/main/.git/worktrees/branch"), 0644))

		result := IsMainRepo(repoDir)
		s.False(result)
	})

	s.Run("returns_true_when_gitdir_file_does_not_exist", func() {
		repoDir := filepath.Join(s.tmpDir, "ismain5")
		gitDir := filepath.Join(repoDir, ".git")
		s.Require().NoError(os.MkdirAll(gitDir, 0755))

		result := IsMainRepo(repoDir)
		s.True(result)
	})

	s.Run("returns_false_for_nonexistent_path", func() {
		result := IsMainRepo("/nonexistent/path")
		s.False(result)
	})
}

func (s *GitUtilsTestSuite) TestGitDir() {
	s.Run("struct_fields_are_correct", func() {
		gitDir := GitDir{
			Name: "myproject",
			Path: "/home/user/projects/myproject",
		}

		s.Equal("myproject", gitDir.Name)
		s.Equal("/home/user/projects/myproject", gitDir.Path)
	})
}
