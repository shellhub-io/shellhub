package routes

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func handlerQueryUsage(fn *ast.FuncDecl, embeds map[string]queryEmbeds) (queryEmbeds, map[string]bool) {
	var needs queryEmbeds
	calls := make(map[string]bool)

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr:
			if pkg, ok := node.X.(*ast.Ident); ok && pkg.Name == "requests" {
				if e, ok := embeds[node.Sel.Name]; ok {
					needs.filters = needs.filters || e.filters
					needs.sorter = needs.sorter || e.sorter
				}
			}
		case *ast.CallExpr:
			if path := selectorPath(node.Fun); path != "" {
				calls[path] = true
			}
		}

		return true
	})

	return needs, calls
}

func TestListHandlersValidateEveryQueryFieldTheyAccept(t *testing.T) {
	embeds := embeddedQueryTypes(t)
	require.NotEmpty(t, embeds)

	for _, file := range parsePackage(t, ".") {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isHandler(fn) {
				continue
			}

			needs, calls := handlerQueryUsage(fn, embeds)

			if needs.filters && !calls["query.ValidateFilters"] {
				t.Errorf(
					"%s binds a request carrying a client filter but never calls query.ValidateFilters: "+
						"the filter then names any column on the table, including ones the response omits",
					fn.Name.Name,
				)
			}

			if needs.sorter && !calls["query.ValidateSorter"] {
				t.Errorf(
					"%s binds a request carrying a client sort_by but never calls query.ValidateSorter: "+
						"the sort then orders by any column on the table, including ones the response omits",
					fn.Name.Name,
				)
			}
		}
	}
}
