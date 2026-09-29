package routes_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var bodylessSuccesses = map[string]bool{
	"StatusOK": true, "200": true,
	"StatusCreated": true, "201": true,
	"StatusAccepted": true, "202": true,
	"StatusNoContent": true, "204": true,
}

const errorHandlerRenderingEveryRefusal = "api/pkg/echo/handlers/errors.go"

func answersASuccess(arg ast.Expr) bool {
	switch status := arg.(type) {
	case *ast.SelectorExpr:
		return bodylessSuccesses[status.Sel.Name]
	case *ast.BasicLit:
		return bodylessSuccesses[status.Value]
	default:
		return false
	}
}

func isModuleRoot(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))

	return err == nil
}

func bodylessRefusals(t *testing.T) map[string]int {
	t.Helper()

	module := filepath.Join("..", "..")
	found := make(map[string]int)
	fset := token.NewFileSet()

	err := filepath.WalkDir(module, func(path string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case entry.IsDir() && path != module && isModuleRoot(path):
			return fs.SkipDir
		case entry.IsDir(), !strings.HasSuffix(path, ".go"), strings.HasSuffix(path, "_test.go"):
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Logf("skipping %s, which does not parse: %v", path, err)

			return nil
		}

		relative, err := filepath.Rel(module, path)
		require.NoError(t, err, "locating %s inside the module", path)

		if filepath.ToSlash(relative) == errorHandlerRenderingEveryRefusal {
			return nil
		}

		ast.Inspect(file, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall || len(call.Args) != 1 {
				return true
			}

			method, isMethod := call.Fun.(*ast.SelectorExpr)
			if !isMethod || method.Sel.Name != "NoContent" || answersASuccess(call.Args[0]) {
				return true
			}

			found[filepath.ToSlash(relative)]++

			return true
		})

		return nil
	})
	require.NoError(t, err, "walking the server module")

	return found
}

// TestEveryRefusalCarriesABody fails when anything in the server module answers a refusal with no
// body. A refusal returns an error so the error handler renders one, and c.NoContent is left to
// the successes that genuinely carry nothing. It walks the whole module rather than this package,
// because the SSH surface registers its own handlers on the same router.
func TestEveryRefusalCarriesABody(t *testing.T) {
	found := bodylessRefusals(t)

	unexpected := make([]string, 0, len(found))
	for path := range found {
		unexpected = append(unexpected, path)
	}

	sort.Strings(unexpected)

	assert.Empty(t, unexpected,
		"these files refuse a request without a body; return an error instead so the handler renders one")
}
