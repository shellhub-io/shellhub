package keygen

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsurePrivateKeyCreatesAMissingKeyReadableOnlyByItsOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shellhub.key")

	require.NoError(t, EnsurePrivateKey(path))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	_, err = ReadPublicKey(path)
	assert.NoError(t, err)
}

func TestEnsurePrivateKeyRestrictsAnExistingKeyWithoutChangingIt(t *testing.T) {
	cases := []struct {
		name string
		mode os.FileMode
	}{
		{name: "world-readable key is restricted", mode: 0o644},
		{name: "group-writable key is restricted", mode: 0o620},
		{name: "owner-only key is left as is", mode: 0o600},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "shellhub.key")
			require.NoError(t, GeneratePrivateKey(path))
			require.NoError(t, os.Chmod(path, tc.mode))

			before, err := fs.ReadFile(os.DirFS(dir), "shellhub.key")
			require.NoError(t, err)

			require.NoError(t, EnsurePrivateKey(path))

			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

			after, err := fs.ReadFile(os.DirFS(dir), "shellhub.key")
			require.NoError(t, err)
			assert.Equal(t, before, after)
		})
	}
}
