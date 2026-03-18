package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"twiggit/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGitUtilsTest(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "git-utils-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

func TestGitUtils_FindGitRepositories(t *testing.T) {
	t.Run("non_existent_directory_returns_empty", func(t *testing.T) {
		result, err := FindGitRepositories("/nonexistent/path", nil)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty_directory_returns_empty", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		emptyDir := filepath.Join(tmpDir, "empty")
		require.NoError(t, os.MkdirAll(emptyDir, 0755))

		result, err := FindGitRepositories(emptyDir, nil)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("filters_out_non_directories", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "mixed")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "subdir"), 0755))

		result, err := FindGitRepositories(testDir, nil)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "subdir", result[0].Name)
	})

	t.Run("with_nil_git_service_includes_all_dirs", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "nogit")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "dir1"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "dir2"), 0755))

		result, err := FindGitRepositories(testDir, nil)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		names := []string{result[0].Name, result[1].Name}
		assert.Contains(t, names, "dir1")
		assert.Contains(t, names, "dir2")
	})

	t.Run("with_git_service_filters_invalid_repos", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "withvalidation")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "valid-repo"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "invalid-repo"), 0755))

		mockClient := mocks.NewMockGoGitClient()
		mockClient.On("ValidateRepository", filepath.Join(testDir, "valid-repo")).Return(nil)
		mockClient.On("ValidateRepository", filepath.Join(testDir, "invalid-repo")).Return(os.ErrNotExist)

		result, err := FindGitRepositories(testDir, mockClient)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "valid-repo", result[0].Name)
		assert.Equal(t, filepath.Join(testDir, "valid-repo"), result[0].Path)
		mockClient.AssertExpectations(t)
	})

	t.Run("returns_correct_git_dir_structure", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "structure")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "myproject"), 0755))

		result, err := FindGitRepositories(testDir, nil)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "myproject", result[0].Name)
		assert.Equal(t, filepath.Join(testDir, "myproject"), result[0].Path)
	})
}

func TestGitUtils_FindGitRepositories_Errors(t *testing.T) {
	t.Run("unreadable_directory_returns_error", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test as root")
		}

		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "noperm")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.Chmod(testDir, 0000))
		defer os.Chmod(testDir, 0755)

		_, err := FindGitRepositories(testDir, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read directory")
	})
}

func TestGitUtils_FindMainRepoByTraversal(t *testing.T) {
	t.Run("finds_main_repo_at_current_level", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "mainrepo")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		result := FindMainRepoByTraversal(repoDir)
		assert.Equal(t, repoDir, result)
	})

	t.Run("finds_main_repo_by_traversing_up", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "mainrepo2")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		nestedDir := filepath.Join(repoDir, "subdir", "nested")
		require.NoError(t, os.MkdirAll(nestedDir, 0755))

		result := FindMainRepoByTraversal(nestedDir)
		assert.Equal(t, repoDir, result)
	})

	t.Run("returns_empty_when_no_main_repo_found", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "norepo")
		require.NoError(t, os.MkdirAll(testDir, 0755))

		result := FindMainRepoByTraversal(testDir)
		assert.Empty(t, result)
	})

	t.Run("skips_worktree_repos", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		worktreeDir := filepath.Join(tmpDir, "worktree")
		gitDir := filepath.Join(worktreeDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))
		gitdirFile := filepath.Join(gitDir, "gitdir")
		require.NoError(t, os.WriteFile(gitdirFile, []byte("/path/to/main/.git/worktrees/branch"), 0644))

		result := FindMainRepoByTraversal(worktreeDir)
		assert.Empty(t, result)
	})

	t.Run("finds_main_repo_above_worktree", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		mainRepo := filepath.Join(tmpDir, "mainrepo3")
		mainGitDir := filepath.Join(mainRepo, ".git")
		require.NoError(t, os.MkdirAll(mainGitDir, 0755))

		worktreeDir := filepath.Join(mainRepo, "subdir")
		require.NoError(t, os.MkdirAll(worktreeDir, 0755))
		worktreeGitFile := filepath.Join(worktreeDir, ".git")
		require.NoError(t, os.WriteFile(worktreeGitFile, []byte("gitdir: /path/to/main/.git/worktrees/branch"), 0644))

		nestedInWorktree := filepath.Join(worktreeDir, "nested")
		require.NoError(t, os.MkdirAll(nestedInWorktree, 0755))

		result := FindMainRepoByTraversal(nestedInWorktree)
		assert.Equal(t, mainRepo, result)
	})
}

