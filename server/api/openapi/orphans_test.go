package openapi_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fileRef = regexp.MustCompile(`\$ref:\s*['"]?([^'"#\s]+)`)

// TestSpecHasNoUnreferencedFiles fails on a spec file that none of the four roots reaches through
// $ref.
func TestSpecHasNoUnreferencedFiles(t *testing.T) {
	dir := specDir(t)

	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, root.Close()) })

	reached := map[string]bool{}
	queue := []string{
		"openapi.yaml",
		"community-openapi.yaml",
		"cloud-openapi.yaml",
		"enterprise-openapi.yaml",
	}

	for len(queue) > 0 {
		file := filepath.Clean(queue[0])
		queue = queue[1:]

		if reached[file] {
			continue
		}

		reached[file] = true

		data, err := root.ReadFile(file)
		require.NoError(t, err, "a $ref points at a missing file or outside the spec directory")

		for _, match := range fileRef.FindAllStringSubmatch(string(data), -1) {
			queue = append(queue, filepath.Join(filepath.Dir(file), match[1]))
		}
	}

	var orphans []string
	require.NoError(t, filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if !reached[rel] {
			orphans = append(orphans, rel)
		}

		return nil
	}))

	sort.Strings(orphans)
	assert.Empty(t, orphans, "spec files no root reaches through $ref")
}
