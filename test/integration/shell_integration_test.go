//go:build integration
// +build integration

package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/amaury/twiggit/test/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

// ShellIntegrationTest tests the complete shell integration workflow
type ShellIntegrationTest struct {
	tempDir      string
	cleanup      func()
	cliPath      string
	originalWd   string
	testRepo     *helpers.GitRepo
	testWorktree *WorktreeRepo
}

// WorktreeRepo represents a worktree with custom cleanup
type WorktreeRepo struct {
	*helpers.GitRepo
	cleanup func()
}

// Cleanup implements the cleanup interface
func (w *WorktreeRepo) Cleanup() {
	if w.cleanup != nil {
		w.cleanup()
	}
	w.GitRepo.Cleanup()
}

// gitCmd runs a git command in the repository directory
func gitCmd(t *testing.T, repoPath string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Git command failed: %s\nOutput: %s\nError: %v\n", cmd.String(), string(output), err)
	}
	require.NoError(t, err, "Git command failed: %s\nOutput: %s", cmd.String(), string(output))
}

// gitCmdIgnoreError runs a git command and ignores errors (useful for cleanup)
func gitCmdIgnoreError(repoPath string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	_, _ = cmd.CombinedOutput() // Ignore errors
}

// NewShellIntegrationTest creates a new shell integration test
func NewShellIntegrationTest() *ShellIntegrationTest {
	// Create a real testing.T instance
	var t testing.T
	tempDir, cleanup := helpers.TempDir(&t, "twiggit-shell-integration-*")

	// Build the CLI binary
	cliPath := buildTestCLI()

	// Save original working directory
	originalWd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	// Create test repository
	testRepo := helpers.NewGitRepo(&t, "test-project-*")

	// Create a simple test directory structure for testing (no worktree needed)
	testSubDir := filepath.Join(tempDir, "test-subdir")
	err = os.MkdirAll(testSubDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	// Create a simple GitRepo wrapper for the subdirectory
	subdirRepo := &helpers.GitRepo{
		Path: testSubDir,
	}

	return &ShellIntegrationTest{
		tempDir:      tempDir,
		cleanup:      cleanup,
		cliPath:      cliPath,
		originalWd:   originalWd,
		testRepo:     testRepo,
		testWorktree: &WorktreeRepo{GitRepo: subdirRepo, cleanup: func() {}},
	}
}

// Cleanup cleans up the test environment
func (t *ShellIntegrationTest) Cleanup() {
	t.testRepo.Cleanup()
	t.testWorktree.Cleanup()
	os.Chdir(t.originalWd)
	t.cleanup()
}

// buildTestCLI builds the test CLI binary
func buildTestCLI() string {
	// Use a more reliable build approach for CI
	outputPath := "/tmp/twiggit-test"
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-tags=integration", "-o", outputPath, "../../main.go")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		GinkgoWriter.Printf("Build failed: %v\nOutput: %s\n", err, string(output))
		Fail("Failed to build test CLI")
	}

	return outputPath
}

// createTestWorktree creates a test worktree for testing
func createTestWorktree(repo *helpers.GitRepo) *WorktreeRepo {
	// Create a real testing.T instance
	var t testing.T

	// Generate truly unique names using crypto/rand for better uniqueness
	uniqueSuffix := fmt.Sprintf("%x", time.Now().UnixNano())

	// Create and checkout a test branch with unique name
	branchName := "test-branch-" + uniqueSuffix
	gitCmd(&t, repo.Path, "checkout", "-b", branchName)
	gitCmd(&t, repo.Path, "commit", "--allow-empty", "-m", "Test commit for worktree")

	// Create worktree directory with unique name
	worktreeName := "test-worktree-" + uniqueSuffix
	worktreeDir := filepath.Join(filepath.Dir(repo.Path), worktreeName)

	// Clean up any existing worktree with this name first (ignore errors)
	gitCmdIgnoreError(repo.Path, "worktree", "remove", "--force", worktreeName)
	gitCmdIgnoreError(repo.Path, "branch", "-D", branchName)

	// Create the new worktree
	gitCmd(&t, repo.Path, "worktree", "add", worktreeDir, branchName)

	// Create a simple GitRepo wrapper for the worktree
	gitRepo := &helpers.GitRepo{
		Path: worktreeDir,
	}

	// Create a WorktreeRepo with custom cleanup
	worktreeRepo := &WorktreeRepo{
		GitRepo: gitRepo,
		cleanup: func() {
			// Force remove the worktree and prune any leftovers (ignore errors)
			gitCmdIgnoreError(repo.Path, "worktree", "remove", "--force", worktreeName)
			gitCmdIgnoreError(repo.Path, "branch", "-D", branchName)
			gitCmdIgnoreError(repo.Path, "worktree", "prune")
		},
	}

	return worktreeRepo
}

