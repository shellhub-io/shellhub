package openapi_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// specDir returns the absolute path to the openapi spec directory, navigating
// from the test file location up to the repository root and then into openapi/spec.
func specDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")

	// file is <repo>/api/openapi/spec_test.go
	// Navigate: api/openapi -> api -> <repo> -> openapi/spec
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..")

	return filepath.Join(repoRoot, "openapi", "spec")
}

// TestGetSessionsAdvertisesFilterQueryParameter verifies that the GET /api/sessions
// OpenAPI path definition advertises a filterQuery parameter, mirroring the
// pattern already used by GET /api/devices.
func TestGetSessionsAdvertisesFilterQueryParameter(t *testing.T) {
	dir := specDir(t)
	sessionsPath := filepath.Join(dir, "paths", "api@sessions.yaml")
	filterRef := filepath.Join(dir, "components", "parameters", "query", "filterQuery.yaml")

	// The referenced component file must exist.
	_, err := os.Stat(filterRef)
	require.NoError(t, err, "filterQuery.yaml component file should exist at %s", filterRef)

	// Read and parse the sessions path spec.
	data, err := os.ReadFile(sessionsPath) //nolint:gosec // path is constructed from runtime.Caller, not user input.
	require.NoError(t, err, "should be able to read %s", sessionsPath)

	var spec map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &spec), "api@sessions.yaml must be valid YAML")

	// Navigate to get.parameters.
	getOp, ok := spec["get"].(map[string]interface{})
	require.True(t, ok, "spec must have a 'get' operation")

	params, ok := getOp["parameters"].([]interface{})
	require.True(t, ok, "get operation must have a 'parameters' list")

	// Look for a $ref entry pointing to filterQuery.yaml.
	const wantRef = "../components/parameters/query/filterQuery.yaml"

	found := false

	for _, p := range params {
		entry, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		if ref, ok := entry["$ref"].(string); ok && ref == wantRef {
			found = true

			break
		}
	}

	assert.True(t, found, "GET /api/sessions parameters should include $ref: %s", wantRef)
}