func TestGitUtils_FindGitDirByTraversal(t *testing.T) {
	t.Run("finds_git_dir_at_current_level", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "gitdir1")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		result := FindGitDirByTraversal(repoDir)
		require.NotNil(t, result)
		assert.Equal(t, repoDir, *result)
	})

	t.Run("finds_git_dir_by_traversing_up", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "gitdir2")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		nestedDir := filepath.Join(repoDir, "a", "b", "c")
		require.NoError(t, os.MkdirAll(nestedDir, 0755))

		result := FindGitDirByTraversal(nestedDir)
		require.NotNil(t, result)
		assert.Equal(t, repoDir, *result)
	})

	t.Run("returns_nil_when_no_git_dir_found", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		testDir := filepath.Join(tmpDir, "nogitdir")
		require.NoError(t, os.MkdirAll(testDir, 0755))

		result := FindGitDirByTraversal(testDir)
		assert.Nil(t, result)
	})

	t.Run("finds_git_file_in_worktree", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		worktreeDir := filepath.Join(tmpDir, "worktreegit")
		require.NoError(t, os.MkdirAll(worktreeDir, 0755))
		gitFile := filepath.Join(worktreeDir, ".git")
		require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /path/to/main/.git/worktrees/branch"), 0644))

		result := FindGitDirByTraversal(worktreeDir)
		require.NotNil(t, result)
		assert.Equal(t, worktreeDir, *result)
	})

	t.Run("finds_git_dir_in_deeply_nested_structure", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "deep")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		deepPath := filepath.Join(repoDir, "src", "internal", "infrastructure", "service")
		require.NoError(t, os.MkdirAll(deepPath, 0755))

		result := FindGitDirByTraversal(deepPath)
		require.NotNil(t, result)
		assert.Equal(t, repoDir, *result)
	})
}

func TestGitUtils_IsMainRepo(t *testing.T) {
	t.Run("returns_true_for_main_repo_with_git_directory", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "ismain1")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		result := IsMainRepo(repoDir)
		assert.True(t, result)
	})

	t.Run("returns_false_when_git_does_not_exist", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "ismain2")
		require.NoError(t, os.MkdirAll(repoDir, 0755))

		result := IsMainRepo(repoDir)
		assert.False(t, result)
	})

	t.Run("returns_false_when_git_is_file", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "ismain3")
		require.NoError(t, os.MkdirAll(repoDir, 0755))
		gitFile := filepath.Join(repoDir, ".git")
		require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /path/to/main/.git/worktrees/branch"), 0644))

		result := IsMainRepo(repoDir)
		assert.False(t, result)
	})

	t.Run("returns_false_when_gitdir_file_exists", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "ismain4")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))
		gitdirFile := filepath.Join(gitDir, "gitdir")
		require.NoError(t, os.WriteFile(gitdirFile, []byte("/path/to/main/.git/worktrees/branch"), 0644))

		result := IsMainRepo(repoDir)
		assert.False(t, result)
	})

	t.Run("returns_true_when_gitdir_file_does_not_exist", func(t *testing.T) {
		tmpDir := setupGitUtilsTest(t)
		repoDir := filepath.Join(tmpDir, "ismain5")
		gitDir := filepath.Join(repoDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		result := IsMainRepo(repoDir)
		assert.True(t, result)
	})

	t.Run("returns_false_for_nonexistent_path", func(t *testing.T) {
		result := IsMainRepo("/nonexistent/path")
		assert.False(t, result)
	})
}

func TestGitUtils_GitDir(t *testing.T) {
	t.Run("struct_fields_are_correct", func(t *testing.T) {
		gitDir := GitDir{
			Name: "myproject",
			Path: "/home/user/projects/myproject",
		}

		assert.Equal(t, "myproject", gitDir.Name)
		assert.Equal(t, "/home/user/projects/myproject", gitDir.Path)
	})
}
