package testimage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeVersions(t *testing.T, content string) string {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, FileName), []byte(content), 0o600))

	return root
}

func TestBuildArgs(t *testing.T) {
	t.Run("every line becomes an argument", func(t *testing.T) {
		root := writeVersions(t, "GOLANG_VERSION=1.26.9-alpine3.24\n\nALPINE_VERSION=3.24.2\n")

		args, err := BuildArgs(root)
		require.NoError(t, err)

		require.Len(t, args, 2)
		assert.Equal(t, "1.26.9-alpine3.24", *args["GOLANG_VERSION"])
		assert.Equal(t, "3.24.2", *args["ALPINE_VERSION"])
	})

	t.Run("a line without a value is an error", func(t *testing.T) {
		root := writeVersions(t, "GOLANG_VERSION=1.26.9-alpine3.24\nALPINE_VERSION=\n")

		_, err := BuildArgs(root)
		require.ErrorContains(t, err, "versions.env:2")
	})

	t.Run("a missing file is an error", func(t *testing.T) {
		_, err := BuildArgs(t.TempDir())
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestRoot(t *testing.T) {
	t.Run("the nearest directory above holding the file", func(t *testing.T) {
		root := writeVersions(t, "ALPINE_VERSION=3.24.2\n")
		nested := filepath.Join(root, "tests", "environment")
		require.NoError(t, os.MkdirAll(nested, 0o750))
		t.Chdir(nested)

		found, err := Root()
		require.NoError(t, err)
		assert.Equal(t, root, found)
	})

	t.Run("no file above is an error", func(t *testing.T) {
		t.Chdir(t.TempDir())

		_, err := Root()
		require.ErrorContains(t, err, FileName)
	})
}

func TestVersion(t *testing.T) {
	root := writeVersions(t, "REDOCLY_VERSION=2.31.5\n")

	t.Run("a pinned key", func(t *testing.T) {
		version, err := Version(root, "REDOCLY_VERSION")
		require.NoError(t, err)
		assert.Equal(t, "2.31.5", version)
	})

	t.Run("a key the file lacks is an error", func(t *testing.T) {
		_, err := Version(root, "PRISM_VERSION")
		require.ErrorContains(t, err, "PRISM_VERSION is not pinned")
	})
}
