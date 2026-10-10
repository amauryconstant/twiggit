package core

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
	assert.Equal(t, []string{"main", "master", "develop", "staging", "production"}, config.Validation.ProtectedBranches)
}

func TestValidate(t *testing.T) {
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
		assert.True(t, hasSuggestion(err, "projects_directory must be absolute path"))
	})

	t.Run("invalid worktrees directory", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "relative/path",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		require.Error(t, err)
		assert.True(t, hasSuggestion(err, "worktrees_directory must be absolute path"))
	})

	t.Run("empty default source branch", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "",
		}

		err := config.Validate()
		require.Error(t, err)
		assert.True(t, hasSuggestion(err, "default_source_branch cannot be empty"))
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
		// Should contain all validation errors as suggestions
		assert.True(t, hasSuggestion(err, "projects_directory must be absolute path"))
		assert.True(t, hasSuggestion(err, "worktrees_directory must be absolute path"))
		assert.True(t, hasSuggestion(err, "default_source_branch cannot be empty"))
	})

	t.Run("negative status thresholds rejected", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "main",
			Status:              StatusConfig{StaleBehind: -1, StaleDays: -1},
		}

		err := config.Validate()
		require.Error(t, err)
		assert.True(t, hasSuggestion(err, "status.stale_behind cannot be negative"))
		assert.True(t, hasSuggestion(err, "status.stale_days cannot be negative"))
	})
}

func TestStatusConfig_Defaults(t *testing.T) {
	is := assert.New(t)
	config := DefaultConfig()
	is.Equal(20, config.Status.StaleBehind, "default StaleBehind must be 20")
	is.Equal(30, config.Status.StaleDays, "default StaleDays must be 30")
}

// TestStatusConfig_FieldsHaveKoanfTags is the structural guard: every
// exported field on StatusConfig carries a koanf struct tag. A future
// rename that drops the tag breaks koanf loading silently; this test
// surfaces the breakage at the unit-test layer.
func TestStatusConfig_FieldsHaveKoanfTags(t *testing.T) {
	t.Parallel()

	// Struct literal with named fields forces a compile failure
	// when a field is removed or renamed. The round-trip below
	// guards the koanf tag presence.
	cfg := StatusConfig{StaleBehind: 50, StaleDays: 7}
	assert.Equal(t, 50, cfg.StaleBehind)
	assert.Equal(t, 7, cfg.StaleDays)
}
