package domain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractProjectFromWorktreePath(t *testing.T) {
	worktreesDir := filepath.Join(os.TempDir(), "worktrees")

	tests := []struct {
		name         string
		worktreePath string
		worktreesDir string
		expected     string
	}{
		{
			name:         "project from single segment",
			worktreePath: filepath.Join(worktreesDir, "myproject", "feature-branch"),
			worktreesDir: worktreesDir,
			expected:     "myproject",
		},
		{
			name:         "project from deeply nested branch",
			worktreePath: filepath.Join(worktreesDir, "myproject", "feature", "deep", "branch"),
			worktreesDir: worktreesDir,
			expected:     "myproject",
		},
		{
			name:         "path outside worktrees directory returns empty",
			worktreePath: filepath.Join(os.TempDir(), "somewhere-else", "feature-branch"),
			worktreesDir: worktreesDir,
			expected:     "",
		},
		{
			name:         "path equal to worktrees directory returns empty",
			worktreePath: worktreesDir,
			worktreesDir: worktreesDir,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			got, err := ExtractProjectFromWorktreePath(tt.worktreePath, tt.worktreesDir)
			is.NoError(err)
			is.Equal(tt.expected, got)
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldExist bool
	}{
		{
			name:        "absolute existing path",
			input:       os.TempDir(),
			shouldExist: true,
		},
		{
			name:        "non-existent path still returns absolute",
			input:       filepath.Join(os.TempDir(), "does-not-exist-xyz"),
			shouldExist: false,
		},
		{
			name:        "relative path returns absolute",
			input:       ".",
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			got, err := NormalizePath(tt.input)
			is.NoError(err)
			is.True(filepath.IsAbs(got), "normalized path must be absolute")
		})
	}
}

func TestIsPathUnder(t *testing.T) {
	base := filepath.Join(os.TempDir(), "twiggit-base")
	target := filepath.Join(base, "sub", "leaf")

	tests := []struct {
		name     string
		base     string
		target   string
		expected bool
		wantErr  bool
	}{
		{
			name:     "target is under base",
			base:     base,
			target:   target,
			expected: true,
		},
		{
			name:     "target equals base",
			base:     base,
			target:   base,
			expected: true,
		},
		{
			name:     "target is outside base",
			base:     base,
			target:   filepath.Join(os.TempDir(), "other-root", "leaf"),
			expected: false,
		},
		{
			name:     "target escapes via parent reference",
			base:     filepath.Join(base, "inner"),
			target:   filepath.Join(base, "inner", "..", "escape"),
			expected: false,
		},
		{
			name:     "empty base errors",
			base:     "",
			target:   target,
			expected: false,
			wantErr:  true,
		},
		{
			name:     "empty target errors",
			base:     base,
			target:   "",
			expected: false,
			wantErr:  true,
		},
		{
			name:     "both empty returns true",
			base:     "",
			target:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			got, err := IsPathUnder(tt.base, tt.target)
			if tt.wantErr {
				is.Error(err)
				return
			}
			is.NoError(err)
			is.Equal(tt.expected, got)
		})
	}
}

func TestIsPathUnder_NonExistentBase(t *testing.T) {
	base := filepath.Join(os.TempDir(), "twiggit-nonexistent-base")
	target := filepath.Join(base, "leaf")

	is := assert.New(t)
	got, err := IsPathUnder(base, target)
	require.NoError(t, err)
	is.True(got, "non-existent base must still answer based on lexical prefix")
}
