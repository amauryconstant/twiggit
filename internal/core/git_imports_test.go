package core

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCoreGitRoles_ImportOnlyStdlib pins the depguard rule for the
// role-interface declaration file: git.go MUST import only Go
// standard-library packages. Any future change that pulls in
// internal/git/, internal/config/, etc. trips this test and forces a
// conscious decision (the depguard rule in .golangci.yml catches the
// same in CI; this test is the in-tree pin).
func TestCoreGitRoles_ImportOnlyStdlib(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "git.go", nil, parser.ImportsOnly)
	require.NoError(t, err)

	for _, imp := range file.Imports {
		path := imp.Path.Value // quoted, e.g. "context"
		assert.NotContains(t, path, "twiggit/",
			"git.go must not import any twiggit/* package (found %s)", path)
		assert.NotContains(t, path, "github.com/",
			"git.go must not import any non-stdlib module (found %s)", path)
	}
}