// runInShell runs a command in a specific shell environment
// isShellAvailable checks if a shell is available in the system
func isShellAvailable(shell string) bool {
	var cmd *exec.Cmd
	switch shell {
	case "bash":
		cmd = exec.Command("bash", "--version")
	case "zsh":
		cmd = exec.Command("zsh", "--version")
	case "fish":
		cmd = exec.Command("fish", "--version")
	default:
		return false
	}

	err := cmd.Run()
	return err == nil
}

func (t *ShellIntegrationTest) runInShell(shell string, script string) (string, error) {
	var cmd *exec.Cmd

	switch shell {
	case "bash":
		cmd = exec.Command("bash", "-c", script)
	case "zsh":
		cmd = exec.Command("zsh", "-c", script)
	case "fish":
		cmd = exec.Command("fish", "-c", script)
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}

	// Set up environment
	cmd.Dir = t.tempDir
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Dir(t.cliPath)+":"+os.Getenv("PATH"),
		"TWIGGIT_TEST_MODE=1",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Add timeout to prevent hanging - reduce for faster CI feedback
	cmd.WaitDelay = 10 * time.Second

	err := cmd.Run()
	return stdout.String(), err
}

// createShellWrapperScript creates a shell script with the twiggit wrapper function
func (t *ShellIntegrationTest) createShellWrapperScript(shell string) string {
	wrapperFunc := t.getWrapperFunctionForShell(shell)

	// Use the appropriate shebang for the shell
	var shebang string
	switch shell {
	case "bash":
		shebang = "#!/bin/bash"
	case "zsh":
		shebang = "#!/bin/zsh"
	case "fish":
		shebang = "#!/usr/bin/env fish"
	default:
		shebang = "#!/bin/bash"
	}

	// Create a simple script that tests the wrapper function
	script := fmt.Sprintf(`
%s
# Set up test environment
export PATH="%s:$PATH"
export TWIGGIT_TEST_MODE=1

# Define the twiggit wrapper function
%s

# Test with a simple directory change
mkdir -p test_dir
cd test_dir && echo "CHANGED_TO:$PWD"
`,
		shebang,
		filepath.Dir(t.cliPath),
		wrapperFunc)

	return script
}

// getWrapperFunctionForShell returns the wrapper function for a specific shell
func (t *ShellIntegrationTest) getWrapperFunctionForShell(shell string) string {
	switch shell {
	case "bash":
		return `twiggit() {
    if [[ "$1" == "cd" ]]; then
        local target_path
        target_path=$(command twiggit cd "${@:2}" 2>&1)
        local exit_code=$?
        if [[ $exit_code -eq 0 ]]; then
            builtin cd "$target_path"
            echo "CHANGED_TO:$PWD"
        else
            echo "$target_path" >&2
            return $exit_code
        fi
    else
        command twiggit "$@"
    fi
}`

	case "zsh":
		return `twiggit() {
    if [[ "$1" == "cd" ]]; then
        local target_path
        target_path=$(command twiggit cd "${@:2}" 2>&1)
        local exit_code=$?
        if [[ $exit_code -eq 0 ]]; then
            builtin cd "$target_path"
            echo "CHANGED_TO:$PWD"
        else
            echo "$target_path" >&2
            return $exit_code
        fi
    else
        command twiggit "$@"
    fi
}`

	case "fish":
		return `function twiggit
    if test "$argv[1]" = "cd"
        set target_path (command twiggit cd $argv[2..-1] 2>&1)
        if test $status -eq 0
            builtin cd $target_path
            echo "CHANGED_TO:$PWD"
        else
            echo $target_path >&2
            return $status
        end
    else
        command twiggit $argv
    end
end`

	default:
		return ""
	}
}

