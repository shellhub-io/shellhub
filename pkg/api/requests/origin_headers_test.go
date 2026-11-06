package requests

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoRequestBindsWhereTheClientCameFrom(t *testing.T) {
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	var bound []string

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.SkipObjectResolution)
		require.NoError(t, err)

		ast.Inspect(file, func(node ast.Node) bool {
			field, ok := node.(*ast.Field)
			if !ok || field.Tag == nil {
				return true
			}

			raw, err := strconv.Unquote(field.Tag.Value)
			require.NoError(t, err)

			header := strings.ToLower(reflect.StructTag(raw).Get("header"))
			if header == "x-real-ip" || header == "forwarded" || strings.HasPrefix(header, "x-forwarded-") {
				bound = append(bound, name+": "+header)
			}

			return true
		})
	}

	assert.Empty(t, bound, "where the client came from is the router's to resolve, through server/api/pkg/origin")
}
