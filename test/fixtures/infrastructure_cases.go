package fixtures

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GitRepositoryTestCase represents a test case for Git repository operations
type GitRepositoryTestCase struct {
	Name           string
	RepoPath       string
	IsBare         bool
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (string, error)
	CleanupFunc    func() error
}

// GitClientTestCase represents a test case for Git client operations
type GitClientTestCase struct {
	Name           string
	Command        string
	Args           []string
	WorkDir        string
	ExpectedError  error
	ExpectedOutput string
	ExpectedResult any
	SetupFunc      func() (string, error)
	CleanupFunc    func() error
}

// GitWorktreeTestCase represents a test case for Git worktree operations
type GitWorktreeTestCase struct {
	Name           string
	RepoPath       string
	WorktreePath   string
	BranchName     string
	Force          bool
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (string, string, error)
	CleanupFunc    func() error
}

// GitConfigTestCase represents a test case for Git configuration operations
type GitConfigTestCase struct {
	Name           string
	ConfigKey      string
	ConfigValue    string
	Global         bool
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (string, error)
	CleanupFunc    func() error
}

// GitBranchTestCase represents a test case for Git branch operations
type GitBranchTestCase struct {
	Name           string
	RepoPath       string
	BranchName     string
	SourceBranch   string
	Force          bool
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (string, error)
	CleanupFunc    func() error
}

// GetGitRepositoryTestCases returns test cases for Git repository operations
func GetGitRepositoryTestCases() []GitRepositoryTestCase {
	return []GitRepositoryTestCase{
		{
			Name:          "Create bare repository",
			RepoPath:      "/tmp/test-bare-repo",
			IsBare:        true,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"isBare":   true,
				"isValid":  true,
				"branches": []string{"main"},
			},
		},
		{
			Name:          "Create standard repository",
			RepoPath:      "/tmp/test-standard-repo",
			IsBare:        false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"isBare":   false,
				"isValid":  true,
				"branches": []string{"main"},
			},
		},
		{
			Name:           "Create repository in existing directory",
			RepoPath:       "/tmp/existing-dir",
			IsBare:         false,
			ExpectedError:  ErrRepositoryAlreadyExists,
			ExpectedResult: nil,
		},
		{
			Name:           "Create repository with invalid path",
			RepoPath:       "/invalid/path/repo",
			IsBare:         false,
			ExpectedError:  ErrInvalidPath,
			ExpectedResult: nil,
		},
		{
			Name:           "Create repository with insufficient permissions",
			RepoPath:       "/root/test-repo",
			IsBare:         false,
			ExpectedError:  ErrPermissionDenied,
			ExpectedResult: nil,
		},
	}
}

// GetGitClientTestCases returns test cases for Git client operations
func GetGitClientTestCases() []GitClientTestCase {
	return []GitClientTestCase{
		{
			Name:           "Execute git status",
			Command:        "git",
			Args:           []string{"status"},
			WorkDir:        "/tmp/test-repo",
			ExpectedError:  nil,
			ExpectedOutput: "On branch main",
			ExpectedResult: map[string]interface{}{
				"exitCode": 0,
				"success":  true,
			},
		},
		{
			Name:           "Execute git log",
			Command:        "git",
			Args:           []string{"log", "--oneline"},
			WorkDir:        "/tmp/test-repo",
			ExpectedError:  nil,
			ExpectedOutput: "commit message",
			ExpectedResult: map[string]interface{}{
				"exitCode": 0,
				"success":  true,
			},
		},
		{
			Name:           "Execute invalid git command",
			Command:        "git",
			Args:           []string{"invalid-command"},
			WorkDir:        "/tmp/test-repo",
			ExpectedError:  ErrCommandFailed,
			ExpectedOutput: "",
			ExpectedResult: map[string]interface{}{
				"exitCode": 1,
				"success":  false,
			},
		},
		{
			Name:           "Execute git command in non-repository",
			Command:        "git",
			Args:           []string{"status"},
			WorkDir:        "/tmp/not-a-repo",
			ExpectedError:  ErrNotRepository,
			ExpectedOutput: "",
			ExpectedResult: map[string]interface{}{
				"exitCode": 128,
				"success":  false,
			},
		},
		{
			Name:           "Execute git command with invalid working directory",
			Command:        "git",
			Args:           []string{"status"},
			WorkDir:        "/nonexistent/directory",
			ExpectedError:  ErrInvalidPath,
			ExpectedOutput: "",
			ExpectedResult: map[string]interface{}{
				"exitCode": -1,
				"success":  false,
			},
		},
	}
}

