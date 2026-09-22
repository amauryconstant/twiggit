package git

import (
	"os"
	"path/filepath"
	"testing"

	"twiggit/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoFinder_FindGitRepositories(t *testing.T) {
	t.Run("non_existent_directory_returns_empty", func(t *testing.T) {
		finder := NewRepoFinder(nil)
		result, err := finder.FindGitRepositories("/nonexistent/path")
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty_directory_returns_empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyDir := filepath.Join(tmpDir, "empty")
		require.NoError(t, os.MkdirAll(emptyDir, 0755))

		finder := NewRepoFinder(nil)
		result, err := finder.FindGitRepositories(emptyDir)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("filters_out_non_directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "mixed")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "subdir"), 0755))

		finder := NewRepoFinder(nil)
		result, err := finder.FindGitRepositories(testDir)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "subdir", result[0].Name)
	})

	t.Run("with_nil_git_client_includes_all_dirs", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "nogit")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "dir1"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "dir2"), 0755))

		finder := NewRepoFinder(nil)
		result, err := finder.FindGitRepositories(testDir)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		names := []string{result[0].Name, result[1].Name}
		assert.Contains(t, names, "dir1")
		assert.Contains(t, names, "dir2")
	})

	t.Run("with_git_client_filters_invalid_repos", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "withvalidation")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "valid-repo"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "invalid-repo"), 0755))

		mockClient := mocks.NewMockGoGitClient()
		mockClient.On("ValidateRepository", filepath.Join(testDir, "valid-repo")).Return(nil)
		mockClient.On("ValidateRepository", filepath.Join(testDir, "invalid-repo")).Return(os.ErrNotExist)
		t.Cleanup(func() {
			mockClient.AssertExpectations(t)
		})

		finder := NewRepoFinder(mockClient)
		result, err := finder.FindGitRepositories(testDir)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "valid-repo", result[0].Name)
		assert.Equal(t, filepath.Join(testDir, "valid-repo"), result[0].Path)
	})

	t.Run("returns_correct_git_dir_structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "structure")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "myproject"), 0755))

		finder := NewRepoFinder(nil)
		result, err := finder.FindGitRepositories(testDir)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "myproject", result[0].Name)
		assert.Equal(t, filepath.Join(testDir, "myproject"), result[0].Path)
	})

	t.Run("returned_slice_is_defensively_copied", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "clone")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "proj-a"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testDir, "proj-b"), 0755))

		finder := NewRepoFinder(nil)

		first, err := finder.FindGitRepositories(testDir)
		require.NoError(t, err)
		require.Len(t, first, 2)

		first[0].Name = "MUTATED"

		second, err := finder.FindGitRepositories(testDir)
		require.NoError(t, err)
		require.Len(t, second, 2)

		names := []string{second[0].Name, second[1].Name}
		assert.NotContains(t, names, "MUTATED")
		assert.Contains(t, names, "proj-a")
		assert.Contains(t, names, "proj-b")
	})
}