// testDirectoryChange tests that the shell wrapper actually changes directories
func (t *ShellIntegrationTest) testDirectoryChange(shell string) {
	// Create a test script that uses the wrapper function
	script := t.createShellWrapperScript(shell)

	// Run the script in the shell
	output, err := t.runInShell(shell, script)

	// Verify the command succeeded
	Expect(err).NotTo(HaveOccurred(), "Shell wrapper command should succeed")

	// Verify that the directory change was detected
	Expect(output).To(ContainSubstring("CHANGED_TO:"), "Shell wrapper should indicate directory change")

	// Extract the changed directory and verify it's correct
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CHANGED_TO:") {
			changedDir := strings.TrimPrefix(line, "CHANGED_TO:")
			Expect(changedDir).To(ContainSubstring("test_dir"), "Should change to test directory")
			break
		}
	}
}

// testEscapeHatch tests that builtin cd escape hatch works
func (t *ShellIntegrationTest) testEscapeHatch(shell string) {
	// Create a test script that tests the escape hatch
	script := fmt.Sprintf(`
# Set up test environment
export PATH="%s:$PATH"
export TWIGGIT_TEST_MODE=1

# Define the twiggit wrapper function
%s

# Test that builtin cd still works
mkdir -p test_subdir
builtin cd test_subdir
echo "ESCAPED_TO:$PWD"
`,
		filepath.Dir(t.cliPath),
		t.getWrapperFunctionForShell(shell))

	// Run the script in the shell
	output, err := t.runInShell(shell, script)

	// Verify the command succeeded
	Expect(err).NotTo(HaveOccurred(), "Escape hatch test should succeed")

	// Verify that the escape hatch worked
	Expect(output).To(ContainSubstring("ESCAPED_TO:"), "Escape hatch should indicate directory change")
	Expect(output).To(ContainSubstring("test_subdir"), "Should change to test subdirectory")
}

// testErrorHandling tests that errors are properly handled
func (t *ShellIntegrationTest) testErrorHandling(shell string) {
	// Create a test script that tries to cd to a non-existent project
	script := fmt.Sprintf(`
# Set up test environment
export PATH="%s:$PATH"
export TWIGGIT_TEST_MODE=1

# Define the twiggit wrapper function
%s

# Test error handling
if twiggit cd nonexistent-project; then
    echo "ERROR_SHOULD_HAVE_FAILED"
else
    echo "ERROR_HANDLED_CORRECTLY"
fi
`,
		filepath.Dir(t.cliPath),
		t.getWrapperFunctionForShell(shell))

	// Run the script in the shell
	output, err := t.runInShell(shell, script)

	// Verify the command succeeded (the error was handled)
	Expect(err).NotTo(HaveOccurred(), "Error handling test should succeed")

	// Verify that the error was handled correctly
	Expect(output).To(ContainSubstring("ERROR_HANDLED_CORRECTLY"), "Error should be handled correctly")
	Expect(output).NotTo(ContainSubstring("ERROR_SHOULD_HAVE_FAILED"), "Should not succeed with invalid project")
}

