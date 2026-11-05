// Package bindingtest applies the [binding] rule to the source of a package, so that a request type
// no route test reaches still fails a build.
package bindingtest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/binding"
)

// CheckDir returns one message per exported field, of a struct type declared in a non-test file of
// dir, that [binding.SourceOf] refuses, plus one per source tag inside a struct the body fills. An
// embedded field with no json tag is left to the package that declares its type. It returns an
// error when dir cannot be read or a file in it cannot be parsed.
func CheckDir(dir string) ([]string, error) {
	return checkStructs(dir, false, func(*ast.StructType) bool { return true })
}

// CheckTree returns, for every package under root, the messages [CheckDir] would for each struct
// type that names a source of [binding.Sources] on any field, declared at package level or inside a
// function. A struct that names none is not a request, so the rule does not reach it. Directories
// named testdata, vendor or node_modules, or starting with a dot, are skipped. It returns an error
// when a directory cannot be read or parsed.
func CheckTree(root string) ([]string, error) {
	var problems []string

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if path != root && (strings.HasPrefix(name, ".") || slices.Contains([]string{"testdata", "vendor", "node_modules"}, name)) {
			return filepath.SkipDir
		}

		found, err := checkStructs(path, true, namesASource)
		problems = append(problems, found...)

		return err
	})

	return problems, err
}

func checkStructs(dir string, anonymous bool, want func(*ast.StructType) bool) ([]string, error) {
	fset := token.NewFileSet()

	files, err := parseDir(fset, dir)
	if err != nil {
		return nil, err
	}

	var problems []string
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			var (
				name string
				st   *ast.StructType
			)

			switch n := node.(type) {
			case *ast.TypeSpec:
				inner, ok := n.Type.(*ast.StructType)
				if !ok {
					return true
				}

				name, st = n.Name.Name, inner
			case *ast.StructType:
				if !anonymous {
					return true
				}

				name, st = "struct", n
			default:
				return true
			}

			if want(st) {
				problems = append(problems, checkStruct(fset, name, st)...)
			}

			return false
		})
	}

	return problems, nil
}

func parseDir(fset *token.FileSet, dir string) ([]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []*ast.File
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

func namesASource(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if len(binding.NamedSources(tagOf(field))) > 0 {
			return true
		}
	}

	return false
}

func checkStruct(fset *token.FileSet, path string, st *ast.StructType) []string {
	var problems []string
	for _, field := range st.Fields.List {
		tag := tagOf(field)

		if len(field.Names) == 0 {
			if _, tagged := tag.Lookup("json"); !tagged {
				continue
			}
		}

		for _, name := range fieldNames(field) {
			if !ast.IsExported(name) {
				continue
			}

			fieldPath := path + "." + name

			source, err := binding.SourceOf(tag)
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s: %s: %v", fset.Position(field.Pos()), fieldPath, err))

				continue
			}

			if inline, ok := field.Type.(*ast.StructType); ok && source.IsBody() {
				problems = append(problems, sourceTagsInside(fset, fieldPath, inline)...)
			}
		}
	}

	return problems
}

func sourceTagsInside(fset *token.FileSet, path string, st *ast.StructType) []string {
	var problems []string
	for _, field := range st.Fields.List {
		for _, source := range binding.NamedSources(tagOf(field)) {
			problems = append(problems, fmt.Sprintf("%s: %s: %s tag inside a body field", fset.Position(field.Pos()), path, source.Tag))
		}

		if inline, ok := field.Type.(*ast.StructType); ok {
			problems = append(problems, sourceTagsInside(fset, path, inline)...)
		}
	}

	return problems
}

func fieldNames(field *ast.Field) []string {
	if len(field.Names) == 0 {
		return []string{embeddedName(field.Type)}
	}

	names := make([]string, 0, len(field.Names))
	for _, ident := range field.Names {
		names = append(names, ident.Name)
	}

	return names
}

func embeddedName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return embeddedName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	default:
		return ""
	}
}

func tagOf(field *ast.Field) reflect.StructTag {
	if field.Tag == nil {
		return ""
	}

	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}

	return reflect.StructTag(raw)
}
