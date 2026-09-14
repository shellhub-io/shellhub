package routes_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var pathParam = regexp.MustCompile(`\{[^}]+\}`)

func normalizePath(path string, params []string) string {
	for _, p := range params {
		path = strings.ReplaceAll(path, ":"+p, "{}")
	}

	return pathParam.ReplaceAllString(path, "{}")
}

const communityEntrypoint = "community-openapi.yaml"

func specEndpoints(t *testing.T) map[string]bool {
	t.Helper()

	dir := filepath.Join("..", "..", "..", "openapi", "spec")
	endpoints := make(map[string]bool)

	methodsOf := func(name string) []string {
		content, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // name comes from this repository's own spec entrypoints, not from input
		require.NoError(t, err, "reading %s", name)

		doc := map[string]any{}
		require.NoError(t, yaml.Unmarshal(content, &doc), "parsing %s", name)

		methods := make([]string, 0, len(doc))

		for key := range doc {
			switch strings.ToUpper(key) {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
				methods = append(methods, strings.ToUpper(key))
			}
		}

		return methods
	}

	content, err := os.ReadFile(filepath.Join(dir, communityEntrypoint)) //nolint:gosec // the name is this repository's own spec entrypoint, not input
	require.NoError(t, err, "reading %s", communityEntrypoint)

	doc := struct {
		Paths map[string]any `yaml:"paths"`
	}{}
	require.NoError(t, yaml.Unmarshal(content, &doc), "parsing %s", communityEntrypoint)

	for path, value := range doc.Paths {
		entry, isMapping := value.(map[string]any)
		require.True(t, isMapping, "%s: %s must be declared with a $ref to its path file", communityEntrypoint, path)

		ref, isRef := entry["$ref"].(string)
		require.True(t, isRef, "%s: %s must be declared with a $ref to its path file", communityEntrypoint, path)

		for _, method := range methodsOf(ref) {
			endpoints[method+" "+normalizePath(path, nil)] = true
		}
	}

	return endpoints
}

// TestEveryAPIRouteIsInTheCommunitySpec fails when a route registered under /api is missing
// from community-openapi.yaml, which the enterprise and cloud entrypoints inherit and which
// the response validator loads. Method and path are both compared, because that is how the
// validator resolves a request to a schema.
func TestEveryAPIRouteIsInTheCommunitySpec(t *testing.T) {
	router := routes.NewRouter(new(mocks.MockService))
	require.NotNil(t, router, "building the router")

	declared := specEndpoints(t)

	missing := make([]string, 0)
	walked := 0

	for _, route := range router.Router().Routes() {
		if !strings.HasPrefix(route.Path, "/api/") || routes.RewrittenFromRoot(route.Path) {
			continue
		}

		walked++

		if !declared[route.Method+" "+normalizePath(route.Path, route.Parameters)] {
			missing = append(missing, route.Method+" "+route.Path)
		}
	}

	require.NotZero(t, walked, "the router registered no /api route, so this check proves nothing")

	sort.Strings(missing)
	require.Empty(t, missing,
		"community-openapi.yaml does not declare these routes. Add each one to "+
			"openapi/spec/community-openapi.yaml, next to the path file that describes it.")
}
