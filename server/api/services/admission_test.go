package services

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sourceUse struct {
	file string
	fn   string
}

func (u sourceUse) String() string {
	return u.file + ":" + u.fn
}

func scanPackageSource(t *testing.T, visit func(stack []ast.Node, use sourceUse) bool) []sourceUse {
	t.Helper()

	paths, err := filepath.Glob("*.go")
	require.NoError(t, err)

	var found []sourceUse

	fset := token.NewFileSet()
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		require.NoError(t, err)

		var stack []ast.Node
		fn := ""
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]

				return true
			}

			if decl, ok := n.(*ast.FuncDecl); ok {
				fn = decl.Name.Name
			}

			stack = append(stack, n)
			use := sourceUse{file: path, fn: fn}
			if visit(stack, use) {
				found = append(found, use)
			}

			return true
		})
	}

	return found
}

func isSelector(n ast.Node, pkg, name string) bool {
	sel, ok := n.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}

	if pkg == "" {
		return true
	}

	ident, ok := sel.X.(*ast.Ident)

	return ok && ident.Name == pkg
}

func TestAdmission_sourceRules(t *testing.T) {
	t.Run("only admitMember creates a membership", func(t *testing.T) {
		uses := scanPackageSource(t, func(stack []ast.Node, _ sourceUse) bool {
			return isSelector(stack[len(stack)-1], "", "NamespaceCreateMembership")
		})

		require.NotEmpty(t, uses)
		for _, use := range uses {
			assert.Equal(t, "admission.go:admitMember", use.String())
		}
	})

	t.Run("only the admission module builds a membership grant", func(t *testing.T) {
		uses := scanPackageSource(t, func(stack []ast.Node, _ sourceUse) bool {
			lit, ok := stack[len(stack)-1].(*ast.CompositeLit)
			if !ok {
				return false
			}

			ident, ok := lit.Type.(*ast.Ident)

			return ok && ident.Name == "membershipGrant"
		})

		require.NotEmpty(t, uses)
		for _, use := range uses {
			assert.Equal(t, "admission.go", use.file, use.String())
		}
	})
}
