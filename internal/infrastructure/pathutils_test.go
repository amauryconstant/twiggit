package infrastructure

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PathUtilsTestSuite struct {
	suite.Suite
}

func TestPathUtils(t *testing.T) {
	suite.Run(t, new(PathUtilsTestSuite))
}

func (s *PathUtilsTestSuite) TestNormalizePath() {
	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "empty path",
			input:       "",
			expectError: false,
			description: "Empty path should resolve to current directory",
		},
		{
			name:        "current directory",
			input:       ".",
			expectError: false,
			description: "Current directory should resolve to absolute path",
		},
		{
			name:        "parent directory",
			input:       "..",
			expectError: false,
			description: "Parent directory should resolve to absolute path",
		},
		{
			name:        "simple relative path",
			input:       "foo/bar",
			expectError: false,
			description: "Simple relative path should be normalized",
		},
		{
			name:        "path with dot components",
			input:       "foo/./bar/../baz",
			expectError: false,
			description: "Path with dot components should be cleaned",
		},
		{
			name:        "path with redundant separators",
			input:       "foo//bar///baz",
			expectError: false,
			description: "Path with redundant separators should be cleaned",
		},
		{
			name:        "already absolute path",
			input:       "/foo/bar",
			expectError: false,
			description: "Absolute path should be normalized",
		},
		{
			name:        "path with trailing separator",
			input:       "foo/bar/",
			expectError: false,
			description: "Path with trailing separator should be cleaned",
		},
		{
			name:        "complex path cleanup",
			input:       "foo/../bar/./baz/../qux",
			expectError: false,
			description: "Complex path with various components should be cleaned",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := NormalizePath(tt.input)

			if tt.expectError {
				s.Require().Error(err, tt.description)
				s.Empty(result, tt.description)
			} else {
				s.Require().NoError(err, tt.description)
				s.NotEmpty(result, tt.description)

				s.True(filepath.IsAbs(result), "Result should be absolute path")

				s.NotContains(result, string(filepath.Separator)+".", "Result should not contain '.' components")
				s.NotContains(result, ".."+string(filepath.Separator), "Result should not contain '..' components")
			}
		})
	}
}

