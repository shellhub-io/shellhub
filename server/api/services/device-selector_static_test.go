package services

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	servicesPackage = "github.com/shellhub-io/shellhub/server/api/services"
	storePackage    = "github.com/shellhub-io/shellhub/server/api/store"
	modelsPackage   = "github.com/shellhub-io/shellhub/pkg/models"
)

var deviceSelectorInputWriters = []string{
	"attachTag",
	"detachTag",
	"detachUnheldTag",
	"renameDevice",
	"applyProvisioningKeyTags",
	"mergeDevice",
	"reloadName",
}

func exportData(t *testing.T) map[string]string {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), "go", "list", "-export", "-deps", "-f", "{{.ImportPath}}={{.Export}}", ".").Output()
	require.NoError(t, err, "listing the export data of the package's dependencies")

	exports := map[string]string{}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		path, export, _ := strings.Cut(scanner.Text(), "=")
		exports[path] = export
	}

	return exports
}

func typeCheckedServices(t *testing.T) ([]*ast.File, *types.Info) {
	t.Helper()

	exports := exportData(t)
	fset := token.NewFileSet()

	sources, err := filepath.Glob("*.go")
	require.NoError(t, err)

	files := []*ast.File{}

	for _, source := range sources {
		if strings.HasSuffix(source, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, source, nil, parser.SkipObjectResolution)
		require.NoError(t, err)

		files = append(files, file)
	}

	lookup := func(path string) (io.ReadCloser, error) {
		return os.Open(exports[path])
	}

	info := &types.Info{Selections: map[*ast.SelectorExpr]*types.Selection{}, Uses: map[*ast.Ident]types.Object{}}
	config := types.Config{Importer: importer.ForCompiler(fset, "gc", lookup)}

	_, err = config.Check(servicesPackage, fset, files, info)
	require.NoError(t, err)

	return files, info
}

func isStoreTagLinkWrite(info *types.Info, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || (selector.Sel.Name != "TagPushToTarget" && selector.Sel.Name != "TagPullFromTarget") {
		return false
	}

	method := info.Uses[selector.Sel]

	return method != nil && method.Pkg() != nil && method.Pkg().Path() == storePackage
}

func isDeviceNameWrite(info *types.Info, lhs ast.Expr) bool {
	selector, ok := lhs.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Name" {
		return false
	}

	selection, ok := info.Selections[selector]
	if !ok || selection.Kind() != types.FieldVal {
		return false
	}

	receiver := selection.Recv()
	if pointer, isPointer := receiver.(*types.Pointer); isPointer {
		receiver = pointer.Elem()
	}

	named, ok := receiver.(*types.Named)

	return ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == modelsPackage && named.Obj().Name() == "Device"
}

func TestDeviceSelectorInputsAreWrittenOnlyThroughTheGuard(t *testing.T) {
	files, info := typeCheckedServices(t)

	writers := map[string]bool{}

	for _, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}

			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.CallExpr:
					if isStoreTagLinkWrite(info, node) {
						writers[function.Name.Name] = true
					}
				case *ast.AssignStmt:
					for _, lhs := range node.Lhs {
						if isDeviceNameWrite(info, lhs) {
							writers[function.Name.Name] = true
						}
					}
				}

				return true
			})
		}
	}

	found := make([]string, 0, len(writers))
	for writer := range writers {
		found = append(found, writer)
	}

	sort.Strings(found)

	expected := append([]string{}, deviceSelectorInputWriters...)
	sort.Strings(expected)

	assert.Equal(t, expected, found,
		"a device's tags and name are what access policies, public keys and firewall rules select by; "+
			"write them through the device selector guard, which checks the caller may edit those rules")
}
