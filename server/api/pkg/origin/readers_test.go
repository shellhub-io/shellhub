package origin

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var requestHeaderReaders = []string{
	"api/pkg/origin/origin.go:Resolve",
	"api/routes/mcp.go:mcpCredentialHeaders",
	"api/routes/middleware/authn.go:Resolve",
}

func TestOnlyDeclaredSitesNameAClientOrCredentialHeader(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)

	var found []string

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == "mocks" || entry.Name() == "testdata" {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		for _, site := range headerReadingSites(file) {
			found = append(found, filepath.ToSlash(rel)+":"+site)
		}

		return nil
	})
	require.NoError(t, err)

	slices.Sort(found)
	found = slices.Compact(found)

	declared := slices.Clone(requestHeaderReaders)
	slices.Sort(declared)

	assert.Equal(t, declared, found, "a site names X-Real-IP, Forwarded, an X-Forwarded-* or X-API-Key header, or calls echo's Scheme, without being declared in requestHeaderReaders, or a declared one no longer does; read where the client came from through this package")
}

func headerReadingSites(file *ast.File) []string {
	var sites []string

	for _, decl := range file.Decls {
		var (
			name string
			body ast.Node
		)

		switch decl := decl.(type) {
		case *ast.FuncDecl:
			name, body = decl.Name.Name, decl.Body
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok && namesAHeader(value) {
					for _, ident := range value.Names {
						sites = append(sites, ident.Name)
					}
				}
			}

			continue
		}

		if body != nil && namesAHeader(body) {
			sites = append(sites, name)
		}
	}

	return sites
}

func namesAHeader(node ast.Node) bool {
	named := false

	ast.Inspect(node, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.BasicLit:
			if node.Kind == token.STRING {
				if value, err := strconv.Unquote(node.Value); err == nil && isClientOrCredentialHeader(value) {
					named = true
				}
			}
		case *ast.SelectorExpr:
			if strings.HasPrefix(node.Sel.Name, "HeaderXForwarded") || node.Sel.Name == "HeaderXRealIP" || node.Sel.Name == "HeaderForwarded" {
				named = true
			}
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Scheme" && len(node.Args) == 0 {
				named = true
			}
		}

		return !named
	})

	return named
}

func isClientOrCredentialHeader(value string) bool {
	header := strings.ToLower(value)

	return header == "x-real-ip" || header == "forwarded" || header == "x-api-key" || strings.HasPrefix(header, "x-forwarded-")
}
