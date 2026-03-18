//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"

	"twiggit/test/e2e/fixtures"
	e2ehelpers "twiggit/test/e2e/helpers"
)

var _ = Describe("Golden file tests", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *e2ehelpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = e2ehelpers.NewTwiggitCLI()
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		fixture.Cleanup()
	})

	Context("list command golden tests", func() {
		It("matches golden file for text output with worktrees", func() {
			fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "list")
			cli.ShouldSucceed(session)

			output := cli.GetOutput(session)
			sanitized := sanitizeOutput(output, fixture.GetTempDir())
			compareGolden("list/basic_output.golden", sanitized)
		})

		It("matches golden file for empty project", func() {
			fixture.SetupSingleProject("empty-project")

			session := ctxHelper.FromProjectDir("empty-project", "list")
			cli.ShouldSucceed(session)

			output := cli.GetOutput(session)
			compareGolden("list/empty_output.golden", output)
		})
	})

	Context("list command JSON output golden tests", func() {
		It("matches golden file for JSON output", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "list", "--output", "json")
			cli.ShouldSucceed(session)

			output := cli.GetOutput(session)
			sanitized := sanitizeOutput(output, fixture.GetTempDir())
			compareGolden("list/json_output.golden", sanitized)
		})
	})

	Context("error golden tests", func() {
		It("matches golden file for validation errors", func() {
			fixture.SetupSingleProject("test-project")

			// Try to create worktree with invalid branch name
			session := ctxHelper.FromProjectDir("test-project", "create", "invalid@branch")
			cli.ShouldFailWithExit(session, 5) // Validation error

			// Get error output from stderr
			output := string(session.Err.Contents())
			compareGolden("errors/validation_error.golden", output)
		})

		It("matches golden file for not-found errors", func() {
			fixture.SetupSingleProject("test-project")

			// Try to delete non-existent worktree
			session := ctxHelper.FromProjectDir("test-project", "delete", "non-existent-worktree")
			cli.ShouldFailWithExit(session, 1) // General error

			// Get error output from stderr
			output := string(session.Err.Contents())
			sanitized := sanitizeOutput(output, fixture.GetTempDir())
			compareGolden("errors/not_found_error.golden", sanitized)
		})

		It("matches golden file for git errors", func() {
			fixture.SetupSingleProject("test-project")

			// Try to cd to non-existent worktree (will fail with error)
			session := ctxHelper.FromProjectDir("test-project", "cd", "non-existent-worktree")
			cli.ShouldFailWithExit(session, 1) // General error

			// Get error output from stderr
			output := string(session.Err.Contents())
			sanitized := sanitizeOutput(output, fixture.GetTempDir())
			compareGolden("errors/git_error.golden", sanitized)
		})
	})
})

// compareGolden compares actual output with golden file content for E2E tests.
// If UPDATE_GOLDEN environment variable is set to "true", the golden file will be updated.
func compareGolden(goldenFile string, actual string) {
	// Get project root (navigate up from test/e2e)
	goldenPath := filepath.Join("..", "..", "test", "golden", goldenFile)
	// Convert to absolute path
	absGoldenPath, err := filepath.Abs(goldenPath)
	if err != nil {
		GinkgoT().Fatalf("failed to get absolute path for golden file: %v", err)
	}
	goldenPath = absGoldenPath

	// Check if UPDATE_GOLDEN is set to true
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		updateGoldenFileE2E(goldenPath, actual)
		return
	}

	// Read golden file
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		GinkgoT().Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	// Normalize line endings and trim whitespace for cross-platform comparison
	// This matches the behavior of GetOutput which uses TrimSpace
	expectedStr := strings.TrimSpace(normalizeLineEndings(string(expected)))
	actualStr := normalizeLineEndings(actual)
	actualStr = strings.TrimSpace(actualStr)

	// Compare actual with expected
	if expectedStr != actualStr {
		// Generate diff
		diff := generateDiff(expectedStr, actualStr)
		GinkgoT().Errorf("output does not match golden file %s\n\n%s\n\nTo update golden file, run: UPDATE_GOLDEN=true go test ./...", goldenPath, diff)
	}
}

// updateGoldenFileE2E writes actual content to the golden file for E2E tests.
func updateGoldenFileE2E(goldenPath string, actual string) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
		GinkgoT().Fatalf("failed to create directory for golden file %s: %v", goldenPath, err)
	}

	// Write golden file
	if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
		GinkgoT().Fatalf("failed to write golden file %s: %v", goldenPath, err)
	}

	GinkgoT().Logf("updated golden file: %s", goldenPath)
}

// normalizeLineEndings converts all line endings to LF for consistent comparison.
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// generateDiff creates a human-readable diff between expected and actual strings.
func generateDiff(expected, actual string) string {
	var buf bytes.Buffer

	buf.WriteString("--- Expected\n")
	buf.WriteString("+++ Actual\n")

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	maxLines := max(len(expectedLines), len(actualLines))

	for i := 0; i < maxLines; i++ {
		expectedLine := ""
		actualLine := ""

		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine == actualLine {
			// Line matches - show as context
			if expectedLine != "" {
				fmt.Fprintf(&buf, "  %s\n", expectedLine)
			}
		} else {
			// Line differs
			if i < len(expectedLines) && expectedLine != "" {
				fmt.Fprintf(&buf, "- %s\n", expectedLine)
			}
			if i < len(actualLines) && actualLine != "" {
				fmt.Fprintf(&buf, "+ %s\n", actualLine)
			}
		}
	}

	return buf.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sanitizeOutput replaces temporary paths and branch names with stable placeholders
// This allows golden files to be stable across test runs
func sanitizeOutput(output, tempDir string) string {
	if tempDir == "" {
		return output
	}

	result := output

	// Replace the temp directory path with a placeholder
	// The tempDir is something like /tmp/ginkgo1234567890
	// We replace it with /tmp/fixtures to make it stable
	result = strings.ReplaceAll(result, tempDir, "/tmp/fixtures")

	// Replace random commit SHAs in branch names with stable placeholders
	// Branch names are like feature-1-be8e1572723bbe73
	// We use regex to match the SHA pattern (16 hex characters)
	shaPattern := regexp.MustCompile(`-[a-f0-9]{16}`)
	result = shaPattern.ReplaceAllString(result, "-<commit>")

	return result
}
