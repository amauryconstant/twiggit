package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempDir creates a temporary directory and returns a cleanup function.
// The cleanup function should be called with defer to ensure proper cleanup.
func TempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", pattern)
	require.NoError(t, err)

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// MustMkdirAll creates directories with proper error handling for tests.
func MustMkdirAll(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, perm))
}

// WorkingDir temporarily changes the working directory and returns a cleanup function.
func WorkingDir(t *testing.T, dir string) func() {
	t.Helper()

	originalWd, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(dir))

	return func() {
		_ = os.Chdir(originalWd)
	}
}
