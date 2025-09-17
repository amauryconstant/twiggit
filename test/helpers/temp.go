package helpers

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
