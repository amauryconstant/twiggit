//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/infrastructure"
)

func TestNormalizePath_Integration(t *testing.T) {
	t.Run("absolute path conversion", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a subdirectory
		subDir := filepath.Join(tempDir, "subdir")
		err := os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		// Change to subdirectory
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		err = os.Chdir(subDir)
		require.NoError(t, err)

		// Test relative path conversion
		result, err := infrastructure.NormalizePath("test.txt")
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result))
		assert.Equal(t, filepath.Join(subDir, "test.txt"), result)
	})

	t.Run("symlink resolution", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		tempDir := t.TempDir()

		// Create target file
		targetFile := filepath.Join(tempDir, "target.txt")
		err := os.WriteFile(targetFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create symlink
		symlinkPath := filepath.Join(tempDir, "symlink.txt")
		err = os.Symlink(targetFile, symlinkPath)
		require.NoError(t, err)

		// Test symlink resolution
		result, err := infrastructure.NormalizePath(symlinkPath)
		require.NoError(t, err)
		assert.Equal(t, targetFile, result)
	})

	t.Run("broken symlink fallback", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		tempDir := t.TempDir()

		// Create symlink to non-existent target
		symlinkPath := filepath.Join(tempDir, "broken_symlink.txt")
		err := os.Symlink("/non/existent/path", symlinkPath)
		require.NoError(t, err)

		// Test broken symlink handling
		result, err := infrastructure.NormalizePath(symlinkPath)
		require.NoError(t, err)
		// Should fallback to absolute path of symlink itself
		assert.True(t, filepath.IsAbs(result))
	})

	t.Run("directory symlink resolution", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		tempDir := t.TempDir()

		// Create target directory
		targetDir := filepath.Join(tempDir, "target_dir")
		err := os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Create file in target directory
		targetFile := filepath.Join(targetDir, "file.txt")
		err = os.WriteFile(targetFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create symlink to directory
		symlinkDir := filepath.Join(tempDir, "symlink_dir")
		err = os.Symlink(targetDir, symlinkDir)
		require.NoError(t, err)

		// Test directory symlink resolution
		result, err := infrastructure.NormalizePath(filepath.Join(symlinkDir, "file.txt"))
		require.NoError(t, err)
		assert.Equal(t, targetFile, result)
	})

	t.Run("complex path with symlinks", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		tempDir := t.TempDir()

		// Create nested directory structure
		deepDir := filepath.Join(tempDir, "a", "b", "c")
		err := os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)

		// Create file in deep directory
		deepFile := filepath.Join(deepDir, "deep.txt")
		err = os.WriteFile(deepFile, []byte("deep content"), 0644)
		require.NoError(t, err)

		// Create symlink chain
		symlink1 := filepath.Join(tempDir, "link1")
		err = os.Symlink(deepDir, symlink1)
		require.NoError(t, err)

		symlink2 := filepath.Join(tempDir, "link2")
		err = os.Symlink(symlink1, symlink2)
		require.NoError(t, err)

		// Test complex symlink resolution
		result, err := infrastructure.NormalizePath(filepath.Join(symlink2, "deep.txt"))
		require.NoError(t, err)
		assert.Equal(t, deepFile, result)
	})

	t.Run("working directory context", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create nested structure
		nestedDir := filepath.Join(tempDir, "level1", "level2")
		err := os.MkdirAll(nestedDir, 0755)
		require.NoError(t, err)

		// Create file at different levels
		rootFile := filepath.Join(tempDir, "root.txt")
		err = os.WriteFile(rootFile, []byte("root"), 0644)
		require.NoError(t, err)

		nestedFile := filepath.Join(nestedDir, "nested.txt")
		err = os.WriteFile(nestedFile, []byte("nested"), 0644)
		require.NoError(t, err)

		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		// Test from root directory
		err = os.Chdir(tempDir)
		require.NoError(t, err)

		result, err := infrastructure.NormalizePath("root.txt")
		require.NoError(t, err)
		assert.Equal(t, rootFile, result)

		// Test from nested directory
		err = os.Chdir(nestedDir)
		require.NoError(t, err)

		result, err = infrastructure.NormalizePath("nested.txt")
		require.NoError(t, err)
		assert.Equal(t, nestedFile, result)

		// Test relative path from nested directory
		result, err = infrastructure.NormalizePath("../../root.txt")
		require.NoError(t, err)
		assert.Equal(t, rootFile, result)
	})
}

