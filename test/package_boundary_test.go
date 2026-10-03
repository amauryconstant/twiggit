// Package boundary enforcement for the test/ directory. Verifies:
//  1. No IMMEDIATE child directory under test/ has a forbidden name
//     (helpers, util, common, misc, support). Nested paths like
//     test/e2e/helpers are fine — the rule applies to the top-level
//     test-helper boundary, not every nested subdirectory.
//  2. Each test/<domain>/ package name matches its directory name.
//  3. Each test/<domain>/ package is importable.
//
// Per the testing-helpers spec: "Test helper packages are content-named"
// — the forbidden list is enforced forever.
package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"twiggit/test/git"
	"twiggit/test/golden"
	"twiggit/test/perf"
	"twiggit/test/repo"
	"twiggit/test/shell"
	"twiggit/test/worktree"

	"github.com/stretchr/testify/require"
)

// forbiddenTestDirNames lists the test-package anti-patterns from the
// golang-project-layout skill ("util, helper packages say nothing
// about content — use specific names"). Enforced forever at the
// top-level test/ boundary.
var forbiddenTestDirNames = map[string]bool{
	"helpers": true,
	"util":    true,
	"common":  true,
	"misc":    true,
	"support": true,
}

// testPackageDirs lists the immediate test/<domain>/ directories that
// must follow the package-name = directory-name rule.
var testPackageDirs = []string{
	"git",
	"shell",
	"worktree",
	"repo",
	"golden",
	"perf",
	"helpers_test",
}

func TestNoForbiddenTopLevelTestHelperNames(t *testing.T) {
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if forbiddenTestDirNames[e.Name()] {
			t.Errorf("forbidden directory name %q under test/ (testing-helpers spec forbids helpers/util/common/misc/support at the top-level)", e.Name())
		}
	}
}

func TestTestPackageNamesMatchDirectories(t *testing.T) {
	for _, dir := range testPackageDirs {
		entries, err := os.ReadDir(dir)
		require.NoError(t, err, "missing test directory %q", dir)

		var pkgName string
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for ln := range strings.SplitSeq(string(b), "\n") {
				ln = strings.TrimSpace(ln)
				if ln == "" || strings.HasPrefix(ln, "//") {
					continue
				}
				if after, ok := strings.CutPrefix(ln, "package "); ok {
					pkgName = strings.TrimSpace(after)
				}
				break
			}
			if pkgName != "" {
				break
			}
		}
		require.NotEmpty(t, pkgName, "no package declaration found in any .go file under %q", dir)

		dirName := filepath.Base(dir)
		require.Equal(t, dirName, pkgName,
			"directory name %q != package name %q in %s", dirName, pkgName, dir)
	}
}

func TestEachTestPackageIsImportable(t *testing.T) {
	// Touch each import by referencing a type or zero construction from
	// each package. If the package does not import cleanly, the file
	// fails to compile and this test won't run.
	_ = git.GitTestHelper{}
	_ = golden.CompareGolden
	_ = perf.BenchmarkResult{}
	_ = repo.RepoTestHelper{}
	_ = shell.ShellTestHelper{}
	_ = worktree.WorktreeTestHelper{}
	// helpers_test is an external _test package; importing it here
	// requires a sibling non-test file too. Skip the explicit import
	// here — its importability is exercised by `go test ./test/helpers_test/...`.
}
