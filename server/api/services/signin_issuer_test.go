package services

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnlyTheSignInModuleIssuesUserTokens(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)

	issuer := filepath.Join(root, "server", "api", "services", "signin.go")
	issuing := map[string]bool{"EncodeUserClaims": true, "mintUserToken": true, "issueSignIn": true}

	var outside []string

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			switch entry.Name() {
			case "node_modules", "ui", ".git":
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || path == issuer {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok && issuing[selector.Sel.Name] {
				outside = append(outside, path+": "+selector.Sel.Name)
			}

			return true
		})

		return nil
	})
	require.NoError(t, err)

	assert.Empty(t, outside, "user tokens are issued only by server/api/services/signin.go")
}