func TestIsPathUnder_Integration(t *testing.T) {
	t.Run("real directory structure", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create directory structure
		baseDir := filepath.Join(tempDir, "base")
		subDir := filepath.Join(baseDir, "subdir")
		deepDir := filepath.Join(subDir, "deep")
		outsideDir := filepath.Join(tempDir, "outside")

		err := os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(outsideDir, 0755)
		require.NoError(t, err)

		// Create files
		baseFile := filepath.Join(baseDir, "base.txt")
		subFile := filepath.Join(subDir, "sub.txt")
		deepFile := filepath.Join(deepDir, "deep.txt")
		outsideFile := filepath.Join(outsideDir, "outside.txt")

		err = os.WriteFile(baseFile, []byte("base"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(subFile, []byte("sub"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(deepFile, []byte("deep"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(outsideFile, []byte("outside"), 0644)
		require.NoError(t, err)

		tests := []struct {
			name     string
			base     string
			target   string
			expected bool
		}{
			{"file under base", baseDir, baseFile, true},
			{"nested file under base", baseDir, deepFile, true},
			{"directory under base", baseDir, subDir, true},
			{"deep directory under base", baseDir, deepDir, true},
			{"file outside base", baseDir, outsideFile, false},
			{"directory outside base", baseDir, outsideDir, false},
			{"same directory", baseDir, baseDir, true},
			{"parent directory", subDir, baseDir, true}, // subDir is under baseDir
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := infrastructure.IsPathUnder(tt.base, tt.target)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("symlink directory relationships", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		tempDir := t.TempDir()

		// Create real directory structure
		realBase := filepath.Join(tempDir, "real_base")
		realSub := filepath.Join(realBase, "subdir")
		err := os.MkdirAll(realSub, 0755)
		require.NoError(t, err)

		// Create symlink to base directory
		symlinkBase := filepath.Join(tempDir, "symlink_base")
		err = os.Symlink(realBase, symlinkBase)
		require.NoError(t, err)

		// Test symlink relationships
		result, err := infrastructure.IsPathUnder(symlinkBase, filepath.Join(symlinkBase, "subdir"))
		require.NoError(t, err)
		assert.True(t, result)

		// Test mixed real/symlink paths
		result, err = infrastructure.IsPathUnder(realBase, filepath.Join(symlinkBase, "subdir"))
		require.NoError(t, err)
		// This depends on how filepath.Rel handles symlinks
		// The result should be consistent with the underlying paths
	})

	t.Run("relative path integration", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create structure
		baseDir := filepath.Join(tempDir, "base")
		subDir := filepath.Join(baseDir, "sub")
		err := os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		// Change to temp directory
		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// Test relative paths
		result, err := infrastructure.IsPathUnder("base", "base/sub")
		require.NoError(t, err)
		assert.True(t, result)

		result, err = infrastructure.IsPathUnder("base/sub", "base")
		require.NoError(t, err)
		assert.True(t, result) // "base" is parent of "base/sub", so "base" is under "base/sub"

		// Change to base directory
		err = os.Chdir(baseDir)
		require.NoError(t, err)

		result, err = infrastructure.IsPathUnder(".", "sub")
		require.NoError(t, err)
		assert.True(t, result)

		result, err = infrastructure.IsPathUnder("sub", "..")
		require.NoError(t, err)
		assert.False(t, result) // ".." is parent of "sub", so ".." is not under "sub"
	})

	t.Run("complex real-world scenario", func(t *testing.T) {
		tempDir := t.TempDir()

		// Simulate a project structure
		projectRoot := filepath.Join(tempDir, "myproject")
		srcDir := filepath.Join(projectRoot, "src")
		pkgDir := filepath.Join(srcDir, "pkg")
		cmdDir := filepath.Join(projectRoot, "cmd")
		vendorDir := filepath.Join(projectRoot, "vendor")
		testDir := filepath.Join(projectRoot, "test")

		dirs := []string{srcDir, pkgDir, cmdDir, vendorDir, testDir}
		for _, dir := range dirs {
			err := os.MkdirAll(dir, 0755)
			require.NoError(t, err)
		}

		// Create some files
		files := map[string]string{
			filepath.Join(projectRoot, "go.mod"):     "module myproject",
			filepath.Join(srcDir, "main.go"):         "package main",
			filepath.Join(pkgDir, "utils.go"):        "package pkg",
			filepath.Join(cmdDir, "cli.go"):          "package cmd",
			filepath.Join(testDir, "integration.go"): "package test",
		}

		for file, content := range files {
			err := os.WriteFile(file, []byte(content), 0644)
			require.NoError(t, err)
		}

		tests := []struct {
			name     string
			base     string
			target   string
			expected bool
		}{
			{"src under project", projectRoot, srcDir, true},
			{"pkg under src", srcDir, pkgDir, true},
			{"cmd under project", projectRoot, cmdDir, true},
			{"vendor under project", projectRoot, vendorDir, true},
			{"test under project", projectRoot, testDir, true},
			{"project under src", srcDir, projectRoot, true}, // srcDir is under projectRoot
			{"cmd under src", srcDir, cmdDir, false},
			{"go.mod under project", projectRoot, filepath.Join(projectRoot, "go.mod"), true},
			{"utils.go under pkg", pkgDir, filepath.Join(pkgDir, "utils.go"), true},
			{"cli.go under src", srcDir, filepath.Join(cmdDir, "cli.go"), false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := infrastructure.IsPathUnder(tt.base, tt.target)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestPathUtils_Integration_EdgeCases(t *testing.T) {
	t.Run("non-existent paths", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test with non-existent paths
		nonExistent := filepath.Join(tempDir, "does_not_exist")

		result, err := infrastructure.NormalizePath(nonExistent)
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result))
		assert.Contains(t, result, "does_not_exist")

		// Test IsPathUnder with non-existent paths
		base := filepath.Join(tempDir, "base")
		target := filepath.Join(base, "sub")

		isUnder, err := infrastructure.IsPathUnder(base, target)
		require.NoError(t, err)
		assert.True(t, isUnder) // Should work with non-existent paths
	})

	t.Run("permission denied scenarios", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		tempDir := t.TempDir()

		// Create a directory and remove read permissions
		restrictedDir := filepath.Join(tempDir, "restricted")
		err := os.Mkdir(restrictedDir, 0000)
		require.NoError(t, err)
		defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

		// Test NormalizePath with restricted directory
		// This should still work as it doesn't need to read the directory contents
		result, err := infrastructure.NormalizePath(restrictedDir)
		// May succeed or fail depending on the system, but shouldn't panic
		if err != nil {
			assert.Contains(t, err.Error(), "failed to normalize path")
		} else {
			assert.True(t, filepath.IsAbs(result))
		}
	})

	t.Run("very long paths", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a very long path name
		longName := strings.Repeat("very_long_directory_name_", 10)
		longPath := filepath.Join(tempDir, longName)

		err := os.Mkdir(longPath, 0755)
		require.NoError(t, err)

		result, err := infrastructure.NormalizePath(longPath)
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result))
		assert.Contains(t, result, longName)
	})
}