// GetGitWorktreeTestCases returns test cases for Git worktree operations
func GetGitWorktreeTestCases() []GitWorktreeTestCase {
	return []GitWorktreeTestCase{
		{
			Name:          "Create worktree from main branch",
			RepoPath:      "/tmp/test-repo",
			WorktreePath:  "/tmp/test-repo-worktrees/feature-1",
			BranchName:    "main",
			Force:         false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"created":  true,
				"path":     "/tmp/test-repo-worktrees/feature-1",
				"branch":   "main",
				"isLinked": true,
			},
		},
		{
			Name:          "Create worktree from feature branch",
			RepoPath:      "/tmp/test-repo",
			WorktreePath:  "/tmp/test-repo-worktrees/feature-2",
			BranchName:    "develop",
			Force:         false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"created":  true,
				"path":     "/tmp/test-repo-worktrees/feature-2",
				"branch":   "develop",
				"isLinked": true,
			},
		},
		{
			Name:          "Create worktree with force flag",
			RepoPath:      "/tmp/test-repo",
			WorktreePath:  "/tmp/test-repo-worktrees/feature-force",
			BranchName:    "main",
			Force:         true,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"created":  true,
				"path":     "/tmp/test-repo-worktrees/feature-force",
				"branch":   "main",
				"isLinked": true,
			},
		},
		{
			Name:           "Create worktree in existing directory",
			RepoPath:       "/tmp/test-repo",
			WorktreePath:   "/tmp/existing-directory",
			BranchName:     "main",
			Force:          false,
			ExpectedError:  ErrWorktreeAlreadyExists,
			ExpectedResult: nil,
		},
		{
			Name:           "Create worktree from non-existent branch",
			RepoPath:       "/tmp/test-repo",
			WorktreePath:   "/tmp/test-repo-worktrees/feature-invalid",
			BranchName:     "non-existent-branch",
			Force:          false,
			ExpectedError:  ErrBranchNotFound,
			ExpectedResult: nil,
		},
	}
}

// GetGitConfigTestCases returns test cases for Git configuration operations
func GetGitConfigTestCases() []GitConfigTestCase {
	return []GitConfigTestCase{
		{
			Name:          "Set user name config",
			ConfigKey:     "user.name",
			ConfigValue:   "Test User",
			Global:        false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"key":   "user.name",
				"value": "Test User",
				"set":   true,
			},
		},
		{
			Name:          "Set user email config",
			ConfigKey:     "user.email",
			ConfigValue:   "test@example.com",
			Global:        false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"key":   "user.email",
				"value": "test@example.com",
				"set":   true,
			},
		},
		{
			Name:          "Set global config",
			ConfigKey:     "core.editor",
			ConfigValue:   "vim",
			Global:        true,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"key":    "core.editor",
				"value":  "vim",
				"global": true,
				"set":    true,
			},
		},
		{
			Name:           "Set invalid config key",
			ConfigKey:      "invalid..key",
			ConfigValue:    "value",
			Global:         false,
			ExpectedError:  ErrInvalidConfigKey,
			ExpectedResult: nil,
		},
		{
			Name:           "Get non-existent config",
			ConfigKey:      "nonexistent.key",
			ConfigValue:    "",
			Global:         false,
			ExpectedError:  ErrConfigNotFound,
			ExpectedResult: nil,
		},
	}
}

// GetGitBranchTestCases returns test cases for Git branch operations
func GetGitBranchTestCases() []GitBranchTestCase {
	return []GitBranchTestCase{
		{
			Name:          "Create branch from main",
			RepoPath:      "/tmp/test-repo",
			BranchName:    "feature-1",
			SourceBranch:  "main",
			Force:         false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"created":       true,
				"name":          "feature-1",
				"source":        "main",
				"currentBranch": "main",
			},
		},
		{
			Name:          "Create branch from feature branch",
			RepoPath:      "/tmp/test-repo",
			BranchName:    "feature-2",
			SourceBranch:  "develop",
			Force:         false,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"created":       true,
				"name":          "feature-2",
				"source":        "develop",
				"currentBranch": "main",
			},
		},
		{
			Name:          "Create branch with force flag",
			RepoPath:      "/tmp/test-repo",
			BranchName:    "existing-branch",
			SourceBranch:  "main",
			Force:         true,
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"created":       true,
				"name":          "existing-branch",
				"source":        "main",
				"currentBranch": "main",
			},
		},
		{
			Name:           "Create existing branch without force",
			RepoPath:       "/tmp/test-repo",
			BranchName:     "existing-branch",
			SourceBranch:   "main",
			Force:          false,
			ExpectedError:  ErrBranchAlreadyExists,
			ExpectedResult: nil,
		},
		{
			Name:           "Create branch from non-existent source",
			RepoPath:       "/tmp/test-repo",
			BranchName:     "new-branch",
			SourceBranch:   "non-existent-source",
			Force:          false,
			ExpectedError:  ErrBranchNotFound,
			ExpectedResult: nil,
		},
	}
}