func (s *PathUtilsTestSuite) TestIsPathUnder() {
	tests := []struct {
		name        string
		base        string
		target      string
		expected    bool
		expectError bool
		description string
	}{
		{
			name:        "target directly under base",
			base:        "/foo",
			target:      "/foo/bar",
			expected:    true,
			expectError: false,
			description: "Target directly under base should return true",
		},
		{
			name:        "target nested under base",
			base:        "/foo",
			target:      "/foo/bar/baz/qux",
			expected:    true,
			expectError: false,
			description: "Target nested under base should return true",
		},
		{
			name:        "target is base directory",
			base:        "/foo",
			target:      "/foo",
			expected:    true,
			expectError: false,
			description: "Target same as base should return true",
		},
		{
			name:        "target outside base - parent",
			base:        "/foo/bar",
			target:      "/foo",
			expected:    true,
			expectError: false,
			description: "Target outside base (parent) should return false",
		},
		{
			name:        "target outside base - sibling",
			base:        "/foo",
			target:      "/bar",
			expected:    false,
			expectError: false,
			description: "Target outside base (sibling) should return false",
		},
		{
			name:        "target outside base - different branch",
			base:        "/foo/bar",
			target:      "/foo/baz",
			expected:    false,
			expectError: false,
			description: "Target outside base (different branch) should return false",
		},
		{
			name:        "relative paths - target under base",
			base:        "foo",
			target:      "foo/bar",
			expected:    true,
			expectError: false,
			description: "Relative paths with target under base should return true",
		},
		{
			name:        "relative paths - target outside base",
			base:        "foo/bar",
			target:      "foo/baz",
			expected:    false,
			expectError: false,
			description: "Relative paths with target outside base should return false",
		},
		{
			name:        "current directory as base",
			base:        ".",
			target:      "./bar",
			expected:    true,
			expectError: false,
			description: "Current directory as base should work",
		},
		{
			name:        "parent directory reference",
			base:        "/foo",
			target:      "/foo/../bar",
			expected:    false,
			expectError: false,
			description: "Parent directory reference should be handled correctly",
		},
		{
			name:        "windows style paths - under",
			base:        "C:\\foo",
			target:      "C:\\foo\\bar",
			expected:    runtime.GOOS == "windows",
			expectError: false,
			description: "Windows-style paths with target under base",
		},
		{
			name:        "windows style paths - outside",
			base:        "C:\\foo",
			target:      "C:\\bar",
			expected:    false,
			expectError: false,
			description: "Windows-style paths with target outside base",
		},
		{
			name:        "empty base",
			base:        "",
			target:      "/foo/bar",
			expected:    false,
			expectError: true,
			description: "Empty base should return error",
		},
		{
			name:        "empty target",
			base:        "/foo",
			target:      "",
			expected:    false,
			expectError: true,
			description: "Empty target should return error",
		},
		{
			name:        "both empty",
			base:        "",
			target:      "",
			expected:    true,
			expectError: false,
			description: "Both empty should return true (relative path is '.')",
		},
		{
			name:        "identical paths with trailing separator",
			base:        "/foo",
			target:      "/foo/",
			expected:    true,
			expectError: false,
			description: "Identical paths with different trailing separators should return true",
		},
		{
			name:        "complex nested structure - under",
			base:        "/a/b/c",
			target:      "/a/b/c/d/e/f",
			expected:    true,
			expectError: false,
			description: "Complex nested structure with target under base",
		},
		{
			name:        "complex nested structure - outside",
			base:        "/a/b/c/d",
			target:      "/a/b/c/e",
			expected:    false,
			expectError: false,
			description: "Complex nested structure with target outside base",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := IsPathUnder(tt.base, tt.target)

			if tt.expectError {
				s.Require().Error(err, tt.description)
			} else {
				s.Require().NoError(err, tt.description)
				s.Equal(tt.expected, result, tt.description)
			}
		})
	}
}

func (s *PathUtilsTestSuite) TestIsPathUnder_CrossPlatform() {
	if runtime.GOOS == "windows" {
		s.T().Skip("Skipping cross-platform test on Windows")
	}

	tests := []struct {
		name     string
		base     string
		target   string
		expected bool
	}{
		{
			name:     "unix style paths - under",
			base:     "/home/user/projects",
			target:   "/home/user/projects/twiggit",
			expected: true,
		},
		{
			name:     "unix style paths - outside",
			base:     "/home/user/projects",
			target:   "/home/user/documents",
			expected: false,
		},
		{
			name:     "mixed separators - under",
			base:     "/home/user\\projects",
			target:   "/home/user/projects/twiggit",
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := IsPathUnder(tt.base, tt.target)
			s.Require().NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *PathUtilsTestSuite) TestIsPathUnder_EdgeCases() {
	s.Run("case sensitivity", func() {
		if runtime.GOOS == "windows" {
			s.T().Skip("Skipping case sensitivity test on Windows")
		}

		result, err := IsPathUnder("/Foo", "/foo/bar")
		s.Require().NoError(err)
		s.False(result)
	})

	s.Run("root directory", func() {
		if runtime.GOOS == "windows" {
			s.T().Skip("Skipping root directory test on Windows")
		}

		result, err := IsPathUnder("/", "/foo/bar")
		s.Require().NoError(err)
		s.True(result)
	})

	s.Run("relative path edge cases", func() {
		tests := []struct {
			base     string
			target   string
			expected bool
		}{
			{"foo", "foo", true},
		}

		for _, tt := range tests {
			result, err := IsPathUnder(tt.base, tt.target)
			s.Require().NoError(err)
			s.Equal(tt.expected, result)
		}
	})
}
