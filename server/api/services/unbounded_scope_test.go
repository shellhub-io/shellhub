package services

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var unboundedScopeSites = []string{
	"api/pkg/gateway/context.go:AdminOrScope",
	"api/pkg/gateway/route.go:resolveScope",
	"api/services/auth.go:AuthAPIKey",
	"api/services/auth.go:AuthDeviceToken",
	"api/services/auth.go:provisioningKeyTenant",
	"api/services/device-pairing.go:acceptedDevicePairing",
	"api/services/heartbeat.go:flush",
	"api/services/session.go:CreateSession",
	"api/services/task.go:DevicePresenceSweep",
	"api/services/task.go:deviceCleanup",
	"api/services/task.go:ephemeralCleanup",
	"api/store/pg/session.go:SessionCreate",
	"ssh/http/handlers.go:resolveDevice",
	"ssh/session/session.go:NewSession",
	"ssh/web/session.go:getAuth",
}

var unboundedScope = mock.MatchedBy(func(sc scope.Scope) bool { return sc.Kind() == scope.KindUnbounded })

func TestUnboundedScopesAreBuiltOnlyAtDeclaredSites(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	var found []string

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == "mocks" || entry.Name() == "storetest" {
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

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok && isUnboundedScopeCall(call) {
					found = append(found, filepath.ToSlash(rel)+":"+fn.Name.Name)
				}

				return true
			})
		}

		return nil
	})
	require.NoError(t, err)

	slices.Sort(found)
	found = slices.Compact(found)

	declared := slices.Clone(unboundedScopeSites)
	slices.Sort(declared)

	assert.Equal(t, declared, found, "a function builds a scope spanning every namespace without being declared in unboundedScopeSites, or a declared one no longer does")
}

func isUnboundedScopeCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "NewUnbounded" {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)

	return ok && pkg.Name == "scope"
}
