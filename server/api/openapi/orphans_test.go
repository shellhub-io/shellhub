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
// $ref. Redocly's no-unused-components only sees #/components, so a whole file nothing refers to
// passes its lint while still being edited and reviewed as if it were published.
func TestSpecHasNoUnreferencedFiles(t *testing.T) {
	dir := specDir(t)

	reached := map[string]bool{}
	queue := []string{
		filepath.Join(dir, "openapi.yaml"),
		filepath.Join(dir, "community-openapi.yaml"),
		filepath.Join(dir, "cloud-openapi.yaml"),
		filepath.Join(dir, "enterprise-openapi.yaml"),
	}

	for len(queue) > 0 {
		file := filepath.Clean(queue[0])
		queue = queue[1:]

		if reached[file] {
			continue
		}

		reached[file] = true

		data, err := os.ReadFile(file)
		require.NoError(t, err, "a $ref points at a missing file")

		for _, match := range fileRef.FindAllStringSubmatch(string(data), -1) {
			queue = append(queue, filepath.Join(filepath.Dir(file), match[1]))
		}
	}

	var orphans []string
	require.NoError(t, filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}

		if !reached[filepath.Clean(path)] {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			orphans = append(orphans, rel)
		}

		return nil
	}))

	sort.Strings(orphans)
	assert.Empty(t, orphans, "spec files no root reaches through $ref")
}
