package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathValidatorImpl_IsValidGitRepoPath(t *testing.T) {
	validator := NewPathValidator()

	t.Run("should return true for valid absolute paths", func(t *testing.T) {
		validPaths := []string{
			"/valid/path",
			"/home/user/repo",
			"/tmp/test-repo",
			"/a/b/c/d",
		}

		for _, path := range validPaths {
			result := validator.IsValidGitRepoPath(path)
			assert.True(t, result, "Expected valid absolute path '%s' to return true", path)
		}
	})

	t.Run("should return false for empty path", func(t *testing.T) {
		result := validator.IsValidGitRepoPath("")
		assert.False(t, result, "Expected empty path to return false")
	})

	t.Run("should return false for relative paths", func(t *testing.T) {
		relativePaths := []string{
			"relative/path",
			"./repo",
			"../parent",
			"~/home",
		}

		for _, path := range relativePaths {
			result := validator.IsValidGitRepoPath(path)
			assert.False(t, result, "Expected relative path '%s' to return false", path)
		}
	})

	t.Run("should return false for paths with path traversal", func(t *testing.T) {
		invalidPaths := []string{
			"/valid/path/../invalid",
			"/a/./b/../c",
			"/path/../../traversal",
		}

		for _, path := range invalidPaths {
			result := validator.IsValidGitRepoPath(path)
			assert.False(t, result, "Expected path with traversal '%s' to return false", path)
		}
	})

	t.Run("should return false for paths with double slashes", func(t *testing.T) {
		invalidPaths := []string{
			"/valid//path",
			"/path//to//repo",
			"//double/slash",
		}

		for _, path := range invalidPaths {
			result := validator.IsValidGitRepoPath(path)
			assert.False(t, result, "Expected path with double slashes '%s' to return false", path)
		}
	})

	t.Run("should return false for paths longer than 255 characters", func(t *testing.T) {
		// Path with exactly 255 characters should be valid
		longPath := "/" + strings.Repeat("a", 254) // Total length: 255
		result := validator.IsValidGitRepoPath(longPath)
		assert.True(t, result, "Expected 255 character path to return true")

		// Path with 256 characters should be invalid
		tooLongPath := "/" + strings.Repeat("a", 255) // Total length: 256
		result = validator.IsValidGitRepoPath(tooLongPath)
		assert.False(t, result, "Expected path longer than 255 characters to return false")
	})
}

func TestPathValidatorImpl_IsValidWorkspacePath(t *testing.T) {
	validator := NewPathValidator()

	t.Run("should return true for valid paths", func(t *testing.T) {
		validPaths := []string{
			"/valid/workspace",
			"/home/user/workspace",
			"/tmp/test-workspace",
			"/a/b/c/d",
			"relative/path", // Workspace paths can be relative
		}

		for _, path := range validPaths {
			result := validator.IsValidWorkspacePath(path)
			assert.True(t, result, "Expected valid path '%s' to return true", path)
		}
	})

	t.Run("should return false for empty path", func(t *testing.T) {
		result := validator.IsValidWorkspacePath("")
		assert.False(t, result, "Expected empty path to return false")
	})

	t.Run("should return false for paths with path traversal", func(t *testing.T) {
		invalidPaths := []string{
			"/valid/path/../invalid",
			"/a/./b/../c",
			"/path/../../traversal",
			"relative/../invalid",
		}

		for _, path := range invalidPaths {
			result := validator.IsValidWorkspacePath(path)
			assert.False(t, result, "Expected path with traversal '%s' to return false", path)
		}
	})

	t.Run("should return false for paths with double slashes", func(t *testing.T) {
		invalidPaths := []string{
			"/valid//path",
			"/path//to//workspace",
			"relative//path",
			"//double/slash",
		}

		for _, path := range invalidPaths {
			result := validator.IsValidWorkspacePath(path)
			assert.False(t, result, "Expected path with double slashes '%s' to return false", path)
		}
	})

	t.Run("should return false for paths longer than 255 characters", func(t *testing.T) {
		// Path with exactly 255 characters should be valid
		longPath := "/" + strings.Repeat("a", 254) // Total length: 255
		result := validator.IsValidWorkspacePath(longPath)
		assert.True(t, result, "Expected 255 character path to return true")

		// Path with 256 characters should be invalid
		tooLongPath := "/" + strings.Repeat("a", 255) // Total length: 256
		result = validator.IsValidWorkspacePath(tooLongPath)
		assert.False(t, result, "Expected path longer than 255 characters to return false")
	})
}

func TestNewPathValidator(t *testing.T) {
	t.Run("should create non-nil PathValidator", func(t *testing.T) {
		validator := NewPathValidator()
		assert.NotNil(t, validator, "Expected NewPathValidator to return non-nil validator")
	})

	t.Run("should create PathValidator with correct type", func(t *testing.T) {
		validator := NewPathValidator()
		assert.IsType(t, &PathValidatorImpl{}, validator, "Expected NewPathValidator to return PathValidatorImpl")
	})
}
