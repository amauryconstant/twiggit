package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)
	assert.NotEmpty(t, config.ProjectsDirectory)
	assert.NotEmpty(t, config.WorktreesDirectory)
	assert.Equal(t, "main", config.DefaultSourceBranch)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid projects directory", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "relative/path",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "projects_directory must be absolute path")
	})

	t.Run("invalid worktrees directory", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "relative/path",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "worktrees_directory must be absolute path")
	})

	t.Run("empty default source branch", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "",
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_source_branch cannot be empty")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "relative/path",
			WorktreesDirectory:  "another/relative/path",
			DefaultSourceBranch: "",
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config validation failed")
		// Should contain all validation errors
		assert.Contains(t, err.Error(), "projects_directory must be absolute path")
		assert.Contains(t, err.Error(), "worktrees_directory must be absolute path")
		assert.Contains(t, err.Error(), "default_source_branch cannot be empty")
	})
}