// Additional error types for infrastructure tests
var (
	ErrRepositoryAlreadyExists = errors.New("repository already exists")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrCommandFailed           = errors.New("command failed")
	ErrNotRepository           = errors.New("not a git repository")
	ErrWorktreeAlreadyExists   = errors.New("worktree already exists")
	ErrInvalidConfigKey        = errors.New("invalid config key")
	ErrConfigNotFound          = errors.New("config not found")
	ErrBranchAlreadyExists     = errors.New("branch already exists")
)

// Infrastructure-specific helper functions

// CreateTempDir creates a temporary directory for testing
func CreateTempDir() (string, error) {
	return os.MkdirTemp("", "twiggit-test-")
}

// CreateTempFile creates a temporary file for testing
func CreateTempFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "twiggit-test-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// CleanupTempDir removes a temporary directory and all its contents
func CleanupTempDir(path string) error {
	return os.RemoveAll(path)
}

// CleanupTempFile removes a temporary file
func CleanupTempFile(path string) error {
	return os.Remove(path)
}

// GetTestWorkingDir returns the current working directory for tests
func GetTestWorkingDir() (string, error) {
	return os.Getwd()
}

// IsWindows returns true if the current OS is Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsUnix returns true if the current OS is Unix-like
func IsUnix() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
}

// NormalizePath normalizes file paths for the current OS
func NormalizePath(path string) string {
	if IsWindows() {
		return filepath.FromSlash(path)
	}
	return path
}

// JoinPaths joins path components using the correct separator for the current OS
func JoinPaths(elements ...string) string {
	return filepath.Join(elements...)
}

// GetAbsolutePath returns the absolute path for the given path
func GetAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CreateDirectory creates a directory with all necessary parents
func CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// CreateFile creates a file with the given content
func CreateFile(path, content string) error {
	if err := CreateDirectory(filepath.Dir(path)); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFile reads the content of a file
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// AppendToFile appends content to a file
func AppendToFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// GetInfrastructureTestCasesByType returns test cases filtered by type
func GetInfrastructureTestCasesByType(testCaseType string) []any {
	switch testCaseType {
	case "repository":
		cases := GetGitRepositoryTestCases()
		result := make([]any, len(cases))
		for i, v := range cases {
			result[i] = v
		}
		return result
	case "client":
		cases := GetGitClientTestCases()
		result := make([]any, len(cases))
		for i, v := range cases {
			result[i] = v
		}
		return result
	case "worktree":
		cases := GetGitWorktreeTestCases()
		result := make([]any, len(cases))
		for i, v := range cases {
			result[i] = v
		}
		return result
	case "config":
		cases := GetGitConfigTestCases()
		result := make([]any, len(cases))
		for i, v := range cases {
			result[i] = v
		}
		return result
	case "branch":
		cases := GetGitBranchTestCases()
		result := make([]any, len(cases))
		for i, v := range cases {
			result[i] = v
		}
		return result
	default:
		return []any{}
	}
}

// GetInfrastructureTestCasesByError returns test cases filtered by expected error
func GetInfrastructureTestCasesByError(errorType error) []any {
	var result []any

	// Check repository test cases
	for _, tc := range GetGitRepositoryTestCases() {
		if errors.Is(tc.ExpectedError, errorType) {
			result = append(result, tc)
		}
	}

	// Check client test cases
	for _, tc := range GetGitClientTestCases() {
		if errors.Is(tc.ExpectedError, errorType) {
			result = append(result, tc)
		}
	}

	// Check worktree test cases
	for _, tc := range GetGitWorktreeTestCases() {
		if errors.Is(tc.ExpectedError, errorType) {
			result = append(result, tc)
		}
	}

	// Check config test cases
	for _, tc := range GetGitConfigTestCases() {
		if errors.Is(tc.ExpectedError, errorType) {
			result = append(result, tc)
		}
	}

	// Check branch test cases
	for _, tc := range GetGitBranchTestCases() {
		if errors.Is(tc.ExpectedError, errorType) {
			result = append(result, tc)
		}
	}

	return result
}

// GetInfrastructureTestCasesByName returns test cases filtered by name pattern
func GetInfrastructureTestCasesByName(namePattern string) []any {
	var result []any

	// Check repository test cases
	for _, tc := range GetGitRepositoryTestCases() {
		if strings.Contains(tc.Name, namePattern) {
			result = append(result, tc)
		}
	}

	// Check client test cases
	for _, tc := range GetGitClientTestCases() {
		if strings.Contains(tc.Name, namePattern) {
			result = append(result, tc)
		}
	}

	// Check worktree test cases
	for _, tc := range GetGitWorktreeTestCases() {
		if strings.Contains(tc.Name, namePattern) {
			result = append(result, tc)
		}
	}

	// Check config test cases
	for _, tc := range GetGitConfigTestCases() {
		if strings.Contains(tc.Name, namePattern) {
			result = append(result, tc)
		}
	}

	// Check branch test cases
	for _, tc := range GetGitBranchTestCases() {
		if strings.Contains(tc.Name, namePattern) {
			result = append(result, tc)
		}
	}

	return result
}
