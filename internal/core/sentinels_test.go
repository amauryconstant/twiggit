package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSentinels_LiveInSentinelsFile pins the location of the rebase
// sentinels: they MUST be declared in sentinels.go and not redeclared
// elsewhere in the package. Walks the AST so a copy/paste into
// errors.go fails the build.
func TestSentinels_LiveInSentinelsFile(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sentinels.go", nil, parser.AllErrors)
	require.NoError(t, err, "sentinels.go must parse")

	want := map[string]bool{
		"ErrRebaseConflict":   false,
		"ErrRebaseInProgress": false,
		"ErrBaseNotSet":       false,
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl) //nolint:exhaustive // sentinels live in GenDecls only
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec) //nolint:exhaustive // sentinels live in ValueSpecs only
			if !ok {
				continue
			}
			for _, ident := range vs.Names {
				if _, tracked := want[ident.Name]; tracked {
					want[ident.Name] = true
				}
			}
		}
	}

	for name, found := range want {
		assert.True(t, found, "%s must be declared in sentinels.go", name)
	}
}

// TestSentinels_NotRedeclaredElsewhere walks every other .go file in
// the package and asserts none of the rebase sentinels is declared a
// second time. The single-canonical-home invariant protects the
// errors.Is chain from drift if a future contributor copies the
// declaration into errors.go or another helper.
func TestSentinels_NotRedeclaredElsewhere(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go") &&
			!strings.HasSuffix(fi.Name(), "_test.go") &&
			fi.Name() != "sentinels.go"
	}, parser.AllErrors)
	require.NoError(t, err)

	forbidden := map[string]bool{
		"ErrRebaseConflict":   true,
		"ErrRebaseInProgress": true,
		"ErrBaseNotSet":       true,
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl) //nolint:exhaustive // sentinels live in GenDecls only
				if !ok || gen.Tok != token.VAR {
					continue
				}
				for _, spec := range gen.Specs {
					vs, ok := spec.(*ast.ValueSpec) //nolint:exhaustive // sentinels live in ValueSpecs only
					if !ok {
						continue
					}
					for _, ident := range vs.Names {
						assert.False(t, forbidden[ident.Name],
							"%s must not be redeclared in %s", ident.Name, f.Name)
					}
				}
			}
		}
	}
}