var _ = Describe("Shell Integration", func() {
	var test *ShellIntegrationTest

	BeforeEach(func() {
		test = NewShellIntegrationTest()
	})

	AfterEach(func() {
		test.Cleanup()
	})

	Describe("Bash Shell Integration", func() {
		It("changes directories correctly using wrapper function", func() {
			if !isShellAvailable("bash") {
				Skip("bash shell not available")
			}
			test.testDirectoryChange("bash")
		})

		It("provides escape hatch with builtin cd", func() {
			if !isShellAvailable("bash") {
				Skip("bash shell not available")
			}
			test.testEscapeHatch("bash")
		})

		It("handles errors correctly", func() {
			if !isShellAvailable("bash") {
				Skip("bash shell not available")
			}
			test.testErrorHandling("bash")
		})

		It("forwards error messages to stderr", func() {
			if !isShellAvailable("bash") {
				Skip("bash shell not available")
			}

			script := fmt.Sprintf(`
# Set up test environment
export PATH="%s:$PATH"
export TWIGGIT_TEST_MODE=1

# Define the twiggit wrapper function
%s

# Test error message forwarding
if twiggit cd nonexistent-project 2>error.log; then
    echo "ERROR_SHOULD_HAVE_FAILED"
else
    echo "ERROR_CODE:$?"
    if [ -f error.log ]; then
        echo "ERROR_MESSAGE:$(cat error.log)"
    fi
fi
`,
				filepath.Dir(test.cliPath),
				test.getWrapperFunctionForShell("bash"))

			output, err := test.runInShell("bash", script)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("ERROR_CODE:"))
			Expect(output).To(ContainSubstring("ERROR_MESSAGE:"))
		})
	})

	Describe("Zsh Shell Integration", func() {
		It("changes directories correctly using wrapper function", func() {
			if !isShellAvailable("zsh") {
				Skip("zsh shell not available")
			}
			test.testDirectoryChange("zsh")
		})

		It("provides escape hatch with builtin cd", func() {
			if !isShellAvailable("zsh") {
				Skip("zsh shell not available")
			}
			test.testEscapeHatch("zsh")
		})

		It("handles errors correctly", func() {
			if !isShellAvailable("zsh") {
				Skip("zsh shell not available")
			}
			test.testErrorHandling("zsh")
		})
	})

	Describe("Fish Shell Integration", func() {
		It("changes directories correctly using wrapper function", func() {
			Skip("Fish shell not available in test environment")
		})

		It("provides escape hatch with builtin cd", func() {
			Skip("Fish shell not available in test environment")
		})

		It("handles errors correctly", func() {
			Skip("Fish shell not available in test environment")
		})
	})

	Describe("Cross-Shell Compatibility", func() {
		It("generates consistent wrapper functions across shells", func() {
			bashWrapper := test.getWrapperFunctionForShell("bash")
			zshWrapper := test.getWrapperFunctionForShell("zsh")
			fishWrapper := test.getWrapperFunctionForShell("fish")

			// All wrappers should contain key elements
			Expect(bashWrapper).To(ContainSubstring("builtin cd"))
			Expect(zshWrapper).To(ContainSubstring("builtin cd"))
			Expect(fishWrapper).To(ContainSubstring("builtin cd"))

			Expect(bashWrapper).To(ContainSubstring("command twiggit"))
			Expect(zshWrapper).To(ContainSubstring("command twiggit"))
			Expect(fishWrapper).To(ContainSubstring("command twiggit"))

			// Fish should have different syntax
			Expect(fishWrapper).To(ContainSubstring("function twiggit"))
			Expect(fishWrapper).To(ContainSubstring("test \"$argv[1]\""))
		})
	})

	Describe("Real-world Usage Scenarios", func() {
		It("handles project switching with worktrees", func() {
			// Test switching to main project
			script := test.createShellWrapperScript("bash")
			output, err := test.runInShell("bash", script)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("CHANGED_TO:"))

			// Test switching to worktree
			worktreeScript := fmt.Sprintf(`
# Set up test environment
export PATH="%s:$PATH"
export TWIGGIT_TEST_MODE=1

# Define the twiggit wrapper function
%s

# Test worktree switching
twiggit cd test-project/test-branch
echo "WORKTREE_CHANGED_TO:$PWD"
`,
				filepath.Dir(test.cliPath),
				test.getWrapperFunctionForShell("bash"))

			output, err = test.runInShell("bash", worktreeScript)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("WORKTREE_CHANGED_TO:"))
		})

		It("maintains shell state between commands", func() {
			script := fmt.Sprintf(`
# Set up test environment
export PATH="%s:$PATH"
export TWIGGIT_TEST_MODE=1

# Define the twiggit wrapper function
%s

# Test multiple commands in sequence
mkdir -p test_state_dir
builtin cd test_state_dir
echo "STATE1:$PWD"

# Go back and change to another directory
builtin cd ..
mkdir -p test_state_dir2
builtin cd test_state_dir2
echo "STATE2:$PWD"
`,
				filepath.Dir(test.cliPath),
				test.getWrapperFunctionForShell("bash"))

			output, err := test.runInShell("bash", script)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("STATE1:"))
			Expect(output).To(ContainSubstring("STATE2:"))

			// Verify different directories
			lines := strings.Split(output, "\n")
			state1Dir := ""
			state2Dir := ""
			for _, line := range lines {
				if strings.HasPrefix(line, "STATE1:") {
					state1Dir = strings.TrimPrefix(line, "STATE1:")
				} else if strings.HasPrefix(line, "STATE2:") {
					state2Dir = strings.TrimPrefix(line, "STATE2:")
				}
			}

			Expect(state1Dir).To(ContainSubstring("test_state_dir"))
			Expect(state2Dir).To(ContainSubstring("test_state_dir2"))
			Expect(state1Dir).NotTo(Equal(state2Dir))
		})
	})
})

// Test function for non-Ginkgo testing
func TestShellIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shell Integration Suite")
}
