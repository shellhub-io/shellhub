package routes

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const requestsPkgDir = "../../../pkg/api/requests"

type queryEmbeds struct {
	filters bool
	sorter  bool
}

func parsePackage(t *testing.T, dir string) []*ast.File {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		require.NoError(t, err)

		files = append(files, file)
	}

	return files
}

func selectorPath(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}

	return pkg.Name + "." + sel.Sel.Name
}

func embeddedQueryTypes(t *testing.T) map[string]queryEmbeds {
	t.Helper()

	embeds := make(map[string]queryEmbeds)

	for _, file := range parsePackage(t, requestsPkgDir) {
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structType, ok := spec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			var found queryEmbeds
			for _, field := range structType.Fields.List {
				if len(field.Names) > 0 {
					continue
				}

				switch selectorPath(field.Type) {
				case "query.Filters":
					found.filters = true
				case "query.Sorter":
					found.sorter = true
				}
			}

			if found.filters || found.sorter {
				embeds[spec.Name.Name] = found
			}

			return true
		})
	}

	return embeds
}

func isHandler(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || fn.Body == nil {
		return false
	}

	for _, param := range fn.Type.Params.List {
		if star, ok := param.Type.(*ast.StarExpr); ok && selectorPath(star.X) == "gateway.Context" {
			return true
		}
	}

	return false
}

func TestListQueryCasesCoverDeclaredRequests(t *testing.T) {
	embeds := embeddedQueryTypes(t)
	require.NotEmpty(t, embeds)
	covered := make(map[string]queryEmbeds)
	for _, tc := range listQueryCases {
		require.NotContains(t, covered, tc.name)
		covered[tc.name] = queryEmbeds{filters: tc.filter, sorter: tc.sort}
	}

	found := 0
	for _, file := range parsePackage(t, ".") {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isHandler(fn) {
				continue
			}
			var needs queryEmbeds
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if selector, ok := node.(*ast.SelectorExpr); ok {
					if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "requests" {
						embed := embeds[selector.Sel.Name]
						needs.filters = needs.filters || embed.filters
						needs.sorter = needs.sorter || embed.sorter
					}
				}

				return true
			})
			if needs.filters || needs.sorter {
				found++
				assert.Equal(t, needs, covered[fn.Name.Name], "%s needs a behavioral rejection case for each query type", fn.Name.Name)
			}
		}
	}
	require.Positive(t, found)
}
